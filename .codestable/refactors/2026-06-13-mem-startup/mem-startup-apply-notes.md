---
doc_type: refactor-apply-notes
refactor: 2026-06-13-mem-startup
---

# mem-startup apply notes

基线：`go test ./...` 全 ok（执行前）。

## 步骤 1: 启动异步化（#4）
- 完成时间: 2026-06-13
- 改动文件: app.go（startup 中 loadInitialState 改 `go a.loadInitialState()`）
- 验证结果: go build + go test ./... 通过；-race 通过（model 写入受 c.mu 保护，无数据竞争）
- 偏离: 无
- HUMAN 目视项: 待用户确认首屏即时性（见末尾）

## 步骤 2: 删 Display 字段现算（#2）
- 完成时间: 2026-06-13
- 改动文件: model.go（删 Display）、logview.go（appendLogLocked 不再算 Display，formatLogDisplay→FormatLogDisplay 导出）、app_state.go（newAppState 窗口现算）、app_state_test.go（删 Display: raw 行）
- 验证结果: go test ./... 通过；grep 确认无残留读取存储 Display 字段（仅余 LogItemView JSON 契约字段）；app_state_test 的 Raw 断言仍过
- 偏离: 无

## 步骤 3: 删常驻 SearchLower，改按需搜索缓存（#3）
- 完成时间: 2026-06-13
- 改动文件: model.go（删 SearchLower 字段）、log_buffer.go（新增 `searchLower []string` 按需缓存）、search_match.go（保留 `searchLowerText` 统一归一化）、logview.go（搜索开启时 `EnsureSearchCache`，清空搜索时 `ReleaseSearchCache`，匹配回到 `strings.Contains(prelowered, query)`）、controller_test.go（新增 Unicode 折叠回归用例）、logview_bench_test.go（新增 benchmark）
- 验证结果: 搜索相关用例全过；`BenchmarkRebuildVisibleWithSearch` 在 5e4 条下 `811101 ns/op, 0 allocs/op`，与基线 `804806 ns/op, 0 allocs/op` 基本持平；清空搜索后缓存立即释放，不再常驻
- 偏离: 初版曾尝试“匹配时现算小写”，bench 退化到 `6.1ms/op`，不满足“性能不回退”，已回退为按需缓存方案

## 步骤 4: allLogs 容量上限（#1）
- 完成时间: 2026-06-13
- 改动文件: controller.go（`allLogs` 从 slice 改为 `logBuffer`，保留 maxLogEntries/nextSourceIndex）、binding.go/logview.go（清空路径改为 `logBuffer.Reset()`）、log_buffer.go（新增固定容量环形缓冲与顺序迭代/快照）、controller_test.go（新增 2 个刻画测试 + fmt 导入）
- 验证结果: 全量 `go test ./...` 通过；`TestControllerCapsLogEntries`（封顶/最旧淘汰）+ `TestControllerCapDropsSelectedAndShiftsSelection`（选中行淘汰失效/选中行左移）通过；`BenchmarkAppendLogAtCapacity` 为 `18.43 ns/op, 0 allocs/op`，避免了“满容量后每条 make+copy 整段”的 O(n) 退化
- 偏离: 初版上限实现用 `make+copy` 淘汰旧日志，虽然能 GC，但稳态写入退化成每条整段复制；已替换为固定容量环形缓冲
- 行为变更（已知）: 日志超过 10 万条后从最旧端淘汰，不再无限保留。这是内存有界化的必要代价

## 量化对比
- 无搜索流式追加：基线 `BenchmarkAppendLogNoSearch = 492.3 ns/op, 1285 B/op, 7 allocs/op`；当前 `21.80 ns/op, ~0 B/op, 0 allocs/op`
- 搜索 rebuild：基线 `804806 ns/op, 0 allocs/op`；当前 `811101 ns/op, 0 allocs/op`
- 满容量稳态追加：当前 `18.43 ns/op, 0 allocs/op`

## 步骤 5: 状态快照链路减配（新增）
- 完成时间: 2026-06-14
- 改动文件: app.go（`emitAndSnapshot` 直返 RPC 结果时不再额外 `EventsEmit` 同一份状态）、app_state.go（去掉 `BoundPIDs/History/MatchIndexes` 二次拷贝，删 `matchSet map[int]struct{}`，改固定长度切片填充，并移除行级 `display` payload，只保留 `SelectedLog.display`）、internal/app/ui_snapshot.go（`sliceSearchWindow` 改 `sort.SearchInts` 定位窗口起点）、frontend/src/use-app-controller.ts（Wails RPC 返回值直接 `setState(next)`，不再 `AppState.createFrom(next)` 深重建）、app_state_bench_test.go（新增 benchmark）
- 验证结果: `go test ./...` 通过；`npm run build` 通过；`BenchmarkNewAppState` 从 `162573 ns/op, 271189 B/op, 5012 allocs/op` 降到 `150954 ns/op, 259563 B/op, 5006 allocs/op`
- 偏离: 无。事件模式仍保留 `state:updated`，仅移除“RPC 返回值 + 同步事件”这份重复状态推送
- 补充量化: `BenchmarkMarshalAppState` 从 `535843 ns/op, 630936 B/op, 5017 allocs/op` 降到 `444244 ns/op, 555336 B/op, 5012 allocs/op`

## 最终验证（2026-06-14）
- 生产构建：`wails build` 成功，产物 `build/bin/logcatviewer.exe`
- 启动实测：`WaitForInputIdle` 在 73 ms 返回；启动后 1.2s / 3.3s / 10.4s 主进程工作集均为 27.5 MB
- 空闲总内存：主进程与 8 个 WebView2 子进程合计工作集约 383~394 MB；其中 Go 主进程仅 27.5 MB，剩余主要是 WebView2 固有多进程开销
- 有界性：`TestControllerCapsLogEntries`、`TestControllerCapDropsSelectedAndShiftsSelection` 与 `BenchmarkAppendLogAtCapacity` 证明超过 10 万条后环形淘汰，稳态追加 0 allocs/op，不再线性增长
- 回归：`go test ./... -count=1 -timeout 60s`、前端 `npm run build` 均通过
