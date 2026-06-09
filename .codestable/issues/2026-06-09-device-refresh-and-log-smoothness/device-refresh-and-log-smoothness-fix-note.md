---
doc_type: issue-fix
issue: device-refresh-and-log-smoothness
status: fixed
tags: [adb, device-tracking, performance, wails, logcat]
---

# Device Refresh And Log Smoothness Fix Note

## 1. 修复范围

- 设备接入/断开后，设备列表不再依赖手动刷新或重启应用
- 当前选中设备失效时，自动清理绑定和会话；新可用设备接入时自动选中
- 日志状态推送从固定 250ms 轮询改为脏信号驱动的短节流推送，减少明显批量刷新感

## 2. 根因兑现

- `internal/app/controller.go` 的 `Load()` 只在启动时调用一次 `ListDevices()`，后续没有任何设备变化跟踪
- `app.go` 的 `pushStateLoop()` 使用固定 `250ms` ticker，导致日志 UI 更新天然呈现为一批一批推送
- 设备列表状态与选中设备/绑定状态没有统一同步逻辑，当前设备断开后不会主动收敛运行态

## 3. 实际修改

- `internal/adb/service.go`
  - 让 `Service` 同时持有 `PipeRunner`，支持长连接式 adb 设备跟踪
- `internal/adb/device_tracker.go`
  - 新增 `TrackDevices()`，基于 `adb track-devices -l` 读取 length-prefixed 快照流
- `internal/app/controller.go`
  - 新增 `DeviceTracker` 协议和 `Dirty()` 脏信号出口
  - `Load()` 改为统一走设备快照同步
- `internal/app/device_sync.go`
  - 新增设备快照同步、自动选中、失效设备清理、设备更新消费逻辑
- `internal/app/revision.go`
  - `markDirtyLocked()` 除 revision 外，额外发出非阻塞脏信号
- `app.go`
  - 启动时并行开启设备跟踪
  - 状态推送改为“脏信号 + 16ms 节流”模型，不再固定 250ms 轮询
- `internal/adb/service_test.go`
  - 补 `track-devices` 协议解析测试
- `internal/app/controller_test.go`
  - 补设备自动选中与断开清理测试

## 4. 验证

- `go test ./...`
- `frontend` 下 `npm run build`

## 5. 结果

- 手机插上后，设备列表会立即跟随 adb 更新；如果当前没有选中设备，会自动选中首个 `device` 状态设备
- 当前选中设备断开后，会主动停止会话、清空绑定并回到空闲态，不再保留陈旧运行状态
- 日志刷新的节奏从粗颗粒定时批推变成更细的事件驱动推送，观感上更连续
