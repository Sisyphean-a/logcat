package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutDevices(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutDeviceList(gtx, model)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Flexed(2, func(gtx layout.Context) layout.Dimensions {
			return s.layoutPackageProcessPanel(gtx, model)
		}),
	)
}

func (s *Shell) layoutDeviceList(gtx layout.Context, model appstate.Model) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.H6(s.theme, "Devices").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.deviceList.Layout(gtx, len(model.Devices), func(gtx layout.Context, index int) layout.Dimensions {
				device := model.Devices[index]
				label := fmt.Sprintf("%s  %s  %s", device.Model, device.Status, device.ID)
				button := material.Button(s.theme, &s.deviceButtons[index], label)
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, button.Layout)
			})
		}),
	)
}
