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
	selectedLog, ok := appendPatchSelectedLog(prev, snapshot)
	if !ok {
		return StateAppendPatch{}, false
	}

	dropped, appendedStart, ok := diffAppendSnapshotWindow(prev.Logs, snapshot.Model.VisibleLogs, snapshot.Model.Selection)
	if !ok {
		return StateAppendPatch{}, false
	}

	appendedLogs := buildAppendedLogRows(snapshot.Model.VisibleLogs[appendedStart:])

	return StateAppendPatch{
		Revision:      snapshot.Revision,
		TotalLogs:     snapshot.Model.TotalLogs,
		VisibleCount:  snapshot.VisibleCount,
		Dropped:       dropped,
		Appended:      appendedLogs,
		SelectedCount: len(snapshot.Model.Selection.SourceIndexes),
		SelectedLog:   selectedLog,
	}, true
}

func appendPatchSelectedLog(prev AppState, snapshot appstate.UISnapshot) (*SelectedLogView, bool) {
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
		return nil, false
	}
	if prev.SelectedCount != len(model.Selection.SourceIndexes) {
		return nil, false
	}

	selectedLog := buildSnapshotSelectedLog(model.VisibleLogs, model.Selection)
	if !slices.EqualFunc(prev.Devices, model.Devices, sameDeviceItemView) ||
		!slices.EqualFunc(prev.Packages, model.Packages, samePackageInfoView) ||
		!slices.EqualFunc(prev.Filter.Saved, model.Filter.Saved, sameSavedFilterItemView) ||
		!sameSelectedLogView(prev.SelectedLog, selectedLog) {
		return nil, false
	}
	return selectedLog, true
}

func sameAppendPatchSnapshotContext(prev AppState, snapshot appstate.UISnapshot) bool {
	_, ok := appendPatchSelectedLog(prev, snapshot)
	return ok
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

	appendedStart := findFirstSourceIndexAfter(current, prevMax)

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

func buildAppendedLogRows(items []appstate.LogViewItem) []LogItemView {
	if len(items) == 0 {
		return nil
	}
	// 走 append patch 前已验证 selection/selectedLog 没变化，
	// 所以新增尾部行不可能是 focused/selected，直接按原始字段投影即可。
	rows := make([]LogItemView, len(items))
	for index, item := range items {
		rows[index] = LogItemView{
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
		}
	}
	return rows
}

func findFirstSourceIndexAfter(items []appstate.LogViewItem, sourceIndex int) int {
	low := 0
	high := len(items)
	for low < high {
		middle := low + (high-low)/2
		if items[middle].SourceIndex <= sourceIndex {
			low = middle + 1
			continue
		}
		high = middle
	}
	return low
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
	item, ok := findLogViewItemBySourceIndex(items, selection.FocusSourceIndex)
	if !ok {
		return nil
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

func findLogViewItemBySourceIndex(items []appstate.LogViewItem, sourceIndex int) (appstate.LogViewItem, bool) {
	low := 0
	high := len(items) - 1
	for low <= high {
		middle := low + (high-low)/2
		current := items[middle]
		switch {
		case current.SourceIndex == sourceIndex:
			return current, true
		case current.SourceIndex < sourceIndex:
			low = middle + 1
		default:
			high = middle - 1
		}
	}
	return appstate.LogViewItem{}, false
}

func applyAppendPatch(state AppState, patch StateAppendPatch) AppState {
	next := state
	next.Revision = patch.Revision
	next.TotalLogs = patch.TotalLogs
	next.VisibleCount = patch.VisibleCount
	next.SelectedCount = patch.SelectedCount
	next.Logs = mergePatchedLogs(state.Logs, patch.Dropped, patch.Appended)
	if sameSelectedLogView(state.SelectedLog, patch.SelectedLog) {
		next.SelectedLog = state.SelectedLog
		return next
	}
	next.SelectedLog = cloneSelectedLog(patch.SelectedLog)
	return next
}

func mergePatchedLogs(current []LogItemView, dropped int, appended []LogItemView) []LogItemView {
	retainedStart := dropped
	if retainedStart < 0 {
		retainedStart = 0
	}
	if retainedStart > len(current) {
		retainedStart = len(current)
	}
	retainedCount := len(current) - retainedStart
	next := make([]LogItemView, retainedCount+len(appended))
	copy(next, current[retainedStart:])
	copy(next[retainedCount:], appended)
	return next
}
