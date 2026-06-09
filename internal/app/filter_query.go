package app

import (
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

type filterTermKind uint8

const (
	filterTermText filterTermKind = iota
	filterTermTag
	filterTermMessage
	filterTermPackage
	filterTermLevel
)

type compiledFilterQuery struct {
	terms             []compiledFilterTerm
	needsLowerMessage bool
}

type compiledFilterTerm struct {
	kind    filterTermKind
	value   string
	negated bool
}

func compileFilterQuery(query string) compiledFilterQuery {
	parts := splitAndTerms(strings.TrimSpace(query))
	compiled := compiledFilterQuery{terms: make([]compiledFilterTerm, 0, len(parts))}
	for _, part := range parts {
		term := compileFilterTerm(part)
		if term.kind == filterTermMessage || term.kind == filterTermText {
			compiled.needsLowerMessage = true
		}
		compiled.terms = append(compiled.terms, term)
	}
	return compiled
}

func (q compiledFilterQuery) matchAll() bool {
	return len(q.terms) == 0
}

func (q compiledFilterQuery) matches(entry logcat.LogEntry, packageName string) bool {
	if q.matchAll() {
		return true
	}

	lowerMessage := ""
	if q.needsLowerMessage {
		lowerMessage = strings.ToLower(entry.Message)
	}

	for _, term := range q.terms {
		if !term.matches(entry, packageName, lowerMessage) {
			return false
		}
	}
	return true
}

func (t compiledFilterTerm) matches(
	entry logcat.LogEntry,
	packageName string,
	lowerMessage string,
) bool {
	matched := true
	switch t.kind {
	case filterTermTag:
		matched = entry.Tag == "" || strings.EqualFold(entry.Tag, t.value)
	case filterTermMessage:
		matched = strings.Contains(lowerMessage, t.value)
	case filterTermPackage:
		matched = strings.EqualFold(packageName, t.value)
	case filterTermLevel:
		matched = entry.Level == "" || strings.EqualFold(entry.Level, t.value)
	default:
		matched = strings.Contains(lowerMessage, t.value)
	}

	if t.negated {
		return !matched
	}
	return matched
}

func compileFilterTerm(term string) compiledFilterTerm {
	normalized, negated := normalizeFilterTerm(term)
	switch {
	case strings.HasPrefix(normalized, "tag:"):
		return compiledFilterTerm{
			kind:    filterTermTag,
			value:   strings.TrimSpace(strings.TrimPrefix(normalized, "tag:")),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "message:"):
		value := strings.TrimSpace(strings.TrimPrefix(normalized, "message:"))
		return compiledFilterTerm{
			kind:    filterTermMessage,
			value:   strings.ToLower(strings.Trim(value, "\"")),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "package:"):
		return compiledFilterTerm{
			kind:    filterTermPackage,
			value:   strings.TrimSpace(strings.TrimPrefix(normalized, "package:")),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "level:"):
		return compiledFilterTerm{
			kind:    filterTermLevel,
			value:   strings.TrimSpace(strings.TrimPrefix(normalized, "level:")),
			negated: negated,
		}
	default:
		return compiledFilterTerm{
			kind:    filterTermText,
			value:   strings.ToLower(normalized),
			negated: negated,
		}
	}
}

func normalizeFilterTerm(term string) (string, bool) {
	trimmed := strings.TrimSpace(term)
	if !strings.HasPrefix(trimmed, "-") {
		return trimmed, false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "-")), true
}

func (c *Controller) setAppliedFilterLocked(query string) {
	c.model.Filter.Applied = query
	c.compiledFilter = compileFilterQuery(query)
}

func (c *Controller) matchesAppliedFilterLocked(entry logcat.LogEntry) bool {
	return c.compiledFilter.matches(entry, c.model.SelectedPackage)
}
