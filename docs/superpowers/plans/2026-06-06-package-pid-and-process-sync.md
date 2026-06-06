# Package PID And Process Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前单设备 H5 日志查看链路上加入包名选择、前台 App 选择、进程/PID 绑定和 App 重启自动重绑。

**Architecture:** 继续沿用 `adb -> session -> app -> ui` 四层。`adb` 负责包/前台/进程发现，`session` 在现有 H5 预设之后追加本地 PID 过滤，`app` 负责绑定状态和 PID 重绑编排，`ui` 负责左侧包/进程选择入口。

**Tech Stack:** Go 1.26、Gio、标准库 `context/os/exec/strings/time`

---

### Task 1: ADB 包与进程发现骨架

**Files:**
- Modify: `internal/adb/service.go`
- Create: `internal/adb/device_service.go`
- Create: `internal/adb/package_service.go`
- Create: `internal/adb/process_service.go`
- Modify: `internal/adb/service_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestListPackagesSupportsScope(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell pm list packages -3": "package:com.demo.app\npackage:com.demo.host\n",
		},
	}

	service := NewService(runner, "")
	packages, err := service.ListPackages(context.Background(), "emulator-5554", PackageScopeUser)
	if err != nil {
		t.Fatalf("ListPackages returned error: %v", err)
	}
	if len(packages) != 2 || packages[0].Name != "com.demo.app" {
		t.Fatalf("unexpected packages: %#v", packages)
	}
}
```

```go
func TestCurrentForegroundPackageParsesComponent(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell dumpsys activity activities": "mResumedActivity: ActivityRecord{... com.demo.host/.MainActivity t12}",
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
```

```go
func TestListProcessesReturnsOnlyPackageRelatedRows(t *testing.T) {
	runner := stubRunner{
		outputs: map[string]string{
			"adb -s emulator-5554 shell ps -A": "USER PID NAME\nu0_a1 123 com.demo.host\nu0_a1 456 com.demo.host:webview\nu0_a1 789 com.other.app\n",
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
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/adb -run 'Test(ListPackagesSupportsScope|CurrentForegroundPackageParsesComponent|ListProcessesReturnsOnlyPackageRelatedRows)' -v`  
Expected: FAIL with `undefined: PackageScope` / `undefined: ListPackages`

- [ ] **Step 3: 写最小实现**

```go
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
```

```go
func (s Service) ListPackages(ctx context.Context, deviceID string, scope PackageScope) ([]PackageInfo, error) {
	args := []string{"-s", deviceID, "shell", "pm", "list", "packages"}
	switch scope {
	case PackageScopeUser:
		args = append(args, "-3")
	case PackageScopeSystem:
		args = append(args, "-s")
	}
	output, err := s.runner.Run(ctx, s.adbPath, args...)
	if err != nil {
		return nil, err
	}
	return parsePackages(output), nil
}
```

```go
func (s Service) CurrentForegroundPackage(ctx context.Context, deviceID string) (string, error) {
	output, err := s.runner.Run(ctx, s.adbPath, "-s", deviceID, "shell", "dumpsys", "activity", "activities")
	if err != nil {
		return "", err
	}
	return parseForegroundPackage(output)
}
```

```go
func (s Service) ListProcesses(ctx context.Context, deviceID, packageName string) ([]ProcessInfo, error) {
	output, err := s.runner.Run(ctx, s.adbPath, "-s", deviceID, "shell", "ps", "-A")
	if err != nil {
		return nil, err
	}
	return parseProcesses(output, packageName)
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/adb -run 'Test(ListPackagesSupportsScope|CurrentForegroundPackageParsesComponent|ListProcessesReturnsOnlyPackageRelatedRows)' -v`  
Expected: PASS

### Task 2: 会话绑定与本地 PID 过滤

**Files:**
- Modify: `internal/session/supervisor.go`
- Modify: `internal/session/supervisor_test.go`
- Modify: `internal/app/model.go`
- Modify: `internal/app/controller.go`
- Create: `internal/app/binding.go`
- Modify: `internal/app/controller_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestSupervisorFiltersByAllowedPIDs(t *testing.T) {
	source := stubSource{
		lines: []string{
			`06-04 16:42:18.479 111 111 I chromium: [H5] app one`,
			`06-04 16:42:18.480 222 222 I chromium: [H5] app two`,
		},
	}

	supervisor := NewSupervisor(source)
	handle, err := supervisor.Start(context.Background(), Config{
		DeviceID:    "device-1",
		AllowedPIDs: []int{222},
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	event := <-handle.Events()
	if event.Entry == nil || event.Entry.PID != 222 {
		t.Fatalf("expected pid 222 entry, got %#v", event.Entry)
	}
}
```

```go
func TestControllerSelectPackageReplacesBindingAndClearsVisibleLogs(t *testing.T) {
	// 目标：选中包后，旧可见日志被清掉，新绑定 PID 写入 model
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/session ./internal/app -run 'Test(SupervisorFiltersByAllowedPIDs|ControllerSelectPackageReplacesBindingAndClearsVisibleLogs)' -v`  
Expected: FAIL with `unknown field AllowedPIDs` / `undefined: SelectPackage`

- [ ] **Step 3: 写最小实现**

```go
type Config struct {
	DeviceID    string
	PackageName string
	ProcessName string
	AllowedPIDs []int
}
```

```go
func allowPID(cfg Config, pid int) bool {
	if len(cfg.AllowedPIDs) == 0 {
		return true
	}
	for _, allowed := range cfg.AllowedPIDs {
		if allowed == pid {
			return true
		}
	}
	return false
}
```

```go
type SessionBinding struct {
	PackageName string
	ProcessName string
	PIDs        []int
}
```

```go
func (c *Controller) SelectPackage(ctx context.Context, packageName string) error {
	// 解析相关进程 -> 更新 SelectedPackage / Processes / BoundPIDs
	// 清可见日志 -> 起新会话
	return nil
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/session ./internal/app -run 'Test(SupervisorFiltersByAllowedPIDs|ControllerSelectPackageReplacesBindingAndClearsVisibleLogs)' -v`  
Expected: PASS

### Task 3: UI 包 / 进程选择面板

**Files:**
- Modify: `internal/ui/shell.go`
- Modify: `internal/ui/devices_panel.go`
- Create: `internal/ui/package_process_panel.go`
- Modify: `internal/ui/interactions.go`
- Modify: `internal/ui/shell_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestShellHandleActionsTriggersForegroundSelection(t *testing.T) {
	// 目标：前台 App 按钮点击后会调用 controller 对应动作
}
```

```go
func TestShellSyncPackageButtonsMatchesModelPackages(t *testing.T) {
	// 目标：包列表长度变化后，按钮状态同步
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/ui -run 'TestShell(HandleActionsTriggersForegroundSelection|SyncPackageButtonsMatchesModelPackages)' -v`  
Expected: FAIL with missing package/process button state

- [ ] **Step 3: 写最小实现**

```go
type Shell struct {
	// ...
	foregroundButton widget.Clickable
	scopeUserButton  widget.Clickable
	scopeSystemButton widget.Clickable
	scopeAllButton   widget.Clickable
	packageButtons   []widget.Clickable
	processButtons   []widget.Clickable
}
```

```go
func (s *Shell) layoutPackageProcessPanel(gtx layout.Context, model appstate.Model) layout.Dimensions {
	// 设备区下方追加包范围、前台 App、包列表、进程列表
	return layout.Dimensions{}
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/ui -run 'TestShell(HandleActionsTriggersForegroundSelection|SyncPackageButtonsMatchesModelPackages)' -v`  
Expected: PASS

### Task 4: PID 重绑与全量验证

**Files:**
- Modify: `internal/app/binding.go`
- Modify: `internal/app/controller_test.go`
- Test: `./...`

- [ ] **Step 1: 先写失败测试**

```go
func TestControllerRebindsWhenProcessPIDChanges(t *testing.T) {
	// 目标：同一包 PID 集变化后，controller 更新 BoundPIDs 并重启会话
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/app -run TestControllerRebindsWhenProcessPIDChanges -v`  
Expected: FAIL because no watcher / rebind behavior

- [ ] **Step 3: 写最小实现**

```go
func (c *Controller) watchBinding(ctx context.Context, deviceID, packageName, processName string) {
	// 定时刷新进程列表，比较 PID 集
	// 变化时更新状态并重启会话
}
```

- [ ] **Step 4: 全量验证**

Run: `go test -count=1 ./...`  
Expected: PASS

Run: `go build ./cmd/logcatviewer`  
Expected: build succeeds
