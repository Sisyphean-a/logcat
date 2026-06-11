package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
)

func TestWriteVisibleLogsFileWritesTSV(t *testing.T) {
	dir := t.TempDir()
	items := []appstate.LogViewItem{
		{
			Entry: logcat.LogEntry{
				TimeText: "06-04 16:42:18.479",
				Level:    "I",
				Tag:      "chromium",
				Message:  "[H5] hello",
				Source:   "views/apply/index.vue:12",
			},
		},
	}

	path, err := writeVisibleLogsFile(items, time.Date(2026, 6, 9, 14, 30, 45, 0, time.UTC), dir)
	if err != nil {
		t.Fatalf("writeVisibleLogsFile returned error: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("expected export path in temp dir, got %q", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "时间\t级\t标签\t消息\t来源") {
		t.Fatalf("expected header row, got %q", got)
	}
	if !strings.Contains(got, "06-04 16:42:18.479\tI\tchromium\t[H5] hello\tviews/apply/index.vue:12") {
		t.Fatalf("expected log row content, got %q", got)
	}
}

func TestWriteVisibleLogsFileRejectsEmptyItems(t *testing.T) {
	_, err := writeVisibleLogsFile(nil, time.Now(), t.TempDir())
	if err == nil {
		t.Fatal("expected empty export to fail")
	}
}

func TestSaveAndLoadFilterState(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("AppData", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	filters := []appstate.SavedFilter{
		{ID: "h5", Name: "H5 日志", Query: "tag:chromium & message:[H5]"},
	}
	history := []string{"level:E", "tag:chromium"}
	if err := SaveFilterState(filters, history, "h5"); err != nil {
		t.Fatalf("SaveFilterState returned error: %v", err)
	}

	state, err := LoadFilterState()
	if err != nil {
		t.Fatalf("LoadFilterState returned error: %v", err)
	}
	if len(state.Filters) != 1 || state.Filters[0].ID != "h5" {
		t.Fatalf("unexpected filters: %#v", state.Filters)
	}
	if len(state.History) != 2 || state.History[0] != "level:E" {
		t.Fatalf("unexpected history: %#v", state.History)
	}
	if state.DefaultFilterID != "h5" {
		t.Fatalf("unexpected default filter id: %#v", state.DefaultFilterID)
	}

	path, err := filtersPath()
	if err != nil {
		t.Fatalf("filtersPath returned error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	var payload SavedFiltersFile
	if err := json.Unmarshal(content, &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(payload.History) != 2 {
		t.Fatalf("expected persisted history, got %#v", payload)
	}
	if payload.DefaultFilterID != "h5" {
		t.Fatalf("expected persisted default filter id, got %#v", payload)
	}
}
