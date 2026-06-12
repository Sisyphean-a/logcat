package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/storage"
)

const stateEventName = "state:updated"
const uiLogWindowSize = 1000
const emitThrottle = 16 * time.Millisecond

type App struct {
	ctx         context.Context
	controller  *appstate.Controller
	lastEmitRev uint64
}

func NewApp(controller *appstate.Controller) *App {
	return &App{controller: controller}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadInitialState()
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

func (a *App) NextMatch() AppState {
	a.controller.NextMatch()
	return a.emitAndSnapshot()
}

func (a *App) PrevMatch() AppState {
	a.controller.PrevMatch()
	return a.emitAndSnapshot()
}

func (a *App) SelectLog(index int) AppState {
	a.controller.SelectLog(index)
	return a.emitAndSnapshot()
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
	filter := a.controller.FilterStateSnapshot()
	_ = storage.SaveFilterState(filter.Saved, filter.History, filter.DefaultFilterID)
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
	state := newAppState(snapshot)
	runtime.EventsEmit(a.ctx, stateEventName, state)
}

func (a *App) emitAndSnapshot() AppState {
	snapshot := a.controller.UISnapshot(uiLogWindowSize)
	a.lastEmitRev = snapshot.Revision
	state := newAppState(snapshot)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, stateEventName, state)
	}
	return state
}

func newStoppedTimer() *time.Timer {
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	return timer
}
