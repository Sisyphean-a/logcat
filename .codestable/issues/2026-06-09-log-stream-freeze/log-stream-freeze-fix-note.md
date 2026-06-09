---
doc_type: issue-fix
issue: log-stream-freeze
status: fixed
tags: [performance, wails, logcat]
---

# Log Stream Freeze Fix Note

## 1. 修复范围

- 后端日志接入改为增量更新，不再对每条新日志全量重建 `VisibleLogs`
- Wails 状态推送改为基于 revision 的脏检查，并只推送尾部窗口日志
- 前端日志表格改为虚拟渲染，避免大量 DOM 同时挂载

## 2. 根因兑现

- `internal/app/logview.go` 原先每条日志都会 `rebuildVisibleFromAllLogsLocked()`，高频流量下退化成 O(N^2)
- `app.go` 原先固定每 250ms 深拷贝整份模型并向前端推送全量表格快照
- `frontend/src/app-shell.tsx` 原先直接 `logs.map(...)` 渲染全部日志行

## 3. 实际修改

- `internal/app/logview.go`
  - `pushEntry` 改为增量 `appendVisibleLogLocked`
  - 保留全量 rebuild 仅给过滤器/恢复等重算场景
- `internal/app/controller.go`
  - 新增 `revision`
- `internal/app/revision.go`
  - 新增统一脏标记
- `internal/app/ui_snapshot.go`
  - 新增 UI 快照窗口化，只向前端暴露尾部窗口
- `app.go`
  - 状态循环改为按 revision 发事件
- `app_state.go`
  - DTO 增加 `visibleStart`
- `frontend/src/app-shell.tsx`
  - 日志表改为固定行高虚拟渲染
- `frontend/src/use-app-controller.ts`
  - 增加滚动位置与视口高度状态

## 4. 验证

- `go test ./...`
- `frontend` 下 `npm run build`

## 5. 结果

- 大量日志流入时，后端不再每条日志重扫全量列表
- Wails 不再固定频率推送重复全量状态
- 前端主表不会因为日志数变大而一次性创建海量行节点
