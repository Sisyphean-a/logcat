package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/logcat"
)

func BenchmarkNewAppState(b *testing.B) {
	snapshot := benchmarkUISnapshot()

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = newAppState(snapshot)
		}
	})

	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = legacyNewAppState(snapshot)
		}
	})
}

func BenchmarkMarshalAppState(b *testing.B) {
	snapshot := benchmarkUISnapshot()

	b.Run("current", func(b *testing.B) {
		sample, err := json.Marshal(newAppState(snapshot))
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(sample)), "json-bytes")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := json.Marshal(newAppState(snapshot)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("legacy", func(b *testing.B) {
		sample, err := json.Marshal(legacyNewAppState(snapshot))
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(sample)), "json-bytes")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := json.Marshal(legacyNewAppState(snapshot)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMarshalStateAppendPatch(b *testing.B) {
	prev, nextSnapshot := benchmarkAppendPatchPair()
	patch, ok := buildStateAppendPatch(prev, nextSnapshot)
	if !ok {
		b.Fatal("expected append patch")
	}

	sample, err := json.Marshal(patch)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(len(sample)), "json-bytes")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(patch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildStateAppendPatch(b *testing.B) {
	prev, nextSnapshot := benchmarkAppendPatchPair()

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, ok := buildStateAppendPatch(prev, nextSnapshot); !ok {
				b.Fatal("expected append patch")
			}
		}
	})

	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, ok := legacyBuildStateAppendPatch(prev, nextSnapshot); !ok {
				b.Fatal("expected append patch")
			}
		}
	})
}

func BenchmarkBuildSnapshotSelectedLog(b *testing.B) {
	snapshot := benchmarkUISnapshot()
	selection := appstate.SelectionState{
		AnchorSourceIndex: 996,
		FocusSourceIndex:  998,
		SourceIndexes:     []int{996, 997, 998},
	}

	b.Run("current", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = buildSnapshotSelectedLog(snapshot.VisibleLogs, selection)
		}
	})

	b.Run("legacy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = legacyBuildSnapshotSelectedLog(snapshot.Model.VisibleLogs, selection)
		}
	})
}

func BenchmarkMarshalSelectionPatch(b *testing.B) {
	snapshot := benchmarkUISnapshot()
	snapshot.Revision++
	snapshot.Model.Selection = appstate.SelectionState{
		AnchorSourceIndex: 996,
		FocusSourceIndex:  998,
		SourceIndexes:     []int{996, 997, 998},
	}
	patch := buildSelectionPatch(snapshot)

	sample, err := json.Marshal(patch)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(len(sample)), "json-bytes")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(patch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildSelectionPatch(b *testing.B) {
	snapshot := benchmarkSelectionSnapshot()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildSelectionPatchFromSnapshot(snapshot)
	}
}

func benchmarkUISnapshot() appstate.UISnapshot {
	const total = 1000
	logs := make([]appstate.LogViewItem, total)
	for i := range logs {
		logs[i] = appstate.LogViewItem{
			SourceIndex: i,
			Entry: logcat.LogEntry{
				TimeText: "06-10 20:41:45.478",
				Level:    "I",
				Tag:      fmt.Sprintf("tag-%d", i%11),
				Message:  fmt.Sprintf("[H5] benchmark message %d token", i),
				Raw:      fmt.Sprintf("raw line %d", i),
				Source:   "H5",
			},
		}
	}

	return appstate.UISnapshot{
		Revision:    42,
		VisibleLogs: visibleLogSnapshots(logs),
		Model: appstate.Model{
			Status:          "running",
			ADBStatus:       "已连接",
			Devices:         []appstate.DeviceItem{{ID: "dev-1", Model: "Pixel 7", Status: "device"}},
			SelectedDevice:  "dev-1",
			Packages:        []adb.PackageInfo{{Name: "com.demo.host"}},
			SelectedPackage: "com.demo.host",
			TotalLogs:       total,
			VisibleLogs:     logs,
			Filter: appstate.FilterState{
				Draft:           `tag=chromium`,
				Applied:         `tag=chromium`,
				ActiveFilterID:  "chromium",
				DefaultFilterID: "chromium",
				Saved: []appstate.SavedFilter{{
					ID:          "chromium",
					Name:        "chromium",
					PackageName: "com.demo.host",
					Query:       `tag=chromium`,
				}},
			},
			Search: appstate.SearchState{
				Query: "token",
			},
			Pause: appstate.PauseState{Active: false},
			Selection: appstate.SelectionState{
				AnchorSourceIndex: total - 1,
				FocusSourceIndex:  total - 1,
				SourceIndexes:     []int{total - 1},
			},
			SelectedIndex: total - 1,
		},
		VisibleCount: total,
	}
}

func benchmarkAppendPatchPair() (AppState, appstate.UISnapshot) {
	prev := newAppState(benchmarkUISnapshot())
	nextSnapshot := benchmarkUISnapshot()
	nextSnapshot.Revision++
	nextSnapshot.Model.TotalLogs++
	nextSnapshot.VisibleCount++
	nextSnapshot.Model.VisibleLogs = append(nextSnapshot.Model.VisibleLogs, appstate.LogViewItem{
		SourceIndex: len(nextSnapshot.Model.VisibleLogs),
		Entry: logcat.LogEntry{
			TimeText: "06-10 20:41:46.000",
			Level:    "I",
			Tag:      "tag-append",
			Message:  "[H5] appended benchmark token",
			Raw:      "raw appended line",
			Source:   "H5",
		},
	})
	nextSnapshot.VisibleLogs = append(nextSnapshot.VisibleLogs, appstate.VisibleLogSnapshot{
		SourceIndex: len(nextSnapshot.VisibleLogs),
		TimeText:    "06-10 20:41:46.000",
		Level:       "I",
		Tag:         "tag-append",
		Message:     "[H5] appended benchmark token",
		Source:      "H5",
	})
	return prev, nextSnapshot
}

func benchmarkSelectionSnapshot() appstate.SelectionSnapshot {
	full := benchmarkUISnapshot()
	focused := full.Model.VisibleLogs[998]
	return appstate.SelectionSnapshot{
		Selection: appstate.SelectionState{
			AnchorSourceIndex: 996,
			FocusSourceIndex:  998,
			SourceIndexes:     []int{996, 997, 998},
		},
		Focused:  &focused,
		Revision: full.Revision + 1,
	}
}

func legacyNewAppState(snapshot appstate.UISnapshot) AppState {
	model := snapshot.Model
	state := AppState{
		Revision:        snapshot.Revision,
		Status:          model.Status,
		ADBStatus:       model.ADBStatus,
		Devices:         make([]DeviceView, 0, len(model.Devices)),
		SelectedDevice:  model.SelectedDevice,
		PackageScope:    string(model.PackageScope),
		Packages:        make([]PackageView, 0, len(model.Packages)),
		SelectedPackage: model.SelectedPackage,
		TotalLogs:       model.TotalLogs,
		VisibleCount:    snapshot.VisibleCount,
		SelectedCount:   len(model.Selection.SourceIndexes),
		Filter: FilterView{
			Draft:           model.Filter.Draft,
			Applied:         model.Filter.Applied,
			Error:           model.Filter.Error,
			ActiveFilterID:  model.Filter.ActiveFilterID,
			DefaultFilterID: model.Filter.DefaultFilterID,
			Saved:           make([]SavedFilterView, 0, len(model.Filter.Saved)),
		},
		Search: SearchView{
			Query: model.Search.Query,
		},
		Pause: PauseView{
			Active: model.Pause.Active,
		},
		Logs: make([]LogItemView, 0, len(model.VisibleLogs)),
	}

	focusedSourceIndex := model.Selection.FocusSourceIndex

	for _, device := range model.Devices {
		state.Devices = append(state.Devices, DeviceView{
			ID:     device.ID,
			Model:  device.Model,
			Status: device.Status,
		})
	}

	for _, pkg := range model.Packages {
		state.Packages = append(state.Packages, PackageView{Name: pkg.Name})
	}

	for _, filter := range model.Filter.Saved {
		state.Filter.Saved = append(state.Filter.Saved, SavedFilterView{
			ID:          filter.ID,
			Name:        filter.Name,
			PackageName: filter.PackageName,
			Query:       filter.Query,
		})
	}

	for _, item := range model.VisibleLogs {
		isSelected := false
		for _, selectedSourceIndex := range model.Selection.SourceIndexes {
			if selectedSourceIndex == item.SourceIndex {
				isSelected = true
				break
			}
		}
		row := LogItemView{
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
			IsFocused:   item.SourceIndex == focusedSourceIndex,
			IsSelected:  isSelected,
		}
		state.Logs = append(state.Logs, row)
		if row.IsFocused {
			state.SelectedLog = &SelectedLogView{
				SourceIndex: row.SourceIndex,
				TimeText:    row.TimeText,
				Level:       row.Level,
				Tag:         row.Tag,
				Message:     row.Message,
				Source:      item.Entry.Source,
			}
		}
	}

	return state
}

func legacyBuildSnapshotSelectedLog(items []appstate.LogViewItem, selection appstate.SelectionState) *SelectedLogView {
	if selection.FocusSourceIndex < 0 {
		return nil
	}
	for _, item := range items {
		if item.SourceIndex != selection.FocusSourceIndex {
			continue
		}
		row := LogItemView{
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
			IsFocused:   true,
		}
		return buildSelectedLogView(row, item.Entry.Source)
	}
	return nil
}

func legacyBuildStateAppendPatch(prev AppState, snapshot appstate.UISnapshot) (StateAppendPatch, bool) {
	if !sameAppendPatchSnapshotContext(prev, snapshot) {
		return StateAppendPatch{}, false
	}

	dropped, appendedStart, ok := diffAppendSnapshotWindow(prev.Logs, snapshot.VisibleLogs, snapshot.Model.Selection)
	if !ok {
		return StateAppendPatch{}, false
	}

	appendedLogs := make([]LogItemView, len(snapshot.VisibleLogs)-appendedStart)
	cursor := newLogRowCursor(snapshot.Model.Selection)
	for index := range snapshot.VisibleLogs {
		item := snapshot.VisibleLogs[index]
		row := cursor.Next(item.SourceIndex, item.TimeText, item.Level, item.Tag, item.Message)
		if index >= appendedStart {
			appendedLogs[index-appendedStart] = row
		}
	}

	return StateAppendPatch{
		Revision:      snapshot.Revision,
		TotalLogs:     snapshot.Model.TotalLogs,
		VisibleCount:  snapshot.VisibleCount,
		Dropped:       dropped,
		Appended:      appendedLogs,
		SelectedCount: len(snapshot.Model.Selection.SourceIndexes),
		SelectedLog:   legacyBuildSnapshotSelectedLog(snapshot.Model.VisibleLogs, snapshot.Model.Selection),
	}, true
}

func visibleLogSnapshots(items []appstate.LogViewItem) []appstate.VisibleLogSnapshot {
	if len(items) == 0 {
		return nil
	}
	snapshots := make([]appstate.VisibleLogSnapshot, len(items))
	for index, item := range items {
		snapshots[index] = appstate.VisibleLogSnapshot{
			SourceIndex: item.SourceIndex,
			TimeText:    item.Entry.TimeText,
			Level:       item.Entry.Level,
			Tag:         item.Entry.Tag,
			Message:     item.Entry.Message,
			Source:      item.Entry.Source,
		}
	}
	return snapshots
}
