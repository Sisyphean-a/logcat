---
doc_type: issue-fix
issue: device-hot-swap-regression
status: fixed
path: fast-track
fix_date: 2026-06-10
tags: [adb, device-tracking, hot-swap, wails, logcat]
---

# 设备热插拔回归修复记录

## 1. 问题描述

现代界面在已连接设备拔除后，没有自动暂停并移除当前设备；随后插入新设备时，设备列表也不再刷新，无法自动获取新设备并恢复调试。

## 2. 根因

`internal/adb/device_tracker.go` 按 4 字节十六进制头读取 `adb track-devices -l` 的帧，但没有跳过每帧后额外输出的 `LF` 分隔符。首帧读完后，下一轮会把这个换行和后续 3 个字节一起当成头部解析，触发 `invalid adb payload header`，设备跟踪协程因此提前退出，后续拔插事件全部丢失。

## 3. 修复方案

在读取下一帧之前，先跳过连续的 `CR/LF` 分隔符，保证头部读取总是从真正的 4 字节十六进制长度开始；同时补一条多帧回归测试，覆盖“有设备 -> 拔空 -> 换新设备”这条热插拔链路。

## 4. 改动文件清单

- `internal/adb/device_tracker.go`
- `internal/adb/service_test.go`

## 5. 验证结果

- `go test ./internal/adb ./internal/app`
- `go test ./...`
- 新增测试覆盖连续三帧：`device-1`、空快照、`device-2`，确认跟踪不会在首帧后退出

## 6. 遗留事项

无。
