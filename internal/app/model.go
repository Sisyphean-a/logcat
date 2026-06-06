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

type LogViewItem struct {
	Entry   logcat.LogEntry
	Display string
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
	Devices         []DeviceItem
	PackageScope    adb.PackageScope
	Packages        []adb.PackageInfo
	SelectedPackage string
	Processes       []adb.ProcessInfo
	SelectedProcess string
	BoundPIDs       []int
	Logs            []string
	VisibleLogs     []LogViewItem
	Search          SearchState
	Pause           PauseState
	SelectedIndex   int
}

func NewModel() Model {
	return Model{
		Status:          "idle",
		Devices:         []DeviceItem{},
		PackageScope:    adb.PackageScopeUser,
		Packages:        []adb.PackageInfo{},
		SelectedPackage: "",
		Processes:       []adb.ProcessInfo{},
		SelectedProcess: "",
		BoundPIDs:       []int{},
		Logs:            []string{},
		VisibleLogs:     []LogViewItem{},
		Search:          SearchState{MatchIndexes: []int{}, Current: -1},
		Pause:           PauseState{},
		SelectedIndex:   -1,
	}
}
