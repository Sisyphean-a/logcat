package app

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

const defaultPauseBufferCap = 10000

// defaultMaxLogEntries 是 allLogs 的容量上限。超过后从最旧端淘汰，使运行时
// 内存有界。按平均一行 ~200B 估算，10 万条约 20~40MB。
const defaultMaxLogEntries = 100000

type SelectionMode string

const (
	SelectionModeReplace SelectionMode = "replace"
	SelectionModeAdd     SelectionMode = "add"
	SelectionModeRange   SelectionMode = "range"
)

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

	c.allLogs.Reset()
	c.model.TotalLogs = 0
	c.model.VisibleLogs = c.model.VisibleLogs[:0]
	c.clearSelectionLocked()
	c.markDirtyLocked()
}

func (c *Controller) SetSearchQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Search.Query = query
	c.compiledSearch = compileSearchQuery(query)
	c.rebuildVisibleFromAllLogsLocked()
}

func (c *Controller) NextMatch() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.compiledSearch.matchAll() || len(c.model.VisibleLogs) == 0 {
		return
	}

	if c.model.SelectedIndex < 0 || c.model.SelectedIndex >= len(c.model.VisibleLogs)-1 {
		c.setSingleSelectionLocked(0)
		c.markDirtyLocked()
		return
	}
	c.setSingleSelectionLocked(c.model.SelectedIndex + 1)
	c.markDirtyLocked()
}

func (c *Controller) PrevMatch() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.compiledSearch.matchAll() || len(c.model.VisibleLogs) == 0 {
		return
	}

	if c.model.SelectedIndex <= 0 {
		c.setSingleSelectionLocked(len(c.model.VisibleLogs) - 1)
		c.markDirtyLocked()
		return
	}
	c.setSingleSelectionLocked(c.model.SelectedIndex - 1)
	c.markDirtyLocked()
}

func (c *Controller) SelectLog(index int) {
	c.SelectLogWithMode(index, SelectionModeReplace)
}

func (c *Controller) SelectLogWithMode(index int, mode SelectionMode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.model.VisibleLogs) {
		return
	}
	if !c.selectLogLocked(index, mode) {
		return
	}
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

func (c *Controller) SelectedLogs() []LogViewItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.model.Selection.SourceIndexes) == 0 {
		return nil
	}

	selected := make([]LogViewItem, 0, len(c.model.Selection.SourceIndexes))
	for _, sourceIndex := range c.model.Selection.SourceIndexes {
		index := c.findVisibleIndexBySourceLocked(sourceIndex)
		if index == -1 {
			continue
		}
		selected = append(selected, c.model.VisibleLogs[index])
	}
	return selected
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

	item, searchLower := c.appendLogLocked(entry)
	if c.matchesVisibleLogLocked(item, searchLower) {
		c.appendVisibleLogLocked(item)
	}
	c.markDirtyLocked()
}

func (c *Controller) appendLogLocked(entry logcat.LogEntry) (LogViewItem, string) {
	item := LogViewItem{
		SourceIndex: c.nextSourceIndex,
		Entry:       entry,
	}
	c.nextSourceIndex++
	searchLower := ""
	if c.searchCacheActiveLocked() {
		searchLower = searchLowerText(entry)
	}
	dropped, minSource := c.allLogs.Append(item, searchLower, c.maxLogEntries)
	if dropped {
		c.dropVisibleBeforeLocked(minSource)
	}
	c.model.TotalLogs = c.allLogs.Len()
	return item, searchLower
}

// dropVisibleBeforeLocked 丢弃 VisibleLogs 中 SourceIndex < minSource 的前缀
// （VisibleLogs 按 SourceIndex 单调递增），并同步 SelectedIndex 与搜索匹配。
func (c *Controller) dropVisibleBeforeLocked(minSource int) {
	drop := 0
	for drop < len(c.model.VisibleLogs) && c.model.VisibleLogs[drop].SourceIndex < minSource {
		drop++
	}
	if drop == 0 {
		return
	}

	c.model.VisibleLogs = append(c.model.VisibleLogs[:0], c.model.VisibleLogs[drop:]...)
	c.rebuildSelectionFromSourceIndexesLocked()
	c.recomputeSearchLocked()
}

func (c *Controller) appendVisibleLogLocked(item LogViewItem) {
	c.model.VisibleLogs = append(c.model.VisibleLogs, item)
	if c.compiledSearch.matchAll() {
		return
	}
	if c.model.SelectedIndex != -1 {
		return
	}
	c.setSingleSelectionLocked(0)
}

func (c *Controller) recomputeSearchLocked() {
	if c.compiledSearch.matchAll() {
		return
	}

	if len(c.model.VisibleLogs) == 0 {
		c.clearSelectionLocked()
		return
	}

	if c.model.SelectedIndex >= 0 && c.model.SelectedIndex < len(c.model.VisibleLogs) {
		return
	}
	c.setSingleSelectionLocked(0)
}

func (c *Controller) updatePausedStatusLocked() {
	c.model.Status = fmt.Sprintf("Paused，缓存 %d 条新日志", c.model.Pause.BufferedCount)
}

func FormatLogDisplay(entry logcat.LogEntry) string {
	return fmt.Sprintf("%s %s %s %s", entry.TimeText, entry.Level, entry.Tag, entry.Message)
}

func (c *Controller) rebuildVisibleFromAllLogsLocked() {
	selectedSourceIndexes := c.cloneSelectionSourceIndexesLocked()
	focusSourceIndex := c.model.Selection.FocusSourceIndex
	anchorSourceIndex := c.model.Selection.AnchorSourceIndex
	searchActive := !c.compiledSearch.matchAll()
	filterActive := !c.compiledFilter.matchAll()
	c.syncSearchCacheLocked(searchActive)
	if !filterActive && !searchActive {
		c.model.VisibleLogs = c.allLogs.AppendOrdered(c.model.VisibleLogs)
		c.restoreSelectionLocked(selectedSourceIndexes, focusSourceIndex, anchorSourceIndex)
		c.recomputeSearchLocked()
		c.markDirtyLocked()
		return
	}

	c.prepareVisibleLogsBufferLocked()
	switch {
	case !filterActive:
		c.appendSearchedVisibleLogsLocked()
	case !searchActive:
		c.appendFilteredVisibleLogsLocked()
	default:
		c.appendFilteredSearchedVisibleLogsLocked()
	}
	c.restoreSelectionLocked(selectedSourceIndexes, focusSourceIndex, anchorSourceIndex)
	c.recomputeSearchLocked()
	c.markDirtyLocked()
}

func (c *Controller) cloneSelectionSourceIndexesLocked() []int {
	count := len(c.model.Selection.SourceIndexes)
	if count == 0 {
		c.selectionScratch = c.selectionScratch[:0]
		return nil
	}
	if cap(c.selectionScratch) < count {
		c.selectionScratch = make([]int, count)
	} else {
		c.selectionScratch = c.selectionScratch[:count]
	}
	copy(c.selectionScratch, c.model.Selection.SourceIndexes)
	return c.selectionScratch
}

func (c *Controller) prepareVisibleLogsBufferLocked() {
	if cap(c.model.VisibleLogs) < c.allLogs.Len() {
		c.model.VisibleLogs = make([]LogViewItem, 0, c.allLogs.Len())
		return
	}
	c.model.VisibleLogs = c.model.VisibleLogs[:0]
}

func (c *Controller) appendSearchedVisibleLogsLocked() {
	items := c.allLogs.items
	searchLower := c.allLogs.searchLower
	start := c.allLogs.start
	query := c.compiledSearch
	c.model.VisibleLogs = appendSearchedVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		searchLower,
		start,
		len(items),
		query,
	)
	if start == 0 {
		return
	}
	c.model.VisibleLogs = appendSearchedVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		searchLower,
		0,
		start,
		query,
	)
}

func appendSearchedVisibleLogsSegment(
	dst []LogViewItem,
	items []LogViewItem,
	searchLower []string,
	start int,
	end int,
	query compiledSearchQuery,
) []LogViewItem {
	for index := start; index < end; index++ {
		if !query.matches(searchLower[index]) {
			continue
		}
		dst = append(dst, items[index])
	}
	return dst
}

func (c *Controller) appendFilteredVisibleLogsLocked() {
	items := c.allLogs.items
	start := c.allLogs.start
	filter := c.compiledFilter
	packageName := c.model.SelectedPackage
	c.model.VisibleLogs = appendFilteredVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		start,
		len(items),
		filter,
		packageName,
	)
	if start == 0 {
		return
	}
	c.model.VisibleLogs = appendFilteredVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		0,
		start,
		filter,
		packageName,
	)
}

func appendFilteredVisibleLogsSegment(
	dst []LogViewItem,
	items []LogViewItem,
	start int,
	end int,
	filter compiledFilterQuery,
	packageName string,
) []LogViewItem {
	for index := start; index < end; index++ {
		if !filter.matches(items[index].Entry, packageName) {
			continue
		}
		dst = append(dst, items[index])
	}
	return dst
}

func (c *Controller) appendFilteredSearchedVisibleLogsLocked() {
	items := c.allLogs.items
	searchLower := c.allLogs.searchLower
	start := c.allLogs.start
	filter := c.compiledFilter
	query := c.compiledSearch
	packageName := c.model.SelectedPackage
	c.model.VisibleLogs = appendFilteredSearchedVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		searchLower,
		start,
		len(items),
		filter,
		query,
		packageName,
	)
	if start == 0 {
		return
	}
	c.model.VisibleLogs = appendFilteredSearchedVisibleLogsSegment(
		c.model.VisibleLogs,
		items,
		searchLower,
		0,
		start,
		filter,
		query,
		packageName,
	)
}

func appendFilteredSearchedVisibleLogsSegment(
	dst []LogViewItem,
	items []LogViewItem,
	searchLower []string,
	start int,
	end int,
	filter compiledFilterQuery,
	query compiledSearchQuery,
	packageName string,
) []LogViewItem {
	for index := start; index < end; index++ {
		entry := items[index].Entry
		if !filter.matches(entry, packageName) || !query.matches(searchLower[index]) {
			continue
		}
		dst = append(dst, items[index])
	}
	return dst
}

func normalizedSearchQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

func (c *Controller) matchesVisibleLogLocked(item LogViewItem, searchLower string) bool {
	if !c.compiledFilter.matchAll() && !c.matchesAppliedFilterLocked(item.Entry) {
		return false
	}
	return c.compiledSearch.matchAll() || c.compiledSearch.matches(searchLower)
}

func (c *Controller) selectedSourceIndexLocked() int {
	return c.model.Selection.FocusSourceIndex
}

func (c *Controller) syncSearchCacheLocked(searchActive bool) {
	if !searchActive {
		c.allLogs.ReleaseSearchCache()
		return
	}
	c.allLogs.EnsureSearchCache(func(item LogViewItem) string {
		return searchLowerText(item.Entry)
	})
}

func (c *Controller) searchCacheActiveLocked() bool {
	return !c.compiledSearch.matchAll()
}

func (c *Controller) selectLogLocked(index int, mode SelectionMode) bool {
	switch mode {
	case SelectionModeAdd:
		return c.toggleSelectionLocked(index)
	case SelectionModeRange:
		return c.extendSelectionLocked(index)
	default:
		return c.setSingleSelectionLocked(index)
	}
}

func (c *Controller) toggleSelectionLocked(index int) bool {
	sourceIndex := c.model.VisibleLogs[index].SourceIndex
	position := slicesIndex(c.model.Selection.SourceIndexes, sourceIndex)
	if position >= 0 {
		if len(c.model.Selection.SourceIndexes) == 1 && c.model.Selection.FocusSourceIndex == sourceIndex {
			return false
		}
		c.model.Selection.SourceIndexes = append(
			c.model.Selection.SourceIndexes[:position],
			c.model.Selection.SourceIndexes[position+1:]...,
		)
		if sourceIndex == c.model.Selection.AnchorSourceIndex {
			c.model.Selection.AnchorSourceIndex = firstSelectionSource(c.model.Selection.SourceIndexes)
		}
		c.model.Selection.FocusSourceIndex = sourceIndex
		c.rebuildSelectionFromSourceIndexesLocked()
		return true
	}

	c.model.Selection.SourceIndexes = append(c.model.Selection.SourceIndexes, sourceIndex)
	sort.Ints(c.model.Selection.SourceIndexes)
	if c.model.Selection.AnchorSourceIndex < 0 {
		c.model.Selection.AnchorSourceIndex = sourceIndex
	}
	c.model.Selection.FocusSourceIndex = sourceIndex
	c.model.SelectedIndex = index
	return true
}

func (c *Controller) extendSelectionLocked(index int) bool {
	sourceIndex := c.model.VisibleLogs[index].SourceIndex
	anchorSourceIndex := c.model.Selection.FocusSourceIndex
	if anchorSourceIndex < 0 {
		return c.setSingleSelectionLocked(index)
	}

	anchorIndex := c.findVisibleIndexBySourceLocked(anchorSourceIndex)
	if anchorIndex == -1 {
		return c.setSingleSelectionLocked(index)
	}

	start := anchorIndex
	end := index
	if start > end {
		start, end = end, start
	}
	selected := make([]int, 0, end-start+1)
	for current := start; current <= end; current++ {
		selected = append(selected, c.model.VisibleLogs[current].SourceIndex)
	}
	if c.model.SelectedIndex == index &&
		c.model.Selection.FocusSourceIndex == sourceIndex &&
		slices.Equal(c.model.Selection.SourceIndexes, selected) {
		return false
	}
	c.model.Selection.SourceIndexes = selected
	c.model.Selection.FocusSourceIndex = sourceIndex
	c.model.SelectedIndex = index
	return true
}

func (c *Controller) setSingleSelectionLocked(index int) bool {
	sourceIndex := c.model.VisibleLogs[index].SourceIndex
	if c.model.SelectedIndex == index &&
		c.model.Selection.AnchorSourceIndex == sourceIndex &&
		c.model.Selection.FocusSourceIndex == sourceIndex &&
		len(c.model.Selection.SourceIndexes) == 1 &&
		c.model.Selection.SourceIndexes[0] == sourceIndex {
		return false
	}
	c.model.SelectedIndex = index
	c.model.Selection.AnchorSourceIndex = sourceIndex
	c.model.Selection.FocusSourceIndex = sourceIndex
	c.model.Selection.SourceIndexes = append(c.model.Selection.SourceIndexes[:0], sourceIndex)
	return true
}

func (c *Controller) clearSelectionLocked() {
	c.model.SelectedIndex = -1
	c.model.Selection.AnchorSourceIndex = -1
	c.model.Selection.FocusSourceIndex = -1
	c.model.Selection.SourceIndexes = c.model.Selection.SourceIndexes[:0]
}

func (c *Controller) restoreSelectionLocked(selected []int, focus int, anchor int) {
	c.model.Selection.AnchorSourceIndex = anchor
	c.model.Selection.FocusSourceIndex = focus
	c.model.Selection.SourceIndexes = append(c.model.Selection.SourceIndexes[:0], selected...)
	c.rebuildSelectionFromSourceIndexesLocked()
}

func (c *Controller) rebuildSelectionFromSourceIndexesLocked() {
	filtered := c.model.Selection.SourceIndexes[:0]
	for _, sourceIndex := range c.model.Selection.SourceIndexes {
		if c.findVisibleIndexBySourceLocked(sourceIndex) == -1 {
			continue
		}
		filtered = append(filtered, sourceIndex)
	}
	c.model.Selection.SourceIndexes = filtered
	if len(filtered) == 0 {
		c.clearSelectionLocked()
		return
	}
	c.model.SelectedIndex = c.findVisibleIndexBySourceLocked(c.model.Selection.FocusSourceIndex)
	if c.model.SelectedIndex == -1 {
		c.model.Selection.FocusSourceIndex = filtered[len(filtered)-1]
		c.model.SelectedIndex = c.findVisibleIndexBySourceLocked(c.model.Selection.FocusSourceIndex)
	}
	if c.findVisibleIndexBySourceLocked(c.model.Selection.AnchorSourceIndex) == -1 {
		c.model.Selection.AnchorSourceIndex = filtered[0]
	}
}

func (c *Controller) findVisibleIndexBySourceLocked(sourceIndex int) int {
	index := sort.Search(len(c.model.VisibleLogs), func(index int) bool {
		return c.model.VisibleLogs[index].SourceIndex >= sourceIndex
	})
	if index < len(c.model.VisibleLogs) && c.model.VisibleLogs[index].SourceIndex == sourceIndex {
		return index
	}
	return -1
}

func slicesIndex(items []int, target int) int {
	index := sort.SearchInts(items, target)
	if index < len(items) && items[index] == target {
		return index
	}
	return -1
}

func firstSelectionSource(items []int) int {
	if len(items) == 0 {
		return -1
	}
	return items[0]
}
