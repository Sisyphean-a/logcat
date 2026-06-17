package app

import "github.com/xiakn/logcat/internal/adb"

func cloneModel(model Model) Model {
	cloned := model
	cloned.Devices = append([]DeviceItem(nil), model.Devices...)
	cloned.Packages = append([]adb.PackageInfo(nil), model.Packages...)
	cloned.Processes = append([]adb.ProcessInfo(nil), model.Processes...)
	cloned.BoundPIDs = append([]int(nil), model.BoundPIDs...)
	cloned.VisibleLogs = append([]LogViewItem(nil), model.VisibleLogs...)
	cloned.Filter = cloneFilterState(model.Filter)
	cloned.Selection.SourceIndexes = append([]int(nil), model.Selection.SourceIndexes...)
	return cloned
}
