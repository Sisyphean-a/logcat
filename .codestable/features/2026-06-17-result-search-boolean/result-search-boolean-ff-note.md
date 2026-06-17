---
doc_type: feature-ff-note
feature: result-search-boolean
date: 2026-06-17
requirement: h5-logcat-viewing
tags: [frontend, go, search, viewer]
---

## 做了什么
给“搜索结果（message / tag）”这层接入了多关键词布尔语法，支持 `&&`、`||` 和 `-`，用于表达“必须同时包含”“包含任意一个”“排除某词”。
这套语法只作用于当前结果搜索，不影响前面的过滤器查询语言。

## 改了哪些
- `internal/app/search_match.go` — 新增结果搜索表达式编译与匹配逻辑，保留无操作符时的原始子串搜索
- `internal/app/logview.go`、`internal/app/controller.go` — controller 改为使用编译后的结果搜索匹配 visible logs 与增量日志
- `frontend/src/use-app-controller.ts`、`frontend/src/log-text.tsx` — 预览模式与前端高亮改成按正向关键词处理布尔搜索
- `frontend/src/app-shell.tsx`、`frontend/src/App.tsx`、`frontend/src/log-row.tsx`、`frontend/src/log-table.tsx` — 更新输入提示并把高亮参数从整串 query 改为关键词集合
- `internal/app/search_match_test.go` — 补布尔搜索单测覆盖 `&&`、`||`、`-`、高亮词提取

## 怎么验证的
运行了 `go test ./internal/app`、`go test ./...`、`go build .`。
运行了 `npm --prefix frontend run build`，前端生产构建通过。
