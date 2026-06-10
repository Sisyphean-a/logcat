---
doc_type: issue-fix
issue: log-message-token-whitespace
path: fast-track
fix_date: 2026-06-10
tags: [frontend, log-view, token-highlight, whitespace, flex]
---

# 消息列表高亮 token 两侧空格丢失 修复记录

## 1. 问题描述

日志表格的“消息”列里，像 `POST` 这类被高亮的 token 左右空格在列表中看不见，导致 `requestPOST/api...` 这种连在一起的视觉效果；同一条消息在详情面板里显示正常。

## 2. 根因

`TokenText` 会把消息拆成多个兄弟 `span`。详情面板把这些 `span` 放在 `pre` 的普通文本流里，所以空格能正常保留；但日志表格上层有一条通用样式 `frontend/src/log-table.css` 的 `.table-row > span { display: flex; }`，把 `message-cell` 也变成了 flex 容器，导致 token 之间位于 flex item 边界上的空格在表格里被视觉吞掉。

## 3. 修复方案

只对消息列做样式覆盖，让 `.table-row > .message-cell` 脱离这条通用 flex 规则，回到普通 block 文本流；同时保留原有的 `nowrap / overflow / ellipsis`，不改分词规则、不改其他列布局。

## 4. 改动文件清单

- `frontend/src/log-table.css`
- `frontend/src/style.css`

## 5. 验证结果

- `frontend` 下 `npm run build` 通过
- 代码层对照结果：详情面板原本正常，表格列现在与详情面板使用同样的 token 文本流，不再依赖 flex item 排版

## 6. 遗留事项

无。
