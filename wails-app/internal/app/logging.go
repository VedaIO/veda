package app

import (
	"database/sql"
	"strings"
	"sync"
	"time"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/write"
	"wails-app/internal/platform/app_filter"

	"github.com/shirou/gopsutil/v3/process"
)

const processCheckInterval = 2 * time.Second

// loggerState encapsulates the in-memory state of the process monitor.
type loggerState struct {
	runningProcs     map[int32]string // PID -> lowercase process name
	runningAppCounts map[string]int   // lowercase process name -> instance count
	sync.Mutex
}

var resetLoggerCh = make(chan struct{}, 1)

// ResetLoggedApps clears the in-memory cache of logged applications.
func ResetLoggedApps() {
	resetLoggerCh <- struct{}{}
}

// StartProcessEventLogger starts a long-running goroutine that monitors process creation and termination events.
func StartProcessEventLogger(appLogger logger.Logger, db *sql.DB) {
	state := &loggerState{
		runningProcs:     make(map[int32]string),
		runningAppCounts: make(map[string]int),
	}

	go func() {
		initializeRunningProcs(state, db)

		ticker := time.NewTicker(processCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				procs, err := process.Processes()
				if err != nil {
					appLogger.Printf("Failed to get processes: %v", err)
					continue
				}

				currentPids := make(map[int32]bool)
				for _, p := range procs {
					currentPids[p.Pid] = true
				}

				logEndedProcesses(state, currentPids)
				logNewProcesses(state, appLogger, procs)
			case <-resetLoggerCh:
				appLogger.Printf("[Logger] Reset signal received. Clearing in-memory state.")
				state.Lock()
				state.runningProcs = make(map[int32]string)
				state.runningAppCounts = make(map[string]int)
				state.Unlock()
			}
		}
	}()
}

func logEndedProcesses(state *loggerState, currentPids map[int32]bool) {
	state.Lock()
	defer state.Unlock()

	for pid, nameLower := range state.runningProcs {
		if !currentPids[pid] {
			write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)

			delete(state.runningProcs, pid)
			state.runningAppCounts[nameLower]--
			if state.runningAppCounts[nameLower] <= 0 {
				delete(state.runningAppCounts, nameLower)
			}
		}
	}
}

func logNewProcesses(state *loggerState, appLogger logger.Logger, procs []*process.Process) {
	state.Lock()
	defer state.Unlock()

	for _, p := range procs {
		if _, exists := state.runningProcs[p.Pid]; exists {
			continue
		}

		name, err := p.Name()
		if err != nil || name == "" {
			continue // Retry next tick
		}
		nameLower := strings.ToLower(name)

		exePath, err := p.Exe()
		if err != nil {
			continue // Retry next tick
		}

		// Rule 1: Platform-specific system exclusion
		if app_filter.ShouldExclude(exePath, p) {
			state.runningProcs[p.Pid] = nameLower
			continue
		}

		// Rule 2: Deduplication (Reference Counting)
		// If count > 0, it means an instance is already logged and active.
		isAlreadyLogged := state.runningAppCounts[nameLower] > 0

		if isAlreadyLogged {
			state.runningProcs[p.Pid] = nameLower
			state.runningAppCounts[nameLower]++
			continue
		}

		// Rule 3: Must be a trackable user application (e.g. has a window)
		if !app_filter.ShouldTrack(exePath, p) {
			continue // Retry next tick
		}

		// Success: Log it
		parent, _ := p.Parent()
		parentName := ""
		if parent != nil {
			parentName, _ = parent.Name()
		}

		write.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time) VALUES (?, ?, ?, ?, ?)",
			name, p.Pid, parentName, exePath, time.Now().Unix())

		state.runningProcs[p.Pid] = nameLower
		state.runningAppCounts[nameLower]++
	}
}

func initializeRunningProcs(state *loggerState, db *sql.DB) {
	rows, err := db.Query("SELECT pid, process_name FROM app_events WHERE end_time IS NULL")
	if err != nil {
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

	state.Lock()
	defer state.Unlock()

	for rows.Next() {
		var pid int32
		var name string
		if err := rows.Scan(&pid, &name); err == nil {
			if exists, _ := process.PidExists(pid); exists {
				nameLower := strings.ToLower(name)
				state.runningProcs[pid] = nameLower
				state.runningAppCounts[nameLower]++
			} else {
				write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
			}
		}
	}
}
