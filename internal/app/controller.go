package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xiakn/logcat/internal/adb"
	"github.com/xiakn/logcat/internal/logcat"
	"github.com/xiakn/logcat/internal/session"
)

type DeviceService interface {
	DetectADB(ctx context.Context) (adb.Install, error)
	ListDevices(ctx context.Context) ([]adb.DeviceInfo, error)
	ListPackages(ctx context.Context, deviceID string, scope adb.PackageScope) ([]adb.PackageInfo, error)
	CurrentForegroundPackage(ctx context.Context, deviceID string) (string, error)
	ListProcesses(ctx context.Context, deviceID, packageName string) ([]adb.ProcessInfo, error)
}

type SessionStarter interface {
	Start(ctx context.Context, cfg session.Config) (session.Handle, error)
}

type Controller struct {
	deviceService DeviceService
	sessionStart  SessionStarter

	mu                  sync.RWMutex
	model               Model
	allLogs             []LogViewItem
	revision            uint64
	sessionCancel       context.CancelFunc
	watchCancel         context.CancelFunc
	pauseBuffer         []logcat.LogEntry
	pauseBufferCap      int
	bindingPollInterval time.Duration
	binding             SessionBinding
}

const defaultBindingPollInterval = 500 * time.Millisecond

func NewController(deviceService DeviceService, sessionStart SessionStarter) *Controller {
		return &Controller{
			deviceService:       deviceService,
			sessionStart:        sessionStart,
			model:               NewModel(),
			allLogs:             []LogViewItem{},
			revision:            1,
			pauseBuffer:         []logcat.LogEntry{},
			pauseBufferCap:      defaultPauseBufferCap,
			bindingPollInterval: defaultBindingPollInterval,
			binding:             SessionBinding{},
		}
}

func (c *Controller) Model() Model {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return cloneModel(c.model)
}

func (c *Controller) SetStatus(status string) {
	c.updateStatus(status)
}

func (c *Controller) Load(ctx context.Context) error {
	install, err := c.deviceService.DetectADB(ctx)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	devices, err := c.deviceService.ListDevices(ctx)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	c.mu.Lock()
	c.model.Status = "adb " + install.Version
	c.model.ADBStatus = "已连接"
	c.model.Devices = mapDevices(devices)
	if len(c.model.Devices) > 0 {
		c.model.SelectedDevice = c.model.Devices[0].ID
	}
	c.markDirtyLocked()
	c.mu.Unlock()

	return nil
}

func (c *Controller) SelectDevice(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		c.clearDeviceSelection()
		return nil
	}

	device, err := c.findDevice(deviceID)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	if device.Status != "device" {
		err = deviceStatusError(device)
		c.updateStatus(err.Error())
		return err
	}

	c.stopWatcher()
	if c.hasActiveSession() {
		if err := c.startSession(ctx, session.Config{DeviceID: deviceID}); err != nil {
			return err
		}
	}

	c.mu.Lock()
	c.binding = SessionBinding{DeviceID: deviceID}
	c.model.PackageScope = adb.PackageScopeAll
	c.model.SelectedDevice = deviceID
	c.model.Packages = c.model.Packages[:0]
	c.model.SelectedPackage = ""
	c.model.Processes = c.model.Processes[:0]
	c.model.SelectedProcess = ""
	c.model.BoundPIDs = c.model.BoundPIDs[:0]
	c.syncActiveFilterLocked()
	c.markDirtyLocked()
	c.mu.Unlock()

	return c.RefreshPackages(ctx)
}

func (c *Controller) startSession(ctx context.Context, cfg session.Config) error {
	sessionCtx, cancel := context.WithCancel(ctx)
	c.swapSession(cancel)

	handle, err := c.sessionStart.Start(sessionCtx, cfg)
	if err != nil {
		cancel()
		c.updateStatus(err.Error())
		return err
	}

	c.updateStatus("running")
	go c.consume(handle)
	return nil
}

func (c *Controller) swapSession(cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionCancel != nil {
		c.sessionCancel()
	}
	c.sessionCancel = cancel
}

func (c *Controller) consume(handle session.Handle) {
	for event := range handle.Events() {
		if event.Entry != nil {
			c.pushEntry(*event.Entry)
		}
		if event.Problem != nil {
			c.updateStatus(event.Problem.Error())
		}
	}
}

func (c *Controller) updateStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.Status = status
	c.markDirtyLocked()
}

func mapDevices(devices []adb.DeviceInfo) []DeviceItem {
	items := make([]DeviceItem, 0, len(devices))
	for _, device := range devices {
		items = append(items, DeviceItem{
			ID:     device.ID,
			Model:  device.Model,
			Status: device.Status,
		})
	}

	return items
}

func (c *Controller) findDevice(deviceID string) (DeviceItem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, device := range c.model.Devices {
		if device.ID == deviceID {
			return device, nil
		}
	}

	return DeviceItem{}, fmt.Errorf("device_not_found: %s", deviceID)
}

func deviceStatusError(device DeviceItem) error {
	switch device.Status {
	case "unauthorized":
		return fmt.Errorf("device_unauthorized: %s", device.ID)
	case "offline":
		return fmt.Errorf("device_offline: %s", device.ID)
	case "no permissions":
		return fmt.Errorf("device_no_permission: %s", device.ID)
	default:
		return fmt.Errorf("device_unavailable: %s (%s)", device.ID, device.Status)
	}
}
