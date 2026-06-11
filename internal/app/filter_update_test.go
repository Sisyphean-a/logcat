package app

import (
	"context"
	"testing"

	"github.com/xiakn/logcat/internal/adb"
)

func TestControllerUpdateSavedFilterDefinitionRenamesAndApplies(t *testing.T) {
	controller := NewController(
		stubDeviceService{
			install: adb.Install{Path: "adb", Version: "1.0.41"},
			devices: []adb.DeviceInfo{
				{ID: "device-1", Model: "Pixel_7", Status: "device"},
			},
		},
		stubSessionStarter{},
	)

	if err := controller.Load(context.Background()); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := controller.SaveFilterDefinition("H5 Errors", "", `message~:"[H5]"`); err != nil {
		t.Fatalf("SaveFilterDefinition returned error: %v", err)
	}
	if err := controller.UpdateSavedFilterDefinition(
		context.Background(),
		"h5-errors",
		"H5 Submit Errors",
		"",
		`message~:"submit"`,
	); err != nil {
		t.Fatalf("UpdateSavedFilterDefinition returned error: %v", err)
	}

	model := controller.Model()
	if len(model.Filter.Saved) != 1 {
		t.Fatalf("expected 1 saved filter, got %d", len(model.Filter.Saved))
	}

	saved := model.Filter.Saved[0]
	if saved.ID != "h5-submit-errors" {
		t.Fatalf("expected renamed id, got %q", saved.ID)
	}
	if saved.Name != "H5 Submit Errors" {
		t.Fatalf("expected renamed filter, got %q", saved.Name)
	}
	if saved.Query != `message~:"submit"` {
		t.Fatalf("expected updated query, got %q", saved.Query)
	}
	if model.Filter.ActiveFilterID != saved.ID {
		t.Fatalf("expected active filter %q, got %q", saved.ID, model.Filter.ActiveFilterID)
	}
	if model.Filter.Applied != `message~:"submit"` {
		t.Fatalf("expected applied query updated, got %q", model.Filter.Applied)
	}
	if model.Filter.Draft != `message~:"submit"` {
		t.Fatalf("expected draft query updated, got %q", model.Filter.Draft)
	}
}

func TestControllerUpdateSavedFilterDefinitionKeepsDefaultOnRename(t *testing.T) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})

	controller.ReplaceSavedFilters([]SavedFilter{
		{
			ID:    "h5-errors",
			Name:  "H5 Errors",
			Query: `message~:"[H5]"`,
		},
	}, "h5-errors")

	if err := controller.UpdateSavedFilterDefinition(
		context.Background(),
		"h5-errors",
		"H5 Submit Errors",
		"",
		`message~:"submit"`,
	); err != nil {
		t.Fatalf("UpdateSavedFilterDefinition returned error: %v", err)
	}

	model := controller.Model()
	if model.Filter.DefaultFilterID != "h5-submit-errors" {
		t.Fatalf("expected renamed default filter id, got %q", model.Filter.DefaultFilterID)
	}
}

func TestControllerReplaceSavedFilterDefinitionsReordersDefaultAndAppliesSelected(t *testing.T) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})

	controller.ReplaceSavedFilters([]SavedFilter{
		{
			ID:    "bridge-h5",
			Name:  "Bridge H5",
			Query: `message~:"bridge"`,
		},
		{
			ID:    "submit-errors",
			Name:  "Submit Errors",
			Query: `message~:"submit"`,
		},
	}, "bridge-h5")

	err := controller.ReplaceSavedFilterDefinitions(
		context.Background(),
		[]SavedFilterDraft{
			{
				ExistingID: "submit-errors",
				Name:       "Submit Errors",
				Query:      `message~:"submit"`,
			},
			{
				ExistingID: "bridge-h5",
				Name:       "Bridge Ready",
				Query:      `message~:"bridge"`,
			},
		},
		"submit-errors",
		"bridge-h5",
	)
	if err != nil {
		t.Fatalf("ReplaceSavedFilterDefinitions returned error: %v", err)
	}

	model := controller.Model()
	if len(model.Filter.Saved) != 2 {
		t.Fatalf("expected 2 saved filters, got %d", len(model.Filter.Saved))
	}
	if model.Filter.Saved[0].ID != "submit-errors" || model.Filter.Saved[1].ID != "bridge-ready" {
		t.Fatalf("unexpected saved filter order: %#v", model.Filter.Saved)
	}
	if model.Filter.DefaultFilterID != "submit-errors" {
		t.Fatalf("expected default filter submit-errors, got %q", model.Filter.DefaultFilterID)
	}
	if model.Filter.ActiveFilterID != "bridge-ready" {
		t.Fatalf("expected active filter bridge-ready, got %q", model.Filter.ActiveFilterID)
	}
	if model.Filter.Applied != `message~:"bridge"` {
		t.Fatalf("expected applied query updated, got %q", model.Filter.Applied)
	}
}

func TestControllerRestoreSavedFilterSetsPackageAndQueryWithoutDevice(t *testing.T) {
	controller := NewController(stubDeviceService{}, stubSessionStarter{})

	controller.ReplaceSavedFilters([]SavedFilter{
		{
			ID:          "host-bridge",
			Name:        "Host Bridge",
			PackageName: "com.demo.host",
			Query:       `message~:"bridge"`,
		},
	}, "host-bridge")

	if err := controller.RestoreSavedFilter("host-bridge"); err != nil {
		t.Fatalf("RestoreSavedFilter returned error: %v", err)
	}

	model := controller.Model()
	if model.Filter.ActiveFilterID != "host-bridge" {
		t.Fatalf("expected active filter host-bridge, got %q", model.Filter.ActiveFilterID)
	}
	if model.SelectedPackage != "com.demo.host" {
		t.Fatalf("expected selected package restored, got %q", model.SelectedPackage)
	}
	if model.Filter.Draft != `message~:"bridge"` || model.Filter.Applied != `message~:"bridge"` {
		t.Fatalf("expected restored query, got draft=%q applied=%q", model.Filter.Draft, model.Filter.Applied)
	}
}
