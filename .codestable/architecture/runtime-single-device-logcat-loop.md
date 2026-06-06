---
doc_type: architecture
slug: single-device-logcat-loop
scope: 当前 Go + Gio 单设备 H5 logcat 查看、基础控制与 package/pid 绑定链路
summary: Gio 壳通过 controller 驱动 adb 检测、包/进程发现、单会话采集和 PID 绑定重启，只展示 chromium 与 [H5] 日志
status: current
last_reviewed: 2026-06-06
tags: [go, gio, adb, logcat, pid]
depends_on: []
implements: [h5-logcat-viewing]
---

# 单设备 H5 Logcat 查看链路

## 0. 术语

- **设备列表投影**：`adb.Service` 的 `DeviceInfo` 被 `app.Controller` 映射成 UI 使用的 `DeviceItem`，前者保留命令解析细节，后者只保留展示需要的字段。
- **包范围**：加载包列表时使用的 `user` / `system` / `all` 范围，由 `adb.PackageScope` 承载。
- **包列表快照**：一次 `ListPackages` 返回的 `[]PackageInfo`，只保留包名，不夹带 label / icon / version 元数据。
- **进程快照**：一次 `ListProcesses` 返回的 `[]ProcessInfo`，只保留 `PID` 与 `processName`，并按“包名完全相等或 `package:` 前缀”判定相关性。
- **会话绑定**：当前日志会话绑定的 `device + package + process + active PIDs` 组合，由 `app.SessionBinding` 和 `app.Model` 共同表达。
- **PID 重绑**：目标包或进程的活动 PID 集变化后，controller 取消旧会话并用新 PID 集重启会话。
- **可见日志**：`app.Model.VisibleLogs` 中当前列表实际渲染的日志项，供主列表、搜索和详情面板共用。
- **暂停缓冲**：暂停后仍被采集、但暂不进入 `VisibleLogs` 的 `logcat.LogEntry` 队列。
- **当前匹配**：`SearchState.Current` 指向当前高亮的搜索命中，跳转时同步更新 `SelectedIndex`。
- **自动跟随**：`ui.Shell` 基于 `widget.List.Position.BeforeEnd` 判断列表是否仍应贴底。

## 1. 定位与受众

- 这份文档覆盖当前仓库已经落地的单设备运行链路：启动桌面程序、检测本机 `adb`、列出设备、加载包列表、切换前台包/进程绑定、按 PID 收窄 H5 日志，并在当前视图内做暂停/恢复、搜索、复制、详情查看与自动跟随。
- 读者主要是后续 feature 设计、问题分析和新接手代码的人。
- 读完后应该能知道当前系统里包/进程绑定落在哪层、日志何时被 PID 过滤、PID 变化如何自动重绑，以及有哪些不能破坏的硬约束。

## 2. 结构与交互

- `cmd/logcatviewer/main.go` 负责装配依赖，把 `adb.ExecRunner`、`adb.Service`、`adb.LogcatSource`、`session.Supervisor`、`app.Controller` 和 `ui.Shell` 串起来。代码锚点：[main.go](/E:/github/logcat/cmd/logcatviewer/main.go:1)
- `ui.Shell` 在启动时触发 `controller.Load()`，把 `adb version` 与 `adb devices -l` 的结果投影到界面；设备按钮、包范围、前台 App、包列表、进程列表、暂停/恢复、搜索、复制与跟随开关都在这里转成 controller 动作或 Gio 剪贴板命令。代码锚点：[shell.go](/E:/github/logcat/internal/ui/shell.go:1)、[interactions.go](/E:/github/logcat/internal/ui/interactions.go:1)、[package_process_panel.go](/E:/github/logcat/internal/ui/package_process_panel.go:1)
- `app.Controller` 是当前唯一的编排层：它读取设备服务结果、维护 `Model`、在设备选中后启动默认设备级 H5 会话并加载用户包列表，负责前台包/包名/进程选择、绑定状态、PID watcher、暂停缓冲、可见日志、搜索匹配和选中态。代码锚点：[controller.go](/E:/github/logcat/internal/app/controller.go:1)、[package_catalog.go](/E:/github/logcat/internal/app/package_catalog.go:1)、[binding.go](/E:/github/logcat/internal/app/binding.go:1)、[binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1)、[logview.go](/E:/github/logcat/internal/app/logview.go:1)
- `session.Supervisor` 负责把 `adb.LogcatSource` 产生的文本行转成事件流，先走 `logcat.ParseThreadtimeLine`，再走 `logcat.MatchesH5Preset`，最后按 `AllowedPIDs` 做本地 PID 过滤，只把符合 `chromium + [H5] + 当前绑定 PID 集` 的条目推给上层。代码锚点：[supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1)
- `adb.Service` 继续屏蔽 `adb` 命令细节，除 `DetectADB` / `ListDevices` 外，还负责 `ListPackages`、`CurrentForegroundPackage`、`ListProcesses` 三类发现行为。代码锚点：[service.go](/E:/github/logcat/internal/adb/service.go:1)、[device_service.go](/E:/github/logcat/internal/adb/device_service.go:1)、[package_service.go](/E:/github/logcat/internal/adb/package_service.go:1)、[process_service.go](/E:/github/logcat/internal/adb/process_service.go:1)
- `adb.LogcatSource` 仍固定使用 `adb -s <device> logcat -v threadtime -s chromium:I *:S` 启动子进程；这一层不探测 `--pid`，也不在命令层拼 PID 参数。代码锚点：[logcat_source.go](/E:/github/logcat/internal/adb/logcat_source.go:1)

## 3. 数据与状态

- `adb.Install` 表示当前机器上的 `adb` 安装信息，当前字段是 `Path` 与 `Version`。定义位置：[service.go](/E:/github/logcat/internal/adb/service.go:1)
- `adb.DeviceInfo` 表示一次 `adb devices -l` 解析结果，包含 `ID`、`Status`、`Model`、`Transport`。定义位置：[service.go](/E:/github/logcat/internal/adb/service.go:1)
- `adb.PackageScope`、`adb.PackageInfo`、`adb.ProcessInfo` 表示包范围、包列表项和进程快照，是包/进程绑定能力的统一数据契约。定义位置：[service.go](/E:/github/logcat/internal/adb/service.go:1)
- `app.Model` 是当前 UI 的唯一状态容器；除了状态文案、设备列表、可见日志、搜索态、暂停态和选中行外，还新增 `PackageScope`、`Packages`、`SelectedPackage`、`Processes`、`SelectedProcess`、`BoundPIDs`。定义位置：[model.go](/E:/github/logcat/internal/app/model.go:1)
- `app.SessionBinding` 保存当前运行态绑定的 `device / package / process / active PIDs`，是 PID 重绑的比较基准。定义位置：[binding.go](/E:/github/logcat/internal/app/binding.go:1)
- `session.Config` 现在承载 `DeviceID`、`PackageName`、`ProcessName`、`AllowedPIDs`；controller 通过它把“设备级 / 包级 / 进程级”三种绑定模式下推到 `session.Supervisor`。定义位置：[supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1)
- `logcat.LogEntry` 继续承载解析后的 threadtime 行，`Raw` 保留原始输入，解析失败则直接把原始行写进错误消息，不丢现场。定义位置：[parser.go](/E:/github/logcat/internal/logcat/parser.go:1)

## 4. 关键决策

- 包名 / 前台包 / 进程发现统一留在 `internal/adb`，UI 和 controller 不处理 adb 原始文本。代码锚点：[package_service.go](/E:/github/logcat/internal/adb/package_service.go:1)、[process_service.go](/E:/github/logcat/internal/adb/process_service.go:1)
- 默认态仍是设备级 H5 视图：设备选中后先起不带 PID 约束的 H5 会话，再额外加载用户包列表；只有用户显式选择包或进程后才收窄绑定。代码锚点：[SelectDevice](/E:/github/logcat/internal/app/controller.go:72)
- 包/进程切换必须清掉当前可见日志、选中态、搜索命中和暂停缓冲，避免不同绑定上下文混流。代码锚点：[clearBindingViewLocked](/E:/github/logcat/internal/app/binding.go:129)
- PID 过滤先在 `session.Supervisor` 的 H5 预设之后本地执行，不把 `--pid` 探测和命令层优化塞进当前 feature。代码锚点：[allowPID](/E:/github/logcat/internal/session/supervisor.go:84)
- PID 重绑由 controller 自己轮询并编排“停旧会话 + 起新会话”，不把长期 watcher 塞进 `session.Supervisor`。代码锚点：[watchBinding](/E:/github/logcat/internal/app/binding_watcher.go:45)、[applyWatcherRunningBinding](/E:/github/logcat/internal/app/binding_watcher.go:127)
- “前台 App”是显式一次性动作，不做持续自动前台跟随。代码锚点：[SelectForegroundPackage](/E:/github/logcat/internal/app/package_catalog.go:43)

## 5. 代码锚点

- [main.go](/E:/github/logcat/cmd/logcatviewer/main.go:1) `main` — 程序入口和依赖装配
- [controller.go](/E:/github/logcat/internal/app/controller.go:1) `Load` / `SelectDevice` — ADB 检测、设备列表加载与默认设备级会话
- [package_catalog.go](/E:/github/logcat/internal/app/package_catalog.go:1) `SetPackageScope` / `RefreshPackages` / `SelectForegroundPackage` — 包范围与前台包编排
- [binding.go](/E:/github/logcat/internal/app/binding.go:1) `SelectPackage` / `SelectProcess` — 包级 / 进程级绑定编排
- [binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1) `watchBinding` / `refreshBinding` — PID 变化监测与自动重绑
- [supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1) `forward` / `allowPID` — H5 预设后追加 PID 过滤
- [package_service.go](/E:/github/logcat/internal/adb/package_service.go:1) `ListPackages` / `CurrentForegroundPackage` — 包范围与前台包发现
- [process_service.go](/E:/github/logcat/internal/adb/process_service.go:1) `ListProcesses` / `processMatchesPackage` — 相关进程发现
- [package_process_panel.go](/E:/github/logcat/internal/ui/package_process_panel.go:1) `layoutPackageProcessPanel` — 左侧绑定面板渲染
- [interactions.go](/E:/github/logcat/internal/ui/interactions.go:1) `handleScopeClicks` / `handleForegroundClicks` / `handlePackageClicks` / `handleProcessClicks` — 绑定面板动作分发

## 6. 已知约束 / 边界情况

- 当前只支持 `threadtime` 输入格式，其他格式不会静默兼容。来源：[parser.go](/E:/github/logcat/internal/logcat/parser.go:1)
- 当前只允许 `device` 状态设备启动会话；`offline` / `unauthorized` / `no permissions` 都直接进入错误状态。来源：[controller.go](/E:/github/logcat/internal/app/controller.go:1)
- 当前同一时刻只允许一个活跃会话；切换设备、包、进程或 PID 重绑都会取消旧会话。来源：[controller.go](/E:/github/logcat/internal/app/controller.go:1)、[binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1)
- 当前只有 `processName == packageName` 或 `strings.HasPrefix(processName, packageName + \":\")` 的进程会被视为该包相关进程。来源：[process_service.go](/E:/github/logcat/internal/adb/process_service.go:1)
- 当前目标包未运行时会显式显示 `app_not_running`，目标进程消失时会显式显示 `process_not_running`，不会继续沿用旧 PID 集假装成功。来源：[binding.go](/E:/github/logcat/internal/app/binding.go:1)、[binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1)
- 当前不解析 App label、图标、版本号或 APK 元数据。来源：`rg -n "label|icon|versionName|versionCode|APK" cmd internal`
- 当前没有高级查询、导出、工作区、多会话、Wi‑Fi ADB、多窗口或多设备并行能力；也还没有命令层 `--pid` 优化，这些都属于后续 feature。来源：[logcat-viewer-roadmap.md](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-roadmap.md:1)

## 7. 相关文档

- [ARCHITECTURE.md](/E:/github/logcat/.codestable/architecture/ARCHITECTURE.md:1)
- [h5-logcat-viewing.md](/E:/github/logcat/.codestable/requirements/h5-logcat-viewing.md:1)
- [package-pid-and-process-sync-design.md](/E:/github/logcat/.codestable/features/2026-06-06-package-pid-and-process-sync/package-pid-and-process-sync-design.md:1)
- [2026-06-06-progress-08.md](/E:/github/logcat/.codestable/compound/2026-06-06-progress-08.md:1)
