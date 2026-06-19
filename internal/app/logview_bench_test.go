package app

import (
	"fmt"
	"testing"

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
