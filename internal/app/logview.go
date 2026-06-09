package app

import (
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
	c.updatePausedStatusLocked()
}

func (c *Controller) ResumeKeep() {
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
	c.rebuildVisibleFromAllLogsLocked()
	c.model.Status = "running"
}

func (c *Controller) ResumeDiscard() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.model.Pause.Active {
		return
	}

	c.pauseBuffer = nil
	c.model.Pause.Active = false
	c.model.Pause.BufferedCount = 0
	c.model.Pause.DroppedCount = 0
	c.model.Status = "running"
}

func (c *Controller) ClearVisible() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.allLogs = c.allLogs[:0]
	c.model.TotalLogs = 0
	c.model.Logs = c.model.Logs[:0]
	c.model.VisibleLogs = c.model.VisibleLogs[:0]
	c.model.SelectedIndex = -1
	c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]
	c.model.Search.Current = -1
}

func (c *Controller) SetSearchQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Search.Query = query
	c.recomputeSearchLocked()
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
}

func (c *Controller) SelectLog(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.model.VisibleLogs) {
		return
	}

	c.model.SelectedIndex = index
	c.syncCurrentMatchToSelectionLocked()
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
		return
	}

	c.appendLogLocked(entry)
	c.rebuildVisibleFromAllLogsLocked()
}

func (c *Controller) appendLogLocked(entry logcat.LogEntry) {
	item := LogViewItem{
		Entry:   entry,
		Display: formatLogDisplay(entry),
	}
	c.allLogs = append(c.allLogs, item)
	c.model.TotalLogs = len(c.allLogs)
}

func (c *Controller) recomputeSearchLocked() {
	query := strings.TrimSpace(c.model.Search.Query)
	c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]

	if query == "" {
		c.model.Search.Current = -1
		return
	}

	normalizedQuery := strings.ToLower(query)
	for index, item := range c.model.VisibleLogs {
		if strings.Contains(strings.ToLower(item.Display), normalizedQuery) {
			c.model.Search.MatchIndexes = append(c.model.Search.MatchIndexes, index)
		}
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
	c.model.VisibleLogs = c.model.VisibleLogs[:0]
	c.model.Logs = c.model.Logs[:0]
	for _, item := range c.allLogs {
		if !matchesFilter(item.Entry, c.model.SelectedPackage, c.model.Filter.Applied) {
			continue
		}
		c.model.VisibleLogs = append(c.model.VisibleLogs, item)
		c.model.Logs = append(c.model.Logs, item.Display)
	}
	if c.model.SelectedIndex >= len(c.model.VisibleLogs) {
		c.model.SelectedIndex = -1
	}
	c.recomputeSearchLocked()
}
