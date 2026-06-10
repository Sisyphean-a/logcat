package app

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/xiakn/logcat/internal/adb"
)

func (c *Controller) stopSession() {
	c.swapSession(nil)
}

func (c *Controller) stopWatcher() {
	c.swapWatcher(nil)
}

func (c *Controller) startBindingWatcher(
	deviceID string,
	packageName string,
	processName string,
) {
	if packageName == "" {
		c.stopWatcher()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.swapWatcher(cancel)
	go c.watchBinding(ctx, deviceID, packageName, processName)
}

func (c *Controller) swapWatcher(cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watchCancel != nil {
		c.watchCancel()
	}
	c.watchCancel = cancel
}

func (c *Controller) watchBinding(
	ctx context.Context,
	deviceID string,
	packageName string,
	processName string,
) {
	ticker := time.NewTicker(c.bindingPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshBinding(ctx, deviceID, packageName, processName)
		}
	}
}

func (c *Controller) refreshBinding(
	ctx context.Context,
	deviceID string,
	packageName string,
	processName string,
) {
	processes, err := c.deviceService.ListProcesses(ctx, deviceID, packageName)
	if err != nil || ctx.Err() != nil {
		if err != nil && ctx.Err() == nil {
			c.updateStatus(err.Error())
		}
		return
	}

	pids := boundPIDsForProcess(processes, processName)
	if !c.shouldRebind(deviceID, packageName, processName, pids) {
		return
	}
	if len(pids) == 0 {
		c.applyWatcherStoppedBinding(deviceID, packageName, processName, processes)
		return
	}
	if !c.hasActiveSession() {
		if c.currentSessionIntent() == sessionIntentRunning {
			c.applyWatcherRunningBinding(ctx, deviceID, packageName, processName, processes, pids)
			return
		}
		c.applyWatcherPreparedBinding(deviceID, packageName, processName, processes, pids)
		return
	}

	c.applyWatcherRunningBinding(ctx, deviceID, packageName, processName, processes, pids)
}

func (c *Controller) shouldRebind(
	deviceID string,
	packageName string,
	processName string,
	pids []int,
) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.binding.DeviceID != deviceID ||
		c.binding.PackageName != packageName ||
		c.binding.ProcessName != processName {
		return false
	}

	return !samePIDs(c.binding.PIDs, pids)
}

func (c *Controller) applyWatcherStoppedBinding(
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
) {
	c.stopSession()

	c.mu.Lock()
	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
	}
	c.updateBoundModelLocked(packageName, processName, processes, nil)
	c.mu.Unlock()

	c.updateStatus(notRunningError(packageName, processName).Error())
}

func (c *Controller) applyWatcherPreparedBinding(
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	pids []int,
) {
	c.mu.Lock()
	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
		PIDs:        append([]int(nil), pids...),
	}
	c.updateBoundModelLocked(packageName, processName, processes, pids)
	if len(pids) == 0 {
		c.model.Status = notRunningError(packageName, processName).Error()
	} else {
		c.model.Status = ""
	}
	c.markDirtyLocked()
	c.mu.Unlock()
}

func (c *Controller) applyWatcherRunningBinding(
	ctx context.Context,
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	pids []int,
) {
	if err := c.startSession(ctx, sessionConfig(deviceID, packageName, processName, pids)); err != nil {
		return
	}

	c.mu.Lock()
	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
		PIDs:        append([]int(nil), pids...),
	}
	c.updateBoundModelLocked(packageName, processName, processes, pids)
	c.mu.Unlock()

	c.updateStatus(rebindStatus(packageName, processName, pids))
}

func boundPIDsForProcess(processes []adb.ProcessInfo, processName string) []int {
	if processName == "" {
		return collectPIDs(processes)
	}

	process, err := findProcess(processes, processName)
	if err != nil {
		return nil
	}

	return []int{process.PID}
}

func samePIDs(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}

	leftCopy := append([]int(nil), left...)
	rightCopy := append([]int(nil), right...)
	slices.Sort(leftCopy)
	slices.Sort(rightCopy)
	return slices.Equal(leftCopy, rightCopy)
}

func rebindStatus(packageName string, processName string, pids []int) string {
	target := packageName
	if processName != "" {
		target = processName
	}

	return fmt.Sprintf("rebound: %s %v", target, pids)
}
