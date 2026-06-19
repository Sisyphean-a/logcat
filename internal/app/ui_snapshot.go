package app

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
	cloned.Packages = append(cloned.Packages[:0], model.Packages...)
	cloned.Processes = nil
	cloned.SelectedProcess = ""
	cloned.BoundPIDs = nil
	cloned.Filter = cloneUISnapshotFilterState(model.Filter)
	cloned.Search = SearchState{Query: model.Search.Query}
	cloned.VisibleLogs = append([]LogViewItem(nil), visibleWindow(model.VisibleLogs, limit)...)
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
