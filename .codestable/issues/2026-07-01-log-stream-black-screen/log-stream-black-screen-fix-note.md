---
doc_type: issue-fix
issue: log-stream-black-screen
date: 2026-07-01
status: fixed
tags: [frontend, wails, streaming, race]
---

## 现象
运行日志后，界面会先短暂出现日志，随后整页黑屏。

## 根因
`frontend/src/use-app-controller.ts` 的流式事件处理分成了两条路径：

- `state:updated` 先写入 `pendingEventStateRef`，等下一帧再 `setState`
- `state:append` 可能直接基于当前渲染态或旧的 `stateRef` 继续合并

高频日志流下，这两条路径会交替抢占“下一份基线状态”。即便前一轮已修掉 `dropped` 越界，组件顶层仍会在重新渲染时把 `stateRef` 回写成尚未追上最新 revision 的旧 state，随后 append patch 又从旧基线继续叠加，最终把前端事件流拖回不一致状态，表现为日志短暂刷出后整页黑屏。

## 修复
- 新增 `frontend/src/state-stream.ts`
  - 抽出 `applyStateAppendPatch`
  - 统一 append patch 的前端合并逻辑
- 修改 `frontend/src/use-app-controller.ts`
  - 把 `state:updated` / `state:append` 统一收敛到同一个 RAF 队列，只保留最新 revision 的待提交 state
  - `state:append` 不再直接 `setState`，而是始终基于“当前已知最新 state”合并后再排队提交
  - 组件顶层只在没有更高 revision 待提交时才回写 `stateRef`，避免把新基线覆盖回旧渲染态
- `mergeAppendedLogs` 改为与 Go 端 `mergePatchedLogs` 一致，对 `dropped` 做边界钳制

## 验证
- `npm --prefix frontend run build`
- `go test ./...`
