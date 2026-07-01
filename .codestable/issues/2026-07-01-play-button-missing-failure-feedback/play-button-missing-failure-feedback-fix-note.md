---
doc_type: issue-fix
issue: play-button-missing-failure-feedback
status: fixed
path: fast-track
fix_date: 2026-07-01
tags: [frontend, wails, device-tracking, session-state, status-feedback]
---

# 开始按钮缺少失败反馈 修复记录

## 1. 问题描述

设备断开再重连后，工具栏会把“开始/暂停”入口维持在错误状态。应用未真正恢复会话时，界面仍可能显示暂停态入口，用户看起来像是“没有对应的开始按钮”或“点了也不知道为什么起不来”。

## 2. 根因

前端一直用 `pause.active` 推断工具栏该显示开始还是暂停，但这只表示“当前是否处于暂停展示态”，不等于“当前是否真的存在活跃日志会话”。

热插拔后如果后端保留了运行意图、但暂时没有活跃会话（例如应用未运行，`app_not_running`），会出现：

- `pause.active = false`
- 实际 `sessionCancel == nil`

这会让前端把“无活跃会话”误判成“正在运行”，从而把开始入口和点击行为都导向错误分支。

## 3. 修复方案

- 在 `UISnapshot` / `AppState` 中显式暴露 `sessionActive`
- 工具栏根据 `sessionActive + pause.active` 判断显示开始还是暂停
- 点击播放/暂停按钮时，前端优先按 `sessionActive` 选择 `ResumeKeep()` 还是 `Pause()`
- 状态栏把常见机器状态码翻译成可读提示，明确告诉用户为什么当前无法开始

## 4. 改动文件清单

- `internal/app/ui_snapshot.go`
- `app_state.go`
- `app_state_patch.go`
- `app_state_test.go`
- `internal/app/controller_test.go`
- `frontend/src/toolbar.tsx`
- `frontend/src/use-app-controller.ts`
- `frontend/src/App.tsx`
- `frontend/src/mock-state.ts`
- `frontend/wailsjs/go/models.ts`

## 5. 验证结果

- `go test ./internal/app -run 'TestController(SelectDeviceDoesNotStartSessionUntilResume|SyncDevicesRestoresPackageContextAcrossReplacementDevice|SyncDevicesKeepsLogsAndPackageContextWhenDeviceBecomesUnavailable|SyncDevicesDeduplicatesOfflineAndReadyEntries|ReconcileTrackedDevicesPromotesOfflineDeviceToReadySnapshot)' -v`
- `go test . -run 'TestNewAppState(SelectedLogOmitsRawPayload|IncludesSessionActive)' -v`
- `go test ./...`
- `npm --prefix frontend run build`

## 6. 遗留事项

- 本地缺少 `codestable-worktree-gate.py`，本次无法执行 start / commit gate；已按现有仓库工具集完成代码与测试验证。
