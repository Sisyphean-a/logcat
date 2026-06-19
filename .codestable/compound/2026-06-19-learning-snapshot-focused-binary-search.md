---
doc_type: learning
track: knowledge
date: 2026-06-19
slug: snapshot-focused-binary-search
component: app-selection-snapshot
tags: [performance, snapshot, selection, binary-search, source-index]
---

# 背景

`SelectionSnapshot` 只需要返回当前焦点行和选中 source index 集，但旧实现仍会在窗口化后的 `VisibleLogs` 里线性扫描 `focusedSourceIndex`。当窗口固定在 1000 行量级时，这种线性扫在逻辑上不大，却会稳定出现在高频热路径里。

# 指导原则

只要列表按稳定键单调有序，就不要在 snapshot 热路径里继续线性查找；默认改成按稳定键二分定位。

# 为什么重要

- snapshot 会在键盘导航、鼠标点选、多选时高频触发，直接影响交互流畅性。
- `VisibleLogs` 天然按 `SourceIndex` 单调递增，数据结构已经给出了更优查找方式。
- 这类优化不改协议、不改状态语义，只是把“找同一行”的成本从 O(n) 变成 O(log n)。

# 何时适用

- 日志/行列表已经按 `SourceIndex`、时间戳或其他稳定键有序。
- 只需要定位单个目标项，而不是遍历整段做过滤。
- snapshot/patch 路径对延迟敏感，值得为查找单独抽一个 helper。

# 示例

本次默认做法：

- `cloneFocusedLogItem` 不再循环扫描窗口行
- 抽出 `findFocusedLogItemIndex(items, focusedSourceIndex)` 做标准二分查找
- 找到下标后再复制单个 `LogViewItem`

量化结果：

- `BenchmarkSelectionSnapshot/current` 从约 `518–524 ns/op` 降到约 `58 ns/op`
- 旧版 legacy 仍在 `15.8–17.6 us/op`

结论：

只要 `VisibleLogs` 的有序性已经是系统不变量，snapshot/patch 这种高频单点定位路径就应默认使用二分查找，而不是保留“简单但线性”的扫描实现。
