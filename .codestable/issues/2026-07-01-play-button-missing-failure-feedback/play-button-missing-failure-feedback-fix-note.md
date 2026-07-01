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

问题有两层：

1. 前端一直用 `pause.active` 推断工具栏该显示开始还是暂停，但这只表示“当前是否处于暂停展示态”，不等于“当前是否真的存在活跃日志会话”。
2. 后端 session 实际结束时，没有把当前活跃 session 标记清掉；这样 `hasActiveSession()` 还会继续返回 true，恢复逻辑会误以为还有一条可继续的会话。

这会造成两种错误表现：

- 热插拔后，界面把“无活跃会话”误判成“正在运行”
- 用户点击播放时，恢复逻辑可能只是在一条已失效会话上切换暂停态，而不会真正重启 session

## 3. 修复方案

- 在 `UISnapshot` / `AppState` 中显式暴露 `sessionActive`
- session 消费协程结束时，若它仍是当前活跃 session，则同步清掉活跃标记
- 工具栏根据 `sessionActive + pause.active` 判断显示开始还是暂停
- 点击播放时，若本次恢复后仍未建立活跃会话，则在顶部直接提示失败原因，而不是只把信息挤在底部状态栏

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
- `frontend/src/style.css`
- `frontend/wailsjs/go/models.ts`

## 5. 验证结果

- `go test ./internal/app -run 'TestController(SelectDeviceDoesNotStartSessionUntilResume|ResumeKeepRestartsWhenPreviousSessionAlreadyExited|SyncDevicesRestoresPackageContextAcrossReplacementDevice|SyncDevicesKeepsLogsAndPackageContextWhenDeviceBecomesUnavailable|SyncDevicesDeduplicatesOfflineAndReadyEntries|ReconcileTrackedDevicesPromotesOfflineDeviceToReadySnapshot)' -v`
- `go test . -run 'TestNewAppState(SelectedLogOmitsRawPayload|IncludesSessionActive)' -v`
- `go test ./...`
- `npm --prefix frontend run build`

## 6. 遗留事项

- 本地缺少 `codestable-worktree-gate.py`，本次无法执行 start / commit gate；已按现有仓库工具集完成代码与测试验证。
