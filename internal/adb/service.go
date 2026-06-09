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
	runner     Runner
	pipeRunner PipeRunner
	adbPath    string
}

func NewService(runner Runner, adbPath string) Service {
	if adbPath == "" {
		adbPath = "adb"
	}

	var pipeRunner PipeRunner
	if candidate, ok := runner.(PipeRunner); ok {
		pipeRunner = candidate
	}

	return Service{
		runner:     runner,
		pipeRunner: pipeRunner,
		adbPath:    adbPath,
	}
}
