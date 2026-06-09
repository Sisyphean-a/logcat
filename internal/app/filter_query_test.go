package app

import (
	"testing"

	"github.com/xiakn/logcat/internal/logcat"
)

func TestCompileFilterQueryMatchesEquivalentToLegacySemantics(t *testing.T) {
	entry := logcat.LogEntry{
		Level:   "I",
		Tag:     "chromium",
		Message: `[H5] token payload`,
	}

	cases := []struct {
		name        string
		packageName string
		query       string
		want        bool
	}{
		{name: "empty query", packageName: "com.demo.app", query: "", want: true},
		{name: "tag match", packageName: "com.demo.app", query: "tag:chromium", want: true},
		{name: "tag mismatch", packageName: "com.demo.app", query: "tag:ActivityManager", want: false},
		{name: "message contains", packageName: "com.demo.app", query: `message:"token"`, want: true},
		{name: "text contains", packageName: "com.demo.app", query: "payload", want: true},
		{name: "package match", packageName: "com.demo.app", query: "package:com.demo.app", want: true},
		{name: "package mismatch", packageName: "com.demo.app", query: "package:com.other.app", want: false},
		{name: "level match", packageName: "com.demo.app", query: "level:i", want: true},
		{name: "negated term", packageName: "com.demo.app", query: "-crash", want: true},
		{name: "and terms", packageName: "com.demo.app", query: `tag:chromium & message:"token"`, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := compileFilterQuery(tc.query).matches(entry, tc.packageName)
			if got != tc.want {
				t.Fatalf("expected %v, got %v for query %q", tc.want, got, tc.query)
			}
		})
	}
}

func TestCompileFilterQueryPreservesEmptyTagAndLevelBehavior(t *testing.T) {
	entry := logcat.LogEntry{Message: "plain"}
	if !compileFilterQuery("tag:anything").matches(entry, "") {
		t.Fatal("expected empty tag to match tag filter")
	}
	if !compileFilterQuery("level:E").matches(entry, "") {
		t.Fatal("expected empty level to match level filter")
	}
}
