package app

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/logcat"
	"github.com/xiakn/logcat/internal/session"
)

type stubDeviceService struct {
	install            adb.Install
	devices            []adb.DeviceInfo
	packagesByScope    map[adb.PackageScope][]adb.PackageInfo
	foregroundPackage  string
	processesByPackage map[string][]adb.ProcessInfo
	listPackagesFunc   func(adb.PackageScope) ([]adb.PackageInfo, error)
	foregroundFunc     func() (string, error)
	listProcessesFunc  func(string) ([]adb.ProcessInfo, error)
	listDevicesFunc    func() ([]adb.DeviceInfo, error)
	err                error
}

func (s stubDeviceService) DetectADB(context.Context) (adb.Install, error) {
	return s.install, s.err
}

func (s stubDeviceService) ListDevices(context.Context) ([]adb.DeviceInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.listDevicesFunc != nil {
		return s.listDevicesFunc()
	}
	return s.devices, s.err
}

func (s stubDeviceService) ListPackages(
	_ context.Context,
	_ string,
	scope adb.PackageScope,
) ([]adb.PackageInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.listPackagesFunc != nil {
		return s.listPackagesFunc(scope)
	}
	return append([]adb.PackageInfo(nil), s.packagesByScope[scope]...), nil
}

func (s stubDeviceService) CurrentForegroundPackage(context.Context, string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if s.foregroundFunc != nil {
		return s.foregroundFunc()
	}
	return s.foregroundPackage, s.err
}

func (s stubDeviceService) ListProcesses(
	_ context.Context,
	_ string,
	packageName string,
) ([]adb.ProcessInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.listProcessesFunc != nil {
		return s.listProcessesFunc(packageName)
	}
	return append([]adb.ProcessInfo(nil), s.processesByPackage[packageName]...), nil
}

type stubSessionHandle struct {
	events chan session.Event
}

type stubSessionStarter struct {
	handle stubSessionHandle
	err    error
}

func (s stubSessionStarter) Start(context.Context, session.Config) (session.Handle, error) {
	return session.NewHandle(s.handle.events), s.err
}

type recordingSessionStarter struct {
	mu       sync.Mutex
	contexts []context.Context
	configs  []session.Config
}

func (s *recordingSessionStarter) Start(ctx context.Context, cfg session.Config) (session.Handle, error) {
	s.mu.Lock()
	s.contexts = append(s.contexts, ctx)
	s.configs = append(s.configs, cfg)
	s.mu.Unlock()

	events := make(chan session.Event)
	close(events)
	return session.NewHandle(events), nil
}

func TestControllerLoadUpdatesDevicesAndStatus(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	model := controller.Model()
	if model.Status != "adb 1.0.41" {
		t.Fatalf("expected adb status, got %q", model.Status)
	}
	if len(model.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(model.Devices))
	}
}

func TestControllerNewModelStartsWithoutAppliedFilter(t *testing.T) {
	model := NewModel()

	if model.Filter.Draft != "" {
		t.Fatalf("expected empty draft filter, got %q", model.Filter.Draft)
	}
	if model.Filter.Applied != "" {
		t.Fatalf("expected empty applied filter, got %q", model.Filter.Applied)
	}
	if model.Filter.ActiveFilterID != "" {
		t.Fatalf("expected no active preset, got %q", model.Filter.ActiveFilterID)
	}
	if len(model.Filter.Saved) != 0 {
		t.Fatalf("expected no saved presets by default, got %#v", model.Filter.Saved)
	}
}

func TestControllerApplyEmptyFilterShowsAllLogs(t *testing.T) {
	events := make(chan session.Event, 2)
	controller := newStreamingController(t, events)

	events <- session.Event{Entry: &logcat.LogEntry{TimeText: "06-04 16:42:18.479", Level: "I", Tag: "ActivityManager", Message: "plain one", Raw: "plain one"}}
	events <- session.Event{Entry: &logcat.LogEntry{TimeText: "06-04 16:42:18.480", Level: "I", Tag: "chromium", Message: "[H5] two", Raw: "[H5] two"}}

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 2
	})

	controller.SetFilterDraft("")
	if err := controller.ApplyFilterDraft(); err != nil {
		t.Fatalf("ApplyFilterDraft returned error: %v", err)
	}

	model := controller.Model()
	if len(model.VisibleLogs) != 2 {
		t.Fatalf("expected all logs visible, got %d", len(model.VisibleLogs))
	}
	if model.Filter.Applied != "" {
		t.Fatalf("expected applied filter cleared, got %q", model.Filter.Applied)
	}
}

func TestControllerLoadADBMissingSetsExplicitStatus(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			err: errors.New("adb_not_found: adb missing"),
		},
		stubSessionStarter{},
	)

	err := controller.Load(context.Background())
	if err == nil {
		t.Fatal("expected load error")
	}

	model := controller.Model()
	if model.Status != "adb_not_found: adb missing" {
		t.Fatalf("expected explicit adb error, got %q", model.Status)
	}
	if len(model.Devices) != 0 {
		t.Fatalf("expected no devices on adb failure, got %d", len(model.Devices))
	}
}

func TestControllerSelectDeviceAppendsLogEvents(t *testing.T) {
	events := make(chan session.Event, 1)
	events <- session.Event{
		Entry: &logcat.LogEntry{Message: `[H5] hello`},
	}
	close(events)

	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
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
	controller.ResumeKeep()

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 1
	})

	model := controller.Model()
	if model.Status != "running" {
		t.Fatalf("expected running status, got %q", model.Status)
	}
	if model.VisibleLogs[0].Entry.Message != `[H5] hello` {
		t.Fatalf("expected log message, got %q", model.VisibleLogs[0].Entry.Message)
	}
}

func TestControllerSelectDeviceLoadsUserPackages(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
				adb.PackageScopeAll: {
					{Name: "com.demo.host"},
					{Name: "com.demo.app"},
				},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}

	model := controller.Model()
	if model.PackageScope != adb.PackageScopeAll {
		t.Fatalf("expected default package scope all, got %q", model.PackageScope)
	}
	if len(model.Packages) != 2 {
		t.Fatalf("expected 2 user packages, got %d", len(model.Packages))
	}
	if model.Packages[0].Name != "com.demo.host" {
		t.Fatalf("unexpected first package: %#v", model.Packages[0])
	}
	if !model.Pause.Active {
		t.Fatal("expected device selection to stay paused before start")
	}
}

func TestControllerSelectDeviceDoesNotStartSessionUntilResume(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}

	starter.mu.Lock()
	configCount := len(starter.configs)
	starter.mu.Unlock()
	if configCount != 0 {
		t.Fatalf("expected no session before resume, got %d", configCount)
	}

	controller.ResumeKeep()

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.configs) == 1
	})
}

func TestControllerRejectsNonReadyDevices(t *testing.T) {
	cases := []struct {
		name           string
		status         string
		expectedStatus string
	}{
		{
			name:           "unauthorized device",
			status:         "unauthorized",
			expectedStatus: "device_unauthorized",
		},
		{
			name:           "offline device",
			status:         "offline",
			expectedStatus: "device_offline",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewController(
				stubDeviceService{
					install: adb.Install{Path: "adb", Version: "1.0.41"},
					devices: []adb.DeviceInfo{
						{ID: "emulator-5554", Model: "Pixel_7", Status: tc.status},
					},
				},
				stubSessionStarter{},
			)

			if err := controller.Load(context.Background()); err != nil {
				t.Fatalf("Load returned error: %v", err)
			}

			err := controller.SelectDevice(context.Background(), "emulator-5554")
			if err == nil {
				t.Fatal("expected select device error")
			}

			model := controller.Model()
			if !strings.Contains(model.Status, tc.expectedStatus) {
				t.Fatalf("expected status containing %q, got %q", tc.expectedStatus, model.Status)
			}
		})
	}
}

func TestControllerCancelsPreviousSessionBeforeStartingNext(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
				{ID: "device-2", Model: "SM_A217F", Status: "device"},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("first SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()
	if err := controller.SelectDevice(context.Background(), "device-2"); err != nil {
		t.Fatalf("second SelectDevice returned error: %v", err)
	}

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.contexts) == 2
	})

	starter.mu.Lock()
	first := starter.contexts[0]
	starter.mu.Unlock()

	select {
	case <-first.Done():
	default:
		t.Fatal("expected first session context to be cancelled")
	}
}

func TestControllerPauseBuffersIncomingLogs(t *testing.T) {
	events := make(chan session.Event, 2)
	controller := newStreamingController(t, events)

	controller.Pause()
	events <- session.Event{Entry: makeEntry("[H5] buffered")}

	waitFor(t, func() bool {
		return controller.Model().Pause.BufferedCount == 1
	})

	model := controller.Model()
	if !model.Pause.Active {
		t.Fatal("expected pause state to be active")
	}
	if len(model.VisibleLogs) != 0 {
		t.Fatalf("expected visible logs to stay unchanged, got %d", len(model.VisibleLogs))
	}
	if !strings.Contains(model.Status, "Paused") {
		t.Fatalf("expected paused status, got %q", model.Status)
	}
}

func TestControllerResumeKeepFlushesBufferedLogs(t *testing.T) {
	events := make(chan session.Event, 2)
	controller := newStreamingController(t, events)

	controller.Pause()
	events <- session.Event{Entry: makeEntry("[H5] keep me")}

	waitFor(t, func() bool {
		return controller.Model().Pause.BufferedCount == 1
	})

	controller.ResumeKeep()

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 1
	})

	model := controller.Model()
	if model.Pause.Active {
		t.Fatal("expected pause state to be cleared")
	}
	if model.Pause.BufferedCount != 0 {
		t.Fatalf("expected buffered count to be reset, got %d", model.Pause.BufferedCount)
	}
	if model.VisibleLogs[0].Entry.Message != "[H5] keep me" {
		t.Fatalf("expected flushed log, got %q", model.VisibleLogs[0].Entry.Message)
	}
}

func TestControllerResumeDiscardDropsBufferedLogs(t *testing.T) {
	events := make(chan session.Event, 2)
	controller := newStreamingController(t, events)

	controller.Pause()
	events <- session.Event{Entry: makeEntry("[H5] discard me")}

	waitFor(t, func() bool {
		return controller.Model().Pause.BufferedCount == 1
	})

	controller.ResumeDiscard()

	model := controller.Model()
	if model.Pause.Active {
		t.Fatal("expected pause state to be cleared")
	}
	if model.Pause.BufferedCount != 0 {
		t.Fatalf("expected buffered count to be reset, got %d", model.Pause.BufferedCount)
	}
	if len(model.VisibleLogs) != 0 {
		t.Fatalf("expected buffered logs to be dropped, got %d", len(model.VisibleLogs))
	}
}

func TestControllerClearVisibleResetsListSelectionAndSearch(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newStreamingController(t, events)

	events <- session.Event{Entry: makeEntry("[H5] token one")}
	events <- session.Event{Entry: makeEntry("[H5] token two")}

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 2
	})

	controller.SetSearchQuery("token")
	controller.NextMatch()
	controller.ClearVisible()

	model := controller.Model()
	if len(model.VisibleLogs) != 0 {
		t.Fatalf("expected visible logs to be cleared, got %d", len(model.VisibleLogs))
	}
	if model.SelectedIndex != -1 {
		t.Fatalf("expected selection reset, got %d", model.SelectedIndex)
	}
	if len(model.Search.MatchIndexes) != 0 {
		t.Fatalf("expected matches to be cleared, got %d", len(model.Search.MatchIndexes))
	}
	if model.Search.Current != -1 {
		t.Fatalf("expected current match reset, got %d", model.Search.Current)
	}
	if model.Search.Query != "token" {
		t.Fatalf("expected query to stay for future logs, got %q", model.Search.Query)
	}
}

func TestControllerSearchTracksMatchesAndCurrentSelection(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newStreamingController(t, events)

	events <- session.Event{Entry: makeEntry("[H5] token one")}
	events <- session.Event{Entry: makeEntry("[H5] miss")}
	events <- session.Event{Entry: makeEntry("[H5] token two")}

	waitFor(t, func() bool {
		return len(controller.Model().VisibleLogs) == 3
	})

	controller.SetSearchQuery("token")

	model := controller.Model()
	if len(model.Search.MatchIndexes) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(model.Search.MatchIndexes))
	}
	if model.Search.Current != 0 {
		t.Fatalf("expected first match selected, got %d", model.Search.Current)
	}
	if model.SelectedIndex != model.Search.MatchIndexes[0] {
		t.Fatalf("expected selection to follow first match, got %d", model.SelectedIndex)
	}

	controller.NextMatch()
	model = controller.Model()
	if model.Search.Current != 1 {
		t.Fatalf("expected second match selected, got %d", model.Search.Current)
	}
	if model.SelectedIndex != model.Search.MatchIndexes[1] {
		t.Fatalf("expected selection to follow second match, got %d", model.SelectedIndex)
	}

	controller.PrevMatch()
	model = controller.Model()
	if model.Search.Current != 0 {
		t.Fatalf("expected previous match to wrap back to first, got %d", model.Search.Current)
	}
}

func TestControllerConsumeRecomputesMatchesWhenLogsArrive(t *testing.T) {
	events := make(chan session.Event, 4)
	controller := newStreamingController(t, events)

	controller.SetSearchQuery("token")
	events <- session.Event{Entry: makeEntry("[H5] token late")}

	waitFor(t, func() bool {
		return len(controller.Model().Search.MatchIndexes) == 1
	})

	model := controller.Model()
	if model.Search.Current != 0 {
		t.Fatalf("expected current match to move to first result, got %d", model.Search.Current)
	}
	if model.SelectedIndex != 0 {
		t.Fatalf("expected selection to follow incoming match, got %d", model.SelectedIndex)
	}
}

func TestControllerUISnapshotWindowsLargeVisibleLogSet(t *testing.T) {
	events := make(chan session.Event, 1105)
	controller := newStreamingController(t, events)

	for index := 0; index < 1105; index++ {
		events <- session.Event{Entry: makeEntry("row")}
	}

	waitFor(t, func() bool {
		return controller.Model().TotalLogs == 1105
	})

	snapshot := controller.UISnapshot(1000)
	if snapshot.VisibleCount != 1105 {
		t.Fatalf("expected full visible count, got %d", snapshot.VisibleCount)
	}
	if snapshot.VisibleStart != 105 {
		t.Fatalf("expected visible window start 105, got %d", snapshot.VisibleStart)
	}
	if len(snapshot.Model.VisibleLogs) != 1000 {
		t.Fatalf("expected 1000 windowed logs, got %d", len(snapshot.Model.VisibleLogs))
	}
}

func TestControllerSelectPackageReplacesBindingAndClearsVisibleLogs(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			processesByPackage: map[string][]adb.ProcessInfo{
				"com.demo.host": {
					{PID: 111, Name: "com.demo.host"},
					{PID: 222, Name: "com.demo.host:webview"},
				},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()

	controller.pushEntry(*makeEntry("[H5] before bind"))
	controller.SetSearchQuery("before")
	controller.NextMatch()

	err := controller.SelectPackage(context.Background(), "com.demo.host")
	if err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.configs) == 2
	})

	model := controller.Model()
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected package updated, got %q", model.SelectedPackage)
	}
	if len(model.Processes) != 2 {
		t.Fatalf("expected 2 related processes, got %d", len(model.Processes))
	}
	if len(model.BoundPIDs) != 2 || model.BoundPIDs[0] != 111 || model.BoundPIDs[1] != 222 {
		t.Fatalf("unexpected bound pids: %#v", model.BoundPIDs)
	}
	if len(model.VisibleLogs) != 0 {
		t.Fatalf("expected visible logs cleared after package bind, got %d", len(model.VisibleLogs))
	}
	if model.SelectedIndex != -1 {
		t.Fatalf("expected selection reset after package bind, got %d", model.SelectedIndex)
	}

	starter.mu.Lock()
	second := starter.configs[1]
	starter.mu.Unlock()
	if second.PackageName != "com.demo.host" {
		t.Fatalf("expected package session config, got %q", second.PackageName)
	}
	if len(second.AllowedPIDs) != 2 || second.AllowedPIDs[1] != 222 {
		t.Fatalf("unexpected allowed pids: %#v", second.AllowedPIDs)
	}
}

func TestControllerSelectProcessNarrowsBindingToSinglePID(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			processesByPackage: map[string][]adb.ProcessInfo{
				"com.demo.host": {
					{PID: 111, Name: "com.demo.host"},
					{PID: 222, Name: "com.demo.host:webview"},
				},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()
	if err := controller.SelectPackage(context.Background(), "com.demo.host"); err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}

	err := controller.SelectProcess(context.Background(), "com.demo.host:webview")
	if err != nil {
		t.Fatalf("SelectProcess returned error: %v", err)
	}

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.configs) == 3
	})

	model := controller.Model()
	if model.SelectedProcess != "com.demo.host:webview" {
		t.Fatalf("expected selected process updated, got %q", model.SelectedProcess)
	}
	if len(model.BoundPIDs) != 1 || model.BoundPIDs[0] != 222 {
		t.Fatalf("unexpected narrowed bound pids: %#v", model.BoundPIDs)
	}

	starter.mu.Lock()
	third := starter.configs[2]
	starter.mu.Unlock()
	if third.ProcessName != "com.demo.host:webview" {
		t.Fatalf("expected process session config, got %q", third.ProcessName)
	}
	if len(third.AllowedPIDs) != 1 || third.AllowedPIDs[0] != 222 {
		t.Fatalf("unexpected process allowed pids: %#v", third.AllowedPIDs)
	}
}

func TestControllerSetPackageScopeRefreshesPackages(t *testing.T) {
	controller := NewController(
		stubDeviceService{
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
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	if err := controller.SetPackageScope(context.Background(), adb.PackageScopeSystem); err != nil {
		t.Fatalf("SetPackageScope returned error: %v", err)
	}

	model := controller.Model()
	if model.PackageScope != adb.PackageScopeSystem {
		t.Fatalf("expected system scope, got %q", model.PackageScope)
	}
	if len(model.Packages) != 1 || model.Packages[0].Name != "com.android.systemui" {
		t.Fatalf("unexpected system packages: %#v", model.Packages)
	}
}

func TestControllerSelectForegroundPackageUsesForegroundPackage(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			foregroundPackage: "com.demo.host",
			processesByPackage: map[string][]adb.ProcessInfo{
				"com.demo.host": {
					{PID: 111, Name: "com.demo.host"},
				},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()
	if err := controller.SelectForegroundPackage(context.Background()); err != nil {
		t.Fatalf("SelectForegroundPackage returned error: %v", err)
	}

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.configs) == 2
	})

	model := controller.Model()
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected foreground package, got %q", model.SelectedPackage)
	}
	if len(model.BoundPIDs) != 1 || model.BoundPIDs[0] != 111 {
		t.Fatalf("unexpected foreground bound pids: %#v", model.BoundPIDs)
	}
}

func TestControllerSelectPackageNotRunningStopsOldBinding(t *testing.T) {
	starter := &recordingSessionStarter{}
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
		},
		starter,
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()

	waitFor(t, func() bool {
		starter.mu.Lock()
		defer starter.mu.Unlock()
		return len(starter.contexts) == 1
	})

	err := controller.SelectPackage(context.Background(), "com.demo.host")
	if err == nil {
		t.Fatal("expected SelectPackage error")
	}

	starter.mu.Lock()
	first := starter.contexts[0]
	configCount := len(starter.configs)
	starter.mu.Unlock()

	select {
	case <-first.Done():
	default:
		t.Fatal("expected prior session to be cancelled")
	}

	model := controller.Model()
	if !strings.Contains(model.Status, "app_not_running") {
		t.Fatalf("expected app_not_running status, got %q", model.Status)
	}
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected package to stay on target, got %q", model.SelectedPackage)
	}
	if len(model.BoundPIDs) != 0 {
		t.Fatalf("expected bound pids to clear, got %#v", model.BoundPIDs)
	}
	if configCount != 1 {
		t.Fatalf("expected no replacement session on not running app, got %d configs", configCount)
	}
}

func TestControllerSelectDeviceClearsBindingState(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
				{ID: "device-2", Model: "SM_A217F", Status: "device"},
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
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("first SelectDevice returned error: %v", err)
	}
	if err := controller.SelectPackage(context.Background(), "com.demo.host"); err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-2"); err != nil {
		t.Fatalf("second SelectDevice returned error: %v", err)
	}

	model := controller.Model()
	if model.SelectedPackage != "" {
		t.Fatalf("expected package selection cleared, got %q", model.SelectedPackage)
	}
	if model.SelectedProcess != "" {
		t.Fatalf("expected process selection cleared, got %q", model.SelectedProcess)
	}
	if len(model.Processes) != 0 {
		t.Fatalf("expected processes cleared, got %#v", model.Processes)
	}
	if len(model.BoundPIDs) != 0 {
		t.Fatalf("expected bound pids cleared, got %#v", model.BoundPIDs)
	}
}

func TestControllerSelectDeviceAllowsClearingSelection(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), ""); err != nil {
		t.Fatalf("clearing device returned error: %v", err)
	}

	model := controller.Model()
	if model.SelectedDevice != "" {
		t.Fatalf("expected device cleared, got %q", model.SelectedDevice)
	}
	if model.PackageScope != "" {
		t.Fatalf("expected package scope cleared, got %q", model.PackageScope)
	}
	if model.Status != "idle" {
		t.Fatalf("expected idle status, got %q", model.Status)
	}
}

func TestControllerSyncDevicesAutoSelectsFirstReadyDevice(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			packagesByScope: map[adb.PackageScope][]adb.PackageInfo{
				adb.PackageScopeAll: {{Name: "com.demo.host"}},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.syncDevices(context.Background(), []adb.DeviceInfo{
		{ID: "device-1", Model: "Pixel_7", Status: "device"},
	}); err != nil {
		t.Fatalf("syncDevices returned error: %v", err)
	}

	model := controller.Model()
	if model.SelectedDevice != "device-1" {
		t.Fatalf("expected auto-selected device-1, got %q", model.SelectedDevice)
	}
	if len(model.Packages) != 1 || model.Packages[0].Name != "com.demo.host" {
		t.Fatalf("expected packages refreshed after auto-select, got %#v", model.Packages)
	}
}

func TestControllerSyncDevicesClearsUnavailableSelection(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "device-1"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	if err := controller.syncDevices(context.Background(), nil); err != nil {
		t.Fatalf("syncDevices returned error: %v", err)
	}

	model := controller.Model()
	if model.SelectedDevice != "" {
		t.Fatalf("expected selected device cleared, got %q", model.SelectedDevice)
	}
	if len(model.Packages) != 0 {
		t.Fatalf("expected packages cleared, got %#v", model.Packages)
	}
	if len(model.BoundPIDs) != 0 {
		t.Fatalf("expected bound pids cleared, got %#v", model.BoundPIDs)
	}
}

func TestControllerReconcileTrackedDevicesPromotesOfflineDeviceToReadySnapshot(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			listDevicesFunc: func() ([]adb.DeviceInfo, error) {
				return []adb.DeviceInfo{
					{ID: "bc82a570", Model: "24122RKC7C", Status: "device"},
				}, nil
			},
		},
		stubSessionStarter{},
	)
	controller.deviceReconcileDelay = 0

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.syncDevices(context.Background(), []adb.DeviceInfo{
		{ID: "bc82a570", Status: "offline"},
	}); err != nil {
		t.Fatalf("syncDevices returned error: %v", err)
	}

	controller.reconcileTrackedDevices()

	model := controller.Model()
	if model.SelectedDevice != "bc82a570" {
		t.Fatalf("expected reconciled selected device, got %q", model.SelectedDevice)
	}
	if len(model.Devices) != 1 || model.Devices[0].Status != "device" || model.Devices[0].Model != "24122RKC7C" {
		t.Fatalf("unexpected reconciled devices: %#v", model.Devices)
	}
}

func TestControllerSelectPackageAllowsClearingSelection(t *testing.T) {
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
		stubSessionStarter{},
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
	if err := controller.SelectPackage(context.Background(), ""); err != nil {
		t.Fatalf("clearing package returned error: %v", err)
	}

	model := controller.Model()
	if model.SelectedPackage != "" {
		t.Fatalf("expected package cleared, got %q", model.SelectedPackage)
	}
	if len(model.Processes) != 0 {
		t.Fatalf("expected process list cleared, got %#v", model.Processes)
	}
	if len(model.BoundPIDs) != 0 {
		t.Fatalf("expected bound pids cleared, got %#v", model.BoundPIDs)
	}
}

func TestControllerRebindsWhenProcessPIDChanges(t *testing.T) {
	starter := &recordingSessionStarter{}
	calls := 0
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
			},
			listProcessesFunc: func(packageName string) ([]adb.ProcessInfo, error) {
				if packageName != "com.demo.host" {
					return nil, nil
				}
				calls++
				if calls == 1 {
					return []adb.ProcessInfo{{PID: 111, Name: "com.demo.host"}}, nil
				}
				return []adb.ProcessInfo{{PID: 222, Name: "com.demo.host"}}, nil
			},
		},
		starter,
	)
	controller.bindingPollInterval = 10 * time.Millisecond

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SelectDevice(context.Background(), "emulator-5554"); err != nil {
		t.Fatalf("SelectDevice returned error: %v", err)
	}
	controller.ResumeKeep()
	if err := controller.SelectPackage(context.Background(), "com.demo.host"); err != nil {
		t.Fatalf("SelectPackage returned error: %v", err)
	}

	waitFor(t, func() bool {
		model := controller.Model()
		return len(model.BoundPIDs) == 1 && model.BoundPIDs[0] == 222
	})

	starter.mu.Lock()
	configCount := len(starter.configs)
	latest := starter.configs[configCount-1]
	starter.mu.Unlock()

	if configCount != 3 {
		t.Fatalf("expected device + package + rebind sessions, got %d", configCount)
	}
	if len(latest.AllowedPIDs) != 1 || latest.AllowedPIDs[0] != 222 {
		t.Fatalf("unexpected rebound allowed pids: %#v", latest.AllowedPIDs)
	}
}

func newStreamingController(t *testing.T, events chan session.Event) *Controller {
	t.Helper()

	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "emulator-5554", Model: "Pixel_7", Status: "device"},
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
	controller.ResumeKeep()

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

func waitFor(t *testing.T, check func() bool) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not met before timeout")
}
