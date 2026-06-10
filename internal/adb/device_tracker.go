package adb

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func (s Service) TrackDevices(ctx context.Context) (<-chan []DeviceInfo, <-chan error, error) {
	if s.pipeRunner == nil {
		return nil, nil, fmt.Errorf("adb_track_devices_unsupported")
	}

	reader, done, err := s.pipeRunner.Start(ctx, s.adbPath, "track-devices", "-l")
	if err != nil {
		return nil, nil, err
	}

	updates := make(chan []DeviceInfo, 8)
	errs := make(chan error, 1)
	go readDeviceUpdates(reader, done, updates, errs)
	return updates, errs, nil
}

func readDeviceUpdates(
	reader io.ReadCloser,
	done <-chan error,
	updates chan<- []DeviceInfo,
	errs chan<- error,
) {
	defer close(updates)
	defer close(errs)
	defer reader.Close()

	buffered := bufio.NewReader(reader)

	for {
		payload, err := readAdbPayload(buffered)
		if err != nil {
			if err == io.EOF || isExpectedCommandExit(err) {
				break
			}
			errs <- err
			return
		}

		devices, err := parseTrackedDevices(payload)
		if err != nil {
			errs <- err
			return
		}
		updates <- devices
	}

	if err := <-done; err != nil && !isExpectedCommandExit(err) {
		errs <- err
	}
}

func readAdbPayload(reader *bufio.Reader) (string, error) {
	if err := skipFrameDelimiters(reader); err != nil {
		return "", err
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return "", err
	}

	size, err := decodeHexSize(header)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", nil
	}

	payload := make([]byte, size)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return "", err
	}
	return string(payload), nil
}

func skipFrameDelimiters(reader *bufio.Reader) error {
	for {
		next, err := reader.Peek(1)
		if err != nil {
			return err
		}
		if next[0] != '\r' && next[0] != '\n' {
			return nil
		}
		if _, err := reader.ReadByte(); err != nil {
			return err
		}
	}
}

func decodeHexSize(header []byte) (int, error) {
	decoded := make([]byte, 2)
	if _, err := hex.Decode(decoded, header); err != nil {
		return 0, fmt.Errorf("invalid adb payload header %q: %w", string(header), err)
	}
	return int(decoded[0])<<8 | int(decoded[1]), nil
}

func parseTrackedDevices(output string) ([]DeviceInfo, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return []DeviceInfo{}, nil
	}

	lines := strings.Split(trimmed, "\n")
	devices := make([]DeviceInfo, 0, len(lines))
	for _, line := range lines {
		device, err := parseDeviceLine(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func isExpectedCommandExit(err error) bool {
	return strings.Contains(err.Error(), "operation was canceled") ||
		strings.Contains(err.Error(), "context canceled") ||
		strings.Contains(err.Error(), "file already closed") ||
		strings.Contains(err.Error(), "The pipe has been ended")
}
