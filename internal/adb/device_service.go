package adb

import (
	"context"
	"fmt"
	"strings"
)

func (s Service) DetectADB(ctx context.Context) (Install, error) {
	output, err := s.runner.Run(ctx, s.adbPath, "version")
	if err != nil {
		return Install{}, err
	}

	return Install{
		Path:    s.adbPath,
		Version: parseVersion(output),
	}, nil
}

func (s Service) ListDevices(ctx context.Context) ([]DeviceInfo, error) {
	output, err := s.runner.Run(ctx, s.adbPath, "devices", "-l")
	if err != nil {
		return nil, err
	}

	return parseDevices(output)
}

func parseVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "Android Debug Bridge version") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			break
		}

		return parts[len(parts)-1]
	}

	return ""
}

func parseDevices(output string) ([]DeviceInfo, error) {
	lines := strings.Split(output, "\n")
	devices := make([]DeviceInfo, 0, len(lines))

	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		device, err := parseDeviceLine(trimmed)
		if err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func parseDeviceLine(line string) (DeviceInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return DeviceInfo{}, fmt.Errorf("invalid device line: %s", line)
	}

	return DeviceInfo{
		ID:        fields[0],
		Status:    fields[1],
		Model:     parseModel(fields[2:]),
		Transport: parseTransport(fields[0]),
	}, nil
}

func parseModel(fields []string) string {
	for _, field := range fields {
		if !strings.HasPrefix(field, "model:") {
			continue
		}

		return strings.TrimPrefix(field, "model:")
	}

	return ""
}

func parseTransport(deviceID string) string {
	if strings.Contains(deviceID, ":") {
		return "wifi"
	}

	return "usb"
}
