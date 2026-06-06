package ui

import (
	"image/color"

	"gioui.org/widget/material"
)

func NewTheme() *material.Theme {
	theme := material.NewTheme()
	theme.Palette.Bg = color.NRGBA{R: 18, G: 18, B: 18, A: 255}
	theme.Palette.Fg = color.NRGBA{R: 236, G: 239, B: 244, A: 255}
	theme.Palette.ContrastBg = color.NRGBA{R: 52, G: 84, B: 136, A: 255}
	theme.Palette.ContrastFg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	return theme
}
