---
doc_type: refactor-design
refactor: 2026-06-13-mem-startup
status: approved
scope: internal/app（日志存储与快照）+ 根 package main（启动时序、app_state 映射）+ internal/storage（导出）
summary: 删冗余派生字段（Display/SearchLower）降单条内存、给 allLogs 加容量上限把内存从无界变有界、启动异步化解除首屏阻塞
---

# mem-startup refactor design

## 1. 本次范围

从 scan 勾选 4 条全做（行为等价 3 条直接做，行为变更 1 条取合理默认并标注）：

- #4 启动异步化（行为等价：只改时序，不改最终状态）
- #2 删 `LogViewItem.Display` 字段，窗口渲染时现算
- #3 删 `LogViewItem.SearchLower` 字段，匹配时现算
- #1 给 `allLogs` 加容量上限 `maxLogEntries`，超限淘汰最旧（**唯一行为变更**：不再无限保留日志）

明确不做：
- `VisibleLogs` 无筛选时的整份拷贝去重——Go string 共享底层字节，拷的是 ~header 不是内容，收益有限且与 #1 索引淘汰强耦合、风险高。留待 profile 证明后单开。

风险档位：#1 中、#3 中、#2 低、#4 低。

## 2. 前置依赖

- 测试覆盖：`controller_test.go` 已覆盖 VisibleLogs/TotalLogs/搜索/筛选/暂停/1000 窗口；`app_state_test.go` 覆盖 Raw 保真。核心路径有覆盖，无需补 characterization test。
- #1 改 `SourceIndex` 语义需搜全部引用——已搜：`logview.go` 内 5 处 + `model.go` 定义，无包外引用。

## 3. 执行顺序

### 步骤 1：启动异步化（#4）
- 引用方法：M-L4-06 Async & Cancellation
- 具体操作：`app.go:startup` 把 `a.loadInitialState()` 同步调用改为 `go a.loadInitialState()`。`loadInitialState` 内部已是 `Load()` + `emitState()`，事件机制会在数据就绪后推给前端；前端 `use-app-controller.ts` 已订阅 `state:updated` 事件，能接住异步到达的状态。
- 退出信号：AI 跑全量 go test 通过；HUMAN 启动应用确认首屏立即出现（不再卡在 ADB detect）、设备列表稍后填充。
- 验证责任：AI 自证 + HUMAN 目视
- 回滚：git revert 本步（改回同步调用）

### 步骤 2：删 Display 字段现算（#2）
- 引用方法：M-L2-02 Inline Function（把驻留值内联为渲染时计算）
- 具体操作：
  1. 导出 `formatLogDisplay` → `FormatLogDisplay`（app 包，供 main 包调用）
  2. 删 `LogViewItem.Display` 字段（model.go）
  3. `appendLogLocked` 不再算 Display
  4. `app_state.go` 的 `newAppState` 循环里用 `appstate.FormatLogDisplay(item.Entry)` 现算（≤1000 条窗口）
  5. `app_state_test.go` 去掉 `Display: raw` 行（断言的是 Raw 不受影响）
- 退出信号：AI 跑 go test ./... 全通过 + go build 通过
- 验证责任：AI 自证
- 回滚：git revert 本步

### 步骤 3：删 SearchLower 字段现算（#3）
- 引用方法：M-L2-03 Replace Temp with Query（以查询取代驻留临时值）
- 具体操作：
  1. 删 `LogViewItem.SearchLower` 字段（model.go）+ `searchLowerText` 函数
  2. `matchesVisibleLogLocked` 里的 `strings.Contains(item.SearchLower, searchQuery)` 改为对 `item.Entry.Tag + Message` 现算小写后匹配（提取 helper `entryMatchesSearch(entry, query)`）
- 退出信号：AI 跑 go test ./internal/app/... 全通过（搜索相关用例：TestController*Search*）+ 加一个 benchmark 确认现算开销在可见集（≤全量遍历一次）可接受
- 验证责任：AI 自证（含 benchmark）
- 回滚：git revert 本步

### 步骤 4：allLogs 容量上限（#1）
- 引用方法：M-L4-02 Batching 的对偶——有界缓冲（环形淘汰）
- 具体操作：
  1. Controller 加字段 `maxLogEntries int`，`NewController` 默认 `defaultMaxLogEntries = 100000`（沿用 `pauseBufferCap` 既有模式）
  2. `appendLogLocked` append 后若 `len(allLogs) > maxLogEntries`，丢弃最旧：`allLogs = allLogs[len-max:]`
  3. `SourceIndex` 当前 = `len(allLogs)`（位置索引），淘汰后会与 VisibleLogs 的 SourceIndex 错位。改为单调递增序号：Controller 加 `nextSeq uint64`，`appendLogLocked` 用 `nextSeq++` 作 SourceIndex；`restoreSelectionLocked`/`selectedSourceIndexLocked` 逻辑不变（仍按 SourceIndex 匹配，只是值不再等于位置）
  4. 淘汰 allLogs 时同步淘汰 VisibleLogs 中 SourceIndex 已不存在的项 + rebuild search match indexes
- 退出信号：AI 跑全量 test + 新增用例 TestControllerCapsLogEntries（push 超过上限后 TotalLogs 封顶、最旧被丢、选中态不串位）
- 验证责任：AI 自证 + HUMAN（长时间挂日志观察内存平稳不涨）
- 回滚：git revert 本步

## 4. 风险与看点

- **步骤 4 最高风险**：SourceIndex 从"位置"变"序号"，所有依赖它的地方（restoreSelection / selectedSourceIndex）必须确认仍正确。淘汰发生时若当前选中行被淘汰，选中态应优雅失效（SelectedIndex = -1）而非 panic。
- **步骤 3 中风险**：搜索匹配从预算改现算，热路径（每条 incoming log 都过 matchesVisibleLogLocked）性能要确认不退化——benchmark 把关。
- 步骤 1/2 低风险，可快速验证。
