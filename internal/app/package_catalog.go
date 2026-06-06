package app

import (
	"context"

	"github.com/xiakn/logcat/internal/adb"
)

func (c *Controller) SetPackageScope(
	ctx context.Context,
	scope adb.PackageScope,
) error {
	c.mu.Lock()
	c.model.PackageScope = scope
	c.mu.Unlock()

	return c.RefreshPackages(ctx)
}

func (c *Controller) RefreshPackages(ctx context.Context) error {
	deviceID, err := c.currentDeviceID()
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	c.mu.RLock()
	scope := c.model.PackageScope
	c.mu.RUnlock()

	packages, err := c.deviceService.ListPackages(ctx, deviceID, scope)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	c.mu.Lock()
	c.model.Packages = append(c.model.Packages[:0], packages...)
	c.mu.Unlock()
	return nil
}

func (c *Controller) SelectForegroundPackage(ctx context.Context) error {
	deviceID, err := c.currentDeviceID()
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	packageName, err := c.deviceService.CurrentForegroundPackage(ctx, deviceID)
	if err != nil {
		c.updateStatus(err.Error())
		return err
	}

	return c.SelectPackage(ctx, packageName)
}
