package app

import "strings"

const maxFilterHistory = 50

func (c *Controller) ReplaceFilterHistory(history []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Filter.History = normalizeFilterHistory(history)
}

func (c *Controller) ApplyHistoryQuery(query string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	query = strings.TrimSpace(query)
	c.model.Filter.Draft = query
	return c.applyFilterQueryLocked(query, true)
}

func (c *Controller) recordFilterHistoryLocked(query string) {
	normalized := strings.TrimSpace(query)
	if normalized == "" {
		return
	}

	history := []string{normalized}
	for _, item := range c.model.Filter.History {
		if strings.EqualFold(item, normalized) {
			continue
		}
		history = append(history, item)
		if len(history) == maxFilterHistory {
			break
		}
	}
	c.model.Filter.History = history
}

func (c *Controller) syncActiveFilterLocked() {
	c.model.Filter.ActiveFilterID = ""
	for _, filter := range c.model.Filter.Saved {
		if filter.Query != c.model.Filter.Applied {
			continue
		}
		if filter.PackageName != c.model.SelectedPackage {
			continue
		}
		c.model.Filter.ActiveFilterID = filter.ID
		return
	}
}

func normalizeFilterHistory(history []string) []string {
	items := make([]string, 0, min(len(history), maxFilterHistory))
	seen := make(map[string]struct{}, len(history))
	for _, item := range history {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, normalized)
		if len(items) == maxFilterHistory {
			break
		}
	}
	return items
}
