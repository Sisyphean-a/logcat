package app

import "github.com/xiakn/logcat/internal/adb"

type UISnapshot struct {
	Model        Model
	Revision     uint64
	VisibleCount int
	VisibleStart int
}

func (c *Controller) UISnapshot(limit int) UISnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return UISnapshot{
		Model:        cloneUISnapshotModel(c.model, limit),
		Revision:     c.revision,
		VisibleCount: len(c.model.VisibleLogs),
		VisibleStart: visibleWindowStart(len(c.model.VisibleLogs), limit),
	}
}

func cloneUISnapshotModel(model Model, limit int) Model {
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

func visibleWindow(items []LogViewItem, limit int) []LogViewItem {
	start := visibleWindowStart(len(items), limit)
	return items[start:]
}

func visibleWindowStart(size int, limit int) int {
	if limit <= 0 || size <= limit {
		return 0
	}
	return size - limit
}
