package adb

import (
	"reflect"
	"testing"
)

func TestBuildLogcatArgsUsesThreadtimeStream(t *testing.T) {
	args := buildLogcatArgs("emulator-5554")
	expected := []string{
		"-s",
		"emulator-5554",
		"logcat",
		"-v",
		"threadtime",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args: %#v", args)
	}
}
