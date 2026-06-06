# single-device-h5-loop 验收报告

> 阶段：阶段 3（验收闭环）
> 验收日期：2026-06-06
> 关联方案 doc：[single-device-h5-loop-design.md](/E:/github/logcat/.codestable/features/2026-06-05-single-device-h5-loop/single-device-h5-loop-design.md:1)

## 1. 接口契约核对

- [x] `DetectADB` / `ListDevices` 示例：代码实际入口是 [controller.go](/E:/github/logcat/internal/app/controller.go:45) `Load`，会先调检测再调设备列表，行为与方案一致。
- [x] `SessionConfig` 示例：当前实现只落地了最小字段 `DeviceID`，代码位置 [supervisor.go](/E:/github/logcat/internal/session/supervisor.go:9)。这属于方案里最小闭环的已实现子集，没有出现相反行为。
- [x] `LogEntry` 名词变化：代码中已存在 [LogEntry](/E:/github/logcat/internal/logcat/parser.go:11)，并承载 `DeviceID`、`TimeText`、`PID`、`TID`、`Level`、`Tag`、`Message`、`Raw`。
- [x] 主流程图核对：`main -> ui.Shell.bootstrap -> controller.Load / SelectDevice -> session.Supervisor -> logcat.ParseThreadtimeLine / MatchesH5Preset` 均有实际代码落点，分别在 [main.go](/E:/github/logcat/cmd/logcatviewer/main.go:1)、[shell.go](/E:/github/logcat/internal/ui/shell.go:46)、[controller.go](/E:/github/logcat/internal/app/controller.go:45)、[supervisor.go](/E:/github/logcat/internal/session/supervisor.go:36)、[parser.go](/E:/github/logcat/internal/logcat/parser.go:27)、[preset.go](/E:/github/logcat/internal/logcat/preset.go:5)。

## 2. 行为与决策核对

- [x] 应用启动后显式展示 `adb` 结果：`Load` 成功时状态文案写成 `adb <version>`，见 [controller.go](/E:/github/logcat/internal/app/controller.go:58)。
- [x] 先做单活跃会话：切换设备前先取消旧会话，见 [controller.go](/E:/github/logcat/internal/app/controller.go:95)。
- [x] 命令层先收窄到 `chromium`：`buildLogcatArgs` 固定生成 `chromium:I` 过滤，见 [logcat_source.go](/E:/github/logcat/internal/adb/logcat_source.go:45)。
- [x] 本地层再过滤 `[H5]`：`MatchesH5Preset` 只接受 `chromium` 且 message 含 `[H5]` 的条目，见 [preset.go](/E:/github/logcat/internal/logcat/preset.go:5)。
- [x] 错误显式上浮：`adb` 缺失、设备未授权、设备离线和解析失败都会进入状态或问题事件，而不是静默吞掉，见 [controller.go](/E:/github/logcat/internal/app/controller.go:47)、[controller.go](/E:/github/logcat/internal/app/controller.go:76)、[parser.go](/E:/github/logcat/internal/logcat/parser.go:30)。
- [x] 挂载点反向核对：当前实现挂载点实际落在 `cmd/logcatviewer/main.go`、`internal/ui/shell.go`、`internal/app/controller.go`、`internal/adb/logcat_source.go`、`internal/logcat/*`。设计中的 `internal/app/bootstrap` 和 `internal/app/session` 是概念性挂入点，实际代码由 `controller.go` 与 `supervisor.go` 承载，没有清单外残留引用。

## 3. 验收场景核对

- [x] 启动成功路径
  - 证据来源：`TestControllerLoadUpdatesDevicesAndStatus`
  - 结果：通过
- [x] 设备启动会话
  - 证据来源：`TestControllerSelectDeviceAppendsLogEvents`
  - 结果：通过
- [x] H5 过滤
  - 证据来源：`TestParseThreadtimeLineParsesChromiumConsole`、`TestSupervisorStreamsMatchingEntries`
  - 结果：通过
- [x] ADB 缺失
  - 证据来源：`TestControllerLoadADBMissingSetsExplicitStatus`
  - 结果：通过
- [x] 设备未授权 / 离线
  - 证据来源：`TestControllerRejectsNonReadyDevices`
  - 结果：通过
- [x] 解析失败保底
  - 证据来源：`TestSupervisorReportsParseErrorWithRawLineAndContinues`
  - 结果：通过

## 4. 术语一致性

- [x] `ADBInstall` / `DeviceInfo` / `LogEntry` / `H5 预设` 这些术语在代码与文档中的语义一致。
- [x] 代码层未出现多窗口、多 Tab、包名选择等超范围概念。
- [x] 范围外词汇检索：`rg -n "logcat -c|pidof|CurrentForeground|SavedFilter|QueryHistory|workspace|tab|foreground" cmd internal` 无命中。

## 5. 架构归并

- [x] 已新增 [runtime-single-device-logcat-loop.md](/E:/github/logcat/.codestable/architecture/runtime-single-device-logcat-loop.md:1)，把本次 feature 的名词、链路、约束和代码锚点写回架构层。
- [x] 已更新 [ARCHITECTURE.md](/E:/github/logcat/.codestable/architecture/ARCHITECTURE.md:1)，让总入口不再是空骨架，并能直接定位当前已落地链路。

## 6. requirement 回写

- [x] 方案 frontmatter 的 `requirement` 为空，但本 feature 新增了用户可感能力，因此按 backfill 补了一份 current requirement。
- [x] 已新增 [h5-logcat-viewing.md](/E:/github/logcat/.codestable/requirements/h5-logcat-viewing.md:1)。
- [x] 已新增 [VISION.md](/E:/github/logcat/.codestable/requirements/VISION.md:1) 并把该能力登记到 `current` 分组。

## 7. roadmap 回写

- [x] 已将 [logcat-viewer-items.yaml](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-items.yaml:1) 中 `single-device-h5-loop` 的状态从 `in-progress` 改为 `done`。
- [x] 已将 [logcat-viewer-roadmap.md](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-roadmap.md:1) 中对应条目标记为 `done`，并补上相关 requirement / architecture 引用。

## 8. attention.md 候选盘点

- [x] 候选 1：当前机器拉 Gio 依赖时需要在命令里临时关闭 `GOSUMDB`，否则可能因为 `sum.golang.org` 握手超时卡住。是否写入 `attention.md` 待用户决定。

## 9. 遗留

- 后续优化点：暂停/恢复、清空视图、复制日志、基础搜索、自动滚动和详情面板尚未实现，对应 roadmap 条目 `viewer-controls-and-search`。
- 已知限制：当前仍是单窗口、单会话、无持久化、无导出、无多设备并行。
- 实现阶段顺手发现：设计里的挂载点名偏概念化，后续 feature 继续写设计时应直接按当前代码中的 `controller` / `supervisor` / `shell` 命名落点。
