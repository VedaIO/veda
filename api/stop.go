package api

import (
	"net/http"
	"os"
	"time"
)

// handleStop handles the graceful shutdown of the application.
func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.Logger.Println("Received stop request. Shutting down...")

	// Respond to the client before shutting down.
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Give the response a moment to be sent before exiting.
	go func() {
		time.Sleep(1 * time.Second)
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			s.Logger.Printf("Failed to close database: %v", err)
		}
		os.Exit(0)
	}()
}
