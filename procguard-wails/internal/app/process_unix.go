//go:build !windows

package app

import (
	"database/sql"
	"procguard-wails/internal/data"
)

// StartProcessEventLogger is a dummy implementation for non-Windows platforms.
func StartProcessEventLogger(appLogger data.Logger, db *sql.DB) {}

// StartBlocklistEnforcer is a dummy implementation for non-Windows platforms.
func StartBlocklistEnforcer(appLogger data.Logger) {}
