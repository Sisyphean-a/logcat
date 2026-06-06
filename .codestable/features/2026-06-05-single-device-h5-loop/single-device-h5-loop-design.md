---
doc_type: feature-design
feature: 2026-06-05-single-device-h5-loop
requirement:
roadmap: logcat-viewer
roadmap_item: single-device-h5-loop
status: approved
summary: 用 Go + Gio 建立单设备 H5 logcat 查看最小闭环
tags: [go, gio, adb, logcat, mvp]
---

# single-device-h5-loop design

## 0. 术语约定

| 术语 | 定义 | 防冲突结论 |
| --- | --- | --- |
| ADB 安装 | 当前机器可执行的 `adb` 二进制及版本信息 | 仓库内无现有代码，沿用 [文档.md](</E:/github/logcat/文档.md:1>) 叫法 |
| 设备快照 | 一次 `adb devices -l` 的解析结果 | 与后续工作区快照区分，前者只描述在线设备 |
| 会话 | 一次从某设备读取 logcat 的运行实例 | 与“窗口 / Tab”分离，窗口后续可承载多个会话 |
| 可见缓冲 | 当前 UI 正在展示的日志集合 | 与暂停期间的后台缓冲区分 |
| H5 预设 | 内置的 `chromium` + `[H5]` 组合过滤规则 | 当前 feature 只支持内置预设，不开放自定义查询 |

## 1. 决策与约束

### 需求摘要

- **做什么**：提供一个可运行的 Go + Gio 桌面程序，能检测本机 `adb`、列出 Android 设备、选择单台设备并实时显示 `chromium` + `[H5]` 日志。
- **为谁做**：H5 / WebView 调试为主的 Android 开发者与测试人员。
- **成功标准**：
  - 应用启动后能显式展示 `adb` 检测结果；
  - 至少一台设备处于 `device` 状态时，用户可启动实时日志会话；
  - 只显示 `chromium` 且 message 含 `[H5]` 的日志；
  - 缺少 `adb`、设备未授权、设备离线时，错误在 UI 中直接可见。
- **明确不做**：
  - 不做包名选择、PID 过滤、前台应用识别；
  - 不做多 Tab / 多窗口 / 分屏；
  - 不做保存过滤器、查询语言、导出、崩溃识别；
  - 不做 `adb logcat -c` 真正清空设备缓冲。

### 复杂度档位

走桌面工具默认档位，无偏离。

### 关键决策

1. **先做单会话最小闭环，不在首个 feature 引入多窗口或多会话**
   - 否则 Gio 壳、ADB 运行时、会话状态、列表渲染会同时扩散，首批代码很难验证。
2. **会话过滤采取“ADB 端先 `chromium`，本地再 `[H5]`”**
   - 对应 [文档.md](</E:/github/logcat/文档.md:1>) 的 MVP 推荐命令，先减少传输噪声，再保留 H5 专属过滤。
3. **暂停 / 搜索 / 复制等交互留给后续 feature**
   - 本 feature 只保证采集主链路和错误态真正跑通，不把 UI 操作堆进第一轮。
4. **所有错误显式上浮到状态区**
   - 不做静默回退，不在 `adb` 缺失或设备异常时伪造空列表。

## 2. 名词与编排

### 2.1 名词层

#### 现状

无现状，全新模块。

#### 变化

1. **新增 `ADBInstall`**
   - 表达 `adb` 路径、版本、来源、server 状态。
2. **新增 `DeviceInfo`**
   - 表达设备 ID、型号、连接方式、状态。
3. **新增 `SessionConfig` / `SessionSnapshot`**
   - 分离会话输入参数与 UI 可见状态。
4. **新增 `LogEntry` / `WebViewConsoleInfo`**
   - 对齐 roadmap 契约，承载解析后的日志。
5. **新增 `ViewModel`**
   - 聚合 ADB 状态、设备列表、选中设备、会话快照、可见日志与问题信息。

#### 接口示例

```go
install, err := deviceSvc.DetectADB(ctx)
if err != nil {
	return Problem{Code: "adb_not_found", Message: err.Error()}
}

devices, err := deviceSvc.ListDevices(ctx)
if err != nil {
	return Problem{Code: "device_list_failed", Message: err.Error()}
}
```

```go
handle, err := supervisor.Start(ctx, SessionConfig{
	SessionID:       "default",
	DeviceID:        "emulator-5554",
	Buffers:         []string{"main"},
	Query:           "",
	UseChromiumOnly: true,
	UseH5Only:       true,
	PauseBufferCap:  10000,
})
```

### 2.2 编排层

```mermaid
flowchart TD
    A[应用启动] --> B[检测 adb]
    B -->|成功| C[加载设备列表]
    B -->|失败| X[状态区展示错误]
    C --> D[用户选择 device 状态设备]
    D --> E[启动 logcat 会话]
    E --> F[读取 threadtime 行]
    F --> G[解析为 LogEntry]
    G --> H[过滤 chromium + [H5]]
    H --> I[追加到可见缓冲]
    I --> J[Gio 列表重绘]
    E -->|设备离线/未授权| X
```

#### 现状

无现状，全新编排。

#### 变化

1. 新增应用启动编排：初始化 Gio 窗口后立即触发 `DetectADB -> ListDevices`。
2. 新增单设备会话编排：用户点击设备后，构造 `SessionConfig` 并启动唯一会话。
3. 新增事件驱动刷新：会话以事件流向 UI 推送 `entry / snapshot / problem`，UI 只消费状态，不直接读子进程。
4. 新增内置过滤流程：先在命令层只看 `chromium`，再在本地判定 message 是否含 `[H5]`。

#### 流程级约束

- `adb` 检测失败时，设备列表区不可伪装成“暂无设备”，必须展示错误原因。
- 只有 `device` 状态的设备允许启动会话；`offline` / `unauthorized` 只能进入错误提示。
- 会话停止前只能存在一个活跃 reader；切换设备必须先停旧会话再起新会话。
- 日志解析失败时保留原始文本并上报问题，不丢弃整行输入。

### 2.3 挂载点清单

- `cmd/logcatviewer/main.go`：新增桌面应用入口
- `internal/app/bootstrap`：新增启动时的 ADB 检测与设备拉取编排
- `internal/app/session`：新增单会话启动与事件消费挂入点
- `internal/ui/shell`：新增深色主题主界面与设备列表/日志列表双栏
- `internal/config/defaults`：新增默认主题与会话预设

### 2.4 推进策略

1. **桌面壳骨架**：先把 Gio 窗口、深色主题、设备区/日志区/状态区静态布局跑起来  
   退出信号：程序可启动，看到完整主界面骨架和默认状态文案
2. **ADB 检测与设备枚举编排**：接入 `DetectADB` 与 `ListDevices`，把结果投影到 UI  
   退出信号：可根据真实或测试 runner 展示 `adb` 状态和设备列表
3. **单会话采集骨架**：选择设备后启动唯一 logcat 会话，事件能从 runner 流向 UI  
   退出信号：选中设备后状态从 idle 变为 running，日志区收到事件
4. **解析与 H5 预设过滤**：实现 `threadtime` 解析、`chromium`/`[H5]` 双层过滤与可见缓冲  
   退出信号：示例日志中仅保留符合条件的 H5 行
5. **错误态与端到端收尾**：补齐 `adb` 缺失、设备未授权、设备离线的显式呈现与测试  
   退出信号：关键错误路径都能在 UI 状态区被观察到，测试通过

### 2.5 结构健康度与微重构

#### 评估

- 文件级：无现有源码文件，本 feature 不涉及修改旧文件
- 目录级：目标目录全为新建目录，当前仓库无平铺拥挤问题

#### 结论：不做

本 feature 从零起步，不需要先做“只搬不改行为”的微重构。

## 3. 验收契约

### 关键场景清单

1. **启动成功路径**
   - 输入 / 触发：机器存在可执行 `adb`，启动应用
   - 期望可观察结果：状态区显示 `adb` 版本，设备区出现 `adb devices -l` 解析后的列表
2. **设备启动会话**
   - 输入 / 触发：至少一台设备状态为 `device`，用户点击该设备
   - 期望可观察结果：状态区显示 `running`，日志区开始追加解析后的日志
3. **H5 过滤**
   - 输入 / 触发：logcat 中同时出现普通 `chromium` 行与含 `[H5]` 的 `chromium` 行
   - 期望可观察结果：日志区只显示含 `[H5]` 的条目
4. **ADB 缺失**
   - 输入 / 触发：系统无 `adb` 或配置路径无效，启动应用
   - 期望可观察结果：状态区显示 `adb_not_found` 类错误，设备区不可点击
5. **设备未授权 / 离线**
   - 输入 / 触发：设备状态为 `unauthorized` 或 `offline`
   - 期望可观察结果：设备区显示对应状态，点击后状态区展示明确问题，不启动会话
6. **解析失败保底**
   - 输入 / 触发：收到非标准 `threadtime` 行
   - 期望可观察结果：问题区记录解析失败，原始行仍可被追踪，不导致会话崩溃

### 明确不做的反向核对项

- 代码中不应出现包名选择器或前台应用识别按钮
- 代码中不应出现多 Tab / 多窗口容器
- 代码中不应调用 `adb logcat -c` 清空设备日志
- 代码中不应保存自定义过滤器或查询历史

## 4. 与项目级架构文档的关系

本 feature 验收后，需要把以下内容回填到架构层：

- **名词**：`ADBInstall`、`DeviceInfo`、`SessionConfig`、`LogEntry`
- **动词骨架**：启动检测、设备选择、单会话采集、事件流刷新
- **流程级约束**：显式错误上浮、单活跃会话、H5 双层过滤策略

当前仓库尚无系统级架构文档细分条目；验收时至少应补 `ARCHITECTURE.md` 的模块索引与主流程说明。
