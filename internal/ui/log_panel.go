package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutLogs(gtx layout.Context, model appstate.Model) layout.Dimensions {
	s.syncSelectedLog(model)
	s.syncAutoFollow(len(model.VisibleLogs))

	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.H6(s.theme, "Logs").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			dims := s.logList.Layout(gtx, len(model.VisibleLogs), func(gtx layout.Context, index int) layout.Dimensions {
				return s.layoutLogRow(gtx, model, index)
			})
			s.syncAutoFollow(len(model.VisibleLogs))
			return dims
		}),
	)
}

func (s *Shell) layoutLogRow(gtx layout.Context, model appstate.Model, index int) layout.Dimensions {
	text := material.Body2(s.theme, model.VisibleLogs[index].Display)
	fill := s.logRowFill(model, index)

	return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return s.logButtons[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return paintRowBackground(gtx, fill, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, text.Layout)
			})
		})
	})
}

func (s *Shell) logRowFill(model appstate.Model, index int) color.NRGBA {
	current := model.Search.Current >= 0 &&
		model.Search.Current < len(model.Search.MatchIndexes) &&
		model.Search.MatchIndexes[model.Search.Current] == index
	selected := model.SelectedIndex == index
	matched := containsIndex(model.Search.MatchIndexes, index)

	switch {
	case selected || current:
		return s.theme.Palette.ContrastBg
	case matched:
		return color.NRGBA{R: 38, G: 54, B: 78, A: 255}
	default:
		return color.NRGBA{R: 24, G: 26, B: 32, A: 255}
	}
}

func paintRowBackground(gtx layout.Context, fill color.NRGBA, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rr := gtx.Dp(unit.Dp(6))
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, rr).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, fill)
	call.Add(gtx.Ops)
	return dims
}

func containsIndex(indexes []int, target int) bool {
	for _, index := range indexes {
		if index == target {
			return true
		}
	}
	return false
}
