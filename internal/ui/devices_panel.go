package ui

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutDeviceSelector(gtx layout.Context, model appstate.Model) layout.Dimensions {
	label := "未选择设备"
	for _, device := range model.Devices {
		if device.ID == model.SelectedDevice {
			label = device.Model
			break
		}
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return iconGlyph(gtx, "device", s.colors.textMuted)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 6)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return selectorBoxStyled(gtx, &s.deviceSelectorButton, selectorStyle{
				fill:      s.colors.panelBar,
				border:    s.colors.panelBorder,
				label:     label,
				labelColor: s.colors.textSecondary,
				minWidth:  103,
			})
		}),
	)
}

func (s *Shell) layoutSavedFilterSelector(gtx layout.Context, model appstate.Model) layout.Dimensions {
	label := "过滤器"
	for _, filter := range model.Filter.Saved {
		if filter.ID == model.Filter.ActiveFilterID {
			label = filter.Name
			break
		}
	}
	if model.Filter.ActiveFilterID == "" && strings.TrimSpace(model.Filter.Applied) != "" {
		label = "自定义"
	}
	return selectorBoxStyled(gtx, &s.filterSelectorButton, selectorStyle{
		fill:       s.colors.panelBar,
		border:     s.colors.panelBorder,
		label:      label,
		labelColor: s.colors.accent,
		minWidth:   100,
	})
}

func (s *Shell) layoutPackageSelector(gtx layout.Context, model appstate.Model) layout.Dimensions {
	label := model.SelectedPackage
	if label == "" {
		label = "选择包名"
	}
	return selectorBoxStyled(gtx, &s.packageSelectorButton, selectorStyle{
		fill:       s.colors.panelBar,
		border:     s.colors.panelBorder,
		label:      label,
		labelColor: appColor{R: 83, G: 234, B: 253, A: 178},
		minWidth:   130,
	})
}

type selectorStyle struct {
	fill       appColor
	border     appColor
	label      string
	labelColor appColor
	minWidth   int
}

func selectorBoxStyled(gtx layout.Context, trigger *widget.Clickable, style selectorStyle) layout.Dimensions {
	minWidth := gtx.Dp(unit.Dp(style.minWidth))
	return trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return borderedBox(
			gtx,
			style.fill,
			style.border,
			unit.Dp(4),
			layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
				Left:   unit.Dp(9),
				Right:  unit.Dp(9),
			},
			func(gtx layout.Context) layout.Dimensions {
				if gtx.Constraints.Min.X < minWidth {
					gtx.Constraints.Min.X = minWidth
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
					gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						text := material.Label(NewTheme(), unit.Sp(11), style.label)
						text.Color = style.labelColor
						return text.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layoutSpacerWidth(gtx, 6)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return iconGlyph(gtx, "chevron-down", appColor{R: 136, G: 136, B: 136, A: 255})
					}),
				)
			},
		)
	})
}

func (s *Shell) layoutDeviceMenu(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return menuBox(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			deviceMenuItems(s, gtx, model)...,
		)
	})
}

func deviceMenuItems(s *Shell, gtx layout.Context, model appstate.Model) []layout.FlexChild {
	items := make([]layout.FlexChild, 0, len(model.Devices))
	for index, device := range model.Devices {
		idx := index
		label := device.Model
		items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return menuItem(gtx, s.theme, &s.deviceButtons[idx], label, device.ID == model.SelectedDevice)
		}))
	}
	return items
}

func (s *Shell) layoutSavedFilterMenu(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return menuBox(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, len(model.Filter.Saved))
		for index, filter := range model.Filter.Saved {
			idx := index
			label := filter.Name
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return menuItem(gtx, s.theme, &s.filterButtons[idx], label, filter.ID == model.Filter.ActiveFilterID)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

func (s *Shell) layoutPackageMenu(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return menuBox(gtx, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return menuItem(gtx, s.theme, &s.packageRootButton, "设备级 H5", model.SelectedPackage == "")
			}),
		}
		for index, pkg := range model.Packages {
			idx := index
			name := pkg.Name
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return menuItem(gtx, s.theme, &s.packageButtons[idx], name, name == model.SelectedPackage)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

func menuBox(gtx layout.Context, child layout.Widget) layout.Dimensions {
	return borderedBox(
		gtx,
		appColor{R: 30, G: 30, B: 30, A: 255},
		appColor{R: 45, G: 45, B: 45, A: 255},
		unit.Dp(4),
		layout.UniformInset(unit.Dp(6)),
		child,
	)
}

func menuItem(gtx layout.Context, theme *material.Theme, button *widget.Clickable, label string, selected bool) layout.Dimensions {
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return menuStaticItem(gtx, theme, label, selected)
	})
}

func menuStaticItem(gtx layout.Context, theme *material.Theme, label string, selected bool) layout.Dimensions {
	gtx.Constraints.Min.X = max(gtx.Constraints.Min.X, gtx.Dp(unit.Dp(160)))
	fill := appColor{}
	textColor := appColor{R: 192, G: 192, B: 192, A: 255}
	if selected {
		fill = appColor{R: 22, G: 36, B: 86, A: 255}
		textColor = appColor{R: 81, G: 162, B: 255, A: 255}
	}
	return paintBox(gtx, fill, appColor{}, unit.Dp(4), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			text := material.Label(theme, unit.Sp(11), label)
			text.Color = textColor
			return text.Layout(gtx)
		})
	})
}

func (s *Shell) layoutOverlayMenus(gtx layout.Context, model appstate.Model) layout.Dimensions {
	dims := layout.Dimensions{}
	if s.deviceMenuOpen {
		dims = maxDims(dims, s.layoutMenuAt(gtx, 250, 32, func(gtx layout.Context) layout.Dimensions {
			return s.layoutDeviceMenu(gtx, model)
		}))
	}
	if s.filterMenuOpen {
		dims = maxDims(dims, s.layoutMenuAt(gtx, 394, 32, func(gtx layout.Context) layout.Dimensions {
			return s.layoutSavedFilterMenu(gtx, model)
		}))
	}
	if s.packageMenuOpen {
		dims = maxDims(dims, s.layoutMenuAt(gtx, 8, 72, func(gtx layout.Context) layout.Dimensions {
			return s.layoutPackageMenu(gtx, model)
		}))
	}
	if s.historyMenuOpen {
		dims = maxDims(dims, s.layoutMenuAt(gtx, 730, 72, func(gtx layout.Context) layout.Dimensions {
			return s.layoutHistoryMenu(gtx, model)
		}))
	}
	return dims
}

func (s *Shell) layoutHistoryMenu(gtx layout.Context, model appstate.Model) layout.Dimensions {
	history := visibleHistory(model)
	return menuBox(gtx, func(gtx layout.Context) layout.Dimensions {
		if len(history) == 0 {
			return menuStaticItem(gtx, s.theme, "暂无历史", false)
		}

		children := make([]layout.FlexChild, 0, len(history))
		for index, query := range history {
			idx := index
			label := historyLabel(query)
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return menuItem(gtx, s.theme, &s.historyButtons[idx], label, false)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

func (s *Shell) layoutMenuAt(gtx layout.Context, xDp int, yDp int, child layout.Widget) layout.Dimensions {
	return s.layoutMenuAtPx(gtx, gtx.Dp(unit.Dp(xDp)), gtx.Dp(unit.Dp(yDp)), child)
}

func (s *Shell) layoutMenuAtPx(gtx layout.Context, xPx int, yPx int, child layout.Widget) layout.Dimensions {
	stack := op.Offset(image.Pt(xPx, yPx)).Push(gtx.Ops)
	defer stack.Pop()
	return child(gtx)
}

func maxDims(left layout.Dimensions, right layout.Dimensions) layout.Dimensions {
	if right.Size.X > left.Size.X {
		left.Size.X = right.Size.X
	}
	if right.Size.Y > left.Size.Y {
		left.Size.Y = right.Size.Y
	}
	return left
}

func (s *Shell) layoutDetailToggleRail(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(24)), gtx.Constraints.Max.Y)
	return paintBox(gtx, s.colors.panel, s.colors.panelBorder, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := "chevron-right"
			if s.detailCollapsed {
				label = "chevron-left"
			}
			return iconButton(gtx, s.theme, &s.toggleDetailButton, label, s.colors.textMuted)
		})
	})
}

func deviceSummary(model appstate.Model) string {
	if model.SelectedDevice == "" {
		return "未连接设备"
	}
	return fmt.Sprintf("设备 %s", model.SelectedDevice)
}
