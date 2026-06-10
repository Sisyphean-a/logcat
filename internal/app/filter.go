package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

func (c *Controller) SetFilterDraft(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Filter.Draft = strings.TrimSpace(query)
	c.markDirtyLocked()
}

func (c *Controller) ReplaceSavedFilters(filters []SavedFilter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Filter.Saved = append(c.model.Filter.Saved[:0], filters...)
	if _, ok := findSavedFilter(c.model.Filter.Saved, c.model.Filter.ActiveFilterID); !ok {
		c.model.Filter.ActiveFilterID = ""
	}
	c.rebuildVisibleFromAllLogsLocked()
}

func (c *Controller) ApplyFilterDraft() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	query := strings.TrimSpace(c.model.Filter.Draft)
	c.model.Filter.Draft = query
	return c.applyFilterQueryLocked(query, true)
}

func (c *Controller) SelectSavedFilter(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filter, ok := findSavedFilter(c.model.Filter.Saved, id)
	if !ok {
		err := fmt.Errorf("saved_filter_not_found: %s", id)
		c.model.Filter.Error = err.Error()
		c.markDirtyLocked()
		return err
	}

	compiled, err := compileFilterQuery(filter.Query)
	if err != nil {
		c.model.Filter.Error = err.Error()
		c.markDirtyLocked()
		return err
	}

	c.model.Filter.ActiveFilterID = filter.ID
	c.model.Filter.Draft = filter.Query
	c.setCompiledFilterLocked(filter.Query, compiled)
	c.model.Filter.Error = ""
	if filter.PackageName != "" {
		c.model.SelectedPackage = filter.PackageName
	}
	c.rebuildVisibleFromAllLogsLocked()
	return nil
}

func (c *Controller) ApplySavedFilter(ctx context.Context, id string) error {
	if id == "" {
		c.clearSavedFilterSelection()
		return nil
	}

	c.mu.RLock()
	filter, ok := findSavedFilter(c.model.Filter.Saved, id)
	currentPackage := c.model.SelectedPackage
	currentDevice := c.binding.DeviceID
	if currentDevice == "" {
		currentDevice = c.model.SelectedDevice
	}
	c.mu.RUnlock()

	if !ok {
		err := fmt.Errorf("saved_filter_not_found: %s", id)
		c.updateFilterError(err.Error())
		return err
	}

	compiled, err := compileFilterQuery(filter.Query)
	if err != nil {
		c.updateFilterError(err.Error())
		return err
	}

	var bindErr error
	switch {
	case filter.PackageName != "" && filter.PackageName != currentPackage:
		bindErr = c.SelectPackage(ctx, filter.PackageName)
	case filter.PackageName == "" && currentPackage != "" && currentDevice != "":
		bindErr = c.SelectDevice(ctx, currentDevice)
	}

	c.mu.Lock()
	c.model.Filter.ActiveFilterID = filter.ID
	c.model.Filter.Draft = filter.Query
	c.setCompiledFilterLocked(filter.Query, compiled)
	c.model.Filter.Error = ""
	if filter.PackageName == "" {
		c.model.SelectedPackage = ""
	} else {
		c.model.SelectedPackage = filter.PackageName
	}
	c.recordFilterHistoryLocked(filter.Query)
	c.rebuildVisibleFromAllLogsLocked()
	c.mu.Unlock()

	return bindErr
}

func (c *Controller) SaveCurrentFilter(name string) error {
	c.mu.RLock()
	packageName := c.model.SelectedPackage
	query := c.model.Filter.Draft
	c.mu.RUnlock()

	return c.SaveFilterDefinition(name, packageName, query)
}

func (c *Controller) SaveFilterDefinition(name string, packageName string, query string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	trimmedName := strings.TrimSpace(name)
	trimmedPackageName := strings.TrimSpace(packageName)
	query = strings.TrimSpace(query)
	if trimmedName == "" {
		err := fmt.Errorf("saved_filter_name_required")
		c.model.Filter.Error = err.Error()
		c.markDirtyLocked()
		return err
	}
	compiled, err := compileFilterQuery(query)
	if err != nil {
		c.model.Filter.Error = err.Error()
		c.markDirtyLocked()
		return err
	}

	id := strings.ToLower(trimmedName)
	id = strings.ReplaceAll(id, " ", "-")
	saved := SavedFilter{
		ID:          id,
		Name:        trimmedName,
		PackageName: trimmedPackageName,
		Query:       query,
	}
	c.model.Filter.Saved = upsertSavedFilter(c.model.Filter.Saved, saved)
	c.model.Filter.ActiveFilterID = saved.ID
	c.setCompiledFilterLocked(query, compiled)
	c.model.Filter.Draft = query
	c.model.Filter.Error = ""
	c.recordFilterHistoryLocked(query)
	c.rebuildVisibleFromAllLogsLocked()
	return nil
}

func (c *Controller) applyFilterQueryLocked(query string, recordHistory bool) error {
	compiled, err := compileFilterQuery(query)
	if err != nil {
		c.model.Filter.Error = err.Error()
		c.markDirtyLocked()
		return err
	}

	c.setCompiledFilterLocked(query, compiled)
	c.model.Filter.Error = ""
	c.syncActiveFilterLocked()
	if recordHistory {
		c.recordFilterHistoryLocked(query)
	}
	c.rebuildVisibleFromAllLogsLocked()
	return nil
}

func validateFilterQuery(query string) error {
	_, err := compileFilterQuery(query)
	return err
}

func matchesFilter(entry logcat.LogEntry, packageName string, query string) bool {
	compiled, err := compileFilterQuery(query)
	if err != nil {
		return false
	}
	return compiled.matches(entry, packageName)
}

func findSavedFilter(filters []SavedFilter, id string) (SavedFilter, bool) {
	for _, filter := range filters {
		if filter.ID == id {
			return filter, true
		}
	}
	return SavedFilter{}, false
}

func upsertSavedFilter(filters []SavedFilter, saved SavedFilter) []SavedFilter {
	for index, filter := range filters {
		if filter.ID == saved.ID {
			filters[index] = saved
			return filters
		}
	}
	return append(filters, saved)
}

func (c *Controller) updateFilterError(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model.Filter.Error = message
	c.markDirtyLocked()
}
