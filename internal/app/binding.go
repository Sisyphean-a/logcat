package app

import (
	"context"
	"fmt"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/session"
)

type SessionBinding struct {
	DeviceID    string
	PackageName string
	ProcessName string
	PIDs        []int
}

func (c *Controller) SelectPackage(ctx context.Context, packageName string) error {
	return c.selectPackage(ctx, packageName, false)
}

func (c *Controller) selectPackage(
	ctx context.Context,
	packageName string,
	preserveLogs bool,
) error {
	if packageName == "" {
		return c.clearPackageSelection(ctx)
	}

	deviceID, err := c.currentDeviceID()
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	processes, err := c.deviceService.ListProcesses(ctx, deviceID, packageName)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	switch c.currentSessionIntent() {
	case sessionIntentNone:
		if len(processes) == 0 {
			err := c.prepareStoppedBinding(deviceID, packageName, "", nil, preserveLogs)
			c.startBindingWatcher(deviceID, packageName, "")
			return err
		}
		c.prepareBindingSelection(deviceID, packageName, "", processes, collectPIDs(processes), preserveLogs)
		c.startBindingWatcher(deviceID, packageName, "")
		return nil
	case sessionIntentPaused:
		if len(processes) == 0 {
			return c.activateStoppedBinding(deviceID, packageName, "", nil, true, preserveLogs)
		}
		return c.activateRunningBinding(
			ctx,
			deviceID,
			packageName,
			"",
			processes,
			collectPIDs(processes),
			true,
			preserveLogs,
		)
	default:
		if len(processes) == 0 {
			return c.activateStoppedBinding(deviceID, packageName, "", nil, false, preserveLogs)
		}
		return c.activateRunningBinding(
			ctx,
			deviceID,
			packageName,
			"",
			processes,
			collectPIDs(processes),
			false,
			preserveLogs,
		)
	}
}

func (c *Controller) SelectProcess(ctx context.Context, processName string) error {
	return c.selectProcess(ctx, processName, false)
}

func (c *Controller) selectProcess(
	ctx context.Context,
	processName string,
	preserveLogs bool,
) error {
	c.mu.RLock()
	deviceID := c.binding.DeviceID
	packageName := c.binding.PackageName
	processes := append([]adb.ProcessInfo(nil), c.model.Processes...)
	c.mu.RUnlock()

	if deviceID == "" {
		err := fmt.Errorf("device_not_selected")
		c.updateStatus(err.Error())
		return err
	}
	if packageName == "" {
		err := fmt.Errorf("package_not_selected")
		c.updateStatus(err.Error())
		return err
	}

	process, err := findProcess(processes, processName)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	if c.currentSessionIntent() == sessionIntentNone {
		c.prepareBindingSelection(
			deviceID,
			packageName,
			process.Name,
			processes,
			[]int{process.PID},
			preserveLogs,
		)
		c.startBindingWatcher(deviceID, packageName, process.Name)
		return nil
	}

	return c.activateRunningBinding(
		ctx,
		deviceID,
		packageName,
		process.Name,
		processes,
		[]int{process.PID},
		c.currentSessionIntent() == sessionIntentPaused,
		preserveLogs,
	)
}

func (c *Controller) activateRunningBinding(
	ctx context.Context,
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	pids []int,
	paused bool,
	preserveLogs bool,
) error {
	c.stopWatcher()

	if err := c.startSession(ctx, sessionConfig(deviceID, packageName, processName, pids)); err != nil {
		return err
	}

	c.mu.Lock()
	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
		PIDs:        append([]int(nil), pids...),
	}
	c.resetBindingViewLocked(!preserveLogs)
	c.model.Pause.Active = paused
	c.rememberBindingLocked(c.binding)
	c.updateBoundModelLocked(packageName, processName, processes, pids)
	if paused {
		c.updatePausedStatusLocked()
	}
	c.mu.Unlock()

	c.startBindingWatcher(deviceID, packageName, processName)
	return nil
}

func (c *Controller) activateStoppedBinding(
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	paused bool,
	preserveLogs bool,
) error {
	c.stopWatcher()
	c.stopSession()

	c.mu.Lock()
	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
	}
	c.resetBindingViewLocked(!preserveLogs)
	c.model.Pause.Active = paused
	c.rememberBindingLocked(c.binding)
	c.updateBoundModelLocked(packageName, processName, processes, nil)
	c.mu.Unlock()

	c.startBindingWatcher(deviceID, packageName, processName)
	err := notRunningError(packageName, processName)
	c.updateStatus(err.Error())
	return err
}

func (c *Controller) updateBoundModelLocked(
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	pids []int,
) {
	c.model.SelectedPackage = packageName
	c.model.Processes = append(c.model.Processes[:0], processes...)
	c.model.SelectedProcess = processName
	c.model.BoundPIDs = append(c.model.BoundPIDs[:0], pids...)
	c.syncActiveFilterLocked()
	c.markDirtyLocked()
}

func (c *Controller) currentDeviceID() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.binding.DeviceID != "" {
		return c.binding.DeviceID, nil
	}
	if c.model.SelectedDevice == "" {
		return "", fmt.Errorf("device_not_selected")
	}
	return c.model.SelectedDevice, nil
}

func (c *Controller) clearBindingViewLocked() {
	c.resetBindingViewLocked(true)
}

func (c *Controller) resetBindingViewLocked(clearLogs bool) {
	if clearLogs {
		c.allLogs.Reset()
		c.model.TotalLogs = 0
		c.model.VisibleLogs = c.model.VisibleLogs[:0]
		c.model.SelectedIndex = -1
		c.model.Search.MatchIndexes = c.model.Search.MatchIndexes[:0]
		c.model.Search.Current = -1
	}
	c.pauseBuffer = c.pauseBuffer[:0]
	c.model.Pause.BufferedCount = 0
	c.model.Pause.DroppedCount = 0
	c.markDirtyLocked()
}

func collectPIDs(processes []adb.ProcessInfo) []int {
	pids := make([]int, 0, len(processes))
	for _, process := range processes {
		pids = append(pids, process.PID)
	}
	return pids
}

func findProcess(processes []adb.ProcessInfo, processName string) (adb.ProcessInfo, error) {
	for _, process := range processes {
		if process.Name == processName {
			return process, nil
		}
	}

	return adb.ProcessInfo{}, fmt.Errorf("process_not_found: %s", processName)
}

func sessionConfig(
	deviceID string,
	packageName string,
	processName string,
	pids []int,
) session.Config {
	return session.Config{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
		AllowedPIDs: append([]int(nil), pids...),
	}
}

func notRunningError(packageName string, processName string) error {
	if processName != "" {
		return fmt.Errorf("process_not_running: %s", processName)
	}

	return fmt.Errorf("app_not_running: %s", packageName)
}
