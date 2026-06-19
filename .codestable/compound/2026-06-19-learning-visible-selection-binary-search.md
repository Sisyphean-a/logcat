---
doc_type: learning
track: knowledge
date: 2026-06-19
slug: visible-selection-binary-search
component: app-visible-selection
tags: [performance, selection, visible-logs, binary-search, source-index]
---

# 背景

`VisibleLogs` 一直按 `SourceIndex` 单调递增，但选择相关路径里仍残留了两个线性查找：`findVisibleIndexBySourceLocked` 在可见列表里扫 source，`slicesIndex` 在已排序的 `Selection.SourceIndexes` 里扫目标 source。单个操作不大，但一旦遇到大选择集，就会退化成明显的 O(n²) 组合成本。

# 指导原则

只要“列表已排序”是明确不变量，就不要保留线性 membership / index lookup；默认换成标准库二分查找。

# 为什么重要

- 多选复制、选择恢复、范围选中都依赖 `SourceIndex` 查找。
- `SelectedLogs()` 这种路径在大选择集场景下会把“外层遍历选中项”叠加成“内层再扫整个可见列表”，成本会急剧放大。
- 这里不改协议，不改状态结构，只是让现有有序数据真正发挥作用。

# 何时适用

- `VisibleLogs` 或选中 source 列表天然有序。
- 查找目标是“某个 source 在不在、在哪一位”，不是复杂谓词过滤。
- 热路径里会反复做同一种定位动作。

# 示例

本次默认做法：

- `findVisibleIndexBySourceLocked` 改为 `sort.Search`
- `slicesIndex` 改为 `sort.SearchInts`
- 保持原有返回语义不变：命中返回下标，未命中返回 `-1`

量化结果：

- `BenchmarkSelectedLogsLargeSelection/current` 约 `0.04–0.06 ms/op`
- `BenchmarkSelectedLogsLargeSelection/legacy` 约 `2.30–2.32 ms/op`

结论：

当 source index 已经是系统级稳定键时，选择相关路径应把它当作“可二分索引”来用，而不是继续拿它做线性比对键。否则一到大选择集场景，流畅性损失会非常明显。
