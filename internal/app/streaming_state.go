package app

import (
	"context"
	"fmt"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/session"
)

func (c *Controller) hasActiveSession() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionCancel != nil
}

func (c *Controller) currentSessionConfig() (session.Config, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	deviceID := c.binding.DeviceID
	if deviceID == "" {
		deviceID = c.model.SelectedDevice
	}
	if deviceID == "" {
		return session.Config{}, fmt.Errorf("device_not_selected")
	}
	if c.binding.PackageName != "" && len(c.binding.PIDs) == 0 {
		return session.Config{}, notRunningError(c.binding.PackageName, c.binding.ProcessName)
	}

	return sessionConfig(
		deviceID,
		c.binding.PackageName,
		c.binding.ProcessName,
		c.binding.PIDs,
	), nil
}

func (c *Controller) prepareBindingSelection(
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	pids []int,
	preserveLogs bool,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
		PIDs:        append([]int(nil), pids...),
	}
	c.rememberBindingLocked(c.binding)
	c.resetBindingViewLocked(!preserveLogs)
	c.model.Pause.Active = true
	c.updateBoundModelLocked(packageName, processName, processes, pids)
}

func (c *Controller) prepareStoppedBinding(
	deviceID string,
	packageName string,
	processName string,
	processes []adb.ProcessInfo,
	preserveLogs bool,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.binding = SessionBinding{
		DeviceID:    deviceID,
		PackageName: packageName,
		ProcessName: processName,
	}
	c.rememberBindingLocked(c.binding)
	c.resetBindingViewLocked(!preserveLogs)
	c.model.Pause.Active = true
	c.updateBoundModelLocked(packageName, processName, processes, nil)
	return notRunningError(packageName, processName)
}

func (c *Controller) startCurrentSelection(ctx context.Context) error {
	return c.startCurrentSelectionWithPause(ctx, false)
}

func (c *Controller) startCurrentSelectionWithPause(
	ctx context.Context,
	paused bool,
) error {
	cfg, err := c.currentSessionConfig()
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}
	if err := c.startSession(ctx, cfg); err != nil {
		return err
	}

	c.mu.Lock()
	c.model.Pause.Active = paused
	if paused {
		c.updatePausedStatusLocked()
	} else {
		c.resumeStreaming = true
	}
	c.markDirtyLocked()
	c.mu.Unlock()
	return nil
}
