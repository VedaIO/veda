package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"log"
	"procguard-wails/api"
	"procguard-wails/internal/daemon"
	"procguard-wails/internal/data"
)

//go:embed all:frontend/dist
var assets embed.FS

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	a.db = db
	data.NewLogger(db)
	a.logger = data.GetLogger()
	a.Server = api.NewServer(db)

	daemon.StartDaemon(a.logger, a.db)
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ProcGuard",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
