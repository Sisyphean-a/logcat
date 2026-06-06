package ui

import (
	"context"
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
)

type copyMode uint8

const (
	copyLine copyMode = iota
	copyRaw
	copyMessage
)

func (s *Shell) handleActions(gtx layout.Context, model appstate.Model) {
	if s.controller == nil {
		return
	}

	s.syncBindingButtons(model)
	s.handleDeviceClicks(gtx, model)
	model = s.controller.Model()
	s.syncBindingButtons(model)
	s.handleScopeClicks(gtx)
	s.handleForegroundClicks(gtx)
	model = s.controller.Model()
	s.syncBindingButtons(model)
	s.handlePackageClicks(gtx, model)
	model = s.controller.Model()
	s.syncBindingButtons(model)
	s.handleProcessClicks(gtx, model)
	model = s.controller.Model()
	s.syncBindingButtons(model)
	s.handleSearchUpdate(model)
	model = s.controller.Model()
	s.syncBindingButtons(model)
	s.handlePauseClicks(gtx, model)
	s.handleSearchNavClicks(gtx)
	s.handleLogClicks(gtx, model)
	s.handleCopyClicks(gtx)
}

func (s *Shell) syncBindingButtons(model appstate.Model) {
	s.syncDeviceButtons(len(model.Devices))
	s.syncPackageButtons(len(model.Packages))
	s.syncProcessButtons(len(model.Processes))
	s.syncLogButtons(len(model.VisibleLogs))
}

func (s *Shell) handleDeviceClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Devices {
		for s.deviceButtons[index].Clicked(gtx) {
			_ = s.controller.SelectDevice(context.Background(), model.Devices[index].ID)
		}
	}
}

func (s *Shell) handleScopeClicks(gtx layout.Context) {
	for s.scopeUserButton.Clicked(gtx) {
		_ = s.controller.SetPackageScope(context.Background(), adb.PackageScopeUser)
	}
	for s.scopeSystemButton.Clicked(gtx) {
		_ = s.controller.SetPackageScope(context.Background(), adb.PackageScopeSystem)
	}
	for s.scopeAllButton.Clicked(gtx) {
		_ = s.controller.SetPackageScope(context.Background(), adb.PackageScopeAll)
	}
}

func (s *Shell) handleForegroundClicks(gtx layout.Context) {
	for s.foregroundButton.Clicked(gtx) {
		_ = s.controller.SelectForegroundPackage(context.Background())
	}
}

func (s *Shell) handlePackageClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Packages {
		for s.packageButtons[index].Clicked(gtx) {
			_ = s.controller.SelectPackage(context.Background(), model.Packages[index].Name)
		}
	}
}

func (s *Shell) handleProcessClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.Processes {
		for s.processButtons[index].Clicked(gtx) {
			_ = s.controller.SelectProcess(context.Background(), model.Processes[index].Name)
		}
	}
}

func (s *Shell) handleSearchUpdate(model appstate.Model) {
	query := strings.TrimSpace(s.searchEditor.Text())
	if query == model.Search.Query {
		return
	}
	s.controller.SetSearchQuery(query)
}

func (s *Shell) handlePauseClicks(gtx layout.Context, model appstate.Model) {
	if model.Pause.Active {
		for s.resumeKeepButton.Clicked(gtx) {
			s.controller.ResumeKeep()
		}
		for s.resumeDropButton.Clicked(gtx) {
			s.controller.ResumeDiscard()
		}
	} else {
		for s.pauseButton.Clicked(gtx) {
			s.controller.Pause()
		}
	}

	for s.clearButton.Clicked(gtx) {
		s.controller.ClearVisible()
	}
	for s.followButton.Clicked(gtx) {
		s.followLogs = true
		s.logList.Position.BeforeEnd = false
	}
}

func (s *Shell) handleSearchNavClicks(gtx layout.Context) {
	for s.prevMatchButton.Clicked(gtx) {
		s.followLogs = false
		s.controller.PrevMatch()
	}
	for s.nextMatchButton.Clicked(gtx) {
		s.followLogs = false
		s.controller.NextMatch()
	}
}

func (s *Shell) handleLogClicks(gtx layout.Context, model appstate.Model) {
	for index := range model.VisibleLogs {
		for s.logButtons[index].Clicked(gtx) {
			s.followLogs = false
			s.controller.SelectLog(index)
		}
	}
}

func (s *Shell) handleCopyClicks(gtx layout.Context) {
	for s.copyLineButton.Clicked(gtx) {
		s.writeClipboard(gtx, copyLine)
	}
	for s.copyRawButton.Clicked(gtx) {
		s.writeClipboard(gtx, copyRaw)
	}
	for s.copyMessageButton.Clicked(gtx) {
		s.writeClipboard(gtx, copyMessage)
	}
}

func (s *Shell) writeClipboard(gtx layout.Context, mode copyMode) {
	text, ok := s.selectedClipboardText(s.controller.Model(), mode)
	if !ok {
		return
	}
	gtx.Execute(clipboard.WriteCmd{
		Type: "application/text",
		Data: io.NopCloser(strings.NewReader(text)),
	})
}

func (s *Shell) syncLogButtons(count int) {
	for len(s.logButtons) < count {
		s.logButtons = append(s.logButtons, widgetClickable())
	}
}

func (s *Shell) syncPackageButtons(count int) {
	s.packageButtons = syncClickables(s.packageButtons, count)
}

func (s *Shell) syncProcessButtons(count int) {
	s.processButtons = syncClickables(s.processButtons, count)
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

func (s *Shell) selectedClipboardText(model appstate.Model, mode copyMode) (string, bool) {
	item, ok := s.selectedLogItem(model)
	if !ok {
		return "", false
	}

	switch mode {
	case copyRaw:
		return item.Entry.Raw, true
	case copyMessage:
		return item.Entry.Message, true
	default:
		return item.Display, true
	}
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
