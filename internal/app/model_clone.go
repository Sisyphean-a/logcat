package app

import "github.com/xiakn/logcat/internal/adb"

func cloneModel(model Model) Model {
	cloned := model
	cloned.Devices = append([]DeviceItem(nil), model.Devices...)
	cloned.Packages = append([]adb.PackageInfo(nil), model.Packages...)
	cloned.Processes = append([]adb.ProcessInfo(nil), model.Processes...)
	cloned.BoundPIDs = append([]int(nil), model.BoundPIDs...)
	cloned.Logs = append([]string(nil), model.Logs...)
	cloned.VisibleLogs = append([]LogViewItem(nil), model.VisibleLogs...)
	cloned.Filter.Saved = append([]SavedFilter(nil), model.Filter.Saved...)
	cloned.Filter.History = append([]string(nil), model.Filter.History...)
	cloned.Search.MatchIndexes = append([]int(nil), model.Search.MatchIndexes...)
	return cloned
}
