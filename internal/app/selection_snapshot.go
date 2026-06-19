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
	index := findFocusedLogItemIndex(items, focusedSourceIndex)
	if index == -1 {
		return nil
	}
	item := items[index]
	return &item
}

func findFocusedLogItemIndex(items []LogViewItem, focusedSourceIndex int) int {
	low := 0
	high := len(items) - 1
	for low <= high {
		middle := low + (high-low)/2
		current := items[middle].SourceIndex
		switch {
		case current == focusedSourceIndex:
			return middle
		case current < focusedSourceIndex:
			low = middle + 1
		default:
			high = middle - 1
		}
	}
	return -1
}
