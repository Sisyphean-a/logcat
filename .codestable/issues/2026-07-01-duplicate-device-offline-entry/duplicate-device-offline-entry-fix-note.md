---
doc_type: issue-fix
issue: duplicate-device-offline-entry
status: fixed
path: fast-track
fix_date: 2026-07-01
tags: [adb, device-tracking, duplicate-device, offline, hot-swap]
---

# 重复设备离线条目导致选错设备 修复记录

## 1. 问题描述

同一台手机在设备下拉中同时出现一条 `OFFLINE` 和一条可用态记录；界面看起来可以切到第二条在线设备，但后端按设备 ID 选中时会命中前一条离线记录，导致运行按钮行为异常，热插拔后还可能触发黑屏式空白界面。

## 2. 根因

`internal/app/device_sync.go` 会把来自 `adb track-devices -l` / `adb devices -l` 的快照原样映射到 `model.Devices`。当同一序列号在短时间内同时出现 `offline` 与 `device` 两条记录时：

- 前端下拉两项都存在，但它们的 `value` 都是同一个 `device.id`
- 后端 `findDevice()` 按 ID 从前往后返回第一条匹配项，容易命中离线记录
- 结果是用户点了“在线设备”，实际选中的仍可能是离线态

## 3. 修复方案

在设备快照进入 UI 状态前先按 `ID` 去重；若同一序列号存在多条记录，优先保留 `device` 状态，其次保留更完整的 `model/transport` 信息。这样前后端都只会看到一条最终设备记录。

## 4. 改动文件清单

- `internal/app/device_sync.go`
- `internal/app/controller_test.go`

## 5. 验证结果

- `go test ./internal/app -run 'TestController(SyncDevicesDeduplicatesOfflineAndReadyEntries|ReconcileTrackedDevicesPromotesOfflineDeviceToReadySnapshot|SyncDevicesRestoresPackageContextAcrossReplacementDevice|SyncDevicesKeepsLogsAndPackageContextWhenDeviceBecomesUnavailable)' -v`
- `go test ./internal/adb -run 'TestTrackDevicesParsesLengthPrefixedSnapshots|TestTrackDevicesParsesMultipleSnapshotsSeparatedByNewline' -v`
- `go test ./...`

新增回归测试覆盖：同一序列号同时出现 `offline` / `device` 两条快照时，只保留一条 `device` 记录并自动选中该设备。

## 6. 遗留事项

本次未直接复现实机黑屏；先修复重复设备根因。若黑屏仍可复现，再单开 issue 跟前端渲染异常链路。
