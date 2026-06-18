---
doc_type: issue-fix
issue: saved-filter-clear-sync
date: 2026-06-18
status: fixed
tags: [frontend, backend, saved-filter, query]
---

## 现象
清空顶部“已保存过滤器”后，UI 里的过滤器名称会消失，但实际生效的筛选规则没有被清掉，导致底部日志总数持续增长而列表仍为空。用户还期望该清空动作能同步清掉下方包名和过滤规则输入。

## 根因
`ApplySavedFilter("")` 只调用了 `clearSavedFilterSelection`，原实现仅清空 `ActiveFilterID` 和错误信息，没有重置 `Filter.Draft`、`Filter.Applied`、已编译过滤条件，也没有重建 `VisibleLogs`。因此 UI 看起来已清空，但后端仍按旧 query 过滤。

## 修复
- [internal/app/filter.go](/E:/github/logcat/internal/app/filter.go:58) 在清空保存过滤器时改为走完整清理流程
- [internal/app/selection_clear.go](/E:/github/logcat/internal/app/selection_clear.go:53) 让 `clearSavedFilterSelection` 同步清空 draft/applied query、包名/进程选择，并在有设备上下文时恢复到设备级绑定
- [internal/app/saved_filter_management.go](/E:/github/logcat/internal/app/saved_filter_management.go:51) 管理弹窗把活动过滤器清空时复用同一条清理逻辑
- [frontend/src/use-app-controller.ts](/E:/github/logcat/frontend/src/use-app-controller.ts:557) 预览态对齐真实行为，清空保存过滤器时同步清包名和 query
- [internal/app/filter_update_test.go](/E:/github/logcat/internal/app/filter_update_test.go:176) 新增回归测试，覆盖“清空后恢复可见日志”和“清空后包名归零”

## 验证
- `go test ./internal/app -count=1`
- `go test ./... -count=1`
- `frontend/npm run build`
