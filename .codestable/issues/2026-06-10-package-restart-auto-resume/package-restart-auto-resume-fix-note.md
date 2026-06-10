---
doc_type: issue-fix
issue: package-restart-auto-resume
path: fast-track
fix_date: 2026-06-10
tags: [adb, logcat, binding-watcher, pid-rebind, wails]
---

# 应用重启后监控未自动恢复 修复记录

## 1. 问题描述

设备已连接、已绑定目标包并处于运行监控状态时，手动关闭目标应用再重新打开，日志列表不再继续刷新；但点击一次暂停再恢复后，积压日志会瞬间补到 UI，说明问题出在自动恢复链路，而不是日志源完全断开。

## 2. 根因

`internal/app/binding_watcher.go` 在 watcher 轮询到“目标进程消失”时，会通过 `applyWatcherStoppedBinding` 停掉当前 session；等进程重新出现后，同一轮询链路又因为 `hasActiveSession()==false` 走进 `applyWatcherPreparedBinding`，只更新了绑定 PID 和状态，没有按照既有的 `sessionIntentRunning` 意图重新启动 logcat session，所以 UI 后续收不到新事件，直到用户手动点一次恢复。

## 3. 修复方案

让 watcher 在“当前没有活跃 session，但运行意图仍是 running，且新的 PID 已经出现”时，直接走 `applyWatcherRunningBinding` 重启 session，而不是只做 prepared binding；同时补一条回归测试覆盖“进程消失 -> 重新拉起 -> 自动续监控”。

## 4. 改动文件清单

- `internal/app/binding_watcher.go`
- `internal/app/controller_test.go`

## 5. 验证结果

- `go test ./internal/app`
- `go test ./...`
- 新增回归测试覆盖：绑定包后首次 PID 为 `111`，随后 watcher 观察到进程消失，再观察到新 PID `222` 出现时，会自动重建 session，而不是停留在仅更新绑定信息的假恢复状态

## 6. 遗留事项

无。
