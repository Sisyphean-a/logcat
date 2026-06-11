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
		filters := stripBuiltinFilters(state.Filters)
		if len(filters) > 0 || len(state.Filters) > 0 {
			controller.ReplaceSavedFilters(filters, state.DefaultFilterID)
		}
		if len(state.History) > 0 {
			controller.ReplaceFilterHistory(state.History)
		}
		if filterState := controller.FilterStateSnapshot(); filterState.DefaultFilterID != "" {
			_ = controller.RestoreSavedFilter(filterState.DefaultFilterID)
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

func stripBuiltinFilters(filters []appstate.SavedFilter) []appstate.SavedFilter {
	clean := make([]appstate.SavedFilter, 0, len(filters))
	for _, filter := range filters {
		if filter.ID == "builtin-h5" {
			continue
		}
		clean = append(clean, filter)
	}
	return clean
}
