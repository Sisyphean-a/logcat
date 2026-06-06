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
)

type Shell struct {
	controller *appstate.Controller
	theme      *material.Theme

	searchEditor      widget.Editor
	deviceButtons     []widget.Clickable
	packageButtons    []widget.Clickable
	processButtons    []widget.Clickable
	logButtons        []widget.Clickable
	foregroundButton  widget.Clickable
	scopeUserButton   widget.Clickable
	scopeSystemButton widget.Clickable
	scopeAllButton    widget.Clickable
	pauseButton       widget.Clickable
	resumeKeepButton  widget.Clickable
	resumeDropButton  widget.Clickable
	clearButton       widget.Clickable
	prevMatchButton   widget.Clickable
	nextMatchButton   widget.Clickable
	followButton      widget.Clickable
	copyLineButton    widget.Clickable
	copyRawButton     widget.Clickable
	copyMessageButton widget.Clickable
	deviceList        widget.List
	packageList       widget.List
	processList       widget.List
	logList           widget.List
	followLogs        bool
	lastLogCount      int
	lastSelected      int
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
	searchEditor := widget.Editor{SingleLine: true, Submit: true}
	return &Shell{
		controller:   controller,
		theme:        NewTheme(),
		searchEditor: searchEditor,
		deviceList:   widget.List{List: layout.List{Axis: layout.Vertical}},
		packageList:  widget.List{List: layout.List{Axis: layout.Vertical}},
		processList:  widget.List{List: layout.List{Axis: layout.Vertical}},
		logList:      widget.List{List: layout.List{Axis: layout.Vertical}},
		followLogs:   true,
		lastSelected: -1,
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
	s.syncDeviceButtons(len(model.Devices))
	s.syncLogButtons(len(model.VisibleLogs))
	s.handleActions(gtx, model)
	model = s.controller.Model()
	s.syncDeviceButtons(len(model.Devices))
	s.syncLogButtons(len(model.VisibleLogs))

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = gtx.Dp(260)
				return s.layoutDevices(gtx, model)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutControls(gtx, model)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutStatus(gtx, model)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(
							gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return s.layoutLogs(gtx, model)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = gtx.Dp(320)
								return s.layoutDetail(gtx, model)
							}),
						)
					}),
				)
			}),
		)
	})
}

func (s *Shell) syncDeviceButtons(count int) {
	for len(s.deviceButtons) < count {
		s.deviceButtons = append(s.deviceButtons, widget.Clickable{})
	}
}

func (s *Shell) layoutStatus(gtx layout.Context, model appstate.Model) layout.Dimensions {
	card := material.Body1(s.theme, "Status: "+model.Status)
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, card.Layout)
}
