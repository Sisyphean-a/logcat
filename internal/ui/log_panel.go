package ui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutFilterEditor(gtx layout.Context) layout.Dimensions {
	return borderedBox(
		gtx,
		s.colors.panelBar,
		s.colors.panelBorder,
		unit.Dp(4),
		layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(7),
			Right:  unit.Dp(9),
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconGlyph(gtx, "search", appColor{R: 136, G: 136, B: 136, A: 255})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 4)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(s.theme, &s.filterEditor, "tag:chromium & message:[H5]")
					editor.Color = s.colors.textPrimary
					editor.HintColor = s.colors.textMuted
					return editor.Layout(gtx)
				}),
			)
		},
	)
}

func (s *Shell) layoutFollowToggle(gtx layout.Context) layout.Dimensions {
	return s.followButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return paintBox(gtx, switchTrackColor(s.followLogs), appColor{}, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(28)), gtx.Dp(unit.Dp(14)))
					gtx.Constraints.Max = gtx.Constraints.Min
					return layout.Stack{}.Layout(
						gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							return layout.Dimensions{Size: gtx.Constraints.Min}
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							x := 2
							if s.followLogs {
								x = 14
							}
							offset := op.Offset(image.Pt(gtx.Dp(unit.Dp(x)), gtx.Dp(unit.Dp(2)))).Push(gtx.Ops)
							defer offset.Pop()
							return paintBox(gtx, appColor{R: 255, G: 255, B: 255, A: 255}, appColor{}, unit.Dp(5), func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(10)), gtx.Dp(unit.Dp(10)))
								gtx.Constraints.Max = gtx.Constraints.Min
								return layout.Dimensions{Size: gtx.Constraints.Min}
							})
						}),
					)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutSpacerWidth(gtx, 6)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(s.theme, unit.Sp(11), "滚动")
				label.Color = appColor{R: 153, G: 153, B: 153, A: 255}
				return label.Layout(gtx)
			}),
		)
	})
}

func (s *Shell) layoutHistorySelector(gtx layout.Context) layout.Dimensions {
	return historyBox(gtx, &s.historyMenuButton)
}

func (s *Shell) layoutSaveButton(gtx layout.Context) layout.Dimensions {
	return borderedBox(
		gtx,
		s.colors.panelBar,
		s.colors.panelBorder,
		unit.Dp(4),
		layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(7),
			Right:  unit.Dp(9),
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconGlyph(gtx, "save", s.colors.textSecondary)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 4)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Label(s.theme, unit.Sp(11), "保存")
					label.Color = appColor{R: 176, G: 176, B: 176, A: 255}
					return label.Layout(gtx)
				}),
			)
		},
	)
}

func (s *Shell) layoutLogs(gtx layout.Context, model appstate.Model) layout.Dimensions {
	s.syncSelectedLog(model)
	s.syncAutoFollow(len(model.VisibleLogs))

	return paintBox(gtx, s.colors.panel, appColor{}, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutLogHeader(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				dims := s.logList.Layout(gtx, len(model.VisibleLogs), func(gtx layout.Context, index int) layout.Dimensions {
					return s.layoutLogRow(gtx, model, index)
				})
				s.syncAutoFollow(len(model.VisibleLogs))
				return dims
			}),
		)
	})
}

func (s *Shell) layoutLogHeader(gtx layout.Context) layout.Dimensions {
	return paintBox(gtx, s.colors.panelBar, s.colors.panelBorder, 0, func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(14),
			Right:  unit.Dp(8),
		}
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return headerCell(gtx, s.theme, s.colors.textMuted, "时间", 96) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return headerCell(gtx, s.theme, s.colors.textMuted, "级", 32) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return headerCell(gtx, s.theme, s.colors.textMuted, "标签", 80) }),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return headerText(gtx, s.theme, s.colors.textMuted, "消息") }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return headerCell(gtx, s.theme, s.colors.textMuted, "来源", 160) }),
			)
		})
	})
}

func headerCell(gtx layout.Context, theme *material.Theme, color appColor, text string, width int) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Dp(unit.Dp(width))
	gtx.Constraints.Max.X = gtx.Constraints.Min.X
	return headerText(gtx, theme, color, text)
}

func headerText(gtx layout.Context, theme *material.Theme, color appColor, text string) layout.Dimensions {
	label := material.Label(theme, unit.Sp(11), text)
	label.Color = color
	return label.Layout(gtx)
}

func (s *Shell) layoutLogRow(gtx layout.Context, model appstate.Model, index int) layout.Dimensions {
	item := model.VisibleLogs[index]
	fill := s.logRowFill(model, index)
	return s.logButtons[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return paintRowBackground(gtx, fill, func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
				Left:   unit.Dp(0),
				Right:  unit.Dp(8),
			}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return colorStripe(gtx, severityStripe(item.Entry.Level))
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return rowCell(gtx, s.theme, s.colors.textMuted, item.Entry.TimeText, 96) }),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return severityBadge(gtx, s.theme, s.colors, item.Entry.Level, 32) }),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return rowCell(gtx, s.theme, severityTextColor(s.colors, item.Entry.Level), item.Entry.Tag, 80) }),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return rowText(gtx, s.theme, messageColor(s.colors, item.Entry.Level), item.Entry.Message) }),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return rowCell(gtx, s.theme, s.colors.textMuted, item.Entry.Source, 160) }),
				)
			})
		})
	})
}

func rowCell(gtx layout.Context, theme *material.Theme, color appColor, text string, width int) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Dp(unit.Dp(width))
	gtx.Constraints.Max.X = gtx.Constraints.Min.X
	return rowText(gtx, theme, color, text)
}

func rowText(gtx layout.Context, theme *material.Theme, color appColor, text string) layout.Dimensions {
	label := material.Label(theme, unit.Sp(11), text)
	label.Color = color
	return label.Layout(gtx)
}

func colorStripe(gtx layout.Context, fill appColor) layout.Dimensions {
	gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(4)), gtx.Dp(unit.Dp(22)))
	gtx.Constraints.Max = gtx.Constraints.Min
	paint.FillShape(gtx.Ops, fill, clip.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Op())
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func severityBadge(gtx layout.Context, theme *material.Theme, colors appPalette, level string, width int) layout.Dimensions {
	bg, fg := severityColors(colors, level)
	gtx.Constraints.Min.X = gtx.Dp(unit.Dp(width))
	gtx.Constraints.Max.X = gtx.Constraints.Min.X
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return borderedBox(
			gtx,
			bg,
			appColor{},
			unit.Dp(4),
			layout.Inset{
				Top:    unit.Dp(1),
				Bottom: unit.Dp(1),
				Left:   unit.Dp(5),
				Right:  unit.Dp(5),
			},
			func(gtx layout.Context) layout.Dimensions {
				label := material.Label(theme, unit.Sp(10), level)
				label.Color = fg
				return label.Layout(gtx)
			},
		)
	})
}

func (s *Shell) logRowFill(model appstate.Model, index int) color.NRGBA {
	if model.SelectedIndex == index {
		return appColor{R: 26, G: 32, B: 44, A: 255}
	}
	return s.colors.panel
}

func paintRowBackground(gtx layout.Context, fill color.NRGBA, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rect := image.Rectangle{Max: dims.Size}
	defer clip.Rect(rect).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, fill)
	call.Add(gtx.Ops)
	return dims
}

func severityColors(colors appPalette, level string) (appColor, appColor) {
	switch level {
	case "E", "F":
		return colors.errorBg, colors.errorFg
	case "W":
		return colors.warnBg, colors.warnFg
	default:
		return colors.infoBg, colors.infoFg
	}
}

func severityTextColor(colors appPalette, level string) appColor {
	switch level {
	case "E", "F":
		return colors.errorFg
	case "W":
		return colors.warnFg
	default:
		return colors.infoFg
	}
}

func severityStripe(level string) appColor {
	switch level {
	case "E", "F":
		return appColor{R: 251, G: 44, B: 54, A: 255}
	case "W":
		return appColor{R: 255, G: 201, B: 24, A: 255}
	default:
		return appColor{R: 43, G: 127, B: 255, A: 255}
	}
}

func messageColor(colors appPalette, level string) appColor {
	switch level {
	case "E", "F":
		return appColor{R: 255, G: 162, B: 162, A: 255}
	default:
		return colors.textPrimary
	}
}

func historyBox(gtx layout.Context, trigger *widget.Clickable) layout.Dimensions {
	width := gtx.Dp(unit.Dp(58))
	return trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return borderedBox(
			gtx,
			appColor{R: 30, G: 30, B: 30, A: 255},
			appColor{R: 45, G: 45, B: 45, A: 255},
			unit.Dp(4),
			layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
				Left:   unit.Dp(9),
				Right:  unit.Dp(9),
			},
			func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = width
				gtx.Constraints.Max.X = width
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return iconGlyph(gtx, "chevron-down", appColor{R: 136, G: 136, B: 136, A: 255})
				})
			},
		)
	})
}

func visibleHistory(model appstate.Model) []string {
	history := model.Filter.History
	if len(history) > 5 {
		history = history[:5]
	}
	return history
}

func historyLabel(query string) string {
	trimmed := strings.TrimSpace(query)
	if len(trimmed) <= 32 {
		return trimmed
	}
	return trimmed[:32] + "..."
}

func switchTrackColor(on bool) appColor {
	if on {
		return appColor{R: 0, G: 146, B: 184, A: 255}
	}
	return appColor{R: 45, G: 45, B: 45, A: 255}
}
