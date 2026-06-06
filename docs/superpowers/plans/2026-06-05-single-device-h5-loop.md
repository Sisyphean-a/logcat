# Single Device H5 Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用 Go + Gio 建立单设备 H5 logcat 查看最小闭环，完成 ADB 检测、设备选择、实时日志采集和 `chromium + [H5]` 预设过滤。

**Architecture:** 代码分成 `adb`、`logcat`、`session`、`app`、`ui` 五层。`adb` 负责命令与解析，`logcat` 负责 threadtime 解析和预设过滤，`session` 负责会话与事件流，`app` 负责 ViewModel 和交互编排，`ui` 负责 Gio 界面。

**Tech Stack:** Go 1.26、Gio、标准库 `context/os/exec/time/bufio`

---

### Task 1: 项目骨架与 ViewModel

**Files:**
- Create: `go.mod`
- Create: `cmd/logcatviewer/main.go`
- Create: `internal/app/model.go`
- Create: `internal/app/model_test.go`
- Create: `internal/ui/theme.go`

- [ ] **Step 1: Write the failing test**

```go
package app

import "testing"

func TestNewModelStartsIdle(t *testing.T) {
	model := NewModel()

	if model.Status != "idle" {
		t.Fatalf("expected idle status, got %q", model.Status)
	}
	if len(model.Devices) != 0 {
		t.Fatalf("expected empty devices, got %d", len(model.Devices))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app -run TestNewModelStartsIdle -v`  
Expected: FAIL with `undefined: NewModel`

- [ ] **Step 3: Write minimal implementation**

```go
package app

type DeviceItem struct {
	ID     string
	Model  string
	Status string
}

type Model struct {
	Status  string
	Devices []DeviceItem
	Logs    []string
}

func NewModel() Model {
	return Model{
		Status:  "idle",
		Devices: []DeviceItem{},
		Logs:    []string{},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/app -run TestNewModelStartsIdle -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod cmd/logcatviewer/main.go internal/app/model.go internal/app/model_test.go internal/ui/theme.go
git commit -m "feat: scaffold gio app model"
```

### Task 2: ADB 检测与设备列表解析

**Files:**
- Create: `internal/adb/runner.go`
- Create: `internal/adb/locator.go`
- Create: `internal/adb/devices.go`
- Create: `internal/adb/locator_test.go`
- Create: `internal/adb/devices_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package adb

import (
	"context"
	"testing"
)

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
}
```

```go
func TestParseDevicesListParsesTransportAndStatus(t *testing.T) {
	raw := "List of devices attached\nemulator-5554 device product:sdk model:Pixel_7 transport_id:1\n192.168.0.10:5555 unauthorized model:SM_A217F transport_id:2\n"

	devices, err := ParseDevices(raw)
	if err != nil {
		t.Fatalf("ParseDevices returned error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adb -v`  
Expected: FAIL with `undefined: NewService` / `undefined: ParseDevices`

- [ ] **Step 3: Write minimal implementation**

```go
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type Service struct {
	runner  Runner
	adbPath string
}

func NewService(runner Runner, adbPath string) Service {
	if adbPath == "" {
		adbPath = "adb"
	}
	return Service{runner: runner, adbPath: adbPath}
}
```

```go
func ParseDevices(raw string) ([]DeviceInfo, error) {
	lines := strings.Split(raw, "\n")
	devices := make([]DeviceInfo, 0, len(lines))
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		devices = append(devices, DeviceInfo{
			ID:     fields[0],
			Status: fields[1],
		})
	}
	return devices, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adb -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adb/runner.go internal/adb/locator.go internal/adb/devices.go internal/adb/locator_test.go internal/adb/devices_test.go
git commit -m "feat: add adb detection and device parsing"
```

### Task 3: threadtime 解析与 H5 预设过滤

**Files:**
- Create: `internal/logcat/parser.go`
- Create: `internal/logcat/preset.go`
- Create: `internal/logcat/parser_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package logcat

import "testing"

func TestParseThreadtimeLineParsesChromiumConsole(t *testing.T) {
	line := `06-04 16:42:18.479 10665 10665 I chromium: [INFO:CONSOLE(618)] "[H5] connected", source: http://127.0.0.1/app.js (618)`

	entry, err := ParseThreadtimeLine("device-1", line)
	if err != nil {
		t.Fatalf("ParseThreadtimeLine returned error: %v", err)
	}
	if entry.Tag != "chromium" {
		t.Fatalf("expected chromium tag, got %q", entry.Tag)
	}
	if !MatchesH5Preset(entry) {
		t.Fatalf("expected entry to match H5 preset")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/logcat -run TestParseThreadtimeLineParsesChromiumConsole -v`  
Expected: FAIL with `undefined: ParseThreadtimeLine`

- [ ] **Step 3: Write minimal implementation**

```go
func MatchesH5Preset(entry Entry) bool {
	return entry.Tag == "chromium" && strings.Contains(entry.Message, "[H5]")
}
```

```go
var threadtimePattern = regexp.MustCompile(`^(\d\d-\d\d \d\d:\d\d:\d\d\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEF])\s+([^:]+):\s(.*)$`)

func ParseThreadtimeLine(deviceID, line string) (Entry, error) {
	match := threadtimePattern.FindStringSubmatch(line)
	if match == nil {
		return Entry{}, fmt.Errorf("invalid threadtime line")
	}
	return Entry{
		DeviceID: deviceID,
		TimeText: match[1],
		PID:      mustAtoi(match[2]),
		TID:      mustAtoi(match[3]),
		Level:    match[4],
		Tag:      strings.TrimSpace(match[5]),
		Message:  match[6],
		Raw:      line,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/logcat -run TestParseThreadtimeLineParsesChromiumConsole -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/logcat/parser.go internal/logcat/preset.go internal/logcat/parser_test.go
git commit -m "feat: add threadtime parsing and h5 preset"
```

### Task 4: 会话监督器与 Gio 集成

**Files:**
- Create: `internal/session/supervisor.go`
- Create: `internal/session/supervisor_test.go`
- Create: `internal/app/controller.go`
- Create: `internal/ui/shell.go`
- Modify: `cmd/logcatviewer/main.go`

- [ ] **Step 1: Write the failing test**

```go
package session

import (
	"context"
	"testing"
)

func TestSupervisorStreamsMatchingEntries(t *testing.T) {
	source := stubSource{
		lines: []string{
			`06-04 16:42:18.479 10665 10665 I chromium: [H5] ok`,
			`06-04 16:42:18.480 10665 10665 I chromium: ignored`,
		},
	}

	supervisor := NewSupervisor(source)
	handle, err := supervisor.Start(context.Background(), Config{DeviceID: "d1"})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	event := <-handle.Events()
	if event.Entry == nil || event.Entry.Message != "[H5] ok" {
		t.Fatalf("expected first matching H5 entry, got %#v", event.Entry)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session -run TestSupervisorStreamsMatchingEntries -v`  
Expected: FAIL with `undefined: NewSupervisor`

- [ ] **Step 3: Write minimal implementation**

```go
type Source interface {
	Stream(ctx context.Context, cfg Config) (<-chan string, <-chan error)
}

type Supervisor struct {
	source Source
}

func NewSupervisor(source Source) Supervisor {
	return Supervisor{source: source}
}
```

```go
func (s Supervisor) Start(ctx context.Context, cfg Config) (Handle, error) {
	events := make(chan Event, 16)
	lines, errs := s.source.Stream(ctx, cfg)
	go func() {
		defer close(events)
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					return
				}
				entry, err := logcat.ParseThreadtimeLine(cfg.DeviceID, line)
				if err == nil && logcat.MatchesH5Preset(entry) {
					events <- Event{Entry: &entry}
				}
			case err := <-errs:
				if err != nil {
					events <- Event{Problem: err}
				}
				return
			}
		}
	}()
	return Handle{events: events}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/session -run TestSupervisorStreamsMatchingEntries -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/supervisor.go internal/session/supervisor_test.go internal/app/controller.go internal/ui/shell.go cmd/logcatviewer/main.go
git commit -m "feat: wire session supervisor into gio shell"
```

### Task 5: 全量验证

**Files:**
- Test: `./...`

- [ ] **Step 1: Run full test suite**

Run: `go test ./...`  
Expected: PASS

- [ ] **Step 2: Run formatter**

Run: `gofmt -w cmd internal`  
Expected: no output

- [ ] **Step 3: Re-run tests after formatting**

Run: `go test ./...`  
Expected: PASS

- [ ] **Step 4: Smoke-run desktop app**

Run: `go run ./cmd/logcatviewer`  
Expected: Gio window opens and shows dark shell with status area

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: deliver single device h5 loop"
```
