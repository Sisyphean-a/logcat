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
	if len(model.VisibleLogs) != 0 {
		t.Fatalf("expected empty visible logs, got %d", len(model.VisibleLogs))
	}
	if !model.Pause.Active {
		t.Fatal("expected initial state paused")
	}
}
