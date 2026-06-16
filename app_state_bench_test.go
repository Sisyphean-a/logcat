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

func benchmarkUISnapshot() appstate.UISnapshot {
	const total = 1000
	logs := make([]appstate.LogViewItem, total)
	matchIndexes := make([]int, 0, total/4)
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
		if i%4 == 0 {
			matchIndexes = append(matchIndexes, i)
		}
	}

	return appstate.UISnapshot{
		Model: appstate.Model{
			Status:          "running",
			ADBStatus:       "已连接",
			Devices:         []appstate.DeviceItem{{ID: "dev-1", Model: "Pixel 7", Status: "device"}},
			SelectedDevice:  "dev-1",
			Packages:        []adb.PackageInfo{{Name: "com.demo.host"}},
			SelectedPackage: "com.demo.host",
			Processes:       []adb.ProcessInfo{{PID: 1234, Name: "com.demo.host"}},
			SelectedProcess: "com.demo.host",
			BoundPIDs:       []int{111, 222},
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
				History: []string{"tag=chromium", `message~:"token"`},
			},
			Search: appstate.SearchState{
				Query:        "token",
				MatchIndexes: matchIndexes,
				Current:      len(matchIndexes) / 2,
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
		VisibleStart: 0,
	}
}

func legacyNewAppState(snapshot appstate.UISnapshot) AppState {
	model := snapshot.Model
	state := AppState{
		Status:          model.Status,
		ADBStatus:       model.ADBStatus,
		Devices:         make([]DeviceView, 0, len(model.Devices)),
		SelectedDevice:  model.SelectedDevice,
		PackageScope:    string(model.PackageScope),
		Packages:        make([]PackageView, 0, len(model.Packages)),
		SelectedPackage: model.SelectedPackage,
		Processes:       make([]ProcessView, 0, len(model.Processes)),
		SelectedProcess: model.SelectedProcess,
		BoundPIDs:       append([]int(nil), model.BoundPIDs...),
		TotalLogs:       model.TotalLogs,
		VisibleCount:    snapshot.VisibleCount,
		VisibleStart:    snapshot.VisibleStart,
		SelectedIndex:   model.SelectedIndex,
		SelectedCount:   len(model.Selection.SourceIndexes),
		Filter: FilterView{
			Draft:           model.Filter.Draft,
			Applied:         model.Filter.Applied,
			Error:           model.Filter.Error,
			ActiveFilterID:  model.Filter.ActiveFilterID,
			DefaultFilterID: model.Filter.DefaultFilterID,
			Saved:           make([]SavedFilterView, 0, len(model.Filter.Saved)),
			History:         append([]string(nil), model.Filter.History...),
		},
		Search: SearchView{
			Query:        model.Search.Query,
			MatchIndexes: append([]int(nil), model.Search.MatchIndexes...),
			Current:      model.Search.Current,
		},
		Pause: PauseView{
			Active:        model.Pause.Active,
			BufferedCount: model.Pause.BufferedCount,
			DroppedCount:  model.Pause.DroppedCount,
		},
		Logs: make([]LogItemView, 0, len(model.VisibleLogs)),
	}

	matchSet := make(map[int]struct{}, len(model.Search.MatchIndexes))
	for _, index := range model.Search.MatchIndexes {
		matchSet[index] = struct{}{}
	}

	currentMatch := -1
	if model.Search.Current >= 0 && model.Search.Current < len(model.Search.MatchIndexes) {
		currentMatch = model.Search.MatchIndexes[model.Search.Current]
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

	for _, process := range model.Processes {
		state.Processes = append(state.Processes, ProcessView{
			PID:  process.PID,
			Name: process.Name,
		})
	}

	for _, filter := range model.Filter.Saved {
		state.Filter.Saved = append(state.Filter.Saved, SavedFilterView{
			ID:          filter.ID,
			Name:        filter.Name,
			PackageName: filter.PackageName,
			Query:       filter.Query,
		})
	}

	for offset, item := range model.VisibleLogs {
		index := snapshot.VisibleStart + offset
		_, isMatch := matchSet[index]
		isSelected := false
		for _, selectedSourceIndex := range model.Selection.SourceIndexes {
			if selectedSourceIndex == item.SourceIndex {
				isSelected = true
				break
			}
		}
		display := appstate.FormatLogDisplay(item.Entry)
		row := LogItemView{
			Index:      index,
			TimeText:   item.Entry.TimeText,
			Level:      item.Entry.Level,
			Tag:        item.Entry.Tag,
			Message:    item.Entry.Message,
			Source:     item.Entry.Source,
			Raw:        item.Entry.Raw,
			Display:    display,
			IsMatch:    isMatch,
			IsCurrent:  currentMatch == index,
			IsFocused:  item.SourceIndex == focusedSourceIndex,
			IsSelected: isSelected,
		}
		state.Logs = append(state.Logs, row)
		if row.IsFocused {
			state.SelectedLog = &SelectedLogView{
				Index:    index,
				TimeText: row.TimeText,
				Level:    row.Level,
				Tag:      row.Tag,
				Message:  row.Message,
				Source:   row.Source,
				Raw:      row.Raw,
				Display:  row.Display,
			}
		}
	}

	return state
}
