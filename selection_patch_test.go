package main

import (
	"testing"

	appstate "github.com/xiakn/logcat/internal/app"
)

func TestBuildSelectionPatch(t *testing.T) {
	logs := []appstate.LogViewItem{
		logViewItem(1, "t1", "I", "tag", "one"),
		logViewItem(2, "t2", "W", "tag", "two"),
		logViewItem(3, "t3", "E", "tag", "three"),
	}
	snapshot := appstate.UISnapshot{
		Revision:     9,
		VisibleCount: 3,
		VisibleLogs:  visibleLogSnapshots(logs),
		Model: appstate.Model{
			Selection: appstate.SelectionState{
				AnchorSourceIndex: 2,
				FocusSourceIndex:  2,
				SourceIndexes:     []int{2},
			},
		},
	}

	patch := buildSelectionPatch(snapshot)
	if patch.Revision != 9 || patch.SelectedCount != 1 || patch.FocusedSourceIndex != 2 {
		t.Fatalf("unexpected patch summary: %+v", patch)
	}
	if len(patch.SelectedSourceIndexes) != 1 || patch.SelectedSourceIndexes[0] != 2 {
		t.Fatalf("unexpected selected indexes: %+v", patch.SelectedSourceIndexes)
	}
	if patch.SelectedLog == nil || patch.SelectedLog.SourceIndex != 2 {
		t.Fatalf("unexpected selected log: %+v", patch.SelectedLog)
	}
}

func TestBuildSelectionPatchFromSelectionSnapshot(t *testing.T) {
	snapshot := appstate.SelectionSnapshot{
		Revision: 12,
		Selection: appstate.SelectionState{
			AnchorSourceIndex: 2,
			FocusSourceIndex:  2,
			SourceIndexes:     []int{2},
		},
		Focused: func() *appstate.LogViewItem {
			item := logViewItem(2, "t2", "W", "tag", "two")
			return &item
		}(),
	}

	patch := buildSelectionPatchFromSnapshot(snapshot)
	if patch.Revision != 12 || patch.SelectedCount != 1 || patch.FocusedSourceIndex != 2 {
		t.Fatalf("unexpected patch summary: %+v", patch)
	}
	if patch.SelectedLog == nil || patch.SelectedLog.SourceIndex != 2 {
		t.Fatalf("unexpected selected log: %+v", patch.SelectedLog)
	}
}

func TestApplySelectionPatch(t *testing.T) {
	state := AppState{
		Revision:      7,
		SelectedCount: 1,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one", IsFocused: true, IsSelected: true},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two"},
			{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three"},
		},
		SelectedLog: &SelectedLogView{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one", Source: "src"},
	}
	patch := SelectionPatch{
		Revision:              8,
		SelectedCount:         2,
		FocusedSourceIndex:    3,
		SelectedSourceIndexes: []int{2, 3},
		SelectedLog:           &SelectedLogView{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three", Source: "src"},
	}

	next := applySelectionPatch(state, patch, []int{1}, 1)
	if next.Revision != 8 || next.SelectedCount != 2 {
		t.Fatalf("unexpected state summary: %+v", next)
	}
	if next.Logs[0].IsSelected || next.Logs[0].IsFocused {
		t.Fatalf("expected first row cleared, got %+v", next.Logs[0])
	}
	if !next.Logs[1].IsSelected || next.Logs[1].IsFocused {
		t.Fatalf("expected second row selected only, got %+v", next.Logs[1])
	}
	if !next.Logs[2].IsSelected || !next.Logs[2].IsFocused {
		t.Fatalf("expected third row selected and focused, got %+v", next.Logs[2])
	}
	if next.SelectedLog == nil || next.SelectedLog.SourceIndex != 3 {
		t.Fatalf("unexpected selected log: %+v", next.SelectedLog)
	}
}

func TestApplySelectionPatchClearsFocusedButUnchangedRows(t *testing.T) {
	state := AppState{
		Revision:      7,
		SelectedCount: 2,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one", IsSelected: true},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", IsFocused: true},
			{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three", IsSelected: true},
		},
		SelectedLog: &SelectedLogView{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", Source: "src"},
	}
	patch := SelectionPatch{
		Revision:              8,
		SelectedCount:         2,
		FocusedSourceIndex:    3,
		SelectedSourceIndexes: []int{1, 3},
		SelectedLog:           &SelectedLogView{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three", Source: "src"},
	}

	next := applySelectionPatch(state, patch, []int{1, 3}, 2)
	if next.Logs[0].IsFocused || !next.Logs[0].IsSelected {
		t.Fatalf("expected first row selected only, got %+v", next.Logs[0])
	}
	if next.Logs[1].IsFocused || next.Logs[1].IsSelected {
		t.Fatalf("expected second row cleared, got %+v", next.Logs[1])
	}
	if !next.Logs[2].IsFocused || !next.Logs[2].IsSelected {
		t.Fatalf("expected third row focused and selected, got %+v", next.Logs[2])
	}
}
