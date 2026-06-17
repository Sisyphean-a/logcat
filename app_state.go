package main

import (
	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
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
	VisibleStart    int              `json:"visibleStart"`
	Filter          FilterView       `json:"filter"`
	Search          SearchView       `json:"search"`
	Pause           PauseView        `json:"pause"`
	SelectedIndex   int              `json:"selectedIndex"`
	SelectedCount   int              `json:"selectedCount"`
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
	Draft           string            `json:"draft"`
	Applied         string            `json:"applied"`
	Error           string            `json:"error"`
	ActiveFilterID  string            `json:"activeFilterId"`
	DefaultFilterID string            `json:"defaultFilterId"`
	Saved           []SavedFilterView `json:"saved"`
	History         []string          `json:"history"`
}

type SavedFilterView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PackageName string `json:"packageName"`
	Query       string `json:"query"`
}

type SearchView struct {
	Query string `json:"query"`
}

type PauseView struct {
	Active        bool `json:"active"`
	BufferedCount int  `json:"bufferedCount"`
	DroppedCount  int  `json:"droppedCount"`
}

type LogItemView struct {
	Index       int    `json:"index"`
	SourceIndex int    `json:"sourceIndex"`
	TimeText    string `json:"timeText"`
	Level       string `json:"level"`
	Tag         string `json:"tag"`
	Message     string `json:"message"`
	IsFocused   bool   `json:"isFocused"`
	IsSelected  bool   `json:"isSelected"`
}

type SelectedLogView struct {
	TimeText string `json:"timeText"`
	Level    string `json:"level"`
	Tag      string `json:"tag"`
	Message  string `json:"message"`
	Source   string `json:"source"`
	Raw      string `json:"raw"`
	Display  string `json:"display"`
}

func newAppState(snapshot appstate.UISnapshot) AppState {
	model := snapshot.Model
	state := AppState{
		Status:          model.Status,
		ADBStatus:       model.ADBStatus,
		Devices:         make([]DeviceView, len(model.Devices)),
		SelectedDevice:  model.SelectedDevice,
		PackageScope:    string(model.PackageScope),
		Packages:        make([]PackageView, len(model.Packages)),
		SelectedPackage: model.SelectedPackage,
		Processes:       make([]ProcessView, len(model.Processes)),
		SelectedProcess: model.SelectedProcess,
		BoundPIDs:       model.BoundPIDs,
		TotalLogs:       model.TotalLogs,
		VisibleCount:    snapshot.VisibleCount,
		VisibleStart:    snapshot.VisibleStart,
		SelectedIndex:   model.SelectedIndex,
		SelectedCount:   len(model.Selection.SourceIndexes),
		Filter: FilterView{
			Draft:           model.Filter.Draft,
			Applied:         model.Filter.Applied,
			Error:           model.Filter.Error,
			ActiveFilterID:  model.Filter.ActiveFilterID,
			DefaultFilterID: model.Filter.DefaultFilterID,
			Saved:           make([]SavedFilterView, len(model.Filter.Saved)),
			History:         model.Filter.History,
		},
		Search: SearchView{
			Query: model.Search.Query,
		},
		Pause: PauseView{
			Active:        model.Pause.Active,
			BufferedCount: model.Pause.BufferedCount,
			DroppedCount:  model.Pause.DroppedCount,
		},
		Logs: make([]LogItemView, len(model.VisibleLogs)),
	}

	focusedSourceIndex := model.Selection.FocusSourceIndex
	selectedPos := 0
	nextSelectedSource := -1
	if len(model.Selection.SourceIndexes) > 0 {
		nextSelectedSource = model.Selection.SourceIndexes[0]
	}

	for index, device := range model.Devices {
		state.Devices[index] = DeviceView{
			ID:     device.ID,
			Model:  device.Model,
			Status: device.Status,
		}
	}

	for index, pkg := range model.Packages {
		state.Packages[index] = PackageView{Name: pkg.Name}
	}

	for index, process := range model.Processes {
		state.Processes[index] = ProcessView{
			PID:  process.PID,
			Name: process.Name,
		}
	}

	for index, filter := range model.Filter.Saved {
		state.Filter.Saved[index] = SavedFilterView{
			ID:          filter.ID,
			Name:        filter.Name,
			PackageName: filter.PackageName,
			Query:       filter.Query,
		}
	}

	for offset, item := range model.VisibleLogs {
		index := snapshot.VisibleStart + offset
		isSelected := item.SourceIndex == nextSelectedSource
		if isSelected {
			selectedPos++
			if selectedPos < len(model.Selection.SourceIndexes) {
				nextSelectedSource = model.Selection.SourceIndexes[selectedPos]
			} else {
				nextSelectedSource = -1
			}
		}
		row := LogItemView{
			Index:       index,
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
			IsFocused:   item.SourceIndex == focusedSourceIndex,
			IsSelected:  isSelected,
		}
		state.Logs[offset] = row
		if row.IsFocused {
			state.SelectedLog = &SelectedLogView{
				TimeText: row.TimeText,
				Level:    row.Level,
				Tag:      row.Tag,
				Message:  row.Message,
				Source:   item.Entry.Source,
				Raw:      item.Entry.Raw,
				Display:  appstate.FormatLogDisplay(item.Entry),
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
