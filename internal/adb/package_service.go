package adb

import (
	"context"
	"fmt"
	"strings"
)

func (s Service) ListPackages(
	ctx context.Context,
	deviceID string,
	scope PackageScope,
) ([]PackageInfo, error) {
	args := []string{"-s", deviceID, "shell", "pm", "list", "packages"}
	switch scope {
	case PackageScopeUser:
		args = append(args, "-3")
	case PackageScopeSystem:
		args = append(args, "-s")
	case PackageScopeAll:
	default:
		return nil, fmt.Errorf("invalid_package_scope: %s", scope)
	}

	output, err := s.runner.Run(ctx, s.adbPath, args...)
	if err != nil {
		return nil, err
	}

	return parsePackages(output), nil
}

func (s Service) CurrentForegroundPackage(
	ctx context.Context,
	deviceID string,
) (string, error) {
	output, err := s.runner.Run(
		ctx,
		s.adbPath,
		"-s",
		deviceID,
		"shell",
		"dumpsys",
		"activity",
		"activities",
	)
	if err != nil {
		return "", err
	}

	return parseForegroundPackage(output)
}

func parsePackages(output string) []PackageInfo {
	lines := strings.Split(output, "\n")
	packages := make([]PackageInfo, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "package:") {
			continue
		}

		packages = append(packages, PackageInfo{
			Name: strings.TrimPrefix(trimmed, "package:"),
		})
	}

	return packages
}

func parseForegroundPackage(output string) (string, error) {
	for _, line := range strings.Split(output, "\n") {
		for _, field := range strings.Fields(line) {
			token := strings.Trim(field, "{}")
			if !strings.Contains(token, "/") {
				continue
			}

			pkg, _, _ := strings.Cut(token, "/")
			if strings.Contains(pkg, ".") {
				return pkg, nil
			}
		}
	}

	return "", fmt.Errorf("foreground_package_not_found")
}
