package main

import (
	"log"
	"os"

	gioapp "gioui.org/app"
	"gioui.org/unit"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/session"
	"github.com/xiakn/logcat/internal/storage"
	"github.com/xiakn/logcat/internal/ui"
)

func main() {
	go func() {
		window := new(gioapp.Window)
		window.Option(
			gioapp.Title("Logcat Viewer"),
			gioapp.Size(unit.Dp(1280), unit.Dp(820)),
		)

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

		if err := ui.Run(window, controller); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	gioapp.Main()
}
