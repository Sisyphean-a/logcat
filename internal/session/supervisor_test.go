package session

import (
	"context"
	"testing"
)

type stubSource struct {
	lines []string
	err   error
}

func (s stubSource) Stream(_ context.Context, _ Config) (<-chan string, <-chan error) {
	lines := make(chan string, len(s.lines))
	errs := make(chan error, 1)

	for _, line := range s.lines {
		lines <- line
	}
	close(lines)

	if s.err != nil {
		errs <- s.err
	}
	close(errs)

	return lines, errs
}

func TestSupervisorStreamsMatchingEntries(t *testing.T) {
	source := stubSource{
		lines: []string{
			`06-04 16:42:18.479 10665 10665 I chromium: [H5] ok`,
			`06-04 16:42:18.480 10665 10665 I ActivityManager: ignored`,
		},
	}

	supervisor := NewSupervisor(source)
	handle, err := supervisor.Start(context.Background(), Config{DeviceID: "d1"})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	var messages []string
	for event := range handle.Events() {
		if event.Entry != nil {
			messages = append(messages, event.Entry.Message)
		}
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 entry events, got %d", len(messages))
	}
	if messages[0] != "[H5] ok" {
		t.Fatalf("expected first message [H5] ok, got %q", messages[0])
	}
	if messages[1] != "ignored" {
		t.Fatalf("expected second message ignored, got %q", messages[1])
	}
}

func TestSupervisorReportsParseErrorWithRawLineAndContinues(t *testing.T) {
	source := stubSource{
		lines: []string{
			`broken line`,
			`06-04 16:42:18.479 10665 10665 I chromium: [H5] ok`,
		},
	}

	supervisor := NewSupervisor(source)
	handle, err := supervisor.Start(context.Background(), Config{DeviceID: "d1"})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	first, ok := <-handle.Events()
	if !ok {
		t.Fatal("expected parse error event")
	}
	if first.Problem == nil {
		t.Fatal("expected parse error problem")
	}
	if got := first.Problem.Error(); got != `invalid threadtime line: broken line` {
		t.Fatalf("expected raw line in parse error, got %q", got)
	}

	second, ok := <-handle.Events()
	if !ok {
		t.Fatal("expected follow-up entry event")
	}
	if second.Entry == nil {
		t.Fatal("expected log entry after parse error")
	}
	if second.Entry.Message != "[H5] ok" {
		t.Fatalf("expected [H5] ok, got %q", second.Entry.Message)
	}
}

func TestSupervisorFiltersByAllowedPIDs(t *testing.T) {
	source := stubSource{
		lines: []string{
			`06-04 16:42:18.479 111 111 I chromium: [H5] first`,
			`06-04 16:42:18.480 222 222 I chromium: [H5] second`,
		},
	}

	supervisor := NewSupervisor(source)
	handle, err := supervisor.Start(context.Background(), Config{
		DeviceID:    "d1",
		AllowedPIDs: []int{222},
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	var entries []int
	for event := range handle.Events() {
		if event.Entry != nil {
			entries = append(entries, event.Entry.PID)
		}
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 filtered entry, got %d", len(entries))
	}
	if entries[0] != 222 {
		t.Fatalf("expected pid 222, got %d", entries[0])
	}
}
