package main

import (
	"context"
	"testing"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/session"
)

func TestPersistFiltersSkipsUnchangedWrites(t *testing.T) {
	controller := appstate.NewController(stubDeviceService{}, stubSessionStarter{})
	controller.ReplaceSavedFilters([]appstate.SavedFilter{{
		ID:          "h5",
		Name:        "H5",
		PackageName: "com.demo.host",
		Query:       "tag=chromium",
	}}, "h5")
	controller.ReplaceFilterHistory([]string{"tag=chromium"})

	savedStates := make([]persistedFilterState, 0, 4)
	app := NewApp(controller)
	app.saveFilterState = func(filters []appstate.SavedFilter, history []string, defaultFilterID string) error {
		savedStates = append(savedStates, persistedFilterState{
			Filters:         append([]appstate.SavedFilter(nil), filters...),
			History:         append([]string(nil), history...),
			DefaultFilterID: defaultFilterID,
		})
		return nil
	}

	app.persistFilters()
	if len(savedStates) != 0 {
		t.Fatalf("expected initial unchanged state to skip persistence, got %d writes", len(savedStates))
	}

	if err := controller.ApplyHistoryQuery("level:E"); err != nil {
		t.Fatalf("ApplyHistoryQuery returned error: %v", err)
	}
	app.persistFilters()
	if len(savedStates) != 1 {
		t.Fatalf("expected changed history to persist once, got %d writes", len(savedStates))
	}

	app.persistFilters()
	if len(savedStates) != 1 {
		t.Fatalf("expected unchanged state to skip persistence, got %d writes", len(savedStates))
	}

	if err := controller.SaveFilterDefinition("Bridge", "com.demo.host", "message~:bridge"); err != nil {
		t.Fatalf("SaveFilterDefinition returned error: %v", err)
	}
	app.persistFilters()
	if len(savedStates) != 2 {
		t.Fatalf("expected saved filters change to persist once, got %d writes", len(savedStates))
	}
}

type stubDeviceService struct{}

func (stubDeviceService) DetectADB(context.Context) (adb.Install, error) {
	return adb.Install{}, nil
}

func (stubDeviceService) ListDevices(context.Context) ([]adb.DeviceInfo, error) {
	return nil, nil
}

func (stubDeviceService) ListPackages(context.Context, string, adb.PackageScope) ([]adb.PackageInfo, error) {
	return nil, nil
}

func (stubDeviceService) CurrentForegroundPackage(context.Context, string) (string, error) {
	return "", nil
}

func (stubDeviceService) ListProcesses(context.Context, string, string) ([]adb.ProcessInfo, error) {
	return nil, nil
}

type stubSessionStarter struct{}

func (stubSessionStarter) Start(context.Context, session.Config) (session.Handle, error) {
	events := make(chan session.Event)
	close(events)
	return session.NewHandle(events), nil
}
