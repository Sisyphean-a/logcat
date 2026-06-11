---
doc_type: feature-ff-note
feature: saved-filter-manager-default
date: 2026-06-11
requirement:
tags: [frontend, filter, saved-filter, default, storage]
---

## 做了什么
把已保存筛选的编辑入口改成侧栏管理弹窗，支持切换不同筛选、排序、删除和设默认。  
同时补齐默认筛选持久化和启动恢复，让应用打开后可以自动选中指定筛选，不再总是空态。

## 改了哪些
- `frontend/src/saved-filters-dialog.tsx` / `frontend/src/saved-filter-form.tsx` / `frontend/src/save-filter-dialog.tsx` — 拆出筛选表单与管理弹窗，新增侧栏切换、排序、删除、默认能力
- `frontend/src/App.tsx` / `frontend/src/use-app-controller.ts` / `frontend/src/select-control.tsx` / `frontend/src/toolbar.tsx` / `frontend/src/style.css` — 接入批量保存接口、移除“全部包”清除按钮并补齐弹窗布局
- `app.go` / `main.go` / `internal/app/saved_filter_management.go` / `internal/app/filter.go` / `internal/storage/filters.go` — 增加默认筛选持久化、批量替换 saved filter 定义和启动恢复逻辑
- `internal/app/filter_update_test.go` / `internal/storage/export_test.go` — 补默认筛选与批量管理的回归测试

## 怎么验证的
跑通 `go test ./...`。  
跑通 `frontend/npm run build`。
