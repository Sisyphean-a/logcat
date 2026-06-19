package main

import (
	"testing"

	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
)

func TestBuildStateAppendPatchAppendedRows(t *testing.T) {
	logs := []appstate.LogViewItem{
		logViewItem(1, "t1", "I", "tag", "one"),
		logViewItem(2, "t2", "I", "tag", "two"),
		logViewItem(3, "t3", "I", "tag", "three"),
		logViewItem(4, "t4", "W", "tag", "four"),
		logViewItem(5, "t5", "E", "tag", "five"),
	}
	prev := AppState{
		Revision:     10,
		TotalLogs:    3,
		VisibleCount: 3,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one"},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two"},
			{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three"},
		},
	}
	current := appstate.UISnapshot{
		Revision:     11,
		VisibleCount: 5,
		VisibleLogs:  visibleLogSnapshots(logs),
		Model: appstate.Model{
			TotalLogs: 5,
		},
	}

	patch, ok := buildStateAppendPatch(prev, current)
	if !ok {
		t.Fatal("expected append patch")
	}
	if patch.Dropped != 0 {
		t.Fatalf("expected dropped 0, got %d", patch.Dropped)
	}
	if len(patch.Appended) != 2 {
		t.Fatalf("expected 2 appended rows, got %d", len(patch.Appended))
	}
	if patch.Appended[0].SourceIndex != 4 || patch.Appended[1].SourceIndex != 5 {
		t.Fatalf("unexpected appended rows: %+v", patch.Appended)
	}
}

func TestBuildStateAppendPatchSlidingWindow(t *testing.T) {
	logs := []appstate.LogViewItem{
		logViewItem(8, "t8", "I", "tag", "eight"),
		logViewItem(9, "t9", "I", "tag", "nine"),
		logViewItem(10, "t10", "I", "tag", "ten"),
	}
	prev := AppState{
		Revision:     20,
		TotalLogs:    1000,
		VisibleCount: 3,
		Logs: []LogItemView{
			{SourceIndex: 7, TimeText: "t7", Level: "I", Tag: "tag", Message: "seven"},
			{SourceIndex: 8, TimeText: "t8", Level: "I", Tag: "tag", Message: "eight"},
			{SourceIndex: 9, TimeText: "t9", Level: "I", Tag: "tag", Message: "nine"},
		},
	}
	current := appstate.UISnapshot{
		Revision:     21,
		VisibleCount: 3,
		VisibleLogs:  visibleLogSnapshots(logs),
		Model: appstate.Model{
			TotalLogs: 1001,
		},
	}

	patch, ok := buildStateAppendPatch(prev, current)
	if !ok {
		t.Fatal("expected append patch")
	}
	if patch.Dropped != 1 {
		t.Fatalf("expected dropped 1, got %d", patch.Dropped)
	}
	if len(patch.Appended) != 1 || patch.Appended[0].SourceIndex != 10 {
		t.Fatalf("unexpected appended rows: %+v", patch.Appended)
	}
}

func TestBuildStateAppendPatchRejectsSelectionChange(t *testing.T) {
	logs := []appstate.LogViewItem{
		logViewItem(1, "t1", "I", "tag", "one"),
		logViewItem(2, "t2", "I", "tag", "two"),
		logViewItem(3, "t3", "I", "tag", "three"),
	}
	prev := AppState{
		Revision:      30,
		TotalLogs:     2,
		VisibleCount:  2,
		SelectedCount: 1,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one", IsFocused: true, IsSelected: true},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two"},
		},
	}
	current := appstate.UISnapshot{
		Revision:     31,
		VisibleCount: 3,
		VisibleLogs:  visibleLogSnapshots(logs),
		Model: appstate.Model{
			TotalLogs: 3,
			Selection: appstate.SelectionState{
				AnchorSourceIndex: 2,
				FocusSourceIndex:  2,
				SourceIndexes:     []int{2},
			},
		},
	}

	if _, ok := buildStateAppendPatch(prev, current); ok {
		t.Fatal("expected selection change to force full state")
	}
}

func TestApplyAppendPatch(t *testing.T) {
	prev := AppState{
		Revision:      10,
		TotalLogs:     3,
		VisibleCount:  3,
		SelectedCount: 1,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one"},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", IsFocused: true, IsSelected: true},
			{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three"},
		},
		SelectedLog: &SelectedLogView{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", Source: "src"},
	}
	patch := StateAppendPatch{
		Revision:      11,
		TotalLogs:     4,
		VisibleCount:  3,
		Dropped:       1,
		Appended:      []LogItemView{{SourceIndex: 4, TimeText: "t4", Level: "I", Tag: "tag", Message: "four"}},
		SelectedCount: 1,
		SelectedLog:   &SelectedLogView{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", Source: "src"},
	}

	next := applyAppendPatch(prev, patch)
	if next.Revision != 11 || next.TotalLogs != 4 || next.VisibleCount != 3 {
		t.Fatalf("unexpected summary after patch: %+v", next)
	}
	if len(next.Logs) != 3 || next.Logs[0].SourceIndex != 2 || next.Logs[2].SourceIndex != 4 {
		t.Fatalf("unexpected logs after patch: %+v", next.Logs)
	}
}

func TestApplyAppendPatchKeepsSelectedLogWhenPatchOmitsIt(t *testing.T) {
	prev := AppState{
		Revision:      10,
		TotalLogs:     3,
		VisibleCount:  3,
		SelectedCount: 1,
		Logs: []LogItemView{
			{SourceIndex: 1, TimeText: "t1", Level: "I", Tag: "tag", Message: "one"},
			{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", IsFocused: true, IsSelected: true},
			{SourceIndex: 3, TimeText: "t3", Level: "I", Tag: "tag", Message: "three"},
		},
		SelectedLog: &SelectedLogView{SourceIndex: 2, TimeText: "t2", Level: "I", Tag: "tag", Message: "two", Source: "src"},
	}
	patch := StateAppendPatch{
		Revision:      11,
		TotalLogs:     4,
		VisibleCount:  3,
		Dropped:       1,
		Appended:      []LogItemView{{SourceIndex: 4, TimeText: "t4", Level: "I", Tag: "tag", Message: "four"}},
		SelectedCount: 1,
	}

	next := applyAppendPatch(prev, patch)
	if next.SelectedLog == nil || next.SelectedLog.SourceIndex != 2 {
		t.Fatalf("expected selected log preserved, got %+v", next.SelectedLog)
	}
}

func logViewItem(sourceIndex int, timeText string, level string, tag string, message string) appstate.LogViewItem {
	return appstate.LogViewItem{
		SourceIndex: sourceIndex,
		Entry: logcat.LogEntry{
			TimeText: timeText,
			Level:    level,
			Tag:      tag,
			Message:  message,
		},
	}
}
