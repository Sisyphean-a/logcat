package ui

import (
	"context"
	"time"

	gioapp "gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
	"github.com/xiakn/logcat/internal/storage"
)

type Shell struct {
	controller *appstate.Controller
	theme      *material.Theme
	colors     appPalette
	exportLogs func([]appstate.LogViewItem) (string, error)

	filterEditor         widget.Editor
	saveNameEditor       widget.Editor
	deviceSelectorButton widget.Clickable
	filterSelectorButton widget.Clickable
	packageSelectorButton widget.Clickable
	historyMenuButton    widget.Clickable
	packageRootButton    widget.Clickable
	deviceButtons        []widget.Clickable
	packageButtons       []widget.Clickable
	filterButtons        []widget.Clickable
	historyButtons       []widget.Clickable
	logButtons           []widget.Clickable
	pauseButton          widget.Clickable
	clearButton          widget.Clickable
	exportButton         widget.Clickable
	followButton         widget.Clickable
	saveFilterButton     widget.Clickable
	toggleDetailButton   widget.Clickable
	logList              widget.List
	followLogs           bool
	lastLogCount         int
	lastSelected         int
	detailCollapsed      bool
	deviceMenuOpen       bool
	filterMenuOpen       bool
	packageMenuOpen      bool
	historyMenuOpen      bool
}

func Run(window *gioapp.Window, controller *appstate.Controller) error {
	shell := newShell(controller)
	shell.bootstrap(window)

	var ops op.Ops
	for {
		event := window.Event()
		switch event := event.(type) {
		case gioapp.DestroyEvent:
			return event.Err
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, event)
			shell.layout(gtx)
			event.Frame(gtx.Ops)
		}
	}
}

func newShell(controller *appstate.Controller) *Shell {
	filterEditor := widget.Editor{SingleLine: true, Submit: true}
	saveNameEditor := widget.Editor{SingleLine: true, Submit: true}
	return &Shell{
		controller:      controller,
		theme:           NewTheme(),
		colors:          defaultPalette(),
		exportLogs:      storage.ExportVisibleLogs,
		filterEditor:    filterEditor,
		saveNameEditor:  saveNameEditor,
		logList:         widget.List{List: layout.List{Axis: layout.Vertical}},
		followLogs:      true,
		lastSelected:    -1,
		detailCollapsed: false,
		deviceMenuOpen:  false,
		filterMenuOpen:  false,
		packageMenuOpen: false,
		historyMenuOpen: false,
	}
}

func (s *Shell) bootstrap(window *gioapp.Window) {
	go func() {
		_ = s.controller.Load(context.Background())
		window.Invalidate()
	}()

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			window.Invalidate()
		}
	}()
}

func (s *Shell) layout(gtx layout.Context) layout.Dimensions {
	model := s.controller.Model()
	s.syncButtons(model)
	s.syncEditors(model)
	s.handleActions(gtx, model)
	model = s.controller.Model()
	s.syncButtons(model)
	s.syncEditors(model)

	gtx.Constraints.Min = gtx.Constraints.Max
	return layout.Stack{}.Layout(
		gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return paintBox(gtx, s.colors.window, appColor{}, 0, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(40))
						return s.layoutToolbar(gtx, model)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(
							gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(
									gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(36))
										return s.layoutFilterBar(gtx, model)
									}),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return s.layoutLogs(gtx, model)
									}),
								)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								width := gtx.Dp(unit.Dp(280))
								if s.detailCollapsed {
									width = gtx.Dp(unit.Dp(24))
								}
								gtx.Constraints.Min.X = width
								gtx.Constraints.Max.X = width
								return s.layoutDetail(gtx, model)
							}),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(20))
						return s.layoutStatusBar(gtx, model)
					}),
				)
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return s.layoutOverlayMenus(gtx, model)
		}),
	)
}

func (s *Shell) syncEditors(model appstate.Model) {
	if s.filterEditor.Text() != model.Filter.Draft {
		s.filterEditor.SetText(model.Filter.Draft)
	}
}

func (s *Shell) syncButtons(model appstate.Model) {
	s.deviceButtons = syncClickables(s.deviceButtons, len(model.Devices))
	s.packageButtons = syncClickables(s.packageButtons, len(model.Packages))
	s.filterButtons = syncClickables(s.filterButtons, len(model.Filter.Saved))
	s.historyButtons = syncClickables(s.historyButtons, len(visibleHistory(model)))
	s.logButtons = syncClickables(s.logButtons, len(model.VisibleLogs))
}

func (s *Shell) persistFilters() {
	if s.controller == nil {
		return
	}
	model := s.controller.Model()
	_ = storage.SaveFilterState(model.Filter.Saved, model.Filter.History)
}
