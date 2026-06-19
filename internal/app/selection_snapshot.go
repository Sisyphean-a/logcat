package app

type SelectionSnapshot struct {
	Selection SelectionState
	Focused   *LogViewItem
	Revision  uint64
}

func (c *Controller) SelectionSnapshot(limit int) SelectionSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return SelectionSnapshot{
		Selection: SelectionState{
			AnchorSourceIndex: c.model.Selection.AnchorSourceIndex,
			FocusSourceIndex:  c.model.Selection.FocusSourceIndex,
			SourceIndexes:     append([]int(nil), c.model.Selection.SourceIndexes...),
		},
		Focused:  cloneFocusedLogItem(visibleWindow(c.model.VisibleLogs, limit), c.model.Selection.FocusSourceIndex),
		Revision: c.revision,
	}
}

func cloneFocusedLogItem(items []LogViewItem, focusedSourceIndex int) *LogViewItem {
	if focusedSourceIndex < 0 {
		return nil
	}
	for index := range items {
		if items[index].SourceIndex != focusedSourceIndex {
			continue
		}
		item := items[index]
		return &item
	}
	return nil
}
