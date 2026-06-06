# package-pid-and-process-sync 验收报告

> 阶段：阶段 3（验收闭环）
> 验收日期：2026-06-06
> 关联方案 doc：[package-pid-and-process-sync-design.md](/E:/github/logcat/.codestable/features/2026-06-06-package-pid-and-process-sync/package-pid-and-process-sync-design.md:1)

## 1. 接口契约核对

- [x] `PackageScope`、`PackageInfo`、`ProcessInfo` 已按方案第 2.1 节落地到 [service.go](/E:/github/logcat/internal/adb/service.go:1)，并由 [package_service.go](/E:/github/logcat/internal/adb/package_service.go:1) 与 [process_service.go](/E:/github/logcat/internal/adb/process_service.go:1) 实际驱动。
- [x] `SessionBinding` 已按方案第 2.1 节落地到 [binding.go](/E:/github/logcat/internal/app/binding.go:1)，包含 `DeviceID`、`PackageName`、`ProcessName`、`PIDs`。
- [x] `session.Config` 已扩展出 `PackageName`、`ProcessName`、`AllowedPIDs`，与方案第 2.1 节接口示例一致，见 [supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1)。
- [x] `app.Model` 已扩展出 `PackageScope`、`Packages`、`SelectedPackage`、`Processes`、`SelectedProcess`、`BoundPIDs`，与方案第 2.1 节一致，见 [model.go](/E:/github/logcat/internal/app/model.go:1)。
- [x] controller 操作面 `SetPackageScope`、`RefreshPackages`、`SelectForegroundPackage`、`SelectPackage`、`SelectProcess` 已全部落地，见 [package_catalog.go](/E:/github/logcat/internal/app/package_catalog.go:1) 和 [binding.go](/E:/github/logcat/internal/app/binding.go:1)。
- [x] 主流程图核对：`点击设备 -> 默认 H5 会话 + 加载 user 包列表 -> 选择前台 App / 包 / 进程 -> SessionConfig 带 AllowedPIDs -> Supervisor 本地 PID 过滤 -> watcher 比较 PID 集并重绑` 在 [controller.go](/E:/github/logcat/internal/app/controller.go:1)、[package_catalog.go](/E:/github/logcat/internal/app/package_catalog.go:1)、[binding.go](/E:/github/logcat/internal/app/binding.go:1)、[binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1)、[supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1) 均有实际代码落点。

## 2. 行为与决策核对

- [x] 设备选中后仍保留设备级 H5 视图，同时加载 `user` 包列表：`SelectDevice` 先 `startSession`，再重置绑定态并调用 `RefreshPackages`，见 [controller.go](/E:/github/logcat/internal/app/controller.go:72)。
- [x] 包名 / 前台包 / 进程发现继续放在 `internal/adb`：controller 只消费 `ListPackages` / `CurrentForegroundPackage` / `ListProcesses`，不处理 adb 原始文本，见 [controller.go](/E:/github/logcat/internal/app/controller.go:1) 和 [package_service.go](/E:/github/logcat/internal/adb/package_service.go:1)。
- [x] PID 过滤先在 session/app 层本地执行：`adb.LogcatSource` 仍只起 `chromium:I` 命令，`session.Supervisor` 在 `MatchesH5Preset` 之后追加 `allowPID`，见 [logcat_source.go](/E:/github/logcat/internal/adb/logcat_source.go:1) 和 [supervisor.go](/E:/github/logcat/internal/session/supervisor.go:1)。
- [x] PID 重绑由 controller 维护：`watchBinding -> refreshBinding -> applyWatcherRunningBinding` 负责比较 PID 集、取消旧会话、启动新会话并更新状态，见 [binding_watcher.go](/E:/github/logcat/internal/app/binding_watcher.go:1)。
- [x] UI 继续沿用按钮列表模式：绑定面板通过 scope 按钮、前台 App 按钮、包按钮和进程按钮驱动 controller，没有引入自定义下拉组件，见 [package_process_panel.go](/E:/github/logcat/internal/ui/package_process_panel.go:1)。
- [x] 前台 App 选择是显式动作，不做持续自动切换：代码里只有 [SelectForegroundPackage](/E:/github/logcat/internal/app/package_catalog.go:43) 的一次性查询，没有持续前台轮询；`rg -n "CurrentForeground|foreground.*Ticker|foreground.*watch"` 在 `cmd internal` 无持续跟随实现。
- [x] 流程级约束“切换包或进程必须清掉当前可见日志和暂停缓冲”已落实：`activateRunningBinding` / `activateStoppedBinding` 都调用 [clearBindingViewLocked](/E:/github/logcat/internal/app/binding.go:129)。
- [x] 流程级约束“目标包未运行时显式提示且不沿用旧绑定”已落实：`activateStoppedBinding` 会 `stopSession`、清空 `BoundPIDs` 并上浮 `app_not_running`，见 [binding.go](/E:/github/logcat/internal/app/binding.go:106)。
- [x] 挂载点反向核对：本 feature 的实际挂入点只落在 `internal/adb/*service*.go`、`internal/session/supervisor.go`、`internal/app/{controller,model,package_catalog,binding,binding_watcher}.go`、`internal/ui/{shell,devices_panel,interactions,package_process_panel}.go` 与现有入口 [main.go](/E:/github/logcat/cmd/logcatviewer/main.go:1)；`rg -n "PackageScope|SessionBinding|AllowedPIDs|SelectForegroundPackage|watchBinding|packageButtons|processButtons|foregroundButton" cmd internal` 未发现清单外模块残留。
- [x] 拔除沙盘推演：去掉 `internal/adb` 的包/进程发现后，绑定面板没有数据源；去掉 `binding_watcher.go` 后 PID 变化不会自动重绑；去掉 `package_process_panel.go` / `interactions.go` 的挂入后，前台 App、包和进程入口全部消失。挂载点清单完整。
- [x] 范围外事项核对：`rg -n "label|icon|versionName|versionCode|--pid|multi-session|multi session|workspace|SavedFilter|QueryHistory|APK" cmd internal` 无命中，说明当前没有抢跑 App 元数据、高级查询、工作区、多会话或命令层 `--pid` 优化。

## 3. 验收场景核对

- [x] **S1 设备选中后能加载用户 App 列表，并默认保留设备级 H5 会话**
  - 证据来源：`TestControllerSelectDeviceLoadsUserPackages`
  - 结果：通过
- [x] **S2 一键选择前台 App 后，包和进程上下文会同步切换**
  - 证据来源：`TestControllerSelectForegroundPackageUsesForegroundPackage`、`TestShellHandleActionsTriggersForegroundSelection`
  - 结果：通过
- [x] **S3 选中未运行 App 时显式提示 app_not_running，不继续沿用旧绑定**
  - 证据来源：`TestControllerSelectPackageNotRunningStopsOldBinding`
  - 结果：通过
- [x] **S4 选中包或进程后，只显示匹配 PID 的 H5 日志**
  - 证据来源：`TestSupervisorFiltersByAllowedPIDs`、`TestControllerSelectProcessNarrowsBindingToSinglePID`、`TestShellHandleActionsSelectsPackageAndProcess`
  - 结果：通过
- [x] **S5 PID 变化后会自动重绑并给出可见状态**
  - 证据来源：`TestControllerRebindsWhenProcessPIDChanges`
  - 结果：通过
- [x] **S6 切换设备后会清空包/进程/BoundPIDs 等绑定状态**
  - 证据来源：`TestControllerSelectDeviceClearsBindingState`
  - 结果：通过
- [x] **UI 运行态**
  - 证据来源：`go build ./cmd/logcatviewer`、启动后窗口句柄检查、窗口截图 [ui-window-acceptance-2026-06-06.png](/E:/github/logcat/.artifacts/ui-window-acceptance-2026-06-06.png:1)、[ui-window-after-device-2026-06-06.png](/E:/github/logcat/.artifacts/ui-window-after-device-2026-06-06.png:1)
  - 结果：通过；Gio 窗口可正常拉起，左侧 `Binding` 区和包范围按钮已真实渲染

## 4. 术语一致性

- [x] `PackageScope`、`PackageInfo`、`ProcessInfo`、`SessionBinding`、`AllowedPIDs`、`BoundPIDs` 在设计和代码里的语义一致。
- [x] 包 / 进程 / PID 相关状态统一挂在 `app.Model` 和 `app.SessionBinding`，没有再发明第二套命名。
- [x] 范围外词汇检索：`rg -n "label|icon|versionName|versionCode|CurrentForeground.*loop|--pid|multi-session|workspace|SavedFilter|QueryHistory|APK" cmd internal` 无命中。

## 5. 架构归并

- [x] 已更新 [runtime-single-device-logcat-loop.md](/E:/github/logcat/.codestable/architecture/runtime-single-device-logcat-loop.md:1)，把 `PackageScope`、`PackageInfo`、`ProcessInfo`、`SessionBinding`、本地 PID 过滤、绑定面板和 PID 重绑这些现状写回架构层。
- [x] 已更新 [ARCHITECTURE.md](/E:/github/logcat/.codestable/architecture/ARCHITECTURE.md:1)，把总入口从“基础筛查闭环”升级为“基础筛查 + 包/PID 绑定闭环”，并同步术语、关键决策和硬边界。

## 6. requirement 回写

- [x] 本 feature 沿用现有 requirement `h5-logcat-viewing`，因为它扩展的是同一条“查看并筛查 H5 日志”的用户能力，而不是新开第二条能力线。
- [x] 已更新 [h5-logcat-viewing.md](/E:/github/logcat/.codestable/requirements/h5-logcat-viewing.md:1)，把用户故事、能力说明、边界和变更日志刷新到当前实现。
- [x] 已更新 [VISION.md](/E:/github/logcat/.codestable/requirements/VISION.md:1) 中该能力的一句话 pitch，把“查看并筛查”升级为“查看、锁定并筛查”。

## 7. roadmap 回写

- [x] 已将 [logcat-viewer-items.yaml](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-items.yaml:1) 中 `package-pid-and-process-sync` 的状态从 `in-progress` 改为 `done`。
- [x] 已将 [logcat-viewer-roadmap.md](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-roadmap.md:1) 中对应条目标记为 `done`，并把当前闭环描述刷新为前三条 feature 已完成。

## 8. attention.md 候选盘点

- [x] 本 feature 未暴露新的 `attention.md` 候选。

## 9. 遗留

- 后续优化点：`query-language-and-saved-filters`、`rich-log-details-and-crash-detection`、`export-workspace-and-settings`、`multi-session-wifi-and-layout`、`editor-ai-and-plugin-extensions` 仍未实现。
- 已知限制：当前仍没有高级查询、导出、工作区、多会话、Wi‑Fi ADB，也还没有命令层 `--pid` 优化和 App 元数据解析。
- 实现阶段顺手发现：无额外方案外遗留，当前验收只做了 architecture / requirement / roadmap 回写。
