---
doc_type: explore
type: question
date: 2026-06-10
slug: message-highlight-whitespace-loss
topic: 消息列表里 token 高亮后左右空格为什么会消失
scope: frontend 日志列表的 message token 渲染与表格样式
keywords:
  - log-row
  - log-text
  - message-cell
  - flex
  - whitespace
status: active
confidence: high
---

## 问题与范围

排查日志列表“消息”列中，`POST` 这类被语义高亮的单词前后空格消失的原因。

## 速答

根因不在正则匹配本身，而在布局层：`tokenizeLogText` 会把原字符串切成多个 token，再由 `TokenText` 渲染成多个兄弟 `span`；与此同时，消息单元格 `message-cell` 被表格样式设成了 `display: flex`。这样每个 token `span` 都成了 flex item，原本落在 token 边界上的前导/尾随空格不再按一整段文本连续排版，结果就是高亮 token 两侧的空格被视觉上“吃掉”。

## 关键证据

- `LogRow` 把 `log.message` 先做 `tokenizeLogText(log.message)`，再在 `message-cell` 里渲染 `<TokenText tokens={messageTokens} />`，说明消息列不再是单个文本节点，而是 token 列表。[frontend/src/log-row.tsx](/E:/github/logcat/frontend/src/log-row.tsx:6)
- `TokenText` 对每个 token 单独输出一个 `span`，非 plain token 还会附加 `token-*` 类名；也就是说高亮词和它左右文本天然被拆到了不同元素里。[frontend/src/log-text.tsx](/E:/github/logcat/frontend/src/log-text.tsx:28)
- `tokenizeLogText` 在命中 token 前，会把 `cursor` 到 `match.index` 的原文切成 plain token 原样塞回去，所以空格其实仍在 token 数据里，没有在分词阶段丢失。[frontend/src/log-text.tsx](/E:/github/logcat/frontend/src/log-text.tsx:46)
- `.table-row > span` 统一被设置为 `display: flex`，`message-cell` 正是 `table-row` 的直接子 `span`，因此消息 token 列表被放进了 flex 容器里。[frontend/src/log-table.css](/E:/github/logcat/frontend/src/log-table.css:7)
- `.message-cell` 本身只做 `overflow: hidden; white-space: nowrap; text-overflow: ellipsis;`，没有恢复普通 inline 文本流；这和 token 拆分叠加后，刚好形成“词变色后边界空格消失”的现象。[frontend/src/style.css](/E:/github/logcat/frontend/src/style.css:639)

## 细节展开

以截图里的 `request POST /api/...` 为例，当前实现会得到近似三段：

1. plain: `"[h5-api] request "`
2. http-method: `"POST"`
3. plain: `" /api/..."`

如果这三段处在普通 inline 文本流里，视觉上会连续显示为 `request POST /api/...`。但现在父级 `message-cell` 是 flex 容器，三个 token `span` 成了三个独立 flex item，位于 item 边界的空格不会像单个文本节点那样稳定保留下来，于是就出现了 `requestPOST/api/...` 这种效果。

## 未决问题

无。现象与当前 DOM 结构和样式规则可以直接对应上。

## 后续建议

如需继续处理，可直接围绕 `message-cell` 的布局方式和 token 空白保留策略做定点修正。

## 相关文档

- [viewer-log-semantic-highlighting-design.md](/E:/github/logcat/.codestable/features/2026-06-09-viewer-log-semantic-highlighting/viewer-log-semantic-highlighting-design.md:1)
- [viewer-log-semantic-highlighting-acceptance.md](/E:/github/logcat/.codestable/features/2026-06-09-viewer-log-semantic-highlighting/viewer-log-semantic-highlighting-acceptance.md:1)
