package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

const defaultPauseBufferCap = 10000

func (c *Controller) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.model.Pause.Active {
		return
	}

	c.model.Pause.Active = true
	c.resumeStreaming = false
	c.updatePausedStatusLocked()
	c.markDirtyLocked()
}

func (c *Controller) ResumeKeep() {
	if !c.hasActiveSession() {
		_ = c.startCurrentSelection(context.Background())
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.model.Pause.Active {
		return
	}

	for _, entry := range c.pauseBuffer {
		c.appendLogLocked(entry)
	}

	c.pauseBuffer = nil
	c.model.Pause.Active = false
	c.model.Pause.BufferedCount = 0
	c.model.Pause.DroppedCount = 0
	c.resumeStreaming = true
	c.rebuildVisibleFromAllLogsLocked()
	c.model.Status = "running"
	c.markDirtyLocked()
}

func (c *Controller) ResumeDiscard() {
	if !c.hasActiveSession() {
		_ = c.startCurrentSelection(context.Background())
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.model.Pause.Active {
		return
	}

	c.pauseBuffer = nil
	c.model.Pause.Active = false
	c.model.Pause.BufferedCount = 0
	c.model.Pause.DroppedCount = 0
	c.resumeStreaming = true
	c.model.Status = "running"
	c.markDirtyLocked()
}

func (c *Controller) ClearVisible() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.allLogs = c.allLogs[:0]
	c.model.TotalLogs = 0
	c.model.VisibleLogs = c.model.VisibleLogs[:0]
	c.model.SelectedIndex = -1
	c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]
	c.model.Search.Current = -1
	c.markDirtyLocked()
}

func (c *Controller) SetSearchQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Search.Query = query
	c.rebuildVisibleFromAllLogsLocked()
}

func (c *Controller) NextMatch() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.model.Search.MatchIndexes) == 0 {
		return
	}

	if c.model.Search.Current == -1 {
		c.model.Search.Current = 0
	} else {
		c.model.Search.Current = (c.model.Search.Current + 1) % len(c.model.Search.MatchIndexes)
	}
	c.model.SelectedIndex = c.model.Search.MatchIndexes[c.model.Search.Current]
	c.markDirtyLocked()
}

func (c *Controller) PrevMatch() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.model.Search.MatchIndexes) == 0 {
		return
	}

	if c.model.Search.Current == -1 {
		c.model.Search.Current = len(c.model.Search.MatchIndexes) - 1
	} else {
		c.model.Search.Current--
		if c.model.Search.Current < 0 {
			c.model.Search.Current = len(c.model.Search.MatchIndexes) - 1
		}
	}
	c.model.SelectedIndex = c.model.Search.MatchIndexes[c.model.Search.Current]
	c.markDirtyLocked()
}

func (c *Controller) SelectLog(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.model.VisibleLogs) {
		return
	}

	c.model.SelectedIndex = index
	c.syncCurrentMatchToSelectionLocked()
	c.markDirtyLocked()
}

func (c *Controller) SelectedLog() (LogViewItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.model.SelectedIndex < 0 || c.model.SelectedIndex >= len(c.model.VisibleLogs) {
		return LogViewItem{}, false
	}

	return c.model.VisibleLogs[c.model.SelectedIndex], true
}

func (c *Controller) pushEntry(entry logcat.LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.model.Pause.Active {
		c.pauseBuffer = append(c.pauseBuffer, entry)
		if len(c.pauseBuffer) > c.pauseBufferCap {
			c.pauseBuffer = c.pauseBuffer[1:]
			c.model.Pause.DroppedCount++
		}
		c.model.Pause.BufferedCount = len(c.pauseBuffer)
		c.updatePausedStatusLocked()
		c.markDirtyLocked()
		return
	}

	item := c.appendLogLocked(entry)
	if c.matchesVisibleLogLocked(item, c.searchQueryLocked()) {
		c.appendVisibleLogLocked(item)
	}
	c.markDirtyLocked()
}

func (c *Controller) appendLogLocked(entry logcat.LogEntry) LogViewItem {
	display := formatLogDisplay(entry)
	item := LogViewItem{
		SourceIndex:  len(c.allLogs),
		Entry:        entry,
		Display:      display,
		DisplayLower: strings.ToLower(display),
		SearchLower:  searchLowerText(entry),
	}
	c.allLogs = append(c.allLogs, item)
	c.model.TotalLogs = len(c.allLogs)
	return item
}

func (c *Controller) appendVisibleLogLocked(item LogViewItem) {
	index := len(c.model.VisibleLogs)
	c.model.VisibleLogs = append(c.model.VisibleLogs, item)
	c.appendSearchMatchLocked(index)
}

func (c *Controller) appendSearchMatchLocked(index int) {
	if c.searchQueryLocked() == "" {
		return
	}
	c.model.Search.MatchIndexes = append(c.model.Search.MatchIndexes, index)
	if c.model.Search.Current != -1 {
		return
	}

	c.model.Search.Current = 0
	c.model.SelectedIndex = 0
}

func (c *Controller) recomputeSearchLocked() {
	if c.searchQueryLocked() == "" {
		c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]
		c.model.Search.Current = -1
		return
	}

	if cap(c.model.Search.MatchIndexes) < len(c.model.VisibleLogs) {
		c.model.Search.MatchIndexes = make([]int, 0, len(c.model.VisibleLogs))
	} else {
		c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]
	}
	for index := range c.model.VisibleLogs {
		c.model.Search.MatchIndexes = append(c.model.Search.MatchIndexes, index)
	}

	if len(c.model.Search.MatchIndexes) == 0 {
		c.model.Search.Current = -1
		return
	}

	c.syncCurrentMatchToSelectionLocked()
	if c.model.Search.Current == -1 {
		c.model.Search.Current = 0
		c.model.SelectedIndex = c.model.Search.MatchIndexes[0]
	}
}

func (c *Controller) syncCurrentMatchToSelectionLocked() {
	c.model.Search.Current = -1
	for matchIndex, index := range c.model.Search.MatchIndexes {
		if index == c.model.SelectedIndex {
			c.model.Search.Current = matchIndex
			return
		}
	}
}

func (c *Controller) updatePausedStatusLocked() {
	c.model.Status = fmt.Sprintf("Paused，缓存 %d 条新日志", c.model.Pause.BufferedCount)
}

func formatLogDisplay(entry logcat.LogEntry) string {
	return fmt.Sprintf("%s %s %s %s", entry.TimeText, entry.Level, entry.Tag, entry.Message)
}

func (c *Controller) rebuildVisibleFromAllLogsLocked() {
	selectedSourceIndex := c.selectedSourceIndexLocked()
	searchQuery := c.searchQueryLocked()
	if c.compiledFilter.matchAll() && searchQuery == "" {
		c.model.VisibleLogs = append(c.model.VisibleLogs[:0], c.allLogs...)
		c.restoreSelectionLocked(selectedSourceIndex)
		c.recomputeSearchLocked()
		c.markDirtyLocked()
		return
	}

	if cap(c.model.VisibleLogs) < len(c.allLogs) {
		c.model.VisibleLogs = make([]LogViewItem, 0, len(c.allLogs))
	} else {
		c.model.VisibleLogs = c.model.VisibleLogs[:0]
	}
	for _, item := range c.allLogs {
		if !c.matchesVisibleLogLocked(item, searchQuery) {
			continue
		}
		c.model.VisibleLogs = append(c.model.VisibleLogs, item)
	}
	c.restoreSelectionLocked(selectedSourceIndex)
	c.recomputeSearchLocked()
	c.markDirtyLocked()
}

func normalizedSearchQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

func (c *Controller) matchesVisibleLogLocked(item LogViewItem, searchQuery string) bool {
	if !c.compiledFilter.matchAll() && !c.matchesAppliedFilterLocked(item.Entry) {
		return false
	}
	return searchQuery == "" || strings.Contains(item.SearchLower, searchQuery)
}

func (c *Controller) restoreSelectionLocked(sourceIndex int) {
	c.model.SelectedIndex = -1
	if sourceIndex < 0 {
		return
	}
	for index, item := range c.model.VisibleLogs {
		if item.SourceIndex == sourceIndex {
			c.model.SelectedIndex = index
			return
		}
	}
}

func (c *Controller) searchQueryLocked() string {
	return normalizedSearchQuery(c.model.Search.Query)
}

func (c *Controller) selectedSourceIndexLocked() int {
	if c.model.SelectedIndex < 0 || c.model.SelectedIndex >= len(c.model.VisibleLogs) {
		return -1
	}
	return c.model.VisibleLogs[c.model.SelectedIndex].SourceIndex
}

func searchLowerText(entry logcat.LogEntry) string {
	return strings.ToLower(entry.Tag + "\n" + entry.Message)
}
