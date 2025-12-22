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

	// 检查单实例
	if !app.CheckSingleInstance() {
		// 已有实例在运行，直接退出
		return
	}

	// 初始化系统托盘
	app.InitTray()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "AI工作台",
		Width:            1920,
		Height:           1080,
		WindowStartState: options.Maximised,
		Frameless:        true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
