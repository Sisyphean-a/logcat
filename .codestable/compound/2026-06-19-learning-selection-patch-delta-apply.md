---
doc_type: learning
track: knowledge
date: 2026-06-19
slug: selection-patch-delta-apply
component: app-state-sync
tags: [performance, selection, patch, state-sync, benchmark]
---

# 背景

日志列表的选择变化本来不涉及 adb，也不需要重建整份 UI state，但 Go 侧为了维护 `lastEmitState`，每次 `SelectionPatch` 仍会整窗复制 `logs`，再逐行重算 `isFocused/isSelected`。这条路径直接影响键盘上下选中、鼠标点选和多选时的响应流畅性。

# 指导原则

对于“状态补丁”型更新，优先保留上一版选择跟踪信息，按差量 source index 只改真正变化的行；不要从整窗行状态反推一遍再整窗重算。

# 为什么重要

- 这类选择操作频率高，用户能直接感知卡顿。
- 该成本完全在应用内部，不受 adb 传输限制，优化收益确定。
- 前端已经用 `selectedSourceIndexes/focusedSourceIndex` 做差量更新，Go 侧维护镜像状态也应采用同一策略，避免两边一边差量、一边全量。

# 何时适用

- 有“完整 state + patch”双轨同步，且服务端/宿主端需要维护一份上次已发状态缓存。
- patch 已经携带“变更前后可对比”的稳定键，例如这里的 `sourceIndex`。
- 行列表按稳定键有序，可用二分定位到单行。

# 示例

本次默认做法：

- `App` 持有 `lastSelectedSource`、`lastFocusedSource`、`selectionTrackRev`，和 `lastEmitState` 一起前进。
- `selectionPatchSnapshot` 直接把缓存的上次选择集传给 `applySelectionPatch`，避免再从 `lastEmitState.Logs` 扫描回推。
- `applySelectionPatch` 按“旧选中集 vs 新选中集”的差量 source index 只更新必要行，并复用已有 `SelectedLog` 指针。

量化结果：

- `BenchmarkApplySelectionPatch/current`: 约 `8.3–8.8 us/op`, `82016 B/op`, `2 allocs/op`
- `BenchmarkApplySelectionPatch/legacy`: 约 `9.7–10.0 us/op`, `82016 B/op`, `2 allocs/op`

结论：

选择补丁这类高频交互路径，若后端只是为了维护缓存 state，就不要再走整窗“复制 + contains”模式；应默认走 source-index 驱动的差量更新。
