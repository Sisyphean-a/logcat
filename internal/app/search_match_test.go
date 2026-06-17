package app

import (
	"testing"

	"github.com/xiakn/logcat/internal/logcat"
)

func TestCompileSearchQueryKeepsLiteralSearchWithoutOperators(t *testing.T) {
	compiled := compileSearchQuery("vite-hmr")
	if compiled.literal != "vite-hmr" {
		t.Fatalf("expected literal search, got %#v", compiled)
	}
	if !compiled.matches("chromium\nvite-hmr connected") {
		t.Fatal("expected literal search to match substring")
	}
}

func TestCompileSearchQueryMatchesAndTerms(t *testing.T) {
	compiled := compileSearchQuery("bridge && token")
	if !compiled.matches("jsbridge\nauth token ready") {
		t.Fatal("expected && search to require both keywords")
	}
	if compiled.matches("jsbridge\nplain ready") {
		t.Fatal("expected && search to reject missing keyword")
	}
}

func TestCompileSearchQueryMatchesOrTerms(t *testing.T) {
	compiled := compileSearchQuery("bridge || token")
	if !compiled.matches("chromium\nauth token ready") {
		t.Fatal("expected || search to match token branch")
	}
	if !compiled.matches("jsbridge\nplain ready") {
		t.Fatal("expected || search to match bridge branch")
	}
	if compiled.matches("network\nplain ready") {
		t.Fatal("expected || search to reject unmatched row")
	}
}

func TestCompileSearchQueryMatchesNegatedTerms(t *testing.T) {
	compiled := compileSearchQuery("bridge && -token")
	if !compiled.matches("jsbridge\nplain ready") {
		t.Fatal("expected negated term to allow row without token")
	}
	if compiled.matches("jsbridge\nauth token ready") {
		t.Fatal("expected negated term to reject row with token")
	}
}

func TestCompileSearchQueryCollectsPositiveHighlightTerms(t *testing.T) {
	compiled := compileSearchQuery("bridge && -token || ready")
	if len(compiled.highlightTerms) != 2 {
		t.Fatalf("expected 2 highlight terms, got %#v", compiled.highlightTerms)
	}
	if compiled.highlightTerms[0] != "bridge" || compiled.highlightTerms[1] != "ready" {
		t.Fatalf("unexpected highlight terms %#v", compiled.highlightTerms)
	}
}

func TestSearchLowerTextOnlyUsesTagAndMessage(t *testing.T) {
	entry := logcat.LogEntry{
		TimeText: "06-04 16:42:18.479",
		Level:    "E",
		Tag:      "JSBridge",
		Message:  "Token expired",
	}

	got := searchLowerText(entry)
	if got != "jsbridge\ntoken expired" {
		t.Fatalf("unexpected search text %q", got)
	}
}
