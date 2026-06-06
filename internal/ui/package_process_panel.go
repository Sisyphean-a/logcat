package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/xiakn/logcat/internal/adb"
	appstate "github.com/xiakn/logcat/internal/app"
)

func (s *Shell) layoutPackageProcessPanel(
	gtx layout.Context,
	model appstate.Model,
) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.H6(s.theme, "Binding").Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutScopeButtons(gtx, model.PackageScope)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutForegroundButton(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, material.Body2(s.theme, bindingSummary(model)).Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(s.theme, "Packages").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutPackageList(gtx, model)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(s.theme, "Processes").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return s.layoutProcessList(gtx, model)
		}),
	)
}

func (s *Shell) layoutScopeButtons(
	gtx layout.Context,
	scope adb.PackageScope,
) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.scopeUserButton, scopeLabel("User", scope == adb.PackageScopeUser)).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(s.theme, &s.scopeSystemButton, scopeLabel("System", scope == adb.PackageScopeSystem)).Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(s.theme, &s.scopeAllButton, scopeLabel("All", scope == adb.PackageScopeAll)).Layout(gtx)
			}),
		)
	})
}

func (s *Shell) layoutForegroundButton(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.Button(s.theme, &s.foregroundButton, "Front App").Layout(gtx)
	})
}

func (s *Shell) layoutPackageList(
	gtx layout.Context,
	model appstate.Model,
) layout.Dimensions {
	if len(model.Packages) == 0 {
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, material.Body2(s.theme, "No packages").Layout)
	}

	return s.packageList.Layout(gtx, len(model.Packages), func(gtx layout.Context, index int) layout.Dimensions {
		pkg := model.Packages[index]
		button := material.Button(
			s.theme,
			&s.packageButtons[index],
			selectLabel(pkg.Name, pkg.Name == model.SelectedPackage),
		)
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, button.Layout)
	})
}

func (s *Shell) layoutProcessList(
	gtx layout.Context,
	model appstate.Model,
) layout.Dimensions {
	if len(model.Processes) == 0 {
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, material.Body2(s.theme, "No processes").Layout)
	}

	return s.processList.Layout(gtx, len(model.Processes), func(gtx layout.Context, index int) layout.Dimensions {
		process := model.Processes[index]
		label := fmt.Sprintf("%s (%d)", process.Name, process.PID)
		button := material.Button(
			s.theme,
			&s.processButtons[index],
			selectLabel(label, process.Name == model.SelectedProcess),
		)
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, button.Layout)
	})
}

func bindingSummary(model appstate.Model) string {
	if model.SelectedPackage == "" {
		return "Device-level H5 view"
	}
	if len(model.BoundPIDs) == 0 {
		return fmt.Sprintf("Target: %s (stopped)", model.SelectedPackage)
	}
	if model.SelectedProcess != "" {
		return fmt.Sprintf("Target: %s %v", model.SelectedProcess, model.BoundPIDs)
	}

	return fmt.Sprintf("Target: %s %v", model.SelectedPackage, model.BoundPIDs)
}

func scopeLabel(label string, selected bool) string {
	return selectLabel(label, selected)
}

func selectLabel(label string, selected bool) string {
	if selected {
		return "[x] " + label
	}

	return "[ ] " + label
}
