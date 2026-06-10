---
doc_type: issue-fix
issue: 2026-06-10-clear-visible-breaks-stream
path: fast-track
fix_date: 2026-06-10
tags: [frontend, log-view, virtual-scroll]
---

# 清空日志后看似停流 修复记录

## 1. 问题描述

监听日志进行中点击“清除信息”后，界面后续不再显示新日志，看起来像日志流已经断开。

## 2. 根因

`frontend/src/log-table.tsx:53` 用清空前残留的 `scrollTop` 直接计算虚拟列表窗口起点。清空后 `logs` 变短甚至为空，但旧偏移仍然很大，后续新日志数量不足以追上该偏移时，`logs.slice(start, end)` 会一直切到空区间，导致日志实际已到达但界面不渲染。

## 3. 修复方案

在日志表中把窗口起点限制在当前日志范围内，确保清空后即使保留旧滚动偏移，后续新增日志也会落入可渲染区间。

## 4. 改动文件清单

- `frontend/src/log-table.tsx`

## 5. 验证结果

- 代码检查：`npm run build` 通过
- 回归验证：`go test ./...` 通过
- 场景结论：清空日志后，后续新日志不会再因旧滚动偏移被切成空列表

## 6. 遗留事项

无
