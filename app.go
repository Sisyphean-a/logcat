package main

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/storage"
)

const stateEventName = "state:updated"
const stateAppendEventName = "state:append"
const uiLogWindowSize = 1000
const emitThrottle = 16 * time.Millisecond

type App struct {
	ctx                context.Context
	controller         *appstate.Controller
	lastEmitRev        uint64
	lastEmitState      AppState
	hasEmitState       bool
	lastSelectedSource []int
	lastFocusedSource  int
	selectionTrackRev  uint64
	saveFilterState    func([]appstate.SavedFilter, []string, string) error
	lastPersistedState persistedFilterState
}

type LogSelectionRequest struct {
	Index int                    `json:"index"`
	Mode  appstate.SelectionMode `json:"mode"`
}

func NewApp(controller *appstate.Controller) *App {
	return &App{
		controller:         controller,
		saveFilterState:    storage.SaveFilterState,
		lastPersistedState: newPersistedFilterState(controller.FilterStateSnapshot()),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 异步加载初始状态：Load 内部串行执行多个 adb 子进程调用，同步执行会
	// 阻塞 Wails 首屏。改为 goroutine 后首屏即时渲染，设备信息经 emitState
	// 事件推送补齐。
	go a.loadInitialState()
	go a.trackDevices(ctx)
	go a.pushStateLoop(ctx)
}

func (a *App) GetState() AppState {
	return newAppState(a.controller.UISnapshot(uiLogWindowSize))
}

func (a *App) SelectDevice(deviceID string) error {
	return a.runAction(func() error {
		return a.controller.SelectDevice(context.Background(), deviceID)
	})
}

func (a *App) SelectPackage(packageName string) error {
	return a.runAction(func() error {
		return a.controller.SelectPackage(context.Background(), packageName)
	})
}

func (a *App) SelectProcess(processName string) error {
	return a.runAction(func() error {
		return a.controller.SelectProcess(context.Background(), processName)
	})
}

func (a *App) SelectForegroundPackage() error {
	return a.runAction(func() error {
		return a.controller.SelectForegroundPackage(context.Background())
	})
}

func (a *App) SetPackageScope(scope string) error {
	return a.runAction(func() error {
		return a.controller.SetPackageScope(context.Background(), appstatePackageScope(scope))
	})
}

func (a *App) SetFilterDraft(query string) AppState {
	a.controller.SetFilterDraft(query)
	return a.emitAndSnapshot()
}

func (a *App) ApplyFilterDraft() error {
	return a.runAction(a.controller.ApplyFilterDraft)
}

func (a *App) ApplySavedFilter(filterID string) error {
	return a.runAction(func() error {
		return a.controller.ApplySavedFilter(context.Background(), filterID)
	})
}

func (a *App) ApplyHistoryQuery(query string) error {
	return a.runAction(func() error {
		return a.controller.ApplyHistoryQuery(query)
	})
}

func (a *App) SaveCurrentFilter(name string) error {
	return a.runAction(func() error {
		return a.controller.SaveCurrentFilter(name)
	})
}

func (a *App) SaveFilterDefinition(name string, packageName string, query string) error {
	return a.runAction(func() error {
		return a.controller.SaveFilterDefinition(name, packageName, query)
	})
}

func (a *App) UpdateSavedFilterDefinition(existingID string, name string, packageName string, query string) error {
	return a.runAction(func() error {
		return a.controller.UpdateSavedFilterDefinition(context.Background(), existingID, name, packageName, query)
	})
}

func (a *App) ReplaceSavedFilterDefinitions(
	drafts []appstate.SavedFilterDraft,
	defaultFilterID string,
	activeFilterID string,
) error {
	return a.runAction(func() error {
		return a.controller.ReplaceSavedFilterDefinitions(
			context.Background(),
			drafts,
			defaultFilterID,
			activeFilterID,
		)
	})
}

func (a *App) Pause() AppState {
	a.controller.Pause()
	return a.emitAndSnapshot()
}

func (a *App) ResumeKeep() AppState {
	a.controller.ResumeKeep()
	return a.emitAndSnapshot()
}

func (a *App) ResumeDiscard() AppState {
	a.controller.ResumeDiscard()
	return a.emitAndSnapshot()
}

func (a *App) ClearVisible() AppState {
	a.controller.ClearVisible()
	return a.emitAndSnapshot()
}

func (a *App) SetSearchQuery(query string) AppState {
	a.controller.SetSearchQuery(query)
	return a.emitAndSnapshot()
}

func (a *App) NextMatch() SelectionPatch {
	a.controller.NextMatch()
	return a.selectionPatchSnapshot()
}

func (a *App) PrevMatch() SelectionPatch {
	a.controller.PrevMatch()
	return a.selectionPatchSnapshot()
}

func (a *App) SelectLog(index int) SelectionPatch {
	a.controller.SelectLog(index)
	return a.selectionPatchSnapshot()
}

func (a *App) SelectLogs(request LogSelectionRequest) SelectionPatch {
	a.controller.SelectLogWithMode(request.Index, request.Mode)
	return a.selectionPatchSnapshot()
}

func (a *App) CopySelectedLogs() error {
	return a.CopyText(a.controller.SelectedLogsText())
}

func (a *App) GetSelectedLogRaw() string {
	item, ok := a.controller.SelectedLog()
	if !ok {
		return ""
	}
	return item.Entry.Raw
}

func (a *App) CopyAllVisibleLogs() error {
	return a.CopyText(a.controller.VisibleLogsText())
}

func (a *App) ExportVisibleLogs() (string, error) {
	logs := a.controller.VisibleLogsSnapshot()
	path, err := storage.ExportVisibleLogs(logs)
	if err != nil {
		a.controller.SetStatus(err.Error())
		a.emitState()
		return "", err
	}

	message := fmt.Sprintf("已导出 %d 条到 Downloads/%s", len(logs), filepath.Base(path))
	a.controller.SetStatus(message)
	a.emitState()
	return path, nil
}

func (a *App) CopyText(value string) error {
	if a.ctx == nil {
		return fmt.Errorf("runtime_not_ready")
	}
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return runtime.ClipboardSetText(a.ctx, value)
}

func (a *App) loadInitialState() {
	_ = a.controller.Load(context.Background())
	a.emitState()
}

func (a *App) pushStateLoop(ctx context.Context) {
	timer := newStoppedTimer()
	defer timer.Stop()
	pending := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.controller.Dirty():
			if pending {
				continue
			}
			pending = true
			timer.Reset(emitThrottle)
		case <-timer.C:
			pending = false
			a.emitStateIfDirty()
		}
	}
}

func (a *App) trackDevices(ctx context.Context) {
	if err := a.controller.TrackDevices(ctx); err != nil {
		a.controller.SetStatus(err.Error())
		a.emitState()
	}
}

func (a *App) runAction(action func() error) error {
	err := action()
	a.persistFilters()
	a.emitState()
	return err
}

func (a *App) persistFilters() {
	state := newPersistedFilterState(a.controller.FilterStateSnapshot())
	if a.lastPersistedState.equal(state) {
		return
	}
	if a.saveFilterState(state.Filters, state.History, state.DefaultFilterID) == nil {
		a.lastPersistedState = state
	}
}

func (a *App) emitState() {
	a.emitStateIfDirty()
}

func (a *App) emitStateIfDirty() {
	if a.ctx == nil {
		return
	}
	snapshot := a.controller.UISnapshot(uiLogWindowSize)
	if snapshot.Revision == a.lastEmitRev {
		return
	}
	a.lastEmitRev = snapshot.Revision
	if a.hasEmitState {
		if patch, ok := buildStateAppendPatch(a.lastEmitState, snapshot); ok {
			a.lastEmitState = applyAppendPatch(a.lastEmitState, patch)
			runtime.EventsEmit(a.ctx, stateAppendEventName, patch)
			return
		}
	}
	state := newAppState(snapshot)
	a.lastEmitState = state
	a.hasEmitState = true
	a.updateSelectionTracking(state)
	runtime.EventsEmit(a.ctx, stateEventName, state)
}

func (a *App) emitAndSnapshot() AppState {
	snapshot := a.controller.UISnapshot(uiLogWindowSize)
	a.lastEmitRev = snapshot.Revision
	// 直返型 RPC 的调用方会直接消费返回值；这里再发同一份 state:updated
	// 事件只会造成重复序列化、重复前端 setState 和额外瞬时内存。
	state := newAppState(snapshot)
	a.lastEmitState = state
	a.hasEmitState = true
	a.updateSelectionTracking(state)
	return state
}

func (a *App) selectionPatchSnapshot() SelectionPatch {
	snapshot := a.controller.SelectionSnapshot(uiLogWindowSize)
	patch := buildSelectionPatchFromSnapshot(snapshot)
	a.lastEmitRev = snapshot.Revision
	selection := a.resolveTrackedSelection()
	a.lastEmitState = applySelectionPatch(
		a.lastEmitState,
		patch,
		selection.selectedSourceIndexes,
		selection.focusedSourceIndex,
	)
	a.hasEmitState = true
	a.lastSelectedSource = append(a.lastSelectedSource[:0], patch.SelectedSourceIndexes...)
	a.lastFocusedSource = patch.FocusedSourceIndex
	a.selectionTrackRev = patch.Revision
	return patch
}

func (a *App) updateSelectionTracking(state AppState) {
	a.lastSelectedSource = collectSelectedSourceIndexes(state.Logs, state.SelectedCount)
	a.lastFocusedSource = focusedSourceIndex(state.SelectedLog)
	a.selectionTrackRev = state.Revision
}

func (a *App) resolveTrackedSelection() trackedSelectionState {
	if a.selectionTrackRev == a.lastEmitState.Revision {
		return trackedSelectionState{
			selectedSourceIndexes: a.lastSelectedSource,
			focusedSourceIndex:    a.lastFocusedSource,
		}
	}
	return trackedSelectionState{
		selectedSourceIndexes: collectSelectedSourceIndexes(a.lastEmitState.Logs, a.lastEmitState.SelectedCount),
		focusedSourceIndex:    focusedSourceIndex(a.lastEmitState.SelectedLog),
	}
}

func newStoppedTimer() *time.Timer {
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	return timer
}

type persistedFilterState struct {
	Filters         []appstate.SavedFilter
	History         []string
	DefaultFilterID string
}

func newPersistedFilterState(filter appstate.FilterState) persistedFilterState {
	return persistedFilterState{
		Filters:         append([]appstate.SavedFilter(nil), filter.Saved...),
		History:         append([]string(nil), filter.History...),
		DefaultFilterID: filter.DefaultFilterID,
	}
}

func (s persistedFilterState) equal(other persistedFilterState) bool {
	return s.DefaultFilterID == other.DefaultFilterID &&
		slices.Equal(s.History, other.History) &&
		slices.Equal(s.Filters, other.Filters)
}
