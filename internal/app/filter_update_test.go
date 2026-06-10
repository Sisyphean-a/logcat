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
