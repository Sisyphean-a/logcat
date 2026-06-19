package app

import (
	"fmt"
	"testing"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/logcat"
)

// BenchmarkRebuildVisibleWithSearch 度量搜索缓存启用后的 rebuild 热路径开销。
func BenchmarkRebuildVisibleWithSearch(b *testing.B) {
	const total = 50000
	controller := NewController(stubDeviceService{}, stubSessionStarter{})

	for i := 0; i < total; i++ {
		entry := logcat.LogEntry{
			DeviceID: "dev",
			TimeText: "06-04 16:42:18.479",
			Level:    "I",
			Tag:      "chromium",
			Message:  fmt.Sprintf("[H5] message body number %d token", i),
			Raw:      "raw",
		}
		controller.allLogs.Append(LogViewItem{
			SourceIndex: i,
			Entry:       entry,
		}, "", 0)
	}
	controller.model.TotalLogs = controller.allLogs.Len()
	controller.model.Search.Query = "token"
	controller.compiledSearch = compileSearchQuery("token")
	controller.syncSearchCacheLocked(true)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		controller.rebuildVisibleFromAllLogsLocked()
	}
}

func BenchmarkAppendLogAtCapacity(b *testing.B) {
	const limit = 100000
	controller := NewController(stubDeviceService{}, stubSessionStarter{})
	controller.maxLogEntries = limit

	for i := 0; i < limit; i++ {
		controller.appendLogLocked(logcat.LogEntry{Message: fmt.Sprintf("seed-%d", i)})
	}

	entry := logcat.LogEntry{Message: "steady-state"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.appendLogLocked(entry)
	}
}

func BenchmarkAppendLogNoSearch(b *testing.B) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})
	entry := logcat.LogEntry{
		DeviceID: "dev",
		TimeText: "06-04 16:42:18.479",
		Level:    "I",
		Tag:      "chromium",
		Message:  "[H5] benchmark token payload",
		Raw:      "raw",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.appendLogLocked(entry)
	}
}

func BenchmarkSelectionSnapshot(b *testing.B) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})
	for i := 0; i < 2000; i++ {
		controller.model.VisibleLogs = append(controller.model.VisibleLogs, LogViewItem{
			SourceIndex: i,
			Entry: logcat.LogEntry{
				TimeText: "06-04 16:42:18.479",
				Level:    "I",
				Tag:      "chromium",
				Message:  fmt.Sprintf("[H5] message body number %d token", i),
				Raw:      "raw",
				Source:   "H5",
			},
		})
	}
	controller.model.Selection = SelectionState{
		AnchorSourceIndex: 1997,
		FocusSourceIndex:  1998,
		SourceIndexes:     []int{1997, 1998, 1999},
	}

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = controller.SelectionSnapshot(1000)
		}
	})

	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = legacySelectionSnapshot(controller.model, controller.revision, 1000)
		}
	})
}

func BenchmarkSelectedLogsLargeSelection(b *testing.B) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})
	for i := 0; i < 2000; i++ {
		controller.model.VisibleLogs = append(controller.model.VisibleLogs, LogViewItem{
			SourceIndex: i,
			Entry: logcat.LogEntry{
				TimeText: "06-04 16:42:18.479",
				Level:    "I",
				Tag:      "chromium",
				Message:  fmt.Sprintf("[H5] message body number %d token", i),
				Raw:      "raw",
				Source:   "H5",
			},
		})
	}
	selected := make([]int, 0, 1000)
	for i := 0; i < 2000; i += 2 {
		selected = append(selected, i)
	}
	controller.model.Selection = SelectionState{
		AnchorSourceIndex: selected[0],
		FocusSourceIndex:  selected[len(selected)-1],
		SourceIndexes:     selected,
	}

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = controller.SelectedLogs()
		}
	})

	b.Run("legacy", func(b *testing.B) {
		model := controller.model
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = legacySelectedLogs(model)
		}
	})
}

func legacySelectionSnapshot(model Model, revision uint64, limit int) SelectionSnapshot {
	return SelectionSnapshot{
		Selection: SelectionState{
			AnchorSourceIndex: model.Selection.AnchorSourceIndex,
			FocusSourceIndex:  model.Selection.FocusSourceIndex,
			SourceIndexes:     append([]int(nil), model.Selection.SourceIndexes...),
		},
		Focused:  cloneFocusedLogItem(append([]LogViewItem(nil), visibleWindow(model.VisibleLogs, limit)...), model.Selection.FocusSourceIndex),
		Revision: revision,
	}
}

func legacySelectedLogs(model Model) []LogViewItem {
	if len(model.Selection.SourceIndexes) == 0 {
		return nil
	}
	selected := make([]LogViewItem, 0, len(model.Selection.SourceIndexes))
	for _, sourceIndex := range model.Selection.SourceIndexes {
		index := legacyFindVisibleIndexBySource(model.VisibleLogs, sourceIndex)
		if index == -1 {
			continue
		}
		selected = append(selected, model.VisibleLogs[index])
	}
	return selected
}

func legacyFindVisibleIndexBySource(items []LogViewItem, sourceIndex int) int {
	for index, item := range items {
		if item.SourceIndex == sourceIndex {
			return index
		}
	}
	return -1
}

func BenchmarkUISnapshot(b *testing.B) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})
	controller.model.Devices = []DeviceItem{{ID: "dev-1", Model: "Pixel", Status: "device"}}
	controller.model.Packages = []adb.PackageInfo{{Name: "com.demo.host"}}
	controller.model.Processes = []adb.ProcessInfo{{PID: 111, Name: "com.demo.host"}}
	controller.model.SelectedProcess = "com.demo.host"
	controller.model.BoundPIDs = []int{111}
	controller.model.Filter = FilterState{
		Draft:           "tag=chromium",
		Applied:         "tag=chromium",
		ActiveFilterID:  "chromium",
		DefaultFilterID: "chromium",
		Saved: []SavedFilter{{
			ID:          "chromium",
			Name:        "chromium",
			PackageName: "com.demo.host",
			Query:       "tag=chromium",
		}},
		History: []string{"tag=chromium", "level=error"},
	}
	controller.model.Search = SearchState{Query: "token"}
	for i := 0; i < 2000; i++ {
		controller.model.VisibleLogs = append(controller.model.VisibleLogs, LogViewItem{
			SourceIndex: i,
			Entry: logcat.LogEntry{
				TimeText: "06-04 16:42:18.479",
				Level:    "I",
				Tag:      "chromium",
				Message:  fmt.Sprintf("[H5] message body number %d token", i),
				Raw:      "raw",
				Source:   "H5",
			},
		})
	}
	controller.model.Selection = SelectionState{
		AnchorSourceIndex: 1997,
		FocusSourceIndex:  1998,
		SourceIndexes:     []int{1997, 1998, 1999},
	}

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = controller.UISnapshot(1000)
		}
	})

	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = legacyUISnapshot(controller.model, controller.revision, 1000)
		}
	})
}

func legacyUISnapshot(model Model, revision uint64, limit int) UISnapshot {
	return UISnapshot{
		Model:        legacyCloneUISnapshotModel(model, limit),
		Revision:     revision,
		VisibleCount: len(model.VisibleLogs),
		VisibleStart: visibleWindowStart(len(model.VisibleLogs), limit),
	}
}

func legacyCloneUISnapshotModel(model Model, limit int) Model {
	cloned := model
	cloned.Devices = append([]DeviceItem(nil), model.Devices...)
	cloned.Packages = append([]adb.PackageInfo(nil), model.Packages...)
	cloned.Processes = append([]adb.ProcessInfo(nil), model.Processes...)
	cloned.BoundPIDs = append([]int(nil), model.BoundPIDs...)
	cloned.Filter = cloneFilterState(model.Filter)
	cloned.Search = SearchState{Query: model.Search.Query}
	cloned.VisibleLogs = append([]LogViewItem(nil), visibleWindow(model.VisibleLogs, limit)...)
	return cloned
}
