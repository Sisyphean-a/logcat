---
doc_type: issue-fix
issue: viewer-polish
status: fixed
tags: [ui, log-view, startup]
---

# Viewer Polish Fix Note

## 1. 修复范围

- 日志表格长 tag 截断，避免和消息列重叠
- 时间列只显示时分秒毫秒，不再显示日期
- 包名选择器支持输入筛选
- 默认进入页面不自动拉流，初始保持暂停，点击开始后才启动

## 2. 根因

- 表格列虽然用了 grid，但 tag 单元格没有单独截断保护
- 时间直接使用 threadtime 原始 `MM-DD HH:MM:SS.mmm`
- 包名选择器只有静态下拉，没有搜索能力
- `loadInitialState()` 与 `SelectDevice()` 会自动启动 session

## 3. 实际修改

- `frontend/src/app-shell.tsx`
  - 时间展示改成仅保留空格后的时间部分
  - tag 单元格加独立截断类
- `frontend/src/style.css`
  - 表格单元格增加 `min-width: 0`
  - 新增包名选择器搜索框样式
- `frontend/src/select-control.tsx`
  - 支持 `filterable`
  - 包下拉支持输入筛选
- `internal/app/model.go`
  - 初始状态改为暂停
- `internal/app/controller.go`
  - 设备选择默认只准备绑定，不自动开流
- `internal/app/binding.go`
  - 包/进程选择在未运行时只准备绑定
- `internal/app/logview.go`
  - 首次点击开始时按当前选择启动 session
- `app.go`
  - 启动只加载设备与包列表，不自动恢复 logcat

## 4. 验证

- `go test ./...`
- `frontend` 下 `npm run build`
