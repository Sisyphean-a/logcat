package app

type SelectionSnapshot struct {
	VisibleLogs []LogViewItem
	Selection   SelectionState
	Revision    uint64
}

func (c *Controller) SelectionSnapshot(limit int) SelectionSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return SelectionSnapshot{
		VisibleLogs: append([]LogViewItem(nil), visibleWindow(c.model.VisibleLogs, limit)...),
		Selection: SelectionState{
			AnchorSourceIndex: c.model.Selection.AnchorSourceIndex,
			FocusSourceIndex:  c.model.Selection.FocusSourceIndex,
			SourceIndexes:     append([]int(nil), c.model.Selection.SourceIndexes...),
		},
		Revision: c.revision,
	}
}
