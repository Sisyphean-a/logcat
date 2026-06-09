package adb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
)

type stubRunner struct {
	outputs map[string]string
	errs    map[string]error
}

func (s stubRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	if err, ok := s.errs[key]; ok {
		return "", err
	}
	if out, ok := s.outputs[key]; ok {
		return out, nil
	}
	return "", fmt.Errorf("unexpected command: %s", key)
}

func (s stubRunner) Start(_ context.Context, name string, args ...string) (io.ReadCloser, <-chan error, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	if err, ok := s.errs[key]; ok {
		return nil, nil, err
	}
	if out, ok := s.outputs[key]; ok {
		done := make(chan error, 1)
		done <- nil
		close(done)
		return io.NopCloser(bytes.NewBufferString(out)), done, nil
	}
	return nil, nil, fmt.Errorf("unexpected command: %s", key)
}

func TestDetectADBReturnsInstall(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb version": "Android Debug Bridge version 1.0.41\nVersion 36.0.0-13206524\n",
		},
	}

	service := NewService(runner, "")
	install, err := service.DetectADB(context.Background())
	if err != nil {
		t.Fatalf("DetectADB returned error: %v", err)
	}
	if install.Path != "adb" {
		t.Fatalf("expected adb path, got %q", install.Path)
	}
	if install.Version != "1.0.41" {
		t.Fatalf("expected version 1.0.41, got %q", install.Version)
	}
}

func TestListDevicesParsesTransportAndStatus(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb devices -l": "List of devices attached\nemulator-5554 device product:sdk model:Pixel_7 transport_id:1\n192.168.0.10:5555 unauthorized model:SM_A217F transport_id:2\n",
		},
	}

	service := NewService(runner, "")
	devices, err := service.ListDevices(context.Background())
	if err != nil {
		t.Fatalf("ListDevices returned error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Transport != "usb" {
		t.Fatalf("expected usb transport, got %q", devices[0].Transport)
	}
	if devices[1].Status != "unauthorized" {
		t.Fatalf("expected unauthorized status, got %q", devices[1].Status)
	}
}

func TestListPackagesSupportsScope(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell pm list packages -3": "package:com.demo.app\npackage:com.demo.host\n",
			"adb -s emulator-5554 shell pm list packages -s": "package:android\npackage:com.android.systemui\n",
			"adb -s emulator-5554 shell pm list packages":    "package:android\npackage:com.demo.host\n",
		},
	}

	service := NewService(runner, "")

	userPackages, err := service.ListPackages(context.Background(), "emulator-5554", PackageScopeUser)
	if err != nil {
		t.Fatalf("ListPackages(user) returned error: %v", err)
	}
	if len(userPackages) != 2 || userPackages[0].Name != "com.demo.app" {
		t.Fatalf("unexpected user packages: %#v", userPackages)
	}

	systemPackages, err := service.ListPackages(context.Background(), "emulator-5554", PackageScopeSystem)
	if err != nil {
		t.Fatalf("ListPackages(system) returned error: %v", err)
	}
	if len(systemPackages) != 2 || systemPackages[1].Name != "com.android.systemui" {
		t.Fatalf("unexpected system packages: %#v", systemPackages)
	}

	allPackages, err := service.ListPackages(context.Background(), "emulator-5554", PackageScopeAll)
	if err != nil {
		t.Fatalf("ListPackages(all) returned error: %v", err)
	}
	if len(allPackages) != 2 || allPackages[0].Name != "android" {
		t.Fatalf("unexpected all packages: %#v", allPackages)
	}
}

func TestCurrentForegroundPackageParsesComponent(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell dumpsys activity activities": "mResumedActivity: ActivityRecord{7d1 u0 com.demo.host/.MainActivity t12}",
		},
	}

	service := NewService(runner, "")
	pkg, err := service.CurrentForegroundPackage(context.Background(), "emulator-5554")
	if err != nil {
		t.Fatalf("CurrentForegroundPackage returned error: %v", err)
	}
	if pkg != "com.demo.host" {
		t.Fatalf("expected com.demo.host, got %q", pkg)
	}
}

func TestListProcessesReturnsOnlyPackageRelatedRows(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell ps -A": "USER PID PPID VSZ RSS WCHAN ADDR S NAME\nu0_a1 123 1 0 0 0 0 S com.demo.host\nu0_a1 456 1 0 0 0 0 S com.demo.host:webview\nu0_a1 789 1 0 0 0 0 S com.other.app\n",
		},
	}

	service := NewService(runner, "")
	processes, err := service.ListProcesses(context.Background(), "emulator-5554", "com.demo.host")
	if err != nil {
		t.Fatalf("ListProcesses returned error: %v", err)
	}
	if len(processes) != 2 {
		t.Fatalf("expected 2 related processes, got %d", len(processes))
	}
	if processes[0].PID != 123 || processes[1].Name != "com.demo.host:webview" {
		t.Fatalf("unexpected related processes: %#v", processes)
	}
}

func TestTrackDevicesParsesLengthPrefixedSnapshots(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb track-devices -l": "0032emulator-5554 device model:Pixel_7 transport_id:1\n",
		},
	}

	service := NewService(runner, "")
	updates, errs, err := service.TrackDevices(context.Background())
	if err != nil {
		t.Fatalf("TrackDevices returned error: %v", err)
	}

	devices := <-updates
	if len(devices) != 1 {
		t.Fatalf("expected 1 tracked device, got %d", len(devices))
	}
	if devices[0].ID != "emulator-5554" || devices[0].Status != "device" {
		t.Fatalf("unexpected tracked device: %#v", devices[0])
	}
	if err := <-errs; err != nil {
		t.Fatalf("expected no tracking error, got %v", err)
	}
}
