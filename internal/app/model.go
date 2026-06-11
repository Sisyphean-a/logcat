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
	Entry        logcat.LogEntry
	SourceIndex  int
	Display      string
	DisplayLower string
	SearchLower  string
}

type FilterState struct {
	Draft           string
	Applied         string
	Error           string
	ActiveFilterID  string
	DefaultFilterID string
	Saved           []SavedFilter
	History         []string
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
	VisibleLogs     []LogViewItem
	Filter          FilterState
	Search          SearchState
	Pause           PauseState
	SelectedIndex   int
}

func NewModel() Model {
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
		VisibleLogs:     []LogViewItem{},
		Filter: FilterState{
			Draft:           "",
			Applied:         "",
			ActiveFilterID:  "",
			DefaultFilterID: "",
			Saved:           []SavedFilter{},
		},
		Search:        SearchState{MatchIndexes: []int{}, Current: -1},
		Pause:         PauseState{Active: true},
		SelectedIndex: -1,
	}
}
