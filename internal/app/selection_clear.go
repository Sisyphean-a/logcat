package app

import (
	"context"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/session"
)

func (c *Controller) clearDeviceSelection() {
	c.stopWatcher()
	c.stopSession()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.binding = SessionBinding{}
	c.clearBindingViewLocked()
	c.model.Status = "idle"
	c.model.SelectedDevice = ""
	c.model.PackageScope = ""
	c.model.Packages = c.model.Packages[:0]
	c.model.SelectedPackage = ""
	c.model.Processes = c.model.Processes[:0]
	c.model.SelectedProcess = ""
	c.model.BoundPIDs = c.model.BoundPIDs[:0]
	c.syncActiveFilterLocked()
}

func (c *Controller) clearPackageSelection(ctx context.Context) error {
	deviceID := c.selectedOrBoundDeviceID()
	if deviceID == "" {
		c.clearDevicePackageBinding("")
		return nil
	}

	c.stopWatcher()
	if err := c.startSession(ctx, session.Config{DeviceID: deviceID}); err != nil {
		return err
	}

	c.clearDevicePackageBinding(deviceID)
	return nil
}

func (c *Controller) clearSavedFilterSelection() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Filter.ActiveFilterID = ""
	c.model.Filter.Error = ""
}

func (c *Controller) selectedOrBoundDeviceID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.binding.DeviceID != "" {
		return c.binding.DeviceID
	}
	return c.model.SelectedDevice
}

func (c *Controller) clearDevicePackageBinding(deviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.binding = SessionBinding{DeviceID: deviceID}
	c.clearBindingViewLocked()
	c.model.Status = "running"
	c.model.SelectedPackage = ""
	c.model.Processes = c.model.Processes[:0]
	c.model.SelectedProcess = ""
	c.model.BoundPIDs = c.model.BoundPIDs[:0]
	c.syncActiveFilterLocked()
}

func effectivePackageScope(scope adb.PackageScope) adb.PackageScope {
	if scope == "" {
		return adb.PackageScopeAll
	}
	return scope
}
