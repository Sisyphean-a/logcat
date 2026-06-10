---
doc_type: issue-fix
issue: device-hot-swap-context-retention
status: fixed
path: fast-track
fix_date: 2026-06-10
tags: [adb, hot-swap, package-selection, log-retention, frontend, wails]
---

# 设备热插拔上下文保留修复记录

## 1. 问题描述

设备在运行中被拔除后，当前日志、包选择和包下拉选项会一起被清空；随后插入新设备时，旧包名也不会跨设备恢复，用户需要重新输入包名，热插拔流程不连续。

## 2. 根因

- `internal/app/device_sync.go` 的断开分支直接复用了“清空绑定视图”逻辑，把日志和包上下文一并抹掉了。
- `internal/app/controller.go` 与 `internal/app/session_intent.go` 的恢复逻辑只允许恢复到同一个设备 ID，跨设备自动选中新设备时不会带回旧包绑定。
- `frontend/src/select-control.tsx` 只会展示存在于 `options` 里的值；即使后端保住了已选包名，只要新设备的包列表里没有它，界面也会显示成空。

## 3. 修复方案

- 把“断开时停止会话/清理运行态”和“断开时清空日志/包上下文”拆开，热插拔断开只清理运行态，保留现有日志和包选择。
- 允许带包绑定的恢复上下文跨设备迁移到新选中的设备，并在恢复绑定时保留日志。
- 选择器为“不在当前 options 里的已选值”补一条临时展示项，保证旧包名继续可见。

## 4. 改动文件清单

- `internal/app/binding.go`
- `internal/app/controller.go`
- `internal/app/controller_test.go`
- `internal/app/device_sync.go`
- `internal/app/session_intent.go`
- `internal/app/streaming_state.go`
- `frontend/src/select-control.tsx`

## 5. 验证结果

- `go test ./...`
- `frontend` 下 `npm run build`
- 新增后端回归测试覆盖：
  - 设备断开时保留日志和包上下文
  - 新设备接入时跨设备恢复包上下文且不清日志

## 6. 遗留事项

无。
