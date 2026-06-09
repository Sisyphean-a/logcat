package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	appstate "github.com/xiakn/logcat/internal/app"
)

type SavedFiltersFile struct {
	Filters []appstate.SavedFilter `json:"filters"`
	History []string               `json:"history,omitempty"`
}

func LoadFilterState() (SavedFiltersFile, error) {
	path, err := filtersPath()
	if err != nil {
		return SavedFiltersFile{}, err
	}

	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return SavedFiltersFile{}, nil
	}
	if err != nil {
		return SavedFiltersFile{}, err
	}

	var payload SavedFiltersFile
	if err := json.Unmarshal(content, &payload); err != nil {
		return SavedFiltersFile{}, err
	}
	return payload, nil
}

func LoadSavedFilters() ([]appstate.SavedFilter, error) {
	payload, err := LoadFilterState()
	if err != nil {
		return nil, err
	}
	return payload.Filters, nil
}

func SaveFilterState(filters []appstate.SavedFilter, history []string) error {
	path, err := filtersPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(SavedFiltersFile{
		Filters: filters,
		History: history,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func SaveSavedFilters(filters []appstate.SavedFilter) error {
	return SaveFilterState(filters, nil)
}

func filtersPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "logcat-viewer", "saved-filters.json"), nil
}
