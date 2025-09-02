package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GitHub Dashboard",
		Width:  1200,
		Height: 800,
		MinWidth:  800,
		MinHeight: 600,
		MaxWidth:  1920,
		MaxHeight: 1080,
		DisableResize: false,
		Fullscreen:    false,
		StartHidden:   false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1}, // Light gray background
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// Disable development features in production
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
