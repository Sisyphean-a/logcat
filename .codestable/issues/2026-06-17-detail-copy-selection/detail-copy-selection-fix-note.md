---
doc_type: issue-fix
issue: detail-copy-selection
status: fixed
path: fast-track
fix_date: 2026-06-17
tags: [frontend, keyboard, clipboard, detail-panel]
---

# 详情面板文本选区复制冲突修复记录

## 1. 问题描述

主列表选中日志后，右侧详情面板会展示文本内容。此时如果用户在详情面板里手动选中一段文本再按 `Ctrl+C`，实际复制的是整条已选日志，而不是当前选区文本。

## 2. 根因

`frontend/src/use-app-controller.ts` 注册了全局 `keydown`，对 `Ctrl+C` 直接拦截并转成“复制所选日志”。这段逻辑只排除了输入框等可编辑控件，没有判断页面上是否已经存在真实文本选区，所以详情面板里的普通文本选中也被错误劫持。

## 3. 修复方案

在全局快捷键处理里增加“当前页面是否存在非折叠文本选区”的判断：

- 有文本选区时，不拦截 `Ctrl+C`，交给浏览器默认复制行为
- 没有文本选区时，才继续执行日志级快捷复制

## 4. 改动文件清单

- `frontend/src/use-app-controller.ts`

## 5. 验证结果

- `npm --prefix frontend run build`
- `go test ./...`

## 6. 遗留事项

无。
