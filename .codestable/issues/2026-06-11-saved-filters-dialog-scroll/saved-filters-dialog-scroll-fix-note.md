---
doc_type: issue-fix
issue: saved-filters-dialog-scroll
date: 2026-06-11
status: fixed
tags: [frontend, dialog, scroll, layout, css]
---

## 现象
管理过滤器弹窗中，右侧表单内容超过可视区后无法滚动，底部内容被裁掉。

## 根因
[frontend/src/style.css](/E:/github/logcat/frontend/src/style.css:992) 里的管理弹窗中间层没有形成可收缩的滚动布局：`saved-filters-layout` / `saved-filters-main` / `saved-filters-form-body` 缺少 `flex: 1` 或 `min-height: 0`，内容因此被外层 `dialog-card` 的 `overflow: hidden` 裁掉。

## 修复
- 给 `saved-filters-layout` 增加 `flex: 1`、`min-width: 0`、`min-height: 0`
- 给 `saved-filters-sidebar`、`saved-filters-main`、`saved-filters-form-body` 补 `min-height: 0`

## 验证
- `frontend/npm run build`
