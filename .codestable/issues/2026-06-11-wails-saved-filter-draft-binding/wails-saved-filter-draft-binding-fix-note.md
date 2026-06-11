---
doc_type: issue-fix
issue: wails-saved-filter-draft-binding
date: 2026-06-11
status: fixed
tags: [frontend, wails, typescript, binding]
---

## 现象
`frontend/npm run build` 在 `ReplaceSavedFilterDefinitions` 调用处报类型错误，提示本地 `SavedFilterDefinitionDraft[]` 不能赋给 Wails 生成的 `SavedFilterDraft[]`。

## 根因
Wails 为 Go 结构 `SavedFilterDraft` 生成的是大写字段 `ExistingID/Name/PackageName/Query`，而前端本地草稿使用的是小写字段 `existingID/name/packageName/query`，直接传参时类型不兼容。

## 修复
- [frontend/src/use-app-controller.ts](/E:/github/logcat/frontend/src/use-app-controller.ts:1) 引入 `app.SavedFilterDraft`
- 在调用 `ReplaceSavedFilterDefinitions` 前，把本地 draft 显式映射成 Wails 绑定对象

## 验证
- `frontend/npm run build`
