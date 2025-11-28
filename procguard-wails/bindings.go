package main

import (
	"database/sql"
	"procguard-wails/api"
	"procguard-wails/internal/data"
)

// App struct
type App struct {
	db     *sql.DB
	logger data.Logger
	*api.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}
