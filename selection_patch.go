package main

import (
	"slices"

	appstate "github.com/xiakn/logcat/internal/app"
)

type SelectionPatch struct {
	Revision              uint64           `json:"revision"`
	SelectedCount         int              `json:"selectedCount"`
	FocusedSourceIndex    int              `json:"focusedSourceIndex"`
	SelectedSourceIndexes []int            `json:"selectedSourceIndexes"`
	SelectedLog           *SelectedLogView `json:"selectedLog"`
}

func buildSelectionPatch(snapshot appstate.UISnapshot) SelectionPatch {
	selected := append([]int(nil), snapshot.Model.Selection.SourceIndexes...)
	return SelectionPatch{
		Revision:              snapshot.Revision,
		SelectedCount:         len(selected),
		FocusedSourceIndex:    snapshot.Model.Selection.FocusSourceIndex,
		SelectedSourceIndexes: selected,
		SelectedLog:           buildSnapshotSelectedLog(snapshot.VisibleLogs, snapshot.Model.Selection),
	}
}

func buildSelectionPatchFromSnapshot(snapshot appstate.SelectionSnapshot) SelectionPatch {
	selected := append([]int(nil), snapshot.Selection.SourceIndexes...)
	return SelectionPatch{
		Revision:              snapshot.Revision,
		SelectedCount:         len(selected),
		FocusedSourceIndex:    snapshot.Selection.FocusSourceIndex,
		SelectedSourceIndexes: selected,
		SelectedLog:           buildFocusedSelectedLog(snapshot.Focused, snapshot.Selection.FocusSourceIndex),
	}
}

func applySelectionPatch(state AppState, patch SelectionPatch) AppState {
	next := state
	next.Revision = patch.Revision
	next.SelectedCount = patch.SelectedCount
	next.Logs = append([]LogItemView(nil), state.Logs...)
	selected := patch.SelectedSourceIndexes
	for index := range next.Logs {
		sourceIndex := next.Logs[index].SourceIndex
		next.Logs[index].IsFocused = sourceIndex == patch.FocusedSourceIndex
		next.Logs[index].IsSelected = slices.Contains(selected, sourceIndex)
	}
	next.SelectedLog = cloneSelectedLog(patch.SelectedLog)
	return next
}

func buildFocusedSelectedLog(item *appstate.LogViewItem, focusedSourceIndex int) *SelectedLogView {
	if item == nil || item.SourceIndex != focusedSourceIndex {
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
