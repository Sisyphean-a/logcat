# Query Language And Saved Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前单设备 H5 查看链路上加入显式应用的高级查询语言，以及保存过滤器和查询历史的最小持久化。

**Architecture:** 继续沿用 `adb -> session -> app -> ui` 四层，但新增 `internal/query` 承载查询编译/匹配，新增 `internal/storage` 承载保存过滤器与历史的最小文件存取。`app.Controller` 维护 query 状态和规范日志缓冲，`ui` 负责 query 编辑区与过滤器面板。

**Tech Stack:** Go 1.26、Gio、标准库 `context/encoding/json/os/path/filepath/regexp/strings/time`

---

### Task 1: 查询语法骨架

**Files:**
- Create: `internal/query/compiler.go`
- Create: `internal/query/matcher.go`
- Create: `internal/query/types.go`
- Create: `internal/query/compiler_test.go`
- Create: `internal/query/matcher_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestCompileSupportsLogicAndRegex(t *testing.T) {
	compiler := NewCompiler()
	compiled, err := compiler.Compile(`level:ERROR | (tag:chromium & message~:"接口.*失败")`)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	if !compiled.HasRegex {
		t.Fatal("expected regex flag")
	}
	if len(compiled.Terms) == 0 {
		t.Fatal("expected compiled terms")
	}
}
```

```go
func TestMatchSupportsAgeAndNegation(t *testing.T) {
	matcher := NewMatcher(func() time.Time {
		return time.Date(2026, 6, 6, 12, 0, 0, 0, time.Local)
	})
	compiled := mustCompile(t, `age:5m & -message:"[vite]"`)
	entry := logcat.LogEntry{
		Timestamp: time.Date(2026, 6, 6, 11, 58, 0, 0, time.Local),
		Message:   `[H5] ready`,
	}
	if !matcher.Match(entry, compiled) {
		t.Fatal("expected entry to match")
	}
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/query -run 'Test(CompileSupportsLogicAndRegex|MatchSupportsAgeAndNegation)' -v`  
Expected: FAIL with missing package / missing compiler symbols

- [ ] **Step 3: 写最小实现**

```go
type CompiledQuery struct {
	Raw      string
	Terms    []QueryTerm
	HasRegex bool
}

type QueryTerm struct {
	Kind      string
	Field     string
	Value     string
	Pattern   *regexp.Regexp
	Duration  time.Duration
	Threshold string
}
```

```go
func (c Compiler) Compile(input string) (CompiledQuery, error) {
	// tokenize -> shunting-yard -> postfix terms
	return CompiledQuery{}, nil
}
```

```go
func (m Matcher) Match(entry logcat.LogEntry, compiled CompiledQuery) bool {
	// stack-evaluate postfix terms against Level/Tag/Message/Timestamp
	return false
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/query -v`  
Expected: PASS

### Task 2: controller 过滤状态与重过滤

**Files:**
- Modify: `internal/app/model.go`
- Create: `internal/app/filter_query.go`
- Modify: `internal/app/logview.go`
- Modify: `internal/app/controller.go`
- Create: `internal/app/filter_query_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestControllerApplyFilterQueryRecomputesVisibleLogs(t *testing.T) {
	// 目标：已有缓存日志在应用 query 后立即重算 VisibleLogs
}
```

```go
func TestControllerInvalidQueryKeepsPreviousAppliedQuery(t *testing.T) {
	// 目标：非法 query 只写 Error，不丢旧 Applied 结果
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/app -run 'TestController(ApplyFilterQueryRecomputesVisibleLogs|InvalidQueryKeepsPreviousAppliedQuery)' -v`  
Expected: FAIL with undefined filter query state / methods

- [ ] **Step 3: 写最小实现**

```go
type FilterQueryState struct {
	Draft          string
	Applied        string
	Error          string
	ActiveFilterID string
	SavedFilters   []storage.SavedFilter
	History        []string
}
```

```go
func (c *Controller) ApplyFilterQuery(query string) error {
	// compile -> swap applied query -> rebuild visible logs -> persist history
	return nil
}
```

```go
func (c *Controller) rebuildVisibleLogsLocked() {
	// derive VisibleLogs from retained entries + applied query
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/app -run 'TestController(ApplyFilterQueryRecomputesVisibleLogs|InvalidQueryKeepsPreviousAppliedQuery)' -v`  
Expected: PASS

### Task 3: 保存过滤器与历史持久化

**Files:**
- Create: `internal/storage/query_store.go`
- Create: `internal/storage/query_store_test.go`
- Modify: `cmd/logcatviewer/main.go`
- Modify: `internal/app/controller.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestFileQueryStorePersistsFiltersAndHistory(t *testing.T) {
	dir := t.TempDir()
	store := NewFileQueryStore(filepath.Join(dir, "query-state.json"))
	state := QueryStateFile{
		SavedFilters: []SavedFilter{{ID: "1", Name: "H5", Query: `tag:chromium`}},
		QueryHistory: []string{`tag:chromium`, `level:ERROR`},
	}
	if err := store.Save(context.Background(), state); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(loaded.SavedFilters) != 1 || len(loaded.QueryHistory) != 2 {
		t.Fatalf("unexpected loaded state: %#v", loaded)
	}
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/storage -run TestFileQueryStorePersistsFiltersAndHistory -v`  
Expected: FAIL with missing store package or symbols

- [ ] **Step 3: 写最小实现**

```go
type SavedFilter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Query string `json:"query"`
}

type QueryStateFile struct {
	SavedFilters []SavedFilter `json:"saved_filters"`
	QueryHistory []string      `json:"query_history"`
}
```

```go
func NewFileQueryStore(path string) FileQueryStore {
	return FileQueryStore{path: path}
}
```

```go
func defaultQueryStatePath() string {
	// os.UserConfigDir()/logcat-viewer/query-state.json
	return ""
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/storage -v`  
Expected: PASS

### Task 4: UI 查询与过滤器面板

**Files:**
- Modify: `internal/ui/shell.go`
- Modify: `internal/ui/controls_panel.go`
- Create: `internal/ui/query_panel.go`
- Create: `internal/ui/filter_library_panel.go`
- Modify: `internal/ui/interactions.go`
- Modify: `internal/ui/shell_test.go`

- [ ] **Step 1: 先写失败测试**

```go
func TestShellHandleActionsAppliesFilterQuery(t *testing.T) {
	// 目标：点击“应用查询”会驱动 controller.ApplyFilterQuery
}
```

```go
func TestShellHandleActionsReplaysSavedFilterAndHistory(t *testing.T) {
	// 目标：点击左侧 saved filter / history 按钮会回填并应用 query
}
```

- [ ] **Step 2: 跑测试确认红灯**

Run: `go test ./internal/ui -run 'TestShell(HandleActionsAppliesFilterQuery|HandleActionsReplaysSavedFilterAndHistory)' -v`  
Expected: FAIL with missing query/filter panel state

- [ ] **Step 3: 写最小实现**

```go
type Shell struct {
	// ...
	queryEditor       widget.Editor
	filterNameEditor  widget.Editor
	applyQueryButton  widget.Clickable
	clearQueryButton  widget.Clickable
	saveFilterButton  widget.Clickable
	savedFilterButtons []widget.Clickable
	historyButtons     []widget.Clickable
}
```

```go
func (s *Shell) layoutQueryPanel(gtx layout.Context, model appstate.Model) layout.Dimensions {
	// query editor + apply / clear / save
	return layout.Dimensions{}
}
```

- [ ] **Step 4: 跑测试确认绿灯**

Run: `go test ./internal/ui -run 'TestShell(HandleActionsAppliesFilterQuery|HandleActionsReplaysSavedFilterAndHistory)' -v`  
Expected: PASS

### Task 5: 全量验证

**Files:**
- Modify: `.codestable/features/2026-06-06-query-language-and-saved-filters/query-language-and-saved-filters-checklist.yaml`
- Test: `./...`

- [ ] **Step 1: 跑 feature 关键场景测试**

Run: `go test ./internal/query ./internal/app ./internal/storage ./internal/ui -v`  
Expected: PASS

- [ ] **Step 2: 跑全量测试**

Run: `go test -count=1 ./...`  
Expected: PASS

- [ ] **Step 3: 构建桌面程序**

Run: `go build ./cmd/logcatviewer`  
Expected: build succeeds

- [ ] **Step 4: 校验 checklist 和 roadmap items**

Run: `python .codestable/tools/validate-yaml.py --file .codestable/features/2026-06-06-query-language-and-saved-filters/query-language-and-saved-filters-checklist.yaml --yaml-only`  
Expected: PASS

Run: `python .codestable/tools/validate-yaml.py --file .codestable/roadmap/logcat-viewer/logcat-viewer-items.yaml --yaml-only`  
Expected: PASS
