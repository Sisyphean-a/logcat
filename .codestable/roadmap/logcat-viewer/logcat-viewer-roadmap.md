---
doc_type: roadmap
slug: logcat-viewer
status: active
created: 2026-06-05
last_reviewed: 2026-06-06
tags: [go, gio, adb, logcat, h5]
related_requirements: [h5-logcat-viewing]
related_architecture: [single-device-logcat-loop]
---

# Go + Gio Logcat Viewer

## 1. 背景

项目目标是做一个面向 H5 / WebView / Android 调试的桌面日志工具，替代命令行 `adb logcat` 的低效体验，并逐步覆盖 Android Studio Logcat 的核心能力。

当前仓库已经完成前三条子 feature，具备单设备 H5 查看、基础筛查与包/PID 绑定闭环；后续路线图的重点转为高级查询、导出、多会话和扩展能力。路线图仍需要保留整套需求拆分与跨模块契约，避免后续每条 feature 各自发明接口。

## 2. 范围与明确不做

### 本 roadmap 覆盖

- Go + Gio 桌面端应用骨架
- ADB 发现、设备管理、Wi-Fi ADB、包名/进程解析
- Logcat 实时采集、解析、过滤、查询语言、崩溃识别
- 多标签 / 多窗口 / 多设备 / 工作区 / 导出 / 脱敏
- VSCode 联动、AI 分析、统计、回放、插件化扩展

### 明确不做

- Android 设备端 Agent 或常驻服务
- 非 Android 日志协议接入（如 iOS、浏览器 DevTools 原生协议）
- Android Studio 工程模型解析与 IDE 深度嵌入
- 超出 [文档.md](</E:/github/logcat/文档.md:1>) 的云端同步、账号系统、在线协作后台

## 3. 模块拆分（概设）

```text
Logcat Viewer
├── desktop-shell：Gio 窗口、布局、交互、主题、多 Tab / 多窗口
├── adb-runtime：ADB 定位、设备枚举、Wi-Fi 连接、包名与进程发现
├── session-pipeline：logcat 采集、会话状态、自动重连、缓冲与背压
├── log-intelligence：threadtime 解析、查询语言、H5 / WebView / 崩溃增强识别
├── workspace-storage：设置、过滤器、历史、工作区、导出与脱敏
└── extensions：VSCode 跳转、AI 分析、统计、回放、插件接口
```

### 模块 A · desktop-shell

- **职责**：承载 Gio 应用入口、窗口生命周期、主界面布局、列表/详情双栏、Tab/窗口/分屏、快捷键与主题。
- **承载的子 feature**：`single-device-h5-loop`、`viewer-controls-and-search`、`multi-session-wifi-and-layout`
- **触碰的现有代码 / 模块**：全新

### 模块 B · adb-runtime

- **职责**：屏蔽 `adb` 命令细节，提供 ADB 检测、设备状态、Wi-Fi 连接、包名/前台应用/PID 能力。
- **承载的子 feature**：`single-device-h5-loop`、`package-pid-and-process-sync`、`multi-session-wifi-and-layout`
- **触碰的现有代码 / 模块**：全新

### 模块 C · session-pipeline

- **职责**：管理 logcat 子进程、缓冲、暂停/恢复、自动滚动、自动重连、设备/进程切换与多会话并发。
- **承载的子 feature**：`single-device-h5-loop`、`viewer-controls-and-search`、`package-pid-and-process-sync`、`multi-session-wifi-and-layout`
- **触碰的现有代码 / 模块**：全新

### 模块 D · log-intelligence

- **职责**：解析 `threadtime` 日志、执行本地过滤与查询语言、识别 WebView console、JSON、URL、堆栈、崩溃类型。
- **承载的子 feature**：`single-device-h5-loop`、`query-language-and-saved-filters`、`rich-log-details-and-crash-detection`
- **触碰的现有代码 / 模块**：全新

### 模块 E · workspace-storage

- **职责**：持久化设置、过滤器、查询历史、工作区、导出文件、脱敏规则与恢复状态。
- **承载的子 feature**：`query-language-and-saved-filters`、`export-workspace-and-settings`
- **触碰的现有代码 / 模块**：全新

### 模块 F · extensions

- **职责**：定义编辑器跳转、AI 分析、统计、回放、团队共享过滤器、远程采集、插件加载的扩展边界。
- **承载的子 feature**：`editor-ai-and-plugin-extensions`
- **触碰的现有代码 / 模块**：全新

## 4. 模块间接口契约 / 共享协议（架构层详设）

### 4.1 ADB 发现与设备协议

**方向**：desktop-shell / session-pipeline → adb-runtime  
**形式**：Go 接口调用

**契约**：

```go
type ADBInstall struct {
	Path         string
	Version      string
	Source       string
	ServerStatus string
}

type DeviceStatus string

const (
	DeviceReady        DeviceStatus = "device"
	DeviceOffline      DeviceStatus = "offline"
	DeviceUnauthorized DeviceStatus = "unauthorized"
	DeviceNoPermission DeviceStatus = "no permissions"
)

type DeviceInfo struct {
	ID        string
	Model     string
	Transport string
	Status    DeviceStatus
}

type PackageScope string

const (
	PackageScopeUser   PackageScope = "user"
	PackageScopeSystem PackageScope = "system"
	PackageScopeAll    PackageScope = "all"
)

type PackageInfo struct {
	Name string
}

type ProcessInfo struct {
	PID  int
	Name string
}

type DeviceService interface {
	DetectADB(ctx context.Context) (ADBInstall, error)
	ListDevices(ctx context.Context) ([]DeviceInfo, error)
	ConnectWiFi(ctx context.Context, endpoint string) error
	DisconnectWiFi(ctx context.Context, endpoint string) error
	ListPackages(ctx context.Context, deviceID string, scope PackageScope) ([]PackageInfo, error)
	CurrentForegroundPackage(ctx context.Context, deviceID string) (string, error)
	ListProcesses(ctx context.Context, deviceID, packageName string) ([]ProcessInfo, error)
}
```

**错误语义**：

- `adb_not_found`：未检测到 `adb` 或配置路径无效
- `device_unauthorized`：设备未授权
- `device_offline`：设备离线
- `wifi_connect_failed`：`adb connect` 失败

### 4.2 会话采集协议

**方向**：desktop-shell → session-pipeline → adb-runtime / log-intelligence  
**形式**：Go 接口调用 + 事件流

**契约**：

```go
type SessionStatus string

const (
	SessionRunning SessionStatus = "running"
	SessionPaused  SessionStatus = "paused"
	SessionStopped SessionStatus = "stopped"
	SessionError   SessionStatus = "error"
)

type SessionConfig struct {
	SessionID       string
	DeviceID        string
	PackageName     string
	ProcessName     string
	PID             int
	Buffers         []string
	Query           string
	UseChromiumOnly bool
	UseH5Only       bool
	PauseBufferCap  int
}

type SessionSnapshot struct {
	Status        SessionStatus
	VisibleCount  int
	BufferedCount int
	DroppedCount  int
	AutoScroll    bool
}

type SessionEventKind string

const (
	SessionEventEntry    SessionEventKind = "entry"
	SessionEventSnapshot SessionEventKind = "snapshot"
	SessionEventProblem  SessionEventKind = "problem"
)

type SessionEvent struct {
	Kind     SessionEventKind
	Entry    *LogEntry
	Snapshot *SessionSnapshot
	Problem  *Problem
}

type SessionHandle interface {
	Events() <-chan SessionEvent
	Pause()
	Resume()
	ClearVisible()
	Stop()
}

type SessionSupervisor interface {
	Start(ctx context.Context, cfg SessionConfig) (SessionHandle, error)
}
```

**约束**：

- 暂停只停止 UI 追加，不杀掉 logcat 进程
- `ClearVisible()` 只清界面缓存，不执行 `adb logcat -c`
- 进程重启后允许更新 PID，但必须通过 `SessionEventProblem` 或状态区可观察

### 4.3 日志解析与查询协议

**方向**：session-pipeline → log-intelligence  
**形式**：Go 接口调用

**契约**：

```go
type LogLevel string

const (
	LogVerbose LogLevel = "V"
	LogDebug   LogLevel = "D"
	LogInfo    LogLevel = "I"
	LogWarn    LogLevel = "W"
	LogError   LogLevel = "E"
	LogFatal   LogLevel = "F"
)

type WebViewConsoleInfo struct {
	ConsoleLevel   string
	LineNumber     int
	SourceURL      string
	SourceLine     int
	ConsoleMessage string
}

type LogEntry struct {
	ID          string
	DeviceID    string
	Timestamp   time.Time
	TimeText    string
	PID         int
	TID         int
	Level       LogLevel
	Tag         string
	PackageName string
	ProcessName string
	Message     string
	Raw         string
	Buffer      string
	WebView     *WebViewConsoleInfo
	StackTrace  []string
}

type QueryCompiler interface {
	Compile(input string) (CompiledQuery, error)
}

type CompiledQuery struct {
	Raw      string
	Terms    []QueryTerm
	Negated  bool
	Logic    string
	HasRegex bool
}

type QueryMatcher interface {
	Match(entry LogEntry, query CompiledQuery) bool
}
```

**约束**：

- `threadtime` 是统一输入格式；不支持时必须显式报错，不静默降级
- WebView console 解析失败时保留原始 message，不吞日志
- 查询编译失败要返回可展示错误，不能假装空结果

### 4.4 工作区与导出协议

**方向**：desktop-shell → workspace-storage / extensions  
**形式**：Go 接口调用

**契约**：

```go
type SavedFilter struct {
	ID          string
	Name        string
	Query       string
	Level       string
	Tags        []string
	PackageName string
	Color       string
	Pinned      bool
}

type WorkspaceSnapshot struct {
	ActiveSessions []SessionConfig
	SavedFilters   []SavedFilter
	QueryHistory   []string
	LastDeviceID   string
}

type ExportRequest struct {
	SessionID      string
	Entries        []LogEntry
	MaskSensitive  bool
	BundleWithMeta bool
}

type WorkspaceStore interface {
	Load(ctx context.Context) (WorkspaceSnapshot, error)
	Save(ctx context.Context, snapshot WorkspaceSnapshot) error
	Export(ctx context.Context, req ExportRequest) (string, error)
}
```

**约束**：

- 工作区只保存用户态配置，不保存 adb 凭据
- 导出脱敏规则必须是显式可配置的，默认行为在 UI 中可见

### 4.5 扩展协议

**方向**：desktop-shell / workspace-storage → extensions  
**形式**：Go 插件边界

**契约**：

```go
type ExtensionDescriptor struct {
	ID          string
	Name        string
	Capabilities []string
}

type ExtensionHost interface {
	OpenFile(path string, line int) error
	Analyze(entries []LogEntry) (string, error)
	PublishFilter(filter SavedFilter) error
}

type Extension interface {
	Descriptor() ExtensionDescriptor
	Bind(host ExtensionHost) error
}
```

**约束**：

- 插件只通过 `ExtensionHost` 与宿主交互，不直接触碰 Gio UI 状态
- AI 分析失败必须可见，不允许伪造成功结果

## 5. 子 feature 清单

1. **single-device-h5-loop** — 单窗口深色主题下完成 ADB 检测、设备选择、实时 `chromium` + `[H5]` 日志查看的最小闭环
   - 所属模块：desktop-shell、adb-runtime、session-pipeline、log-intelligence
   - 依赖：无
   - 状态：done
   - 对应 feature：2026-06-05-single-device-h5-loop
   - 备注：MVP 最小闭环
2. **viewer-controls-and-search** — 补齐暂停/恢复、清空视图、复制日志、基础搜索、自动滚动与详情面板
   - 所属模块：desktop-shell、session-pipeline
   - 依赖：single-device-h5-loop
   - 状态：done
   - 对应 feature：2026-06-06-viewer-controls-and-search
3. **package-pid-and-process-sync** — 加入包名选择、前台应用识别、PID/进程过滤与应用重启自动重绑
   - 所属模块：adb-runtime、session-pipeline
   - 依赖：single-device-h5-loop
   - 状态：done
   - 对应 feature：2026-06-06-package-pid-and-process-sync
4. **query-language-and-saved-filters** — 实现高级查询语言、Level/Tag/Message/Age/Regex/否定逻辑、保存过滤器与历史
   - 所属模块：log-intelligence、workspace-storage、desktop-shell
   - 依赖：viewer-controls-and-search, package-pid-and-process-sync
   - 状态：in-progress
   - 对应 feature：2026-06-06-query-language-and-saved-filters
5. **rich-log-details-and-crash-detection** — WebView console 结构化、JSON/URL 识别、多行堆栈、崩溃/ANR/Native Crash 识别
   - 所属模块：log-intelligence、desktop-shell
   - 依赖：single-device-h5-loop
   - 状态：planned
   - 对应 feature：未启动
6. **export-workspace-and-settings** — 导出、日志包、工作区、设置、脱敏和恢复状态
   - 所属模块：workspace-storage、desktop-shell
   - 依赖：query-language-and-saved-filters, rich-log-details-and-crash-detection
   - 状态：planned
   - 对应 feature：未启动
7. **multi-session-wifi-and-layout** — 多 Tab / 多窗口 / 分屏、多设备并行、Wi-Fi ADB、自动重连
   - 所属模块：desktop-shell、adb-runtime、session-pipeline
   - 依赖：viewer-controls-and-search, package-pid-and-process-sync, export-workspace-and-settings
   - 状态：planned
   - 对应 feature：未启动
8. **editor-ai-and-plugin-extensions** — VSCode 跳转、source 映射、AI 分析、统计图、日志回放、共享过滤器、远程采集、插件接口
   - 所属模块：extensions、workspace-storage、log-intelligence
   - 依赖：multi-session-wifi-and-layout, rich-log-details-and-crash-detection, export-workspace-and-settings
   - 状态：planned
   - 对应 feature：未启动

**当前闭环**：前 3 条 feature 做完后，已经可以在桌面端启动程序、检测本机 `adb`、选择一台 Android 设备、实时查看 `chromium` 与 `[H5]` 日志，并在当前视图内做暂停、恢复、搜索、复制、详情查看、包范围切换、前台 App 一键选择、包/进程绑定和 PID 自动重绑。

## 6. 排期思路

先做最窄的单设备 H5 调试闭环，把 Gio 窗口、ADB 检测、实时采集与基础解析真正跑起来；否则后续所有高级过滤、导出、多会话都只能停留在文档层。

中段按“先会话可控，再过滤可用，再数据可存，再并发扩展”的顺序推进。最后再接编辑器、AI、插件与远程能力，避免在采集与解析主链路未稳定前过早扩展边界。

## 7. 观察项

- 当前仓库没有 `requirements/` 能力文档；后续若要长期维护产品边界，应补一份能力愿景文档。
- 当前仓库没有 `architecture/` 现状文档；首个 feature 验收通过后需要把模块边界和契约回填到架构层。
