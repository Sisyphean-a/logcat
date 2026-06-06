package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutControls(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.H6(s.theme, "Controls").Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutPrimaryControls(gtx, model)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutSearchControls(gtx, model)
		}),
	)
}

func (s *Shell) layoutPrimaryControls(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutPauseGroup(gtx, model)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(s.theme, &s.clearButton, "清空视图").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := "恢复跟随"
					if s.followLogs {
						label = "跟随中"
					}
					return material.Button(s.theme, &s.followButton, label).Layout(gtx)
				})
			}),
		)
	})
}

func (s *Shell) layoutPauseGroup(gtx layout.Context, model appstate.Model) layout.Dimensions {
	if !model.Pause.Active {
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Button(s.theme, &s.pauseButton, "暂停").Layout(gtx)
		})
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.resumeKeepButton, "恢复并显示").Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.resumeDropButton, "恢复并丢弃").Layout(gtx)
			})
		}),
	)
}

func (s *Shell) layoutSearchControls(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
			gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return s.layoutSearchEditor(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(s.theme, &s.prevMatchButton, "上一条").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(s.theme, &s.nextMatchButton, "下一条").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := searchSummary(model)
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, material.Body2(s.theme, text).Layout)
			}),
		)
	})
}

func (s *Shell) layoutSearchEditor(gtx layout.Context) layout.Dimensions {
	border := widget.Border{
		Color:        s.theme.Palette.ContrastBg,
		CornerRadius: unit.Dp(6),
		Width:        unit.Dp(1),
	}
	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			editor := material.Editor(s.theme, &s.searchEditor, "搜索当前视图")
			return editor.Layout(gtx)
		})
	})
}

func searchSummary(model appstate.Model) string {
	total := len(model.Search.MatchIndexes)
	if total == 0 {
		return "0 / 0"
	}
	return fmt.Sprintf("%d / %d", model.Search.Current+1, total)
}
