package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "TamagOps",
		Width:  1040,
		Height: 720,

		MinWidth:  860,
		MinHeight: 600,

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		BackgroundColour: &options.RGBA{R: 18, G: 26, B: 24, A: 1},

		OnStartup:  app.startup,
		OnShutdown: app.shutdown,

		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Erro ao iniciar o TamagOps:", err.Error())
	}
}
