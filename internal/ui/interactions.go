package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) handleActions(gtx layout.Context, model appstate.Model) {
	if s.controller == nil {
		return
	}

	s.handleSelectorToggles(gtx)
	s.handleDeviceClicks(gtx, model)
	s.handlePackageClicks(gtx, model)
	s.handleFilterClicks(gtx, model)
	s.handleHistoryClicks(gtx, model)
	s.handleFilterDraft(gtx, model)
	s.handleToolbarClicks(gtx, model)
	s.handleLogClicks(gtx, model)
}

func (s *Shell) handleSelectorToggles(gtx layout.Context) {
	for s.deviceSelectorButton.Clicked(gtx) {
		s.deviceMenuOpen = !s.deviceMenuOpen
		s.filterMenuOpen = false
		s.packageMenuOpen = false
		s.historyMenuOpen = false
	}
	for s.filterSelectorButton.Clicked(gtx) {
		s.filterMenuOpen = !s.filterMenuOpen
		s.deviceMenuOpen = false
		s.packageMenuOpen = false
		s.historyMenuOpen = false
	}
	for s.packageSelectorButton.Clicked(gtx) {
		s.packageMenuOpen = !s.packageMenuOpen
		s.deviceMenuOpen = false
		s.filterMenuOpen = false
		s.historyMenuOpen = false
	}
	for s.historyMenuButton.Clicked(gtx) {
		s.historyMenuOpen = !s.historyMenuOpen
		s.deviceMenuOpen = false
		s.filterMenuOpen = false
		s.packageMenuOpen = false
	}
}

func (s *Shell) handleDeviceClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Devices {
		if index >= len(s.deviceButtons) {
			break
		}
		for s.deviceButtons[index].Clicked(gtx) {
			_ = s.controller.SelectDevice(context.Background(), model.Devices[index].ID)
			s.deviceMenuOpen = false
		}
	}
}

func (s *Shell) handlePackageClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Packages {
		if index >= len(s.packageButtons) {
			break
		}
		for s.packageButtons[index].Clicked(gtx) {
			_ = s.controller.SelectPackage(context.Background(), model.Packages[index].Name)
			s.packageMenuOpen = false
		}
	}
	for s.packageRootButton.Clicked(gtx) {
		if model.SelectedDevice != "" {
			_ = s.controller.SelectDevice(context.Background(), model.SelectedDevice)
		}
		s.packageMenuOpen = false
	}
}

func (s *Shell) handleFilterClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Filter.Saved {
		if index >= len(s.filterButtons) {
			break
		}
		for s.filterButtons[index].Clicked(gtx) {
			if err := s.controller.ApplySavedFilter(context.Background(), model.Filter.Saved[index].ID); err == nil {
				s.persistFilters()
			}
			s.filterMenuOpen = false
		}
	}
}

func (s *Shell) handleHistoryClicks(gtx layout.Context, model appstate.Model) {
	history := visibleHistory(model)
	for index := range history {
		if index >= len(s.historyButtons) {
			break
		}
		for s.historyButtons[index].Clicked(gtx) {
			if err := s.controller.ApplyHistoryQuery(history[index]); err == nil {
				s.persistFilters()
			}
			s.historyMenuOpen = false
		}
	}
}

func (s *Shell) handleFilterDraft(gtx layout.Context, model appstate.Model) {
	query := strings.TrimSpace(s.filterEditor.Text())
	if query != model.Filter.Draft {
		s.controller.SetFilterDraft(query)
	}
	for {
		event, ok := s.filterEditor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.SubmitEvent); ok {
			if err := s.controller.ApplyFilterDraft(); err == nil {
				s.persistFilters()
			}
		}
	}
}

func (s *Shell) handleToolbarClicks(gtx layout.Context, model appstate.Model) {
	for s.pauseButton.Clicked(gtx) {
		if model.Pause.Active {
			s.controller.ResumeKeep()
		} else {
			s.controller.Pause()
		}
	}
	for s.clearButton.Clicked(gtx) {
		s.controller.ClearVisible()
	}
	for s.exportButton.Clicked(gtx) {
		s.exportVisibleLogs(model)
	}
	for s.followButton.Clicked(gtx) {
		s.followLogs = !s.followLogs
		s.logList.Position.BeforeEnd = false
	}
	for s.saveFilterButton.Clicked(gtx) {
		name := strings.TrimSpace(s.saveNameEditor.Text())
		if name == "" {
			name = autoFilterName(s.controller.Model())
		}
		if err := s.controller.SaveCurrentFilter(name); err == nil {
			s.persistFilters()
			s.saveNameEditor.SetText("")
		}
	}
	for s.toggleDetailButton.Clicked(gtx) {
		s.detailCollapsed = !s.detailCollapsed
	}
}

func (s *Shell) handleLogClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.VisibleLogs {
		if index >= len(s.logButtons) {
			break
		}
		for s.logButtons[index].Clicked(gtx) {
			s.followLogs = false
			s.controller.SelectLog(index)
		}
	}
}

func (s *Shell) exportVisibleLogs(model appstate.Model) {
	path, err := s.exportLogs(model.VisibleLogs)
	if err != nil {
		s.controller.SetStatus(err.Error())
		return
	}
	message := fmt.Sprintf("已导出 %d 条到 Downloads/%s", len(model.VisibleLogs), filepath.Base(path))
	s.controller.SetStatus(message)
}

func widgetClickable() widget.Clickable {
	return widget.Clickable{}
}

func syncClickables(buttons []widget.Clickable, count int) []widget.Clickable {
	if len(buttons) > count {
		return buttons[:count]
	}
	for len(buttons) < count {
		buttons = append(buttons, widgetClickable())
	}
	return buttons
}

func (s *Shell) selectedLogItem(model appstate.Model) (appstate.LogViewItem, bool) {
	if model.SelectedIndex < 0 || model.SelectedIndex >= len(model.VisibleLogs) {
		return appstate.LogViewItem{}, false
	}
	return model.VisibleLogs[model.SelectedIndex], true
}

func (s *Shell) syncAutoFollow(logCount int) {
	if s.followLogs && s.logList.Position.BeforeEnd {
		s.followLogs = false
	}
	if s.followLogs {
		s.logList.ScrollToEnd = true
		if logCount != s.lastLogCount {
			s.logList.Position.BeforeEnd = false
		}
	} else {
		s.logList.ScrollToEnd = false
	}
	s.lastLogCount = logCount
}

func (s *Shell) syncSelectedLog(model appstate.Model) {
	if model.SelectedIndex == s.lastSelected {
		return
	}
	s.lastSelected = model.SelectedIndex
	if s.followLogs {
		return
	}
	if model.SelectedIndex < 0 || model.SelectedIndex >= len(model.VisibleLogs) {
		return
	}
	s.logList.ScrollTo(model.SelectedIndex)
}

func autoFilterName(model appstate.Model) string {
	if model.SelectedPackage != "" {
		return model.SelectedPackage
	}
	query := strings.TrimSpace(model.Filter.Draft)
	if query == "" {
		return "H5 日志"
	}
	if len(query) > 24 {
		return query[:24]
	}
	return query
}
