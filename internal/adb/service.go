package adb

import "context"

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type Install struct {
	Path    string
	Version string
}

type DeviceInfo struct {
	ID        string
	Status    string
	Model     string
	Transport string
}

type PackageScope string

const (
	PackageScopeUser   PackageScope = "user"
	PackageScopeSystem PackageScope = "system"
	PackageScopeAll    PackageScope = "all"
)

type PackageInfo struct {
	Name string
}

type ProcessInfo struct {
	PID  int
	Name string
}

type Service struct {
	runner  Runner
	adbPath string
}

func NewService(runner Runner, adbPath string) Service {
	if adbPath == "" {
		adbPath = "adb"
	}

	return Service{
		runner:  runner,
		adbPath: adbPath,
	}
}
