package app

import (
	"context"
	"testing"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/session"
)

func TestControllerPauseBufferedLogsAdvanceRevision(t *testing.T) {
	events := make(chan session.Event, 2)
	controller := newStreamingController(t, events)

	controller.Pause()
	before := controller.UISnapshot(1000).Revision
	events <- session.Event{Entry: makeEntry("[H5] buffered")}

	waitFor(t, func() bool {
		return controller.UISnapshot(1000).Revision > before
	})

	snapshot := controller.UISnapshot(1000)
	if snapshot.Model.Pause.BufferedCount != 1 {
		t.Fatalf("expected buffered count 1, got %d", snapshot.Model.Pause.BufferedCount)
	}
}

func TestControllerReconnectRestoresRunningPackageBinding(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
			},
			packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
				adb.PackageScopeAll: {{Name: "com.demo.host"}},
			},
			processesByPackage: map[string][]adb.ProcessInfo{
				"com.demo.host": {{PID: 111, Name: "com.demo.host"}},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()
	if err := controller.SelectPackage(context.Background(), "com.demo.host"); err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}

	waitFor(t, func() bool {
		model := controller.Model()
		return model.SelectedPackage == "com.demo.host" &&
			len(model.BoundPIDs) == 1 &&
			model.BoundPIDs[0] == 111
	})

	if err := controller.syncDevices(context.Background(), nil); err != nil {
		t.Fatalf("syncDevices disconnect returned error: %v", err)
	}
	waitFor(t, func() bool {
		return controller.Model().SelectedDevice == ""
	})

	if err := controller.syncDevices(context.Background(), []adb.DeviceInfo{
		{ID: "device-1", Model: "Pixel_7", Status: "device"},
	}); err != nil {
		t.Fatalf("syncDevices reconnect returned error: %v", err)
	}

	waitFor(t, func() bool {
		model := controller.Model()
		return model.SelectedDevice == "device-1" &&
			model.SelectedPackage == "com.demo.host" &&
			len(model.BoundPIDs) == 1 &&
			model.BoundPIDs[0] == 111 &&
			!model.Pause.Active
	})

	starter.mu.Lock()
	configCount := len(starter.configs)
	latest := starter.configs[configCount-1]
	starter.mu.Unlock()

	if configCount != 3 {
		t.Fatalf("expected 3 session starts, got %d", configCount)
	}
	if latest.PackageName != "com.demo.host" {
		t.Fatalf("expected rebound package session, got %q", latest.PackageName)
	}
	if len(latest.AllowedPIDs) != 1 || latest.AllowedPIDs[0] != 111 {
		t.Fatalf("unexpected rebound allowed pids: %#v", latest.AllowedPIDs)
	}
}

func TestControllerResumeKeepStartsFallbackSessionForStoppedPackage(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	if err := controller.SelectPackage(context.Background(), "com.demo.host"); err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}

	controller.ResumeKeep()

	model := controller.Model()
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected package kept, got %q", model.SelectedPackage)
	}
	if len(model.BoundPIDs) != 0 {
		t.Fatalf("expected stopped package to have no bound pids, got %#v", model.BoundPIDs)
	}

	starter.mu.Lock()
	configCount := len(starter.configs)
	latest := starter.configs[configCount-1]
	starter.mu.Unlock()

	if configCount != 1 {
		t.Fatalf("expected fallback start from stopped package to create one session, got %d", configCount)
	}
	if latest.PackageName != "" || latest.ProcessName != "" {
		t.Fatalf("expected fallback session without package binding, got package=%q process=%q", latest.PackageName, latest.ProcessName)
	}
	if len(latest.AllowedPIDs) != 0 {
		t.Fatalf("expected fallback session without pid filter, got %#v", latest.AllowedPIDs)
	}
}
