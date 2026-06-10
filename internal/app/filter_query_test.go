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
			compiled, err := compileFilterQuery(tc.query)
			if err != nil {
				t.Fatalf("compile returned error: %v", err)
			}

			got := compiled.matches(entry, tc.packageName)
			if got != tc.want {
				t.Fatalf("expected %v, got %v for query %q", tc.want, got, tc.query)
			}
		})
	}
}

func TestCompileFilterQueryPreservesEmptyTagAndLevelBehavior(t *testing.T) {
	entry := logcat.LogEntry{Message: "plain"}
	compiledTag, err := compileFilterQuery("tag:anything")
	if err != nil {
		t.Fatalf("compile tag query returned error: %v", err)
	}
	if !compiledTag.matches(entry, "") {
		t.Fatal("expected empty tag to match tag filter")
	}

	compiledLevel, err := compileFilterQuery("level:E")
	if err != nil {
		t.Fatalf("compile level query returned error: %v", err)
	}
	if !compiledLevel.matches(entry, "") {
		t.Fatal("expected empty level to match level filter")
	}
}

func TestCompileFilterQuerySupportsOrContainsAndGrouping(t *testing.T) {
	entry := logcat.LogEntry{
		Level:   "W",
		Tag:     "bridge.dispatch",
		Message: "bridge opened h5 channel",
	}

	compiled, err := compileFilterQuery(`(level:E || level:W) && tag~:"bridge" && message~:"h5"`)
	if err != nil {
		t.Fatalf("compile returned error: %v", err)
	}
	if !compiled.matches(entry, "com.demo.app") {
		t.Fatal("expected grouped query to match")
	}
}

func TestCompileFilterQuerySupportsNegatedContains(t *testing.T) {
	entry := logcat.LogEntry{
		Level:   "I",
		Tag:     "bridge.dispatch",
		Message: "bridge opened native channel",
	}

	compiled, err := compileFilterQuery(`tag~:"bridge" && -message~:"h5"`)
	if err != nil {
		t.Fatalf("compile returned error: %v", err)
	}
	if !compiled.matches(entry, "com.demo.app") {
		t.Fatal("expected negated contains query to match")
	}
}

func TestCompileFilterQuerySupportsEqualsAliases(t *testing.T) {
	entry := logcat.LogEntry{
		Level:   "I",
		Tag:     "jsbridge",
		Message: "native channel ready",
	}

	compiled, err := compileFilterQuery(`tag=jsbridge || message="native channel"`)
	if err != nil {
		t.Fatalf("compile returned error: %v", err)
	}
	if !compiled.matches(entry, "com.demo.app") {
		t.Fatal("expected equals aliases to match")
	}
}

func TestCompileFilterQueryRejectsInvalidSyntax(t *testing.T) {
	cases := []string{
		`(`,
		`tag:bridge &&`,
		`|| level:I`,
		`message:"h5`,
	}

	for _, query := range cases {
		t.Run(query, func(t *testing.T) {
			if _, err := compileFilterQuery(query); err == nil {
				t.Fatalf("expected compile error for %q", query)
			}
		})
	}
}
