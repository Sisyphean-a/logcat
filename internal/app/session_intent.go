package app

type sessionIntent uint8

const (
	sessionIntentNone sessionIntent = iota
	sessionIntentPaused
	sessionIntentRunning
)

func (c *Controller) currentSessionIntent() sessionIntent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.sessionCancel != nil {
		if c.model.Pause.Active {
			return sessionIntentPaused
		}
		return sessionIntentRunning
	}
	if c.resumeStreaming {
		return sessionIntentRunning
	}
	return sessionIntentNone
}

func cloneSessionBinding(binding SessionBinding) SessionBinding {
	return SessionBinding{
		DeviceID:    binding.DeviceID,
		PackageName: binding.PackageName,
		ProcessName: binding.ProcessName,
		PIDs:        append([]int(nil), binding.PIDs...),
	}
}

func (c *Controller) rememberBindingLocked(binding SessionBinding) {
	c.resumeBinding = cloneSessionBinding(binding)
}

func (c *Controller) rememberCurrentBindingLocked() {
	binding := cloneSessionBinding(c.binding)
	if binding.DeviceID == "" {
		binding.DeviceID = c.model.SelectedDevice
	}
	c.resumeBinding = binding
}

func (c *Controller) pendingBindingForDevice(deviceID string) SessionBinding {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.resumeBinding.DeviceID != deviceID {
		return SessionBinding{}
	}
	return cloneSessionBinding(c.resumeBinding)
}
