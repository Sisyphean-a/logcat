---
doc_type: feature-ff-note
feature: filter-dialog-figma-style
date: 2026-06-11
requirement:
tags: [frontend, figma, dialog, filter, style]
---

## 做了什么
把过滤器相关弹窗的视觉风格向用户提供的 Figma 参考收拢，重点调整暗色层级、尺寸、间距、圆角和表单节奏。  
这次只借用视觉语言，没有改动现有业务逻辑和交互能力。

## 改了哪些
- `frontend/src/filter-dialog.css` — 重做弹窗基底样式，改成更中性的深灰面板、紧凑 header/body/footer、克制的输入框和按钮风格
- `frontend/src/style.css` — 收口管理过滤器弹窗的侧栏、顶部操作区、默认标签和主编辑区视觉
- `Figma 节点 2:129` — 作为风格参考，抽取尺寸、间距、层级和表单节奏，不照抄结构

## 怎么验证的
跑通 `frontend/npm run build`。  
跑通 `go test ./...`。
