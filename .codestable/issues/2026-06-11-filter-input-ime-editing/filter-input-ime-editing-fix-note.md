---
doc_type: issue-fix
issue: filter-input-ime-editing
path: fast-track
fix_date: 2026-06-11
tags: [frontend, filter, input, ime, caret]
---

# 过滤输入框 IME / 空格 / 光标跳尾 修复记录

## 1. 问题描述

过滤输入框存在三类编辑问题：

- 输入空格时会被吃掉
- 中文输入法组合输入后切回英文，未确认的拼音容易残留
- 在字符串中间插入内容时，光标会跳到末尾

## 2. 根因

`frontend/src/app-shell.tsx` 的过滤输入框直接受控于 `state.filter.draft`，并且每次 `onChange` 都立即调用 `SetFilterDraft` 回写到 Go 层，再用返回状态重新控制 `value`。这让输入法组合态和光标位置持续被外部状态打断。  
同时 `internal/app/filter.go` 的 `SetFilterDraft` 在“草稿阶段”就执行 `TrimSpace`，导致尾部空格输入时被立刻吞掉。

## 3. 修复方案

- 过滤输入框改为前端本地草稿态，输入期间不再每击键回写后端
- 仅在“应用”或“保存”动作发生时同步草稿到状态层
- Enter 键在 IME 组合态下不触发应用
- 后端 `SetFilterDraft` 不再 trim，trim 只保留在真正“应用过滤器”的阶段

## 4. 改动文件清单

- `frontend/src/app-shell.tsx`
- `frontend/src/App.tsx`
- `frontend/src/use-app-controller.ts`
- `internal/app/filter.go`
- `internal/app/controller_test.go`

## 5. 验证结果

- `go test ./...` 通过
- `frontend/npm run build` 通过
- 本地预览页 `http://127.0.0.1:4173` 可正常返回
- 临时 DOM 回归脚本验证了三条关键表现：
  - 中间插入字符后，光标位置保持在插入点后
  - 尾部空格可保留
  - 组合输入结束后继续英文输入时，结果值保持为最终文本而不是拼音残留

## 6. 遗留事项

无。
