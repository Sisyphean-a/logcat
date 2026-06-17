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
