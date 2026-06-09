package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutToolbar(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return barSurface(gtx, s.colors.chromeBar, func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutBrand(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutToolbarSeparator(gtx, 20)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutADBStatus(gtx, model)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutToolbarSeparator(gtx, 20)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutDeviceSelector(gtx, model)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutToolbarSeparator(gtx, 20)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutSavedFilterSelector(gtx, model)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutToolbarActions(gtx, model)
				}),
			)
		})
	})
}

func (s *Shell) layoutFilterBar(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return barSurface(gtx, s.colors.panelBar, func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		}
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutPackageSelector(gtx, model)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 6)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return s.layoutFilterEditor(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 6)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutHistorySelector(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 6)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutFollowToggle(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutSpacerWidth(gtx, 6)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.layoutSaveButton(gtx)
				}),
			)
		})
	})
}

func (s *Shell) layoutBrand(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return tokenBox(gtx, s.colors.accentMuted, s.colors.accentOutline, "H5")
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 8)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(s.theme, unit.Sp(13), "Logcat Viewer")
			label.Color = s.colors.textPrimary
			return label.Layout(gtx)
		}),
	)
}

func (s *Shell) layoutADBStatus(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return paintBox(gtx, s.colors.success, appColor{}, unit.Dp(3), func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(6)), gtx.Dp(unit.Dp(6)))
				gtx.Constraints.Max = gtx.Constraints.Min
				return layout.Dimensions{Size: gtx.Constraints.Min}
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 6)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(s.theme, unit.Sp(11), "adb")
			label.Color = s.colors.textMuted
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 6)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(s.theme, unit.Sp(11), model.ADBStatus)
			label.Color = s.colors.success
			return label.Layout(gtx)
		}),
	)
}

func (s *Shell) layoutToolbarActions(gtx layout.Context, model appstate.Model) layout.Dimensions {
	pauseLabel := "pause"
	pauseColor := s.colors.textMuted
	if model.Pause.Active {
		pauseLabel = "play"
		pauseColor = s.colors.accent
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return iconButton(gtx, s.theme, &s.pauseButton, pauseLabel, pauseColor)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 2)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return iconButton(gtx, s.theme, &s.clearButton, "trash", s.colors.textMuted)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 2)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return iconButton(gtx, s.theme, &s.exportButton, "download", s.colors.textMuted)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutSpacerWidth(gtx, 8)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return iconGlyph(gtx, "settings", s.colors.textMuted)
		}),
	)
}

func barSurface(gtx layout.Context, fillColor appColor, child layout.Widget) layout.Dimensions {
	defer paint.ColorOp{Color: fillColor}.Add(gtx.Ops)
	return child(gtx)
}

func layoutToolbarSeparator(gtx layout.Context, heightDp int) layout.Dimensions {
	return layout.Inset{
		Left:  unit.Dp(8),
		Right: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return paintBox(gtx, appColor{R: 42, G: 42, B: 42, A: 255}, appColor{}, 0, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(1)), gtx.Dp(unit.Dp(heightDp)))
			gtx.Constraints.Max = gtx.Constraints.Min
			return layout.Dimensions{Size: gtx.Constraints.Min}
		})
	})
}
