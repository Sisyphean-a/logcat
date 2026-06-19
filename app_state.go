package main

import (
	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
)

type AppState struct {
	Revision        uint64           `json:"revision"`
	Status          string           `json:"status"`
	ADBStatus       string           `json:"adbStatus"`
	Devices         []DeviceView     `json:"devices"`
	SelectedDevice  string           `json:"selectedDevice"`
	PackageScope    string           `json:"packageScope"`
	Packages        []PackageView    `json:"packages"`
	SelectedPackage string           `json:"selectedPackage"`
	TotalLogs       int              `json:"totalLogs"`
	VisibleCount    int              `json:"visibleCount"`
	Filter          FilterView       `json:"filter"`
	Search          SearchView       `json:"search"`
	Pause           PauseView        `json:"pause"`
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
	Active bool `json:"active"`
}

type LogItemView struct {
	SourceIndex int    `json:"sourceIndex"`
	TimeText    string `json:"timeText"`
	Level       string `json:"level"`
	Tag         string `json:"tag"`
	Message     string `json:"message"`
	IsFocused   bool   `json:"isFocused"`
	IsSelected  bool   `json:"isSelected"`
}

type SelectedLogView struct {
	SourceIndex int    `json:"sourceIndex"`
	TimeText    string `json:"timeText"`
	Level       string `json:"level"`
	Tag         string `json:"tag"`
	Message     string `json:"message"`
	Source      string `json:"source"`
}

func newAppState(snapshot appstate.UISnapshot) AppState {
	model := snapshot.Model
	state := AppState{
		Revision:        snapshot.Revision,
		Status:          model.Status,
		ADBStatus:       model.ADBStatus,
		Devices:         make([]DeviceView, len(model.Devices)),
		SelectedDevice:  model.SelectedDevice,
		PackageScope:    string(model.PackageScope),
		Packages:        make([]PackageView, len(model.Packages)),
		SelectedPackage: model.SelectedPackage,
		TotalLogs:       model.TotalLogs,
		VisibleCount:    snapshot.VisibleCount,
		SelectedCount:   len(model.Selection.SourceIndexes),
		Filter: FilterView{
			Draft:           model.Filter.Draft,
			Applied:         model.Filter.Applied,
			Error:           model.Filter.Error,
			ActiveFilterID:  model.Filter.ActiveFilterID,
			DefaultFilterID: model.Filter.DefaultFilterID,
			Saved:           make([]SavedFilterView, len(model.Filter.Saved)),
		},
		Search: SearchView{
			Query: model.Search.Query,
		},
		Pause: PauseView{
			Active: model.Pause.Active,
		},
	}

	for index := range model.Devices {
		device := &model.Devices[index]
		state.Devices[index] = DeviceView{
			ID:     device.ID,
			Model:  device.Model,
			Status: device.Status,
		}
	}

	for index := range model.Packages {
		pkg := &model.Packages[index]
		state.Packages[index] = PackageView{Name: pkg.Name}
	}

	for index := range model.Filter.Saved {
		filter := &model.Filter.Saved[index]
		state.Filter.Saved[index] = SavedFilterView{
			ID:          filter.ID,
			Name:        filter.Name,
			PackageName: filter.PackageName,
			Query:       filter.Query,
		}
	}

	state.Logs, state.SelectedLog = buildLogRows(snapshot.VisibleLogs, model.Selection)

	return state
}

func buildLogRows(
	items []appstate.VisibleLogSnapshot,
	selection appstate.SelectionState,
) ([]LogItemView, *SelectedLogView) {
	if len(selection.SourceIndexes) <= 1 {
		return buildLogRowsSingleSelection(items, selection.FocusSourceIndex, singleSelectedSourceIndex(selection.SourceIndexes))
	}
	return buildLogRowsMultiSelection(items, selection.FocusSourceIndex, selection.SourceIndexes)
}

func buildLogRowsSingleSelection(
	items []appstate.VisibleLogSnapshot,
	focusedSourceIndex int,
	selectedSourceIndex int,
) ([]LogItemView, *SelectedLogView) {
	logs := make([]LogItemView, len(items))
	var selectedLog *SelectedLogView
	for offset := range items {
		item := &items[offset]
		sourceIndex := item.SourceIndex
		isFocused := sourceIndex == focusedSourceIndex
		logs[offset] = LogItemView{
			SourceIndex: sourceIndex,
			TimeText:    item.TimeText,
			Level:       item.Level,
			Tag:         item.Tag,
			Message:     item.Message,
			IsFocused:   isFocused,
			IsSelected:  sourceIndex == selectedSourceIndex,
		}
		if isFocused {
			selectedLog = buildSelectedLogFromItem(item)
		}
	}
	return logs, selectedLog
}

func buildLogRowsMultiSelection(
	items []appstate.VisibleLogSnapshot,
	focusedSourceIndex int,
	selectedSourceIndexes []int,
) ([]LogItemView, *SelectedLogView) {
	logs := make([]LogItemView, len(items))
	var selectedLog *SelectedLogView
	selectedCount := len(selectedSourceIndexes)
	selectedPos := 0
	nextSelectedSource := selectedSourceIndexes[0]
	for offset := range items {
		item := &items[offset]
		sourceIndex := item.SourceIndex
		isFocused := sourceIndex == focusedSourceIndex
		isSelected := sourceIndex == nextSelectedSource
		if isSelected {
			selectedPos++
			if selectedPos < selectedCount {
				nextSelectedSource = selectedSourceIndexes[selectedPos]
			} else {
				nextSelectedSource = -1
			}
		}
		logs[offset] = LogItemView{
			SourceIndex: sourceIndex,
			TimeText:    item.TimeText,
			Level:       item.Level,
			Tag:         item.Tag,
			Message:     item.Message,
			IsFocused:   isFocused,
			IsSelected:  isSelected,
		}
		if isFocused {
			selectedLog = buildSelectedLogFromItem(item)
		}
	}
	return logs, selectedLog
}

func singleSelectedSourceIndex(sourceIndexes []int) int {
	if len(sourceIndexes) == 1 {
		return sourceIndexes[0]
	}
	return -1
}

func buildSelectedLogFromItem(item *appstate.VisibleLogSnapshot) *SelectedLogView {
	return &SelectedLogView{
		SourceIndex: item.SourceIndex,
		TimeText:    item.TimeText,
		Level:       item.Level,
		Tag:         item.Tag,
		Message:     item.Message,
		Source:      item.Source,
	}
}

type logRowCursor struct {
	focusedSourceIndex  int
	selectedSourceIndex []int
	selectedPos         int
	nextSelectedSource  int
}

func newLogRowCursor(selection appstate.SelectionState) logRowCursor {
	nextSelectedSource := -1
	if len(selection.SourceIndexes) > 0 {
		nextSelectedSource = selection.SourceIndexes[0]
	}
	return logRowCursor{
		focusedSourceIndex:  selection.FocusSourceIndex,
		selectedSourceIndex: selection.SourceIndexes,
		nextSelectedSource:  nextSelectedSource,
	}
}

func (c *logRowCursor) Next(
	sourceIndex int,
	timeText string,
	level string,
	tag string,
	message string,
) LogItemView {
	isSelected := sourceIndex == c.nextSelectedSource
	if isSelected {
		c.selectedPos++
		if c.selectedPos < len(c.selectedSourceIndex) {
			c.nextSelectedSource = c.selectedSourceIndex[c.selectedPos]
		} else {
			c.nextSelectedSource = -1
		}
	}
	return LogItemView{
		SourceIndex: sourceIndex,
		TimeText:    timeText,
		Level:       level,
		Tag:         tag,
		Message:     message,
		IsFocused:   sourceIndex == c.focusedSourceIndex,
		IsSelected:  isSelected,
	}
}

func buildSelectedLogView(row LogItemView, source string) *SelectedLogView {
	return &SelectedLogView{
		SourceIndex: row.SourceIndex,
		TimeText:    row.TimeText,
		Level:       row.Level,
		Tag:         row.Tag,
		Message:     row.Message,
		Source:      source,
	}
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
