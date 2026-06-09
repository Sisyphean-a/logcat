package app

import (
	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/logcat"
)

type DeviceItem struct {
	ID     string
	Model  string
	Status string
}

type SavedFilter struct {
	ID          string
	Name        string
	PackageName string
	Query       string
}

type LogViewItem struct {
	Entry   logcat.LogEntry
	Display string
}

type FilterState struct {
	Draft          string
	Applied        string
	Error          string
	ActiveFilterID string
	Saved          []SavedFilter
	History        []string
}

type SearchState struct {
	Query        string
	MatchIndexes []int
	Current      int
}

type PauseState struct {
	Active        bool
	BufferedCount int
	DroppedCount  int
}

type Model struct {
	Status          string
	ADBStatus       string
	Devices         []DeviceItem
	SelectedDevice  string
	PackageScope    adb.PackageScope
	Packages        []adb.PackageInfo
	SelectedPackage string
	Processes       []adb.ProcessInfo
	SelectedProcess string
	BoundPIDs       []int
	TotalLogs       int
	Logs            []string
	VisibleLogs     []LogViewItem
	Filter          FilterState
	Search          SearchState
	Pause           PauseState
	SelectedIndex   int
}

func NewModel() Model {
	defaults := defaultSavedFilters()
	defaultQuery := defaults[0].Query
	return Model{
		Status:          "idle",
		ADBStatus:       "未连接",
		Devices:         []DeviceItem{},
		SelectedDevice:  "",
		PackageScope:    adb.PackageScopeAll,
		Packages:        []adb.PackageInfo{},
		SelectedPackage: "",
		Processes:       []adb.ProcessInfo{},
		SelectedProcess: "",
		BoundPIDs:       []int{},
		TotalLogs:       0,
		Logs:            []string{},
		VisibleLogs:     []LogViewItem{},
		Filter: FilterState{
			Draft:          defaultQuery,
			Applied:        defaultQuery,
			ActiveFilterID: defaults[0].ID,
			Saved:          defaults,
		},
		Search:          SearchState{MatchIndexes: []int{}, Current: -1},
		Pause:           PauseState{},
		SelectedIndex:   -1,
	}
}

func defaultSavedFilters() []SavedFilter {
	return []SavedFilter{
		{
			ID:          "builtin-h5",
			Name:        "H5 日志",
			PackageName: "",
			Query:       "tag:chromium & message:[H5]",
		},
	}
}
