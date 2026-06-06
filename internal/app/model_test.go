package app

import "testing"

func TestNewModelStartsIdle(t *testing.T) {
	model := NewModel()

	if model.Status != "idle" {
		t.Fatalf("expected idle status, got %q", model.Status)
	}
	if len(model.Devices) != 0 {
		t.Fatalf("expected empty devices, got %d", len(model.Devices))
	}
	if len(model.Logs) != 0 {
		t.Fatalf("expected empty logs, got %d", len(model.Logs))
	}
}
