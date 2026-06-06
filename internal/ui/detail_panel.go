package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutDetail(gtx layout.Context, _ appstate.Model) layout.Dimensions {
	title := material.H6(s.theme, "Detail")
	model := s.controller.Model()
	item, ok := s.selectedLogItem(model)
	if !ok {
		body := material.Body2(s.theme, "Select a log line to inspect details.")
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, title.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, body.Layout)
			}),
		)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, title.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutCopyActions(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return detailSection(gtx, s.theme, "Time", item.Entry.TimeText)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return detailSection(gtx, s.theme, "Level", item.Entry.Level)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return detailSection(gtx, s.theme, "Tag", item.Entry.Tag)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return detailSection(gtx, s.theme, "Message", item.Entry.Message)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return detailSection(gtx, s.theme, "Raw", item.Entry.Raw)
		}),
	)
}

func (s *Shell) layoutCopyActions(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.copyLineButton, "复制当前行").Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(s.theme, &s.copyRawButton, "复制原始日志").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.copyMessageButton, "复制 message").Layout(gtx)
			}),
		)
	})
}

func detailSection(gtx layout.Context, theme *material.Theme, title, value string) layout.Dimensions {
	head := material.Body1(theme, title)
	body := material.Body2(theme, value)

	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(head.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(2)).Layout(gtx, body.Layout)
			}),
		)
	})
}
