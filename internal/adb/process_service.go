package adb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (s Service) ListProcesses(
	ctx context.Context,
	deviceID string,
	packageName string,
) ([]ProcessInfo, error) {
	output, err := s.runner.Run(ctx, s.adbPath, "-s", deviceID, "shell", "ps", "-A")
	if err != nil {
		return nil, err
	}

	return parseProcesses(output, packageName)
}

func parseProcesses(output string, packageName string) ([]ProcessInfo, error) {
	lines := strings.Split(output, "\n")
	processes := make([]ProcessInfo, 0, len(lines))

	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid process line: %s", line)
		}

		name := fields[len(fields)-1]
		if !processMatchesPackage(name, packageName) {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("invalid process pid: %s", fields[1])
		}

		processes = append(processes, ProcessInfo{
			PID:  pid,
			Name: name,
		})
	}

	return processes, nil
}

func processMatchesPackage(processName string, packageName string) bool {
	return processName == packageName ||
		strings.HasPrefix(processName, packageName+":")
}
