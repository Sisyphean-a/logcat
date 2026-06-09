package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/session"
	"github.com/xiakn/logcat/internal/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	runner := adb.ExecRunner{}
	deviceService := adb.NewService(runner, "")
	source := adb.NewLogcatSource(runner, "")
	supervisor := session.NewSupervisor(source)
	controller := appstate.NewController(deviceService, supervisor)

	if state, err := storage.LoadFilterState(); err == nil {
		if len(state.Filters) > 0 {
			controller.ReplaceSavedFilters(state.Filters)
		}
		if len(state.History) > 0 {
			controller.ReplaceFilterHistory(state.History)
		}
	}

	app := NewApp(controller)

	err := wails.Run(&options.App{
		Title:     "Logcat Viewer",
		Width:     1440,
		Height:    900,
		MinWidth:  1200,
		MinHeight: 760,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 22, G: 22, B: 22, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
