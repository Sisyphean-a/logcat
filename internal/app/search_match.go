package app

import (
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

type compiledSearchQuery struct {
	literal        string
	groups         []searchGroup
	highlightTerms []string
}

type searchGroup struct {
	terms []searchTerm
}

type searchTerm struct {
	needle  string
	negated bool
}

func searchLowerText(entry logcat.LogEntry) string {
	return strings.ToLower(entry.Tag + "\n" + entry.Message)
}

func compileSearchQuery(query string) compiledSearchQuery {
	normalized := normalizedSearchQuery(query)
	if normalized == "" {
		return compiledSearchQuery{}
	}
	if !searchQueryUsesOperators(normalized) {
		return newLiteralSearchQuery(normalized)
	}

	compiled := newBooleanSearchQuery(normalized)
	if compiled.matchAll() {
		return newLiteralSearchQuery(normalized)
	}
	return compiled
}

func newLiteralSearchQuery(literal string) compiledSearchQuery {
	return compiledSearchQuery{
		literal:        literal,
		highlightTerms: []string{literal},
	}
}

func newBooleanSearchQuery(query string) compiledSearchQuery {
	groups := buildSearchGroups(query)
	if len(groups) == 0 {
		return compiledSearchQuery{}
	}

	return compiledSearchQuery{
		groups:         groups,
		highlightTerms: collectSearchHighlightTerms(groups),
	}
}

func searchQueryUsesOperators(query string) bool {
	return strings.Contains(query, "&&") ||
		strings.Contains(query, "||") ||
		strings.HasPrefix(query, "-")
}

func buildSearchGroups(query string) []searchGroup {
	parts := strings.Split(query, "||")
	groups := make([]searchGroup, 0, len(parts))
	for _, part := range parts {
		group := buildSearchGroup(part)
		if len(group.terms) == 0 {
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

func buildSearchGroup(query string) searchGroup {
	parts := strings.Split(query, "&&")
	terms := make([]searchTerm, 0, len(parts))
	for _, part := range parts {
		term, ok := buildSearchTerm(part)
		if !ok {
			continue
		}
		terms = append(terms, term)
	}
	return searchGroup{terms: terms}
}

func buildSearchTerm(part string) (searchTerm, bool) {
	term := strings.TrimSpace(part)
	if term == "" {
		return searchTerm{}, false
	}

	negated := false
	for strings.HasPrefix(term, "-") {
		negated = !negated
		term = strings.TrimSpace(term[1:])
	}
	if term == "" {
		return searchTerm{}, false
	}

	return searchTerm{needle: term, negated: negated}, true
}

func collectSearchHighlightTerms(groups []searchGroup) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0, len(groups))
	for _, group := range groups {
		for _, term := range group.terms {
			if term.negated {
				continue
			}
			if _, ok := seen[term.needle]; ok {
				continue
			}
			seen[term.needle] = struct{}{}
			terms = append(terms, term.needle)
		}
	}
	return terms
}

func (q compiledSearchQuery) matchAll() bool {
	return q.literal == "" && len(q.groups) == 0
}

func (q compiledSearchQuery) matches(searchLower string) bool {
	if q.literal != "" {
		return strings.Contains(searchLower, q.literal)
	}
	if len(q.groups) == 0 {
		return true
	}

	for _, group := range q.groups {
		if group.matches(searchLower) {
			return true
		}
	}
	return false
}

func (g searchGroup) matches(searchLower string) bool {
	for _, term := range g.terms {
		matched := strings.Contains(searchLower, term.needle)
		if term.negated && matched {
			return false
		}
		if !term.negated && !matched {
			return false
		}
	}
	return len(g.terms) > 0
}
