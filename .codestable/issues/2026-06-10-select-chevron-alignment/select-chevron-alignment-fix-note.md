---
doc_type: issue-fix
issue: select-chevron-alignment
path: fast-track
fix_date: 2026-06-10
tags: [frontend, ui, select-control, chevron]
---

# 下拉箭头未垂直居中 修复记录

## 1. 问题描述

顶部过滤器和包名选择器里的下拉箭头视觉上偏下，和输入框内容不在同一垂直中心线上，看起来发沉。

## 2. 根因

`frontend/src/style.css` 里 `.select-control .chevron` 被设成了绝对定位，但没有给箭头外层提供稳定的垂直居中锚点；SVG 实际按静态位置落在触发器里，导致视觉上没有居中。

## 3. 修复方案

把箭头外层 `select-control-arrow` 设成一个绝对定位、`top: 50%` 的 14x14 居中容器，再让内部 `.chevron` 回到普通流里，避免 SVG 基线和绝对定位叠加造成的偏移。

## 4. 改动文件清单

- `frontend/src/style.css`

## 5. 验证结果

- `frontend` 下 `npm run build` 通过
- 改动范围仅限选择器箭头定位样式，没有触碰 `select-control.tsx` 当前未提交的逻辑修改

## 6. 遗留事项

无。
