package adb

import (
	"reflect"
	"testing"
)

func TestBuildLogcatArgsUsesChromiumPreset(t *testing.T) {
	args := buildLogcatArgs("emulator-5554")
	expected := []string{
		"-s",
		"emulator-5554",
		"logcat",
		"-v",
		"threadtime",
		"-s",
		"chromium:I",
		"*:S",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args: %#v", args)
	}
}
