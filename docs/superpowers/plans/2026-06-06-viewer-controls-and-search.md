# Viewer Controls And Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为当前单设备 H5 日志视图补齐暂停/恢复、清空视图、复制、基础搜索、自动跟随与详情面板。

**Architecture:** 先把 `internal/ui` 拆成壳、设备区、日志区、详情区、工具栏五块，避免继续把交互堆进单文件。随后在 `internal/app` 增加暂停缓冲、搜索状态和选中态，让 `ui` 只负责渲染与 Gio 剪贴板命令，`session` 继续只做事件流。

**Tech Stack:** Go 1.26、Gio、标准库 `context/sync/strings`

---

### Task 1: UI 微重构

**Files:**
- Create: `internal/ui/devices_panel.go`
- Create: `internal/ui/log_panel.go`
- Create: `internal/ui/detail_panel.go`
- Create: `internal/ui/controls_panel.go`
- Modify: `internal/ui/shell.go`

- [ ] **Step 1: Write the failing build check**

Run: `go test ./...`
Expected: PASS before split, then keep PASS after split

- [ ] **Step 2: Move existing panel methods into focused files**

Keep these signatures stable:

```go
func (s *Shell) layoutDevices(gtx layout.Context, model appstate.Model) layout.Dimensions
func (s *Shell) layoutLogs(gtx layout.Context, model appstate.Model) layout.Dimensions
```

- [ ] **Step 3: Reduce `shell.go` to event loop + orchestration**

Keep `Run`, `newShell`, `bootstrap`, `layout`, `syncDeviceButtons`, `handleClicks` in `shell.go`.

- [ ] **Step 4: Run verification**

Run: `$env:GOSUMDB='off'; go test ./...`
Expected: PASS

### Task 2: Controller 状态与操作

**Files:**
- Modify: `internal/app/model.go`
- Modify: `internal/app/controller.go`
- Create: `internal/app/logview.go`
- Modify: `internal/app/controller_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests for:

```go
func TestControllerPauseBuffersIncomingLogs(t *testing.T)
func TestControllerResumeKeepFlushesBufferedLogs(t *testing.T)
func TestControllerResumeDiscardDropsBufferedLogs(t *testing.T)
func TestControllerClearVisibleResetsListSelectionAndSearch(t *testing.T)
func TestControllerSearchTracksMatchesAndCurrentSelection(t *testing.T)
```

- [ ] **Step 2: Run targeted tests to verify they fail**

Run: `$env:GOSUMDB='off'; go test ./internal/app -run 'TestController(Pause|Resume|ClearVisible|Search)' -v`
Expected: FAIL with missing methods / fields

- [ ] **Step 3: Add model types and controller methods**

Introduce:

```go
type LogViewItem struct {
	Entry   logcat.LogEntry
	Display string
}

type SearchState struct {
	Query        string
	MatchIndexes []int
	Current      int
}
```

And controller methods:

```go
func (c *Controller) Pause()
func (c *Controller) ResumeKeep()
func (c *Controller) ResumeDiscard()
func (c *Controller) ClearVisible()
func (c *Controller) SetSearchQuery(query string)
func (c *Controller) NextMatch()
func (c *Controller) PrevMatch()
func (c *Controller) SelectLog(index int)
```

- [ ] **Step 4: Run targeted tests to verify they pass**

Run: `$env:GOSUMDB='off'; go test ./internal/app -run 'TestController(Pause|Resume|ClearVisible|Search)' -v`
Expected: PASS

### Task 3: 采集与搜索联动

**Files:**
- Modify: `internal/app/controller.go`
- Modify: `internal/session/supervisor_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestControllerConsumeRecomputesMatchesWhenLogsArrive(t *testing.T)
```

The test should prove: query already set to `token`, new matching entry arrives, `MatchIndexes` updates automatically.

- [ ] **Step 2: Run test to verify it fails**

Run: `$env:GOSUMDB='off'; go test ./internal/app -run TestControllerConsumeRecomputesMatchesWhenLogsArrive -v`
Expected: FAIL

- [ ] **Step 3: Update consume path**

When not paused:
- append `LogViewItem`
- recompute search matches
- keep selected entry aligned with current match when appropriate

When paused:
- append to paused buffer
- update buffered count / dropped count

- [ ] **Step 4: Run tests to verify they pass**

Run: `$env:GOSUMDB='off'; go test ./internal/app ./internal/session`
Expected: PASS

### Task 4: 工具栏、详情区与复制

**Files:**
- Modify: `internal/ui/shell.go`
- Modify: `internal/ui/controls_panel.go`
- Modify: `internal/ui/log_panel.go`
- Modify: `internal/ui/detail_panel.go`

- [ ] **Step 1: Write the failing build check**

Run: `$env:GOSUMDB='off'; go test ./...`
Expected: FAIL after wiring references to not-yet-implemented UI fields

- [ ] **Step 2: Add Gio UI state**

Need UI-local state for:

```go
searchEditor widget.Editor
pauseButton widget.Clickable
resumeKeepButton widget.Clickable
resumeDropButton widget.Clickable
clearButton widget.Clickable
followButton widget.Clickable
copyLineButton widget.Clickable
copyRawButton widget.Clickable
copyMessageButton widget.Clickable
nextMatchButton widget.Clickable
prevMatchButton widget.Clickable
```

- [ ] **Step 3: Wire clipboard actions**

Use:

```go
gtx.Execute(clipboard.WriteCmd{
	Type: "application/text",
	Data: io.NopCloser(strings.NewReader(text)),
})
```

- [ ] **Step 4: Run verification**

Run: `$env:GOSUMDB='off'; go test ./...`
Expected: PASS

### Task 5: 自动跟随与收尾

**Files:**
- Modify: `internal/ui/log_panel.go`
- Modify: `internal/ui/shell.go`
- Modify: `.codestable/features/2026-06-06-viewer-controls-and-search/viewer-controls-and-search-checklist.yaml`

- [ ] **Step 1: Add auto-follow state**

Use `widget.List.List.ScrollToEnd` and `Position.BeforeEnd`:

```go
s.logList.List.ScrollToEnd = s.followLogs
if s.followLogs && s.logList.Position.BeforeEnd {
	s.followLogs = false
}
```

- [ ] **Step 2: Add manual recover-follow action**

When follow button clicked:

```go
s.followLogs = true
s.logList.Position.BeforeEnd = false
```

- [ ] **Step 3: Run full verification**

Run:
- `gofmt -w internal cmd`
- `$env:GOSUMDB='off'; go test ./...`
- `$env:GOSUMDB='off'; go build ./cmd/logcatviewer`

Expected: all PASS
