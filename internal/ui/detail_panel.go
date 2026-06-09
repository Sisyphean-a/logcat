package ui

import (
	"fmt"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutDetail(gtx layout.Context, model appstate.Model) layout.Dimensions {
	if s.detailCollapsed {
		return s.layoutDetailToggleRail(gtx)
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutDetailToggleRail(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return paintBox(gtx, s.colors.panel, s.colors.panelBorder, 0, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutDetailHeader(gtx)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						item, ok := s.selectedLogItem(model)
						if !ok {
							return s.layoutDetailEmpty(gtx)
						}
						return s.layoutDetailBody(gtx, item)
					}),
				)
			})
		}),
	)
}

func (s *Shell) layoutDetailHeader(gtx layout.Context) layout.Dimensions {
	return paintBox(gtx, s.colors.panel, s.colors.panelBorder, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(10),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(s.theme, unit.Sp(11), "详情面板")
			label.Color = s.colors.textMuted
			return label.Layout(gtx)
		})
	})
}

func (s *Shell) layoutDetailEmpty(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return borderedBox(
					gtx,
					s.colors.panelBar,
					s.colors.panelBorder,
					unit.Dp(16),
					layout.UniformInset(unit.Dp(8)),
					func(gtx layout.Context) layout.Dimensions {
						return iconGlyph(gtx, "up-circle", appColor{R: 85, G: 85, B: 85, A: 255})
					},
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutSpacerHeight(gtx, 8)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(s.theme, unit.Sp(11), "选择一条日志\n查看详情")
				label.Color = s.colors.textMuted
				return label.Layout(gtx)
			}),
		)
	})
}

func (s *Shell) layoutDetailBody(gtx layout.Context, item appstate.LogViewItem) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(12),
		Bottom: unit.Dp(12),
		Left:   unit.Dp(12),
		Right:  unit.Dp(12),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "时间", item.Entry.TimeText)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "级别", item.Entry.Level)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "标签", item.Entry.Tag)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "来源", item.Entry.Source)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "消息", item.Entry.Message)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return detailField(gtx, s.theme, s.colors, "原始", item.Entry.Raw)
			}),
		)
	})
}

func detailField(gtx layout.Context, theme *material.Theme, colors appPalette, title string, value string) layout.Dimensions {
	if value == "" {
		value = "-"
	}
	return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(theme, unit.Sp(10), title)
				label.Color = colors.textMuted
				return label.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutSpacerHeight(gtx, 4)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Label(theme, unit.Sp(11), value)
				label.Color = colors.textPrimary
				return label.Layout(gtx)
			}),
		)
	})
}

func (s *Shell) layoutStatusBar(gtx layout.Context, model appstate.Model) layout.Dimensions {
	totalLogs := model.TotalLogs
	visibleLogs := len(model.VisibleLogs)
	filterLabel := "H5 日志"
	statusMessage := secondaryStatusMessage(model)
	for _, filter := range model.Filter.Saved {
		if filter.ID == model.Filter.ActiveFilterID {
			filterLabel = filter.Name
			break
		}
	}
	if model.Filter.ActiveFilterID == "" {
		if strings.TrimSpace(model.Filter.Applied) == "" {
			filterLabel = "-"
		} else {
			filterLabel = "自定义"
		}
	}

	return paintBox(gtx, appColor{R: 17, G: 17, B: 17, A: 255}, s.colors.panelBorder, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(3),
			Bottom: unit.Dp(3),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.success, "adb 已连接") }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 12) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "设备 "+selectedDeviceName(model)) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 12) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "包名 "+emptyDash(model.SelectedPackage)) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 12) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "日志") }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 4) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textSecondary, fmt.Sprintf("%d", totalLogs)) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 4) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "/") }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 4) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.accent, fmt.Sprintf("%d", visibleLogs)) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 4) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "条") }),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 12) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.accent, "过滤器 "+filterLabel) }),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if statusMessage == "" {
						return layout.Dimensions{}
					}
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return statusText(gtx, s.theme, s.colors.textMuted, statusMessage)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					scrollState := "关"
					scrollColor := s.colors.textMuted
					if s.followLogs {
						scrollState = "开"
						scrollColor = s.colors.success
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, s.colors.textMuted, "自动滚动") }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layoutSpacerWidth(gtx, 4) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusText(gtx, s.theme, scrollColor, scrollState) }),
					)
				}),
			)
		})
	})
}

func statusText(gtx layout.Context, theme *material.Theme, color appColor, text string) layout.Dimensions {
	label := material.Label(theme, unit.Sp(10), text)
	label.Color = color
	return label.Layout(gtx)
}

func selectedDeviceName(model appstate.Model) string {
	for _, device := range model.Devices {
		if device.ID == model.SelectedDevice {
			return device.Model
		}
	}
	return emptyDash(model.SelectedDevice)
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func secondaryStatusMessage(model appstate.Model) string {
	if model.Pause.Active {
		return fmt.Sprintf("暂停中，缓存 %d 条", model.Pause.BufferedCount)
	}
	switch {
	case model.Status == "", model.Status == "idle", model.Status == "running":
		return ""
	case strings.HasPrefix(model.Status, "adb "):
		return ""
	default:
		return model.Status
	}
}
