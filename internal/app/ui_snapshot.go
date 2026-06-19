package app

import "github.com/xiakn/logcat/internal/adb"

type UISnapshot struct {
	Model        Model
	VisibleLogs  []VisibleLogSnapshot
	Revision     uint64
	VisibleCount int
	VisibleStart int
}

type VisibleLogSnapshot struct {
	SourceIndex int
	TimeText    string
	Level       string
	Tag         string
	Message     string
	Source      string
}

func (c *Controller) UISnapshot(limit int) UISnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	visibleStart := visibleWindowStart(len(c.model.VisibleLogs), limit)
	return UISnapshot{
		Model:        cloneUISnapshotModel(c.model, limit),
		VisibleLogs:  cloneVisibleLogSnapshots(c.model.VisibleLogs[visibleStart:]),
		Revision:     c.revision,
		VisibleCount: len(c.model.VisibleLogs),
		VisibleStart: visibleStart,
	}
}

func cloneUISnapshotModel(model Model, limit int) Model {
	cloned := model
	cloned.Devices = append([]DeviceItem(nil), model.Devices...)
	cloned.Packages = append([]adb.PackageInfo(nil), model.Packages...)
	cloned.Processes = nil
	cloned.SelectedProcess = ""
	cloned.BoundPIDs = nil
	cloned.Filter = cloneUISnapshotFilterState(model.Filter)
	cloned.Search = SearchState{Query: model.Search.Query}
	cloned.VisibleLogs = nil
	cloned.Selection.SourceIndexes = append([]int(nil), model.Selection.SourceIndexes...)
	return cloned
}

func cloneUISnapshotFilterState(state FilterState) FilterState {
	cloned := state
	cloned.Saved = append([]SavedFilter(nil), state.Saved...)
	cloned.History = nil
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

func cloneVisibleLogSnapshots(items []LogViewItem) []VisibleLogSnapshot {
	if len(items) == 0 {
		return nil
	}
	snapshots := make([]VisibleLogSnapshot, len(items))
	for index := range items {
		item := &items[index]
		entry := &item.Entry
		snapshots[index] = VisibleLogSnapshot{
			SourceIndex: item.SourceIndex,
			TimeText:    entry.TimeText,
			Level:       entry.Level,
			Tag:         entry.Tag,
			Message:     entry.Message,
			Source:      entry.Source,
		}
	}
	return snapshots
}
