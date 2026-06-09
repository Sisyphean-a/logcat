package main

import (
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/adb"
)

type AppState struct {
	Status          string           `json:"status"`
	ADBStatus       string           `json:"adbStatus"`
	Devices         []DeviceView     `json:"devices"`
	SelectedDevice  string           `json:"selectedDevice"`
	PackageScope    string           `json:"packageScope"`
	Packages        []PackageView    `json:"packages"`
	SelectedPackage string           `json:"selectedPackage"`
	Processes       []ProcessView    `json:"processes"`
	SelectedProcess string           `json:"selectedProcess"`
	BoundPIDs       []int            `json:"boundPids"`
	TotalLogs       int              `json:"totalLogs"`
	VisibleCount    int              `json:"visibleCount"`
	Filter          FilterView       `json:"filter"`
	Search          SearchView       `json:"search"`
	Pause           PauseView        `json:"pause"`
	SelectedIndex   int              `json:"selectedIndex"`
	Logs            []LogItemView    `json:"logs"`
	SelectedLog     *SelectedLogView `json:"selectedLog"`
}

type DeviceView struct {
	ID     string `json:"id"`
	Model  string `json:"model"`
	Status string `json:"status"`
}

type PackageView struct {
	Name string `json:"name"`
}

type ProcessView struct {
	PID  int    `json:"pid"`
	Name string `json:"name"`
}

type FilterView struct {
	Draft          string            `json:"draft"`
	Applied        string            `json:"applied"`
	Error          string            `json:"error"`
	ActiveFilterID string            `json:"activeFilterId"`
	Saved          []SavedFilterView `json:"saved"`
	History        []string          `json:"history"`
}

type SavedFilterView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PackageName string `json:"packageName"`
	Query       string `json:"query"`
}

type SearchView struct {
	Query        string `json:"query"`
	MatchIndexes []int  `json:"matchIndexes"`
	Current      int    `json:"current"`
}

type PauseView struct {
	Active        bool `json:"active"`
	BufferedCount int  `json:"bufferedCount"`
	DroppedCount  int  `json:"droppedCount"`
}

type LogItemView struct {
	Index     int    `json:"index"`
	TimeText  string `json:"timeText"`
	Level     string `json:"level"`
	Tag       string `json:"tag"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Raw       string `json:"raw"`
	Display   string `json:"display"`
	IsMatch   bool   `json:"isMatch"`
	IsCurrent bool   `json:"isCurrent"`
	IsSelected bool  `json:"isSelected"`
}

type SelectedLogView struct {
	Index    int    `json:"index"`
	TimeText string `json:"timeText"`
	Level    string `json:"level"`
	Tag      string `json:"tag"`
	Message  string `json:"message"`
	Source   string `json:"source"`
	Raw      string `json:"raw"`
	Display  string `json:"display"`
}

func newAppState(model appstate.Model) AppState {
	state := AppState{
		Status:          model.Status,
		ADBStatus:       model.ADBStatus,
		Devices:         make([]DeviceView, 0, len(model.Devices)),
		SelectedDevice:  model.SelectedDevice,
		PackageScope:    string(model.PackageScope),
		Packages:        make([]PackageView, 0, len(model.Packages)),
		SelectedPackage: model.SelectedPackage,
		Processes:       make([]ProcessView, 0, len(model.Processes)),
		SelectedProcess: model.SelectedProcess,
		BoundPIDs:       append([]int(nil), model.BoundPIDs...),
		TotalLogs:       model.TotalLogs,
		VisibleCount:    len(model.VisibleLogs),
		SelectedIndex:   model.SelectedIndex,
		Filter: FilterView{
			Draft:          model.Filter.Draft,
			Applied:        model.Filter.Applied,
			Error:          model.Filter.Error,
			ActiveFilterID: model.Filter.ActiveFilterID,
			Saved:          make([]SavedFilterView, 0, len(model.Filter.Saved)),
			History:        append([]string(nil), model.Filter.History...),
		},
		Search: SearchView{
			Query:        model.Search.Query,
			MatchIndexes: append([]int(nil), model.Search.MatchIndexes...),
			Current:      model.Search.Current,
		},
		Pause: PauseView{
			Active:        model.Pause.Active,
			BufferedCount: model.Pause.BufferedCount,
			DroppedCount:  model.Pause.DroppedCount,
		},
		Logs: make([]LogItemView, 0, len(model.VisibleLogs)),
	}

	matchSet := make(map[int]struct{}, len(model.Search.MatchIndexes))
	for _, index := range model.Search.MatchIndexes {
		matchSet[index] = struct{}{}
	}

	currentMatch := -1
	if model.Search.Current >= 0 && model.Search.Current < len(model.Search.MatchIndexes) {
		currentMatch = model.Search.MatchIndexes[model.Search.Current]
	}

	for _, device := range model.Devices {
		state.Devices = append(state.Devices, DeviceView{
			ID:     device.ID,
			Model:  device.Model,
			Status: device.Status,
		})
	}

	for _, pkg := range model.Packages {
		state.Packages = append(state.Packages, PackageView{Name: pkg.Name})
	}

	for _, process := range model.Processes {
		state.Processes = append(state.Processes, ProcessView{
			PID:  process.PID,
			Name: process.Name,
		})
	}

	for _, filter := range model.Filter.Saved {
		state.Filter.Saved = append(state.Filter.Saved, SavedFilterView{
			ID:          filter.ID,
			Name:        filter.Name,
			PackageName: filter.PackageName,
			Query:       filter.Query,
		})
	}

	for index, item := range model.VisibleLogs {
		_, isMatch := matchSet[index]
		row := LogItemView{
			Index:      index,
			TimeText:   item.Entry.TimeText,
			Level:      item.Entry.Level,
			Tag:        item.Entry.Tag,
			Message:    item.Entry.Message,
			Source:     item.Entry.Source,
			Raw:        item.Entry.Raw,
			Display:    item.Display,
			IsMatch:    isMatch,
			IsCurrent:  currentMatch == index,
			IsSelected: model.SelectedIndex == index,
		}
		state.Logs = append(state.Logs, row)
		if row.IsSelected {
			state.SelectedLog = &SelectedLogView{
				Index:    index,
				TimeText: row.TimeText,
				Level:    row.Level,
				Tag:      row.Tag,
				Message:  row.Message,
				Source:   row.Source,
				Raw:      row.Raw,
				Display:  row.Display,
			}
		}
	}

	return state
}

func appstatePackageScope(scope string) adb.PackageScope {
	if scope == "" {
		return ""
	}

	switch adb.PackageScope(scope) {
	case adb.PackageScopeUser, adb.PackageScopeSystem, adb.PackageScopeAll:
		return adb.PackageScope(scope)
	default:
		return adb.PackageScopeAll
	}
}
