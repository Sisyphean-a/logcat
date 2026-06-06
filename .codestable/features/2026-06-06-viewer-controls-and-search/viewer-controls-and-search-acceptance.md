# viewer-controls-and-search 验收报告

> 阶段：阶段 3（验收闭环）
> 验收日期：2026-06-06
> 关联方案 doc：[viewer-controls-and-search-design.md](/E:/github/logcat/.codestable/features/2026-06-06-viewer-controls-and-search/viewer-controls-and-search-design.md:1)

## 1. 接口契约核对

- [x] `LogViewItem`：代码已把可见列表从字符串升级成 [LogViewItem](/E:/github/logcat/internal/app/model.go:11)，同时持有 `Entry` 与 `Display`，与方案第 2.1 节一致。
- [x] `SearchState`：代码已落地 [SearchState](/E:/github/logcat/internal/app/model.go:16)，保存 `Query`、`MatchIndexes`、`Current`，与接口示例一致。
- [x] `PauseState`：代码已落地 [PauseState](/E:/github/logcat/internal/app/model.go:22)，保存 `Active`、`BufferedCount`、`DroppedCount`，与方案一致。
- [x] 选中态：实现里没有额外的 `SelectionState` 实体，而是用 [SelectedIndex](/E:/github/logcat/internal/app/model.go:35) 显式承载选中态；已回填方案 doc，当前设计与代码一致。
- [x] controller 操作面：`Pause`、`ResumeKeep`、`ResumeDiscard`、`ClearVisible`、`SetSearchQuery`、`NextMatch`、`PrevMatch`、`SelectLog`、`SelectedLog` 已全部落地，见 [logview.go](/E:/github/logcat/internal/app/logview.go:12)。
- [x] 主流程图核对：`日志事件进入 Controller -> paused/visible 分流 -> 搜索重算 / 缓冲计数更新 -> UI 重绘 -> 选中详情联动 -> 跟随开关` 在 [logview.go](/E:/github/logcat/internal/app/logview.go:136)、[interactions.go](/E:/github/logcat/internal/ui/interactions.go:23)、[log_panel.go](/E:/github/logcat/internal/ui/log_panel.go:17)、[detail_panel.go](/E:/github/logcat/internal/ui/detail_panel.go:11) 均有实际代码落点。

## 2. 行为与决策核对

- [x] 暂停/恢复留在 controller 层：暂停只改内存状态和缓冲，不杀 logcat 进程，见 [logview.go](/E:/github/logcat/internal/app/logview.go:12) 和 [logview.go](/E:/github/logcat/internal/app/logview.go:136)。
- [x] 搜索只作用于当前可见日志：命中计算只遍历 `VisibleLogs`，见 [logview.go](/E:/github/logcat/internal/app/logview.go:164)。
- [x] 复制动作由 UI 发起：剪贴板命令只在 [writeClipboard](/E:/github/logcat/internal/ui/interactions.go:111) 里发出，controller 没有平台剪贴板依赖。
- [x] 自动跟随依赖 Gio 列表位置：跟随状态由 [syncAutoFollow](/E:/github/logcat/internal/ui/interactions.go:155) 基于 `ScrollToEnd + Position.BeforeEnd` 维护。
- [x] 详情面板只展示当前已解析字段：`Time`、`Level`、`Tag`、`Message`、`Raw` 都在 [layoutDetail](/E:/github/logcat/internal/ui/detail_panel.go:11) 落地，没有抢跑 JSON / 堆栈深解析。
- [x] `ClearVisible` 只清视图：实现只清 `VisibleLogs`、匹配集和选中态，不调用设备级清空，且 query 保留给后续新日志继续匹配；方案 doc 已同步澄清，见 [logview.go](/E:/github/logcat/internal/app/logview.go:59)。
- [x] 挂载点反向核对：本 feature 的实际挂入点只落在 `internal/app/model.go`、`internal/app/logview.go`、`internal/ui/*.go` 和现有入口 [main.go](/E:/github/logcat/cmd/logcatviewer/main.go:1)；没有清单外的新模块残留。
- [x] 拔除沙盘推演：去掉 `internal/ui` 的工具栏 / 列表 / 详情挂入后，本 feature 的暂停、搜索、复制、跟随、详情能力就全部消失；去掉 `internal/app/logview.go` 后 UI 控制没有后端状态支撑，挂载点清单完整。

## 3. 验收场景核对

- [x] **S1 暂停保持采集**
  - 证据来源：`TestControllerPauseBuffersIncomingLogs`
  - 结果：通过
- [x] **S2 恢复并保留缓冲**
  - 证据来源：`TestControllerResumeKeepFlushesBufferedLogs`
  - 结果：通过
- [x] **S3 恢复并丢弃缓冲**
  - 证据来源：`TestControllerResumeDiscardDropsBufferedLogs`
  - 结果：通过
- [x] **S4 清空视图**
  - 证据来源：`TestControllerClearVisibleResetsListSelectionAndSearch`
  - 结果：通过
- [x] **S5 基础搜索**
  - 证据来源：`TestControllerSearchTracksMatchesAndCurrentSelection`、`TestControllerConsumeRecomputesMatchesWhenLogsArrive`
  - 结果：通过
- [x] **S6 自动跟随**
  - 证据来源：`TestShellSyncAutoFollowStopsAfterLeavingEnd`、`TestShellSyncAutoFollowRestoresEndWhenEnabled`、`TestShellProgrammaticSelectionKeepsAutoFollowEnabled`
  - 结果：通过
- [x] **S7 复制当前日志**
  - 证据来源：`TestShellSelectedClipboardTextUsesCurrentSelection`
  - 结果：通过
- [x] **UI 运行态**
  - 证据来源：`go build -o .\logcatviewer.exe ./cmd/logcatviewer` 后启动进程 2 秒存活检查
  - 结果：通过，进程未秒崩

## 4. 术语一致性

- [x] `LogViewItem`、`SearchState`、`PauseState`、`SelectedIndex` 在代码与设计中的语义一致。
- [x] `SelectionState` 已从方案文档回填为 `SelectedIndex`，代码里无遗留冲突命名。
- [x] 范围外词汇检索：`rg -n "logcat -c|caseSensitive|wholeWord|workspace|multi-session|multi session|source parser|StackTrace|stacktrace" internal\app internal\ui cmd` 无命中。

## 5. 架构归并

- [x] 已更新 [runtime-single-device-logcat-loop.md](/E:/github/logcat/.codestable/architecture/runtime-single-device-logcat-loop.md:1)，把 `LogViewItem`、`SearchState`、`PauseState`、`SelectedIndex`、暂停缓冲、自动跟随和详情面板这些现状写回架构层。
- [x] 已更新 [ARCHITECTURE.md](/E:/github/logcat/.codestable/architecture/ARCHITECTURE.md:1)，让总入口从“最小闭环”升级为“单设备查看与基础筛查闭环”，并同步关键约束与术语。

## 6. requirement 回写

- [x] 本 feature 沿用现有 requirement `h5-logcat-viewing`，因为它扩展的是同一条“查看并筛查 H5 日志”的用户能力。
- [x] 已更新 [h5-logcat-viewing.md](/E:/github/logcat/.codestable/requirements/h5-logcat-viewing.md:1)，把用户故事、能力描述和边界刷新到当前实现，并追加变更日志。
- [x] 已更新 [VISION.md](/E:/github/logcat/.codestable/requirements/VISION.md:1) 中该能力的一句话 pitch。

## 7. roadmap 回写

- [x] 已将 [logcat-viewer-items.yaml](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-items.yaml:1) 中 `viewer-controls-and-search` 的状态从 `in-progress` 改为 `done`。
- [x] 已将 [logcat-viewer-roadmap.md](/E:/github/logcat/.codestable/roadmap/logcat-viewer/logcat-viewer-roadmap.md:1) 中对应条目标记为 `done`，并回写当前闭环描述。

## 8. attention.md 候选盘点

- [x] 本 feature 未暴露新的 `attention.md` 候选。

## 9. 遗留

- 后续优化点：`package-pid-and-process-sync`、`rich-log-details-and-crash-detection`、`query-language-and-saved-filters` 仍未实现。
- 已知限制：当前仍是单设备、单活跃会话、基础子串搜索、无导出、无持久化、无包名 / PID 绑定。
- 实现阶段顺手发现：自动跟随原先会被程序性选中误关；已在验收阶段修复，并由 `TestShellProgrammaticSelectionKeepsAutoFollowEnabled` 兜住。
