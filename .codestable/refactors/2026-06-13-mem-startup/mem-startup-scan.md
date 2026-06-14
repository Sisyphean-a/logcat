---
doc_type: refactor-scan
refactor: 2026-06-13-mem-startup
status: pending-user-selection
scope: internal/app/{logview.go,model.go,controller.go,model_clone.go,ui_snapshot.go}、app.go、app_state.go
summary: 发现 4 条优化点（内存 3 / 启动 1）；按风险 低 2 / 中 2
---

# mem-startup scan

## 总览

- 扫描范围：`internal/app/logview.go`、`model.go`、`controller.go`、`model_clone.go`、`ui_snapshot.go`、根 `app.go`、`app_state.go`，及对应测试 `controller_test.go`
- 发现 4 条优化点：内存 3 / 启动 1（无可读性 / 结构项）
- 按风险：低 2（#2 #4）/ 中 2（#1 #3）
- 建议先做：**#4**（启动异步化，低风险独立）、**#2**（移除 Display 全量驻留，低风险）
- 建议慎做 / 需决策：**#1**（容量上限会改"日志无限保留"这一可观察行为，须用户拍板默认值）、**#3**（搜索热路径，需 benchmark 确认不回退）
- 前置检查 7 条全过：✓

### 内存模型校正（影响优先级，先说清）

Go `string` 是 `{ptr,len}` header，底层字节数组不可变且共享。据此重新排了大头：

- **真·额外字节**：`LogViewItem.Display`（`fmt.Sprintf` 新分配，≈ 一行长度）、`LogViewItem.SearchLower`（`ToLower` 新分配，≈ Tag+Message 长度）。每条日志因此多占约 1.5~2 倍解析后字段的字节。→ #2 #3
- **无界增长**：`allLogs` 只增不减（仅 `ClearVisible` 清空），长时间挂 logcat 线性涨到 OOM。→ #1
- **被高估的项（故未列）**：`VisibleLogs` 无筛选时 `append(VisibleLogs[:0], allLogs...)` 看似翻倍，实际只拷 string header（每 item ~150B），不拷字符串内容；且与 #1 环形淘汰的索引同步交互复杂、风险高、收益有限。本次**不做**，如后续 profile 证明 header 拷贝是瓶颈再单独开。
- **不动 `Entry.Raw`**：被前端 `copySelected("raw")` 与详情面板消费，属功能字段非冗余。

## 条目

### [#1] 给 allLogs 加容量上限（环形淘汰最旧），无界 → 有界

- **位置**：`internal/app/logview.go:181-192`（appendLogLocked）、`controller.go:37`（allLogs 字段）、`model.go:23`（SourceIndex）
- **分类**：性能
- **现状**：`c.allLogs = append(c.allLogs, item)` 只增不减；`SourceIndex` 取 `len(c.allLogs)`（append 前长度）。除手动 `ClearVisible` 外无任何回收。
- **问题**：内存随日志条数线性无上限增长，长时间运行必然 OOM（可度量：稳态内存 = 已收日志总数 × 每条占用，无收敛）。
- **建议**：为 `allLogs` 设容量上限（默认建议 5e4~1e5，可配置），超限丢弃最旧条目（环形）。`SourceIndex` 改为 controller 上单调递增 `nextSeq`（不再等于 slice 下标）；淘汰旧条目时同步从 `VisibleLogs`/`Search.MatchIndexes` 前端清理失效项。
- **建议映射的方法**：M-L4-05（容量/失效策略，最接近的淘汰方法）
- **风险**：中 —— **改变"日志无限保留"这一可观察行为**（超限即丢）；牵涉 `SourceIndex` 语义改写 + VisibleLogs/搜索索引在淘汰时的同步。须用户决策默认上限值。
- **验证**：AI 自证（新增刻画测试：超上限后最旧被丢、TotalLogs 反映上限、淘汰后选中态/搜索索引不越界）+ HUMAN（确认默认上限合理）
- **范围**：约 60 行 / 3 文件

### [#2] 移除 LogViewItem.Display，改 newAppState 按可见窗口现算

- **位置**：`internal/app/model.go:24`（Display 字段）、`logview.go:182-188,255-257`（formatLogDisplay 全量调用）、`app_state.go:187,202`（消费处）
- **分类**：性能
- **现状**：`appendLogLocked` 对每条日志 `fmt.Sprintf("%s %s %s %s", ...)` 生成 Display 并随 item 全量驻留；前端仅需 ≤1000 条可见窗口的 display（`UISnapshot` 已限窗）。
- **问题**：为全部日志预算并常驻一份完整拼接字符串（真·额外字节 ≈ 每条一行长度 × 总条数），但实际只有窗口内 ≤1000 条会被前端读取。
- **建议**：从 `LogViewItem` 删除 `Display`；在 `newAppState` 遍历可见窗口时对每行调用 `formatLogDisplay(item.Entry)` 现算（≤1000 次/快照）。`SelectedLogView.Display` 同样现算。
- **建议映射的方法**：M-L2-03（以查询取代存储的临时值）
- **风险**：低 —— 纯派生值，前端拿到的 `display` 内容不变；窗口现算量受 1000 上限封顶。
- **验证**：AI 自证（`go test ./...`；`app_state_test.go` 断言 display 输出一致；grep `\.Display` 确认无残留读取存储字段）
- **范围**：约 25 行 / 3 文件

### [#3] 移除 LogViewItem.SearchLower，搜索匹配时现算

- **位置**：`internal/app/model.go:25`（SearchLower 字段）、`logview.go:187,290-295,321-323`（searchLowerText / matchesVisibleLogLocked）
- **分类**：性能
- **现状**：`appendLogLocked` 对每条 `strings.ToLower(Tag+"\n"+Message)` 生成 SearchLower 全量驻留；`matchesVisibleLogLocked` 用 `strings.Contains(item.SearchLower, q)` 匹配。
- **问题**：为全部日志常驻一份小写副本（真·额外字节 ≈ Tag+Message 长度 × 总条数），仅服务于子串搜索。
- **建议**：删除 `SearchLower` 字段；匹配改为大小写无关比较（优先 `caselessContains`，避免每次 `ToLower` 分配；退路是匹配时 `ToLower` 现算）。query 仍预先归一化一次。
- **建议映射的方法**：M-L2-03（以查询取代存储的临时值）
- **风险**：中 —— 搜索/筛选是 rebuild 热路径，去掉预算可能增加匹配期计算。须 benchmark 确认大日志量下匹配耗时不明显回退。
- **验证**：AI 自证（搜索相关测试 `controller_test.go` 全过；新增/复用 benchmark 对比 rebuild 在 5e4 条下耗时）
- **范围**：约 30 行 / 2 文件

### [#4] startup 的 loadInitialState 异步化，解除首屏阻塞

- **位置**：`app.go:28-33`（startup）、`app.go:185-188`（loadInitialState）
- **分类**：性能
- **现状**：`startup` 同步调用 `a.loadInitialState()` → `controller.Load()`，内部串行执行 ADB detect + ListDevices + syncDevices（均为外部 `adb` 子进程调用），完成后才返回，期间 Wails 首屏阻塞。
- **问题**：首屏渲染被 N 个 adb 子进程往返阻塞（启动耗时 = adb 启动 + 设备枚举往返，设备未连/adb 冷启时可达秒级），用户看到空白等待。
- **建议**：将 `loadInitialState` 放入 goroutine（`go a.loadInitialState()`），与 `trackDevices`/`pushStateLoop` 一致；首屏先渲染 idle 空状态，设备信息经既有 `emitState` 事件推送补齐。确认 `controller.Load` 内部对 model 的写入已被 `mu` 保护（line 115-119 已加锁）。
- **建议映射的方法**：M-L4-06（异步与取消）
- **风险**：低 —— 最终状态一致，仅"阻塞等待"变"先渲染后填充"；并发安全由既有 `c.mu` 保证。
- **验证**：HUMAN（启动应用，确认窗口立即出现且 ADB 状态/设备列表随后填充，无空指针/竞态）+ AI 自证（`go build`；`go test ./...`；`go vet` / `-race` 跑相关测试）
- **范围**：约 5 行 / 1 文件
