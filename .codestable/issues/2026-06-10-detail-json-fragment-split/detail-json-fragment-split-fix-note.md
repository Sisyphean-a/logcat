---
doc_type: issue-fix
issue: detail-json-fragment-split
path: fast-track
fix_date: 2026-06-10
tags: [frontend, detail-panel, json, url, parsing]
---

# 详情面板 JSON 被中途拆断 修复记录

## 1. 问题描述

详情面板里某些带长 URL 的 JSON 响应会在中途被拆成多个块显示：前半段先作为一个 JSON block 封口，后半段又被当成普通文本和 URL block 继续渲染，导致同一个顶层 JSON 看起来被“截断”。

## 2. 根因

`frontend/src/log-detail.tsx` 原先优先走 `findJsonFragments`，会从文本里扫描所有可独立 `JSON.parse` 的子片段。这样一来，只要整段消息里某个内层对象或数组恰好能单独 parse，就可能被提前抽成一个独立 JSON block；外层 JSON 即使本该整体展示，也会被误拆成“前半 JSON + 后半文本/URL”。

## 3. 修复方案

调整分块策略：

- 先尝试把整段文本当成一个完整 JSON 解析；
- 如果整段能 parse，直接整段作为单个 JSON block 渲染；
- 如果整段长得像 JSON（真正的 `{...}` 或 `[...]` 起始）但整体 parse 失败，就整段按普通文本显示，不再往里挖内层 JSON 片段；
- 对带日志前缀的场景（如 `[h5-api] response ... {json}`），不能把前缀里的 `[` 误判成 JSON 数组起始，仍要继续识别后面的内联 JSON 片段。

这样可以避免“顶层还没结束，内层先被封口”的错分块。

## 4. 改动文件清单

- `frontend/src/log-detail.tsx`

## 5. 验证结果

- `frontend` 下 `npm run build` 通过
- 代码层验证：
  - 你给的纯 JSON 长 URL 样例会先整体识别为单个 JSON block，不再拆出中间子对象
  - 带 `[h5-api] ... {json}` 前缀的响应日志不会再被误判成“整段像 JSON 但 parse 失败”，后续内联 JSON 仍可正常识别

## 6. 遗留事项

无。
