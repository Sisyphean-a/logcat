---
doc_type: issue-fix
issue: save-filter-dialog-or-group-button-hidden
path: fast-track
fix_date: 2026-06-11
tags: [frontend, dialog, css, button, filter]
---

# 保存过滤器弹窗 OR 条件组按钮被隐藏 修复记录

## 1. 问题描述

新增 / 编辑过滤器弹窗里，用户看不到“新增 OR 条件组”入口，只能改预览文本，无法通过结构化编辑器新增条件组。

## 2. 根因

`frontend/src/filter-rule-editor.tsx` 实际已经渲染了 `新增 OR 条件组` 按钮，但 `frontend/src/style.css` 里存在一条全局规则：

```css
.text-button.secondary {
  display: none;
}
```

这条规则把所有二级文本按钮都隐藏了，连弹窗里的 `新增 OR 条件组`、`关闭`、设置里的 `重置` 一并吃掉。

## 3. 修复方案

不改弹窗逻辑，只把隐藏规则从全局缩到主界面过滤栏：保留主界面里原本被隐藏的二级按钮，同时恢复弹窗和设置面板中的二级按钮显示。

## 4. 改动文件清单

- `frontend/src/style.css`

## 5. 验证结果

- `frontend/npm run build` 通过
- 代码层确认：`RuleGroupsEditor` 中的 `新增 OR 条件组` 按钮仍存在，且不再被全局 `.text-button.secondary` 规则隐藏

## 6. 遗留事项

无。
