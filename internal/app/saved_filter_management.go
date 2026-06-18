package app

import (
	"context"
	"fmt"
	"strings"
)

type SavedFilterDraft struct {
	ExistingID  string
	Name        string
	PackageName string
	Query       string
}

func (c *Controller) ReplaceSavedFilters(filters []SavedFilter, defaultFilterID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Filter.Saved = append(c.model.Filter.Saved[:0], filters...)
	c.model.Filter.DefaultFilterID = normalizeSavedFilterID(defaultFilterID, c.model.Filter.Saved)
	if _, ok := findSavedFilter(c.model.Filter.Saved, c.model.Filter.ActiveFilterID); !ok {
		c.model.Filter.ActiveFilterID = ""
	}
	c.rebuildVisibleFromAllLogsLocked()
}

func (c *Controller) ReplaceSavedFilterDefinitions(
	ctx context.Context,
	drafts []SavedFilterDraft,
	defaultFilterID string,
	activeFilterID string,
) error {
	filters, renamedIDs, err := buildSavedFilterDrafts(drafts)
	if err != nil {
		c.updateFilterError(err.Error())
		return err
	}

	nextDefaultID := renamedIDs[strings.TrimSpace(defaultFilterID)]
	nextActiveID := renamedIDs[strings.TrimSpace(activeFilterID)]

	c.mu.Lock()
	c.model.Filter.Saved = filters
	c.model.Filter.DefaultFilterID = normalizeSavedFilterID(nextDefaultID, filters)
	c.model.Filter.Error = ""
	c.markDirtyLocked()
	c.mu.Unlock()

	if nextActiveID == "" {
		return c.clearSavedFilterSelection(ctx)
	}
	return c.applySelectedSavedFilter(ctx, nextActiveID)
}

func (c *Controller) RestoreSavedFilter(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filter, ok := findSavedFilter(c.model.Filter.Saved, id)
	if !ok {
		return fmt.Errorf("saved_filter_not_found: %s", id)
	}

	compiled, err := compileFilterQuery(filter.Query)
	if err != nil {
		return err
	}

	c.model.Filter.ActiveFilterID = filter.ID
	c.model.Filter.Draft = filter.Query
	c.model.SelectedPackage = filter.PackageName
	c.model.Filter.Error = ""
	c.setCompiledFilterLocked(filter.Query, compiled)
	c.rebuildVisibleFromAllLogsLocked()
	c.markDirtyLocked()
	return nil
}

func (c *Controller) SavedFilterPackageName(id string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filter, ok := findSavedFilter(c.model.Filter.Saved, id)
	if !ok {
		return ""
	}
	return filter.PackageName
}

func (c *Controller) applySelectedSavedFilter(ctx context.Context, id string) error {
	c.mu.RLock()
	filter, ok := findSavedFilter(c.model.Filter.Saved, id)
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
	if currentDevice == "" && filter.PackageName != "" {
		return c.RestoreSavedFilter(id)
	}
	return c.ApplySavedFilter(ctx, id)
}

func buildSavedFilterDrafts(drafts []SavedFilterDraft) ([]SavedFilter, map[string]string, error) {
	filters := make([]SavedFilter, 0, len(drafts))
	renamedIDs := make(map[string]string, len(drafts))
	seenExisting := make(map[string]struct{}, len(drafts))
	seenFilterIDs := make(map[string]string, len(drafts))

	for _, draft := range drafts {
		existingID := strings.TrimSpace(draft.ExistingID)
		if existingID == "" {
			return nil, nil, fmt.Errorf("saved_filter_existing_id_required")
		}
		if _, ok := seenExisting[existingID]; ok {
			return nil, nil, fmt.Errorf("saved_filter_duplicate_entry: %s", existingID)
		}
		seenExisting[existingID] = struct{}{}

		filter, err := validatedSavedFilter(draft.Name, draft.PackageName, draft.Query)
		if err != nil {
			return nil, nil, err
		}
		if otherExistingID, ok := seenFilterIDs[filter.ID]; ok {
			return nil, nil, fmt.Errorf("saved_filter_name_conflict: %s,%s", otherExistingID, existingID)
		}
		seenFilterIDs[filter.ID] = existingID
		renamedIDs[existingID] = filter.ID
		filters = append(filters, filter)
	}

	return filters, renamedIDs, nil
}

func normalizeSavedFilterID(id string, filters []SavedFilter) string {
	normalizedID := strings.TrimSpace(id)
	if normalizedID == "" {
		return ""
	}
	if _, ok := findSavedFilter(filters, normalizedID); ok {
		return normalizedID
	}
	return ""
}
