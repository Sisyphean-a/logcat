package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type appColor = color.NRGBA

type appPalette struct {
	window        appColor
	chromeBar     appColor
	panelBar      appColor
	panel         appColor
	panelBorder   appColor
	rowBorder     appColor
	textPrimary   appColor
	textSecondary appColor
	textMuted     appColor
	accent        appColor
	accentMuted   appColor
	accentOutline appColor
	success       appColor
	warn          appColor
	error         appColor
	infoBg        appColor
	infoFg        appColor
	warnBg        appColor
	warnFg        appColor
	errorBg       appColor
	errorFg       appColor
	selectedLine  appColor
}

func NewTheme() *material.Theme {
	theme := material.NewTheme()
	theme.Palette.Bg = color.NRGBA{R: 22, G: 22, B: 22, A: 255}
	theme.Palette.Fg = color.NRGBA{R: 212, G: 212, B: 212, A: 255}
	theme.Palette.ContrastBg = color.NRGBA{R: 43, G: 127, B: 255, A: 255}
	theme.Palette.ContrastFg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	return theme
}

func defaultPalette() appPalette {
	return appPalette{
		window:        appColor{R: 22, G: 22, B: 22, A: 255},
		chromeBar:     appColor{R: 22, G: 22, B: 22, A: 255},
		panelBar:      appColor{R: 26, G: 26, B: 26, A: 255},
		panel:         appColor{R: 22, G: 22, B: 22, A: 255},
		panelBorder:   appColor{R: 42, G: 42, B: 42, A: 255},
		rowBorder:     appColor{R: 30, G: 30, B: 30, A: 255},
		textPrimary:   appColor{R: 224, G: 224, B: 224, A: 255},
		textSecondary: appColor{R: 192, G: 192, B: 192, A: 255},
		textMuted:     appColor{R: 136, G: 136, B: 136, A: 255},
		accent:        appColor{R: 0, G: 211, B: 243, A: 255},
		accentMuted:   appColor{R: 0, G: 184, B: 219, A: 48},
		accentOutline: appColor{R: 0, G: 184, B: 219, A: 96},
		success:       appColor{R: 5, G: 223, B: 114, A: 255},
		warn:          appColor{R: 255, G: 201, B: 24, A: 255},
		error:         appColor{R: 251, G: 44, B: 54, A: 255},
		infoBg:        appColor{R: 22, G: 36, B: 86, A: 255},
		infoFg:        appColor{R: 81, G: 162, B: 255, A: 255},
		warnBg:        appColor{R: 70, G: 57, B: 8, A: 255},
		warnFg:        appColor{R: 255, G: 201, B: 24, A: 255},
		errorBg:       appColor{R: 70, G: 8, B: 9, A: 255},
		errorFg:       appColor{R: 255, G: 100, B: 103, A: 255},
		selectedLine:  appColor{R: 43, G: 127, B: 255, A: 255},
	}
}

func fillRect(gtx layout.Context, fill appColor) layout.Dimensions {
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, fill)
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func borderedBox(gtx layout.Context, fill appColor, border appColor, radius unit.Dp, inset layout.Inset, child layout.Widget) layout.Dimensions {
	return paintBox(gtx, fill, border, radius, func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, child)
	})
}

func paintBox(gtx layout.Context, fill appColor, border appColor, radius unit.Dp, child layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := child(gtx)
	call := macro.Stop()

	rr := gtx.Dp(radius)
	rect := image.Rectangle{Max: dims.Size}
	defer clip.UniformRRect(rect, rr).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, fill)
	call.Add(gtx.Ops)
	if border.A > 0 {
		paint.FillShape(gtx.Ops, border, clip.Stroke{
			Path:  clip.UniformRRect(rect, rr).Path(gtx.Ops),
			Width: float32(gtx.Dp(unit.Dp(1))),
		}.Op())
	}
	return dims
}

func layoutSpacerWidth(gtx layout.Context, dp int) layout.Dimensions {
	return layout.Spacer{Width: unit.Dp(dp)}.Layout(gtx)
}

func layoutSpacerHeight(gtx layout.Context, dp int) layout.Dimensions {
	return layout.Spacer{Height: unit.Dp(dp)}.Layout(gtx)
}

func iconButton(gtx layout.Context, theme *material.Theme, clickable *widget.Clickable, label string, color appColor) layout.Dimensions {
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return iconGlyph(gtx, label, color)
	})
}

func iconGlyph(gtx layout.Context, label string, color appColor) layout.Dimensions {
	return paintBox(gtx, appColor{}, appColor{}, unit.Dp(4), func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(28)), gtx.Dp(unit.Dp(28)))
		gtx.Constraints.Max = gtx.Constraints.Min
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return drawToolbarIcon(gtx, label, color)
		})
	})
}

func drawToolbarIcon(gtx layout.Context, kind string, stroke appColor) layout.Dimensions {
	size := image.Pt(gtx.Dp(unit.Dp(14)), gtx.Dp(unit.Dp(14)))
	gtx.Constraints.Min = size
	gtx.Constraints.Max = size
	switch kind {
	case "pause":
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.08333, 1.75), f32.Pt(4.08333, 12.25))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.91667, 1.75), f32.Pt(9.91667, 12.25))
	case "play":
		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(scalePoint(gtx, 4.08333, 3.20833))
		path.LineTo(scalePoint(gtx, 10.5, 7))
		path.LineTo(scalePoint(gtx, 4.08333, 10.79167))
		path.Close()
		paint.FillShape(gtx.Ops, stroke, clip.Outline{Path: path.End()}.Op())
	case "trash":
		drawLine(gtx, stroke, 1.16667, f32.Pt(1.75, 3.5), f32.Pt(12.25, 3.5))
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.66667, 3.5), f32.Pt(4.66667, 2.33333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.33333, 3.5), f32.Pt(9.33333, 2.33333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(5.83333, 6.41667), f32.Pt(5.83333, 9.91667))
		drawLine(gtx, stroke, 1.16667, f32.Pt(8.16667, 6.41667), f32.Pt(8.16667, 9.91667))
		drawRectStroke(gtx, stroke, 1.16667, 2.91667, 3.5, 11.08333, 12.83333)
		drawRectStroke(gtx, stroke, 1.16667, 4.66667, 1.16667, 9.33333, 3.5)
	case "download":
		drawLine(gtx, stroke, 1.16667, f32.Pt(7, 1.75), f32.Pt(7, 8.75))
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.08333, 5.83333), f32.Pt(7, 8.75))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.91667, 5.83333), f32.Pt(7, 8.75))
		drawLine(gtx, stroke, 1.16667, f32.Pt(1.75, 8.75), f32.Pt(1.75, 11.08333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(12.25, 8.75), f32.Pt(12.25, 11.08333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(1.75, 12.25), f32.Pt(12.25, 12.25))
	case "settings":
		drawCircleStroke(gtx, stroke, 1.16667, 7, 7, 1.75)
		drawLine(gtx, stroke, 1.16667, f32.Pt(7, 1.16667), f32.Pt(7, 2.33333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(7, 11.66667), f32.Pt(7, 12.83333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(1.89583, 4.19417), f32.Pt(2.91667, 4.78333))
		drawLine(gtx, stroke, 1.16667, f32.Pt(11.08333, 9.21667), f32.Pt(12.10417, 9.80583))
		drawLine(gtx, stroke, 1.16667, f32.Pt(1.89583, 9.80583), f32.Pt(2.91667, 9.21667))
		drawLine(gtx, stroke, 1.16667, f32.Pt(11.08333, 4.78333), f32.Pt(12.10417, 4.19417))
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.19417, 1.89583), f32.Pt(4.78333, 2.91667))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.21667, 11.08333), f32.Pt(9.80583, 12.10417))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.21667, 2.91667), f32.Pt(9.80583, 1.89583))
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.19417, 12.10417), f32.Pt(4.78333, 11.08333))
	case "chevron-down":
		drawLine(gtx, stroke, 1.16667, f32.Pt(3.5, 5.25), f32.Pt(6, 7.75))
		drawLine(gtx, stroke, 1.16667, f32.Pt(6, 7.75), f32.Pt(8.5, 5.25))
	case "chevron-left":
		drawLine(gtx, stroke, 1.16667, f32.Pt(8.5, 3.5), f32.Pt(5.5, 7))
		drawLine(gtx, stroke, 1.16667, f32.Pt(5.5, 7), f32.Pt(8.5, 10.5))
	case "chevron-right":
		drawLine(gtx, stroke, 1.16667, f32.Pt(5.5, 3.5), f32.Pt(8.5, 7))
		drawLine(gtx, stroke, 1.16667, f32.Pt(8.5, 7), f32.Pt(5.5, 10.5))
	case "up-circle":
		drawCircleStroke(gtx, stroke, 1.16667, 7, 7, 6)
		drawLine(gtx, stroke, 1.16667, f32.Pt(7, 10), f32.Pt(7, 4))
		drawLine(gtx, stroke, 1.16667, f32.Pt(4.75, 6.25), f32.Pt(7, 4))
		drawLine(gtx, stroke, 1.16667, f32.Pt(9.25, 6.25), f32.Pt(7, 4))
	case "device":
		drawRectStroke(gtx, stroke, 1.16667, 1.75, 2.33333, 12.25, 9.33333)
		drawLine(gtx, stroke, 1.16667, f32.Pt(5.25, 11.66667), f32.Pt(8.75, 11.66667))
		drawLine(gtx, stroke, 1.16667, f32.Pt(7, 9.33333), f32.Pt(7, 11.66667))
	case "search":
		drawCircleStroke(gtx, stroke, 1.16667, 6, 6, 3.5)
		drawLine(gtx, stroke, 1.16667, f32.Pt(8.75, 8.75), f32.Pt(11.25, 11.25))
	case "save":
		drawRectStroke(gtx, stroke, 1.0, 1.5, 1.5, 10.5, 10.5)
		drawRectStroke(gtx, stroke, 1.0, 3.5, 6.5, 8.5, 10.5)
		drawLine(gtx, stroke, 1.0, f32.Pt(3.5, 1.5), f32.Pt(3.5, 4))
		drawLine(gtx, stroke, 1.0, f32.Pt(3.5, 4), f32.Pt(7.5, 4))
	default:
		return layout.Dimensions{Size: size}
	}
	return layout.Dimensions{Size: size}
}

func scalePoint(gtx layout.Context, x float32, y float32) f32.Point {
	return f32.Pt(x*float32(gtx.Dp(unit.Dp(1))), y*float32(gtx.Dp(unit.Dp(1))))
}

func drawLine(gtx layout.Context, stroke appColor, width float32, from f32.Point, to f32.Point) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(scalePoint(gtx, from.X, from.Y))
	path.LineTo(scalePoint(gtx, to.X, to.Y))
	paint.FillShape(gtx.Ops, stroke, clip.Stroke{
		Path:  path.End(),
		Width: width * float32(gtx.Dp(unit.Dp(1))),
	}.Op())
}

func drawRectStroke(gtx layout.Context, stroke appColor, width float32, x0 float32, y0 float32, x1 float32, y1 float32) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(scalePoint(gtx, x0, y0))
	path.LineTo(scalePoint(gtx, x1, y0))
	path.LineTo(scalePoint(gtx, x1, y1))
	path.LineTo(scalePoint(gtx, x0, y1))
	path.Close()
	paint.FillShape(gtx.Ops, stroke, clip.Stroke{
		Path:  path.End(),
		Width: width * float32(gtx.Dp(unit.Dp(1))),
	}.Op())
}

func drawCircleStroke(gtx layout.Context, stroke appColor, width float32, cx float32, cy float32, r float32) {
	diameter := r * 2 * float32(gtx.Dp(unit.Dp(1)))
	size := image.Pt(int(diameter), int(diameter))
	offset := op.Offset(image.Pt(
		int((cx-r)*float32(gtx.Dp(unit.Dp(1)))),
		int((cy-r)*float32(gtx.Dp(unit.Dp(1)))),
	)).Push(gtx.Ops)
	defer offset.Pop()
	paint.FillShape(gtx.Ops, stroke, clip.Stroke{
		Path:  clip.Ellipse(image.Rectangle{Max: size}).Path(gtx.Ops),
		Width: width * float32(gtx.Dp(unit.Dp(1))),
	}.Op())
}

func tokenBox(gtx layout.Context, fill appColor, border appColor, text string) layout.Dimensions {
	return borderedBox(
		gtx,
		fill,
		border,
		unit.Dp(4),
		layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(4),
			Right:  unit.Dp(4),
		},
		func(gtx layout.Context) layout.Dimensions {
			labelWidget := material.Label(NewTheme(), unit.Sp(9), text)
			labelWidget.Color = appColor{R: 0, G: 211, B: 243, A: 255}
			return labelWidget.Layout(gtx)
		},
	)
}
