package main

import (
	"slices"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
)

type StateAppendPatch struct {
	Revision      uint64           `json:"revision"`
	TotalLogs     int              `json:"totalLogs"`
	VisibleCount  int              `json:"visibleCount"`
	Dropped       int              `json:"dropped"`
	Appended      []LogItemView    `json:"appended"`
	SelectedCount int              `json:"selectedCount"`
	SelectedLog   *SelectedLogView `json:"selectedLog"`
}

func buildStateAppendPatch(prev AppState, snapshot appstate.UISnapshot) (StateAppendPatch, bool) {
	if !sameAppendPatchSnapshotContext(prev, snapshot) {
		return StateAppendPatch{}, false
	}

	dropped, appendedStart, ok := diffAppendSnapshotWindow(prev.Logs, snapshot.Model.VisibleLogs, snapshot.Model.Selection)
	if !ok {
		return StateAppendPatch{}, false
	}

	appendedLogs := make([]LogItemView, len(snapshot.Model.VisibleLogs)-appendedStart)
	cursor := newLogRowCursor(snapshot.Model.Selection)
	for index := range snapshot.Model.VisibleLogs {
		row := cursor.Next(snapshot.Model.VisibleLogs[index])
		if index >= appendedStart {
			appendedLogs[index-appendedStart] = row
		}
	}

	return StateAppendPatch{
		Revision:      snapshot.Revision,
		TotalLogs:     snapshot.Model.TotalLogs,
		VisibleCount:  snapshot.VisibleCount,
		Dropped:       dropped,
		Appended:      appendedLogs,
		SelectedCount: len(snapshot.Model.Selection.SourceIndexes),
		SelectedLog:   buildSnapshotSelectedLog(snapshot.Model.VisibleLogs, snapshot.Model.Selection),
	}, true
}

func sameAppendPatchSnapshotContext(prev AppState, snapshot appstate.UISnapshot) bool {
	model := snapshot.Model
	if prev.Status != model.Status ||
		prev.ADBStatus != model.ADBStatus ||
		prev.SelectedDevice != model.SelectedDevice ||
		prev.PackageScope != string(model.PackageScope) ||
		prev.SelectedPackage != model.SelectedPackage ||
		prev.Filter.Draft != model.Filter.Draft ||
		prev.Filter.Applied != model.Filter.Applied ||
		prev.Filter.Error != model.Filter.Error ||
		prev.Filter.ActiveFilterID != model.Filter.ActiveFilterID ||
		prev.Filter.DefaultFilterID != model.Filter.DefaultFilterID ||
		prev.Search.Query != model.Search.Query ||
		prev.Pause.Active != model.Pause.Active {
		return false
	}
	if prev.SelectedCount != len(model.Selection.SourceIndexes) {
		return false
	}

	return slices.EqualFunc(prev.Devices, model.Devices, sameDeviceItemView) &&
		slices.EqualFunc(prev.Packages, model.Packages, samePackageInfoView) &&
		slices.EqualFunc(prev.Filter.Saved, model.Filter.Saved, sameSavedFilterItemView) &&
		sameSelectedLogView(prev.SelectedLog, buildSnapshotSelectedLog(model.VisibleLogs, model.Selection))
}

func sameDeviceItemView(left DeviceView, right appstate.DeviceItem) bool {
	return left.ID == right.ID && left.Model == right.Model && left.Status == right.Status
}

func samePackageInfoView(left PackageView, right adb.PackageInfo) bool {
	return left.Name == right.Name
}

func sameSavedFilterItemView(left SavedFilterView, right appstate.SavedFilter) bool {
	return left.ID == right.ID &&
		left.Name == right.Name &&
		left.PackageName == right.PackageName &&
		left.Query == right.Query
}

func diffAppendSnapshotWindow(
	prev []LogItemView,
	current []appstate.LogViewItem,
	selection appstate.SelectionState,
) (int, int, bool) {
	prevMax := -1
	if len(prev) > 0 {
		prevMax = prev[len(prev)-1].SourceIndex
	}

	appendedStart := len(current)
	for index, row := range current {
		if row.SourceIndex > prevMax {
			appendedStart = index
			break
		}
	}

	appendedCount := len(current) - appendedStart
	dropped := len(prev) + appendedCount - len(current)
	if dropped < 0 || dropped > len(prev) {
		return 0, 0, false
	}

	overlap := len(current) - appendedCount
	if overlap < 0 || dropped+overlap > len(prev) {
		return 0, 0, false
	}

	cursor := newLogRowCursor(selection)
	for index := 0; index < overlap; index++ {
		row := cursor.Next(current[index])
		if prev[dropped+index] != row {
			return 0, 0, false
		}
	}

	return dropped, appendedStart, true
}

func cloneSelectedLog(selected *SelectedLogView) *SelectedLogView {
	if selected == nil {
		return nil
	}
	cloned := *selected
	return &cloned
}

func sameSelectedLogView(left *SelectedLogView, right *SelectedLogView) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func buildSnapshotSelectedLog(items []appstate.LogViewItem, selection appstate.SelectionState) *SelectedLogView {
	if selection.FocusSourceIndex < 0 {
		return nil
	}
	for _, item := range items {
		if item.SourceIndex != selection.FocusSourceIndex {
			continue
		}
		row := LogItemView{
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
			IsFocused:   true,
		}
		return buildSelectedLogView(row, item.Entry.Source)
	}
	return nil
}

func applyAppendPatch(state AppState, patch StateAppendPatch) AppState {
	next := state
	next.Revision = patch.Revision
	next.TotalLogs = patch.TotalLogs
	next.VisibleCount = patch.VisibleCount
	next.SelectedCount = patch.SelectedCount
	if patch.Dropped > 0 {
		next.Logs = append([]LogItemView(nil), next.Logs[patch.Dropped:]...)
	} else {
		next.Logs = append([]LogItemView(nil), next.Logs...)
	}
	next.Logs = append(next.Logs, patch.Appended...)
	next.SelectedLog = cloneSelectedLog(patch.SelectedLog)
	return next
}
