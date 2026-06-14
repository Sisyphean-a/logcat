package main

import (
	"strings"
	"testing"

	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
)

func TestNewAppStatePreservesSelectedRawLog(t *testing.T) {
	message := "[H5] " + strings.Repeat("x", 5000)
	raw := "06-10 20:41:45.478 1234 1234 I chromium: " + message
	snapshot := appstate.UISnapshot{
		Model: appstate.Model{
			SelectedIndex: 0,
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
	if len(state.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(state.Logs))
	}
	if state.Logs[0].Raw != raw {
		t.Fatal("expected row raw to stay unchanged")
	}
	if state.SelectedLog == nil {
		t.Fatal("expected selected log")
	}
	if state.SelectedLog.Raw != raw {
		t.Fatal("expected selected raw to stay unchanged")
	}
}
