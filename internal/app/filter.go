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
		return err
	}

	c.model.Filter.ActiveFilterID = filter.ID
	c.model.Filter.Draft = filter.Query
	c.model.Filter.Applied = filter.Query
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
	c.model.Filter.Applied = filter.Query
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
	c.mu.Lock()
	defer c.mu.Unlock()

	trimmedName := strings.TrimSpace(name)
	query := strings.TrimSpace(c.model.Filter.Draft)
	if trimmedName == "" {
		err := fmt.Errorf("saved_filter_name_required")
		c.model.Filter.Error = err.Error()
		return err
	}
	if err := validateFilterQuery(query); err != nil {
		c.model.Filter.Error = err.Error()
		return err
	}

	id := strings.ToLower(trimmedName)
	id = strings.ReplaceAll(id, " ", "-")
	saved := SavedFilter{
		ID:          id,
		Name:        trimmedName,
		PackageName: c.model.SelectedPackage,
		Query:       query,
	}
	c.model.Filter.Saved = upsertSavedFilter(c.model.Filter.Saved, saved)
	c.model.Filter.ActiveFilterID = saved.ID
	c.model.Filter.Applied = query
	c.model.Filter.Draft = query
	c.model.Filter.Error = ""
	c.recordFilterHistoryLocked(query)
	c.rebuildVisibleFromAllLogsLocked()
	return nil
}

func (c *Controller) applyFilterQueryLocked(query string, recordHistory bool) error {
	if err := validateFilterQuery(query); err != nil {
		c.model.Filter.Error = err.Error()
		return err
	}

	c.model.Filter.Applied = query
	c.model.Filter.Error = ""
	c.syncActiveFilterLocked()
	if recordHistory {
		c.recordFilterHistoryLocked(query)
	}
	c.rebuildVisibleFromAllLogsLocked()
	return nil
}

func validateFilterQuery(query string) error {
	if query == "" {
		return nil
	}

	open := 0
	for _, r := range query {
		switch r {
		case '(':
			open++
		case ')':
			open--
		}
		if open < 0 {
			return fmt.Errorf("filter_query_invalid: unmatched ')'")
		}
	}
	if open != 0 {
		return fmt.Errorf("filter_query_invalid: unmatched '('")
	}
	if strings.Count(query, "\"")%2 != 0 {
		return fmt.Errorf("filter_query_invalid: unmatched quote")
	}
	return nil
}

func matchesFilter(entry logcat.LogEntry, packageName string, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	parts := splitAndTerms(query)
	for _, part := range parts {
		if !matchTerm(entry, packageName, part) {
			return false
		}
	}
	return true
}

func splitAndTerms(query string) []string {
	parts := strings.Split(query, "&")
	terms := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(strings.Trim(part, "()"))
		if trimmed != "" {
			terms = append(terms, trimmed)
		}
	}
	return terms
}

func matchTerm(entry logcat.LogEntry, packageName string, term string) bool {
	negated := strings.HasPrefix(term, "-")
	if negated {
		term = strings.TrimSpace(strings.TrimPrefix(term, "-"))
	}

	matched := true
	switch {
	case strings.HasPrefix(term, "tag:"):
		expected := strings.TrimSpace(strings.TrimPrefix(term, "tag:"))
		matched = entry.Tag == "" || strings.EqualFold(entry.Tag, expected)
	case strings.HasPrefix(term, "message:"):
		expected := strings.Trim(strings.TrimSpace(strings.TrimPrefix(term, "message:")), "\"")
		matched = strings.Contains(strings.ToLower(entry.Message), strings.ToLower(expected))
	case strings.HasPrefix(term, "package:"):
		matched = strings.EqualFold(packageName, strings.TrimSpace(strings.TrimPrefix(term, "package:")))
	case strings.HasPrefix(term, "level:"):
		expected := strings.TrimSpace(strings.TrimPrefix(term, "level:"))
		matched = entry.Level == "" || strings.EqualFold(entry.Level, expected)
	default:
		matched = strings.Contains(strings.ToLower(entry.Message), strings.ToLower(term))
	}

	if negated {
		return !matched
	}
	return matched
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
}
