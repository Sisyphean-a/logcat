package logcat

import (
	"strings"
	"testing"
)

func TestParseThreadtimeLineParsesChromiumConsole(t *testing.T) {
	line := `06-04 16:42:18.479 10665 10665 I chromium: [INFO:CONSOLE(618)] "[H5] connected", source: http://127.0.0.1/app.js (618)`

	entry, err := ParseThreadtimeLine("device-1", line)
	if err != nil {
		t.Fatalf("ParseThreadtimeLine returned error: %v", err)
	}
	if entry.DeviceID != "device-1" {
		t.Fatalf("expected device-1, got %q", entry.DeviceID)
	}
	if entry.Tag != "chromium" {
		t.Fatalf("expected chromium tag, got %q", entry.Tag)
	}
	if entry.Level != "I" {
		t.Fatalf("expected level I, got %q", entry.Level)
	}
	if entry.Message != "[H5] connected" {
		t.Fatalf("expected normalized message, got %q", entry.Message)
	}
	if entry.Source != "app.js" {
		t.Fatalf("expected source app.js, got %q", entry.Source)
	}
	if !MatchesH5Preset(entry) {
		t.Fatalf("expected entry to match H5 preset")
	}
}

func TestParseThreadtimeLineRejectsInvalidInput(t *testing.T) {
	_, err := ParseThreadtimeLine("device-1", "broken line")
	if err == nil {
		t.Fatal("expected parse error for broken line")
	}
}

func TestParseThreadtimeLineAcceptsTagsWithMultipleColons(t *testing.T) {
	line := `06-09 16:30:28.224 25841 26156 D dist:vtcamera:PhoneTemperatureNotifier: tempV=31153, highTemperature_state=false`

	entry, err := ParseThreadtimeLine("device-1", line)
	if err != nil {
		t.Fatalf("ParseThreadtimeLine returned error: %v", err)
	}
	if entry.Level != "D" {
		t.Fatalf("expected level D, got %q", entry.Level)
	}
	if entry.Tag != "dist:vtcamera:PhoneTemperatureNotifier" {
		t.Fatalf("unexpected tag: %q", entry.Tag)
	}
	if entry.Message != "tempV=31153, highTemperature_state=false" {
		t.Fatalf("unexpected message: %q", entry.Message)
	}
}

func TestParseThreadtimeLinePreservesLongRawLine(t *testing.T) {
	payload := strings.Repeat("x", 5000)
	line := `06-10 20:41:45.478 1234 1234 I chromium: [INFO:CONSOLE(1)] "` +
		payload +
		`", source: http://127.0.0.1/app.js (1)`

	entry, err := ParseThreadtimeLine("device-1", line)
	if err != nil {
		t.Fatalf("ParseThreadtimeLine returned error: %v", err)
	}
	if entry.Message != payload {
		t.Fatalf("expected payload length %d, got %d", len(payload), len(entry.Message))
	}
	if entry.Raw != line {
		t.Fatal("expected raw line to round-trip unchanged")
	}
}
