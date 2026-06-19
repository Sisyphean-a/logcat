---
doc_type: learning
track: knowledge
date: 2026-06-19
slug: rebuild-selection-scratch-reuse
component: app-search-rebuild
tags: [performance, search, rebuild, selection, allocation]
---

# 背景

`rebuildVisibleFromAllLogsLocked` 在搜索/过滤切换时会全量重建 `VisibleLogs`，这是日志查看器最敏感的内部热路径之一。此前主循环已经做到按需搜索缓存和有界缓冲，但 rebuild 前仍会用 `append([]int(nil), c.model.Selection.SourceIndexes...)` 临时拷贝选中集。

# 指导原则

对控制器内部“先记住旧状态、重建后再恢复”的短生命周期数据，优先复用 controller 级 scratch buffer，不要为了快照恢复每次新建临时切片。

# 为什么重要

- 这类 rebuild 会在筛选、搜索、恢复上下文时高频触发，属于用户能感知的交互热路径。
- 单次分配虽小，但会稳定出现在 benchmark 里，说明它是确定可消除的内部成本。
- 这里不涉及跨 goroutine 共享，也不暴露给外部接口，适合用受控 scratch buffer 换掉短命分配。

# 何时适用

- 数据只在持锁方法内部临时保存，生命周期不跨调用边界。
- 恢复逻辑只读这份副本，不要求保留历史版本。
- controller/对象实例天然串行访问，scratch buffer 不会被并发读写。

# 示例

这次的默认做法：

- 在 `Controller` 上增加 `selectionScratch []int`
- `rebuildVisibleFromAllLogsLocked` 调 `cloneSelectionSourceIndexesLocked()`，按当前选中数复用/扩容 scratch
- 无选中时直接复位 `selectionScratch[:0]`，避免残留旧长度

量化结果：

- `BenchmarkRebuildVisibleWithSearch`：从约 `634–636 us/op, 1 alloc/op` 降到约 `604–610 us/op, 0 alloc/op`

结论：

对搜索/过滤 rebuild 这种“重建列表前先保存少量状态”的路径，内部 scratch 复用应当是默认选项。只要状态不逃逸到外部，就不要为了临时备份再做一次 `append(nil, ...)`。
