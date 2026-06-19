package main

import (
	"strings"
	"testing"

	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
)

func TestNewAppStateSelectedLogOmitsRawPayload(t *testing.T) {
	message := "[H5] " + strings.Repeat("x", 5000)
	raw := "06-10 20:41:45.478 1234 1234 I chromium: " + message
	snapshot := appstate.UISnapshot{
		Model: appstate.Model{
			SelectedIndex: 0,
			Selection: appstate.SelectionState{
				AnchorSourceIndex: 0,
				FocusSourceIndex:  0,
				SourceIndexes:     []int{0},
			},
			VisibleLogs: []appstate.LogViewItem{{
				Entry: logcat.LogEntry{
					TimeText: "06-10 20:41:45.478",
					Level:    "I",
					Tag:      "chromium",
					Message:  message,
					Raw:      raw,
				},
			}},
		},
		VisibleCount: 1,
	}

	state := newAppState(snapshot)
	if state.Revision != 0 {
		t.Fatalf("expected revision 0, got %d", state.Revision)
	}
	if len(state.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(state.Logs))
	}
	if state.Logs[0].SourceIndex != 0 {
		t.Fatalf("expected row sourceIndex 0, got %d", state.Logs[0].SourceIndex)
	}
	if state.SelectedLog == nil {
		t.Fatal("expected selected log")
	}
	if state.SelectedLog.SourceIndex != 0 {
		t.Fatalf("expected selected sourceIndex 0, got %d", state.SelectedLog.SourceIndex)
	}
	if state.SelectedLog.Message != message {
		t.Fatal("expected selected message to stay unchanged")
	}
}
