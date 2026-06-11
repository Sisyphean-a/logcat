---
doc_type: feature-ff-note
feature: move-package-scope-toolbar
date: 2026-06-11
requirement:
tags: [frontend, toolbar, filter, package-scope]
---

## 做了什么
把包范围选择器从下方过滤栏挪到了顶部工具栏，释放主过滤区横向空间，同时保留“全部包 / 用户包 / 系统包”的切换能力。

## 改了哪些
- `frontend/src/toolbar.tsx` — 在设备选择后新增顶部包范围选择器
- `frontend/src/app-shell.tsx` — 移除过滤栏里的包范围选择器
- `frontend/src/App.tsx` — 把 `setPackageScope` 事件接到工具栏
- `frontend/src/style.css` — 增加顶部选择器宽度样式并移除旧位置样式

## 怎么验证的
跑了 `frontend/npm run build`，通过。代码层确认包范围切换入口只保留顶部这一处。
