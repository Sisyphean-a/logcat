---
doc_type: issue-fix
issue: saved-filters-dialog-black-screen
date: 2026-06-11
status: fixed
tags: [frontend, react, hooks, dialog]
---

## 现象
点击“编辑”打开已保存筛选管理弹窗时，界面直接黑屏。

## 根因
[frontend/src/saved-filters-dialog.tsx](/E:/github/logcat/frontend/src/saved-filters-dialog.tsx:1) 里把 `useCallback` 写在 `if (!open) return null` 后面，导致关闭态和打开态的 hook 数量不一致，React 运行时会抛出 hook 顺序错误并中断渲染。

## 修复
- 把 `handleSelectedDraftChange` 的 `useCallback` 提前到条件返回之前，保证每次渲染的 hook 顺序一致

## 验证
- `frontend/npm run build`
