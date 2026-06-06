package ui

import (
	"context"
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
	"github.com/xiakn/logcat/internal/session"
)

type stubDeviceService struct {
	install            adb.Install
	devices            []adb.DeviceInfo
	packagesByScope    map[adb.PackageScope][]adb.PackageInfo
	foregroundPackage  string
	processesByPackage map[string][]adb.ProcessInfo
}

func (s stubDeviceService) DetectADB(context.Context) (adb.Install, error) {
	return s.install, nil
}

func (s stubDeviceService) ListDevices(context.Context) ([]adb.DeviceInfo, error) {
	return s.devices, nil
}

func (s stubDeviceService) ListPackages(
	_ context.Context,
	_ string,
	scope adb.PackageScope,
) ([]adb.PackageInfo, error) {
	return append([]adb.PackageInfo(nil), s.packagesByScope[scope]...), nil
}

func (s stubDeviceService) CurrentForegroundPackage(context.Context, string) (string, error) {
	return s.foregroundPackage, nil
}

func (s stubDeviceService) ListProcesses(
	_ context.Context,
	_ string,
	packageName string,
) ([]adb.ProcessInfo, error) {
	return append([]adb.ProcessInfo(nil), s.processesByPackage[packageName]...), nil
}

type stubSessionHandle struct {
	events chan session.Event
}

type stubSessionStarter struct {
	handle stubSessionHandle
}

func (s stubSessionStarter) Start(context.Context, session.Config) (session.Handle, error) {
	return session.NewHandle(s.handle.events), nil
}

func TestShellHandleActionsDrivesController(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newTestController(t, events)
	shell := newShell(controller)

	events <- session.Event{Entry: makeEntry("[H5] token one")}
	events <- session.Event{Entry: makeEntry("[H5] miss")}

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 2
	})

	shell.searchEditor.SetText("token")
	shell.pauseButton.Click()
	shell.nextMatchButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())

	model := controller.Model()
	if !model.Pause.Active {
		t.Fatal("expected pause state to be active")
	}
	if model.Search.Query != "token" {
		t.Fatalf("expected search query to sync, got %q", model.Search.Query)
	}
	if model.Search.Current != 0 {
		t.Fatalf("expected first match to be selected, got %d", model.Search.Current)
	}
	if model.SelectedIndex != 0 {
		t.Fatalf("expected selected index to follow current match, got %d", model.SelectedIndex)
	}
}

func TestShellHandleActionsSelectsLogRow(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newTestController(t, events)
	shell := newShell(controller)

	events <- session.Event{Entry: makeEntry("[H5] first")}
	events <- session.Event{Entry: makeEntry("[H5] second")}

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 2
	})

	model := controller.Model()
	shell.syncLogButtons(len(model.VisibleLogs))
	shell.logButtons[1].Click()
	shell.handleActions(testLayoutContext(), model)

	if controller.Model().SelectedIndex != 1 {
		t.Fatalf("expected second log to be selected, got %d", controller.Model().SelectedIndex)
	}
}

func TestShellSelectedClipboardTextUsesCurrentSelection(t *testing.T) {
	shell := newShell(nil)
	model := appstate.NewModel()
	model.VisibleLogs = []appstate.LogViewItem{
		{
			Display: "06-04 16:42:18.479 I chromium [H5] hello",
			Entry: logcat.LogEntry{
				Raw:     "raw line",
				Message: "[H5] hello",
			},
		},
	}
	model.SelectedIndex = 0

	line, ok := shell.selectedClipboardText(model, copyLine)
	if !ok || line != "06-04 16:42:18.479 I chromium [H5] hello" {
		t.Fatalf("expected display line, got %q ok=%v", line, ok)
	}

	raw, ok := shell.selectedClipboardText(model, copyRaw)
	if !ok || raw != "raw line" {
		t.Fatalf("expected raw line, got %q ok=%v", raw, ok)
	}

	message, ok := shell.selectedClipboardText(model, copyMessage)
	if !ok || message != "[H5] hello" {
		t.Fatalf("expected message, got %q ok=%v", message, ok)
	}
}

func TestShellSyncAutoFollowStopsAfterLeavingEnd(t *testing.T) {
	shell := newShell(nil)
	shell.followLogs = true
	shell.lastLogCount = 3
	shell.logList.Position.BeforeEnd = true

	shell.syncAutoFollow(3)

	if shell.followLogs {
		t.Fatal("expected auto follow to disable after leaving end")
	}
	if shell.logList.ScrollToEnd {
		t.Fatal("expected scroll-to-end to be disabled")
	}
}

func TestShellSyncAutoFollowRestoresEndWhenEnabled(t *testing.T) {
	shell := newShell(nil)
	shell.followLogs = true

	shell.syncAutoFollow(2)

	if !shell.logList.ScrollToEnd {
		t.Fatal("expected scroll-to-end to stay enabled")
	}
	if shell.logList.Position.BeforeEnd {
		t.Fatal("expected position to be forced back to end")
	}
}

func TestShellProgrammaticSelectionKeepsAutoFollowEnabled(t *testing.T) {
	shell := newShell(nil)
	shell.followLogs = true

	model := appstate.NewModel()
	model.VisibleLogs = []appstate.LogViewItem{
		{Display: "one"},
		{Display: "two"},
	}
	model.SelectedIndex = 0

	shell.syncSelectedLog(model)
	shell.syncAutoFollow(len(model.VisibleLogs))

	if !shell.followLogs {
		t.Fatal("expected auto follow to stay enabled for programmatic selection")
	}
	if !shell.logList.ScrollToEnd {
		t.Fatal("expected scroll-to-end to remain enabled")
	}
}

func TestShellHandleActionsTriggersForegroundSelection(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
		packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
			adb.PackageScopeUser: {
				{Name: "com.demo.host"},
			},
		},
		foregroundPackage: "com.demo.host",
		processesByPackage: map[string][]adb.ProcessInfo{
			"com.demo.host": {
				{PID: 111, Name: "com.demo.host"},
			},
		},
	})
	shell := newShell(controller)

	shell.foregroundButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())

	model := controller.Model()
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected foreground package selected, got %q", model.SelectedPackage)
	}
	if len(model.BoundPIDs) != 1 || model.BoundPIDs[0] != 111 {
		t.Fatalf("unexpected foreground bound pids: %#v", model.BoundPIDs)
	}
}

func TestShellHandleActionsSelectsPackageAndProcess(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
		packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
			adb.PackageScopeUser: {
				{Name: "com.demo.host"},
			},
		},
		processesByPackage: map[string][]adb.ProcessInfo{
			"com.demo.host": {
				{PID: 111, Name: "com.demo.host"},
				{PID: 222, Name: "com.demo.host:webview"},
			},
		},
	})
	shell := newShell(controller)

	model := controller.Model()
	shell.syncPackageButtons(len(model.Packages))
	shell.packageButtons[0].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected package selected, got %q", model.SelectedPackage)
	}
	if len(model.BoundPIDs) != 2 {
		t.Fatalf("expected package binding to keep 2 pids, got %#v", model.BoundPIDs)
	}

	shell.syncProcessButtons(len(model.Processes))
	shell.processButtons[1].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.SelectedProcess != "com.demo.host:webview" {
		t.Fatalf("expected process selected, got %q", model.SelectedProcess)
	}
	if len(model.BoundPIDs) != 1 || model.BoundPIDs[0] != 222 {
		t.Fatalf("unexpected narrowed bound pids: %#v", model.BoundPIDs)
	}
}

func TestShellHandleActionsSwitchesPackageScope(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
		packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
			adb.PackageScopeUser: {
				{Name: "com.demo.host"},
			},
			adb.PackageScopeSystem: {
				{Name: "com.android.systemui"},
			},
		},
	})
	shell := newShell(controller)

	shell.scopeSystemButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())

	model := controller.Model()
	if model.PackageScope != adb.PackageScopeSystem {
		t.Fatalf("expected system scope, got %q", model.PackageScope)
	}
	if len(model.Packages) != 1 || model.Packages[0].Name != "com.android.systemui" {
		t.Fatalf("unexpected system packages: %#v", model.Packages)
	}
}

func TestShellSyncPackageButtonsMatchesModelPackages(t *testing.T) {
	shell := newShell(nil)

	shell.syncPackageButtons(2)
	if len(shell.packageButtons) != 2 {
		t.Fatalf("expected 2 package buttons, got %d", len(shell.packageButtons))
	}

	shell.syncPackageButtons(1)
	if len(shell.packageButtons) != 1 {
		t.Fatalf("expected package buttons to shrink to 1, got %d", len(shell.packageButtons))
	}
}

func newTestController(t *testing.T, events chan session.Event) *appstate.Controller {
	t.Helper()

	controller := appstate.NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
				adb.PackageScopeUser: {
					{Name: "com.demo.host"},
				},
			},
		},
		stubSessionStarter{
			handle: stubSessionHandle{events: events},
		},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}

	return controller
}

func newControllerWithService(
	t *testing.T,
	service stubDeviceService,
) *appstate.Controller {
	t.Helper()

	events := make(chan session.Event)
	close(events)
	controller := appstate.NewController(
		service,
		stubSessionStarter{
			handle: stubSessionHandle{events: events},
		},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}

	return controller
}

func makeEntry(message string) *logcat.LogEntry {
	return &logcat.LogEntry{
		DeviceID: "emulator-5554",
		TimeText: "06-04 16:42:18.479",
		Level:    "I",
		Tag:      "chromium",
		Message:  message,
		Raw:      message,
	}
}

func testLayoutContext() layout.Context {
	return layout.Context{
		Constraints: layout.Exact(image.Pt(1280, 720)),
		Now:         time.Unix(0, 0),
		Ops:         new(op.Ops),
	}
}

func waitFor(t *testing.T, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not met before timeout")
}
