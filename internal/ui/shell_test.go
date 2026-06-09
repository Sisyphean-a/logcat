package ui

import (
	"context"
	"image"
	"fmt"
	"os"
	"path/filepath"
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

func (s stubDeviceService) ListPackages(_ context.Context, _ string, scope adb.PackageScope) ([]adb.PackageInfo, error) {
	return append([]adb.PackageInfo(nil), s.packagesByScope[scope]...), nil
}

func (s stubDeviceService) CurrentForegroundPackage(context.Context, string) (string, error) {
	return s.foregroundPackage, nil
}

func (s stubDeviceService) ListProcesses(_ context.Context, _ string, packageName string) ([]adb.ProcessInfo, error) {
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

func TestShellHandleActionsSelectsDeviceAndPackage(t *testing.T) {
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
			},
		},
	})
	shell := newShell(controller)

	model := controller.Model()
	shell.syncButtons(model)
	shell.deviceButtons[0].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	shell.syncButtons(model)
	shell.packageButtons[0].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.SelectedDevice != "emulator-5554" {
		t.Fatalf("expected selected device updated, got %q", model.SelectedDevice)
	}
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected package updated, got %q", model.SelectedPackage)
	}
}

func TestShellHandleActionsSavesCurrentFilter(t *testing.T) {
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
	})
	shell := newShell(controller)
	shell.saveNameEditor.SetText("申请流")
	shell.filterEditor.SetText("tag:chromium & message:apply")
	shell.saveFilterButton.Click()

	shell.handleActions(testLayoutContext(), controller.Model())
	model := controller.Model()
	if len(model.Filter.Saved) < 2 {
		t.Fatalf("expected saved filter appended, got %#v", model.Filter.Saved)
	}
	if model.Filter.Saved[len(model.Filter.Saved)-1].Name != "申请流" {
		t.Fatalf("unexpected saved filter tail: %#v", model.Filter.Saved[len(model.Filter.Saved)-1])
	}
}

func TestShellHandleActionsSelectsSavedFilter(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
		packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
			adb.PackageScopeUser: {
				{Name: "com.demo.host"},
				{Name: "com.demo.other"},
			},
		},
		processesByPackage: map[string][]adb.ProcessInfo{
			"com.demo.host":  {{PID: 111, Name: "com.demo.host"}},
			"com.demo.other": {{PID: 222, Name: "com.demo.other"}},
		},
	})
	controller.ReplaceSavedFilters([]appstate.SavedFilter{
		{ID: "builtin-h5", Name: "H5 日志", Query: "tag:chromium & message:[H5]"},
		{ID: "network", Name: "网络错误", PackageName: "com.demo.other", Query: "level:E & message:网络错误"},
	})
	shell := newShell(controller)

	model := controller.Model()
	shell.syncButtons(model)
	shell.filterButtons[1].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.Filter.ActiveFilterID != "network" {
		t.Fatalf("expected active filter updated, got %q", model.Filter.ActiveFilterID)
	}
	if model.Filter.Applied != "level:E & message:网络错误" {
		t.Fatalf("expected applied query updated, got %q", model.Filter.Applied)
	}
	if model.SelectedPackage != "com.demo.other" {
		t.Fatalf("expected saved filter package to bind, got %q", model.SelectedPackage)
	}
}

func TestShellHandleActionsAppliesHistoryQuery(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
	})
	controller.ReplaceFilterHistory([]string{
		"level:E & message:网络错误",
		"tag:chromium & message:[H5]",
	})
	shell := newShell(controller)

	model := controller.Model()
	shell.syncButtons(model)
	shell.historyMenuButton.Click()
	shell.handleActions(testLayoutContext(), model)
	model = controller.Model()
	shell.syncButtons(model)
	shell.historyButtons[0].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.Filter.Draft != "level:E & message:网络错误" {
		t.Fatalf("expected draft from history, got %q", model.Filter.Draft)
	}
	if model.Filter.Applied != "level:E & message:网络错误" {
		t.Fatalf("expected applied from history, got %q", model.Filter.Applied)
	}
	if len(model.Filter.History) == 0 || model.Filter.History[0] != "level:E & message:网络错误" {
		t.Fatalf("expected history to stay normalized, got %#v", model.Filter.History)
	}
}

func TestShellHandleActionsApplyingCustomQueryClearsActiveFilter(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
	})
	shell := newShell(controller)
	shell.filterEditor.SetText("level:E")

	shell.handleActions(testLayoutContext(), controller.Model())
	if err := controller.ApplyFilterDraft(); err != nil {
		t.Fatalf("ApplyFilterDraft returned error: %v", err)
	}

	model := controller.Model()
	if model.Filter.ActiveFilterID != "" {
		t.Fatalf("expected active filter cleared for custom query, got %q", model.Filter.ActiveFilterID)
	}
	if len(model.Filter.History) == 0 || model.Filter.History[0] != "level:E" {
		t.Fatalf("expected custom query recorded in history, got %#v", model.Filter.History)
	}
}

func TestShellHandleActionsSelectingPackageClearsMismatchedActiveFilter(t *testing.T) {
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
			},
		},
	})
	controller.ReplaceSavedFilters([]appstate.SavedFilter{
		{ID: "builtin-h5", Name: "H5 日志", Query: "tag:chromium & message:[H5]"},
		{ID: "network", Name: "网络错误", Query: "level:E", PackageName: "com.demo.other"},
	})
	controller.ReplaceFilterHistory([]string{"level:E"})
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	shell := newShell(controller)

	shell.filterEditor.SetText("level:E")
	shell.handleActions(testLayoutContext(), controller.Model())
	if err := controller.ApplyFilterDraft(); err != nil {
		t.Fatalf("ApplyFilterDraft returned error: %v", err)
	}

	model := controller.Model()
	shell.syncButtons(model)
	shell.packageButtons[0].Click()
	shell.handleActions(testLayoutContext(), model)

	model = controller.Model()
	if model.Filter.ActiveFilterID != "" {
		t.Fatalf("expected active filter cleared after selecting unmatched package, got %q", model.Filter.ActiveFilterID)
	}
}

func TestShellHandleActionsPauseButtonTogglesResumeKeep(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newTestController(t, events)
	shell := newShell(controller)

	controller.Pause()
	events <- session.Event{Entry: makeEntry("[H5] buffered")}
	waitFor(t, func() bool {
		return controller.Model().Pause.BufferedCount == 1
	})

	shell.pauseButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())

	model := controller.Model()
	if model.Pause.Active {
		t.Fatal("expected pause state cleared after second toolbar click")
	}
	if len(model.VisibleLogs) != 1 {
		t.Fatalf("expected buffered log flushed on resume, got %d", len(model.VisibleLogs))
	}
}

func TestShellHandleActionsExportsVisibleLogs(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newTestController(t, events)
	shell := newShell(controller)
	dir := t.TempDir()
	var exportedPath string
	shell.exportLogs = func(items []appstate.LogViewItem) (string, error) {
		if len(items) != 1 {
			return "", fmt.Errorf("unexpected export count: %d", len(items))
		}
		exportedPath = filepath.Join(dir, "logs.tsv")
		return exportedPath, os.WriteFile(exportedPath, []byte(items[0].Display), 0o644)
	}

	events <- session.Event{Entry: makeEntry("[H5] export me")}
	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 1
	})

	shell.exportButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())

	if exportedPath == "" {
		t.Fatal("expected export path recorded")
	}
	content, err := os.ReadFile(exportedPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) == "" {
		t.Fatal("expected export file content")
	}
	if controller.Model().Status == "" {
		t.Fatal("expected status updated after export")
	}
}

func TestShellHandleActionsTogglesSelectorMenus(t *testing.T) {
	controller := newControllerWithService(t, stubDeviceService{
		install: adb.Install{Path: "adb", Version: "1.0.41"},
		devices: []adb.DeviceInfo{
			{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
		},
	})
	shell := newShell(controller)

	shell.deviceSelectorButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())
	if !shell.deviceMenuOpen {
		t.Fatal("expected device menu open")
	}

	shell.filterSelectorButton.Click()
	shell.handleActions(testLayoutContext(), controller.Model())
	if !shell.filterMenuOpen {
		t.Fatal("expected filter menu open")
	}
	if shell.deviceMenuOpen {
		t.Fatal("expected device menu closed when filter menu opens")
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
	shell.syncButtons(model)
	shell.logButtons[1].Click()
	shell.handleActions(testLayoutContext(), model)

	if controller.Model().SelectedIndex != 1 {
		t.Fatalf("expected second log selected, got %d", controller.Model().SelectedIndex)
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

func newControllerWithService(t *testing.T, service stubDeviceService) *appstate.Controller {
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
