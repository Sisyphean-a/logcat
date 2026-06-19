package main

import appstate "github.com/xiakn/logcat/internal/app"

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

type trackedSelectionState struct {
	selectedSourceIndexes []int
	focusedSourceIndex    int
}

func applySelectionPatch(
	state AppState,
	patch SelectionPatch,
	previousSelectedSourceIndexes []int,
	previousFocusedSourceIndex int,
) AppState {
	next := state
	next.Revision = patch.Revision
	next.SelectedCount = patch.SelectedCount
	next.Logs = applySelectionRows(
		state.Logs,
		previousSelectedSourceIndexes,
		patch.SelectedSourceIndexes,
		previousFocusedSourceIndex,
		patch.FocusedSourceIndex,
	)
	if sameSelectedLogView(state.SelectedLog, patch.SelectedLog) {
		next.SelectedLog = state.SelectedLog
		return next
	}
	next.SelectedLog = cloneSelectedLog(patch.SelectedLog)
	return next
}

func focusedSourceIndex(selectedLog *SelectedLogView) int {
	if selectedLog == nil {
		return -1
	}
	return selectedLog.SourceIndex
}

func collectSelectedSourceIndexes(logs []LogItemView, selectedCount int) []int {
	if selectedCount <= 0 {
		return nil
	}
	selected := make([]int, 0, selectedCount)
	for _, row := range logs {
		if !row.IsSelected {
			continue
		}
		selected = append(selected, row.SourceIndex)
		if len(selected) == selectedCount {
			break
		}
	}
	return selected
}

func applySelectionRows(
	logs []LogItemView,
	previousSelected []int,
	nextSelected []int,
	previousFocused int,
	nextFocused int,
) []LogItemView {
	nextLogs := logs
	cloned := false
	previousIndex := 0
	nextIndex := 0
	for previousIndex < len(previousSelected) && nextIndex < len(nextSelected) {
		previousSource := previousSelected[previousIndex]
		nextSource := nextSelected[nextIndex]
		switch {
		case previousSource == nextSource:
			previousIndex++
			nextIndex++
		case previousSource < nextSource:
			nextLogs, cloned = applySelectionRow(logs, nextLogs, cloned, previousSource, nextSelected, nextFocused)
			previousIndex++
		default:
			nextLogs, cloned = applySelectionRow(logs, nextLogs, cloned, nextSource, nextSelected, nextFocused)
			nextIndex++
		}
	}
	for ; previousIndex < len(previousSelected); previousIndex++ {
		nextLogs, cloned = applySelectionRow(logs, nextLogs, cloned, previousSelected[previousIndex], nextSelected, nextFocused)
	}
	for ; nextIndex < len(nextSelected); nextIndex++ {
		nextLogs, cloned = applySelectionRow(logs, nextLogs, cloned, nextSelected[nextIndex], nextSelected, nextFocused)
	}
	nextLogs, cloned = applySelectionRow(logs, nextLogs, cloned, previousFocused, nextSelected, nextFocused)
	nextLogs, _ = applySelectionRow(logs, nextLogs, cloned, nextFocused, nextSelected, nextFocused)
	return nextLogs
}

func applySelectionRow(
	logs []LogItemView,
	nextLogs []LogItemView,
	cloned bool,
	sourceIndex int,
	nextSelected []int,
	nextFocused int,
) ([]LogItemView, bool) {
	if sourceIndex < 0 {
		return nextLogs, cloned
	}
	logIndex := findLogItemViewIndexBySource(logs, sourceIndex)
	if logIndex == -1 {
		return nextLogs, cloned
	}
	row := logs[logIndex]
	isFocused := sourceIndex == nextFocused
	isSelected := hasSortedInt(nextSelected, sourceIndex)
	if row.IsFocused == isFocused && row.IsSelected == isSelected {
		return nextLogs, cloned
	}
	if !cloned {
		nextLogs = append([]LogItemView(nil), logs...)
		cloned = true
	}
	nextLogs[logIndex] = cloneLogItemViewWithSelection(row, isFocused, isSelected)
	return nextLogs, cloned
}

func findLogItemViewIndexBySource(logs []LogItemView, sourceIndex int) int {
	low := 0
	high := len(logs) - 1
	for low <= high {
		middle := low + (high-low)/2
		current := logs[middle].SourceIndex
		switch {
		case current == sourceIndex:
			return middle
		case current < sourceIndex:
			low = middle + 1
		default:
			high = middle - 1
		}
	}
	return -1
}

func hasSortedInt(items []int, target int) bool {
	low := 0
	high := len(items) - 1
	for low <= high {
		middle := low + (high-low)/2
		current := items[middle]
		switch {
		case current == target:
			return true
		case current < target:
			low = middle + 1
		default:
			high = middle - 1
		}
	}
	return false
}

func cloneLogItemViewWithSelection(row LogItemView, isFocused bool, isSelected bool) LogItemView {
	return LogItemView{
		SourceIndex: row.SourceIndex,
		TimeText:    row.TimeText,
		Level:       row.Level,
		Tag:         row.Tag,
		Message:     row.Message,
		IsFocused:   isFocused,
		IsSelected:  isSelected,
	}
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
