package app

import (
	"context"
	"log"
	"time"

	"github.com/xiakn/logcat/internal/adb"
)

const trackedDeviceReconcileDelay = 400 * time.Millisecond

func (c *Controller) TrackDevices(ctx context.Context) error {
	tracker, ok := c.deviceService.(DeviceTracker)
	if !ok {
		return nil
	}

	updates, errs, err := tracker.TrackDevices(ctx)
	if err != nil {
		return err
	}

	go c.consumeDeviceUpdates(ctx, updates, errs)
	return nil
}

func (c *Controller) consumeDeviceUpdates(
	ctx context.Context,
	updates <-chan []adb.DeviceInfo,
	errs <-chan error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case devices, ok := <-updates:
			if !ok {
				return
			}
			log.Printf("consumeDeviceUpdates devices=%v", devices)
			if err := c.syncDevices(context.Background(), devices); err != nil {
				c.updateStatus(err.Error())
			}
			go c.reconcileTrackedDevices()
		case err, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			if err != nil {
				c.updateStatus(err.Error())
			}
		}
	}
}

func (c *Controller) reconcileTrackedDevices() {
	timer := time.NewTimer(c.deviceReconcileDelay)
	defer timer.Stop()

	<-timer.C

	devices, err := c.deviceService.ListDevices(context.Background())
	if err != nil {
		c.updateStatus(err.Error())
		return
	}

	log.Printf("reconcileTrackedDevices devices=%v", devices)
	if err := c.syncDevices(context.Background(), devices); err != nil {
		c.updateStatus(err.Error())
	}
}

func (c *Controller) syncDevices(ctx context.Context, devices []adb.DeviceInfo) error {
	selection, previousDevice, changed := c.applyDeviceSnapshot(devices)
	if !changed {
		return nil
	}
	if selection == "" {
		if previousDevice != "" {
			c.clearUnavailableDeviceSelection()
		}
		return nil
	}
	if selection != previousDevice {
		return c.SelectDevice(ctx, selection)
	}
	return nil
}

func (c *Controller) applyDeviceSnapshot(devices []adb.DeviceInfo) (string, string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	previousDevice := c.model.SelectedDevice
	previousDevices := c.model.Devices
	c.model.Devices = mapDevices(devices)
	selection := resolveSelectedDevice(previousDevice, previousDevices, c.model.Devices)
	if selection == previousDevice && sameDeviceItems(previousDevices, c.model.Devices) {
		return selection, previousDevice, false
	}

	c.model.SelectedDevice = selection
	c.markDirtyLocked()
	return selection, previousDevice, true
}

func resolveSelectedDevice(
	previousSelected string,
	previousDevices []DeviceItem,
	currentDevices []DeviceItem,
) string {
	if keepCurrentSelection(previousSelected, currentDevices) {
		return previousSelected
	}
	if previousSelected == "" && len(previousDevices) == 0 {
		return firstReadyDevice(currentDevices)
	}
	return firstReadyDevice(currentDevices)
}

func keepCurrentSelection(deviceID string, devices []DeviceItem) bool {
	for _, device := range devices {
		if device.ID == deviceID {
			return device.Status == "device"
		}
	}
	return false
}

func firstReadyDevice(devices []DeviceItem) string {
	for _, device := range devices {
		if device.Status == "device" {
			return device.ID
		}
	}
	return ""
}

func sameDeviceItems(left []DeviceItem, right []DeviceItem) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func (c *Controller) clearUnavailableDeviceSelection() {
	c.stopWatcher()
	c.stopSession()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.binding = SessionBinding{}
	c.clearBindingViewLocked()
	c.model.SelectedDevice = ""
	c.model.PackageScope = adb.PackageScopeAll
	c.model.Packages = c.model.Packages[:0]
	c.model.SelectedPackage = ""
	c.model.Processes = c.model.Processes[:0]
	c.model.SelectedProcess = ""
	c.model.BoundPIDs = c.model.BoundPIDs[:0]
	c.model.Pause.Active = true
	c.model.Status = "idle"
	c.syncActiveFilterLocked()
	c.markDirtyLocked()
}
