# Logcat Viewer

一个基于 `Wails + Go + React` 的桌面端 Android Logcat 查看器，重点服务 H5 / WebView 调试场景。

当前版本已经具备可用的实时查看、包名绑定、查询过滤、保存过滤器、详情解析和导出能力，目标是提供一个比直接跑 `adb logcat` 更易用的本地桌面工具。

## 已实现功能

- 自动检测本机 `adb`，启动时读取版本并加载设备列表。
- 监听设备变化，设备插拔或状态变化后自动刷新可选设备。
- 支持按设备查看日志，自动选中首个可用设备。
- 支持包名范围筛选：`全部包`、`用户包`、`系统包`。
- 选择包名后自动绑定该应用相关进程的 PID；应用重启后会轮询并自动重新绑定。
- 实时读取 `adb -s <device> logcat -v threadtime` 输出并持续刷新界面。
- 支持暂停/恢复采集视图；暂停期间会缓存新日志，恢复时可继续显示。
- 支持清空当前视图，不会修改设备上的 log buffer。
- 支持导出当前可见日志为 `TSV`，默认输出到 `Downloads` 目录。
- 支持复制选中日志的展示文本、原始日志和纯消息文本。
- 支持日志详情面板，显示时间、级别、标签、来源、消息解析和原始日志。
- 对日志详情中的 JSON、URL、常见错误关键字、HTTP 方法、堆栈帧做了额外解析和高亮。
- 对 Chromium / WebView console 日志自动提取 `message` 和 `source`。
- 支持保存过滤器、编辑过滤器、排序、删除、设置默认过滤器。
- 已持久化保存过滤器、默认过滤器和查询历史。
- 日志列表使用虚拟滚动，适合持续流式日志场景。
- 支持界面字号设置，并保存在本地配置中。

## 查询语法

顶部筛选框支持逻辑表达式，当前已实现这些能力：

- 字段：`level`、`tag`、`message`、`package`
- 运算：`:` / `=`、`~:` / `~=`、前缀 `-` 取反
- 逻辑：`&` / `&&`、`|` / `||`、`()`
- 值支持双引号
- 不带字段的裸词，按消息文本大小写不敏感包含匹配

示例：

```text
level:E
tag:"chromium" && message~:"[H5]"
package:com.demo.app && -message~:"vite"
(tag~:"bridge" || message~:"jsbridge") && level:W
```

## 适用场景

- 在电脑上实时查看 Android 设备日志
- 只盯某个 App 的日志，不想手工维护 PID
- 查看 WebView / H5 `console` 输出
- 过滤 H5 关键字、接口错误、Bridge 调用等噪声
- 导出当前过滤结果发给同事复现问题

## 技术栈

- 后端：Go 1.26
- 桌面容器：Wails v2
- 前端：React 18 + TypeScript + Vite
- 数据来源：本机 `adb`

## 运行前提

启动前需要本机满足：

- 已安装 `adb`
- `adb` 已加入 `PATH`
- 已安装 Go
- 已安装 Node.js / npm
- 已安装 Wails CLI

可用下面的命令安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 本地开发

安装前端依赖：

```bash
cd frontend
npm install
cd ..
```

启动开发模式：

```bash
wails dev
```

生产构建：

```bash
wails build
```

## 使用方式

1. 连接 Android 设备，并确认 `adb devices -l` 能看到设备。
2. 启动应用后，在顶部选择设备。
3. 按需切换包范围，然后在筛选栏选择包名。
4. 在查询框输入规则并点击“应用”。
5. 需要复用时，将当前规则保存为过滤器。
6. 选中某条日志后，可在右侧详情面板查看解析结果或复制内容。
7. 需要分享时，点击导出按钮生成 `TSV` 文件。

## 本地数据位置

- 保存过滤器：系统配置目录下的 `logcat-viewer/saved-filters.json`
- 导出日志：用户 `Downloads` 目录

## 项目结构

```text
.
├─ main.go                  Wails 启动入口
├─ app.go                   Wails 绑定方法与状态推送
├─ internal/
│  ├─ adb/                  adb 检测、设备/包名/进程/logcat 读取
│  ├─ app/                  应用状态、筛选、会话绑定、暂停恢复
│  ├─ logcat/               threadtime 解析与 chromium source 提取
│  ├─ session/              logcat 会话监督与事件转发
│  └─ storage/              过滤器持久化与日志导出
└─ frontend/src/            React 界面
```

## 当前边界

- 当前是单窗口、单主视图形态，还没有多标签、多分屏。
- UI 目前主要暴露设备、包名、查询、保存过滤器、详情查看等能力。
- 依赖本机 `adb`，项目本身不内置 Android Platform Tools。
- 清空操作只清空本地视图，不会执行 `adb logcat -c`。

