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
	cloned.VisibleLogs = append([]LogViewItem(nil), visibleWindow(model.VisibleLogs, limit)...)
	cloned.Search.MatchIndexes, cloned.Search.Current = sliceSearchWindow(
		model.Search.MatchIndexes,
		model.Search.Current,
		visibleWindowStart(len(model.VisibleLogs), limit),
	)
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

func sliceSearchWindow(matchIndexes []int, current int, start int) ([]int, int) {
	windowMatches := make([]int, 0, len(matchIndexes))
	windowCurrent := -1
	for _, index := range matchIndexes {
		if index < start {
			continue
		}
		windowMatches = append(windowMatches, index)
	}
	if current < 0 || current >= len(matchIndexes) {
		return windowMatches, windowCurrent
	}

	currentIndex := matchIndexes[current]
	for matchPos, index := range windowMatches {
		if index == currentIndex {
			windowCurrent = matchPos
			break
		}
	}
	return windowMatches, windowCurrent
}
