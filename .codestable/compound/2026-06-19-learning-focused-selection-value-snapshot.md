---
doc_type: learning
track: knowledge
date: 2026-06-19
slug: focused-selection-value-snapshot
component: app-selection-snapshot
tags: [performance, selection, snapshot, memory, value-object]
---

# 背景

`SelectionSnapshot` 只需要把“当前焦点行的展示字段”传给补丁构造，但旧实现直接复制整行 `LogViewItem` 并返回 `*LogViewItem`。这会把 `LogEntry` 整个值对象一起带过来，即使后续真正用到的只有 `time/level/tag/message/source`。

# 指导原则

高频 snapshot 路径不要为了方便复用上游重对象；如果下游只消费一个稳定字段子集，就应该显式定义轻量值快照。

# 为什么重要

- `SelectionSnapshot` 会在键盘导航、鼠标点选、多选时频繁触发。
- `LogViewItem` 包含比 patch 所需更多的字段，把整对象带过去会增加无意义复制成本。
- 这类轻量值快照不会改变协议语义，只是把“需要什么”表达得更精确。

# 何时适用

- 上游对象明显比下游需求更重。
- 下游只读字段，不依赖上游对象身份或可变状态。
- 路径足够热，值得为它单独引入一个值对象类型。

# 示例

本次默认做法：

- `SelectionSnapshot.Focused` 从 `*LogViewItem` 改成 `*FocusedLogSnapshot`
- `FocusedLogSnapshot` 只保留 `sourceIndex/timeText/level/tag/message/source`
- `buildFocusedSelectedLog` 直接从轻量快照构造 `SelectedLogView`

量化结果：

- `BenchmarkSelectionSnapshot/current` 从约 `168 B/op` 降到约 `120 B/op`
- 同时耗时从约 `58–59 ns/op` 保持在同一量级并略有改善

结论：

当 patch/snapshot 只需要“展示用字段子集”时，不要继续透传或复制完整领域对象。显式建一个轻量值快照，通常能同时改善内存复制和代码表达。
