package app

import (
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

type filterNodeKind uint8

const (
	filterNodeAnd filterNodeKind = iota
	filterNodeOr
	filterNodeTerm
)

type filterTermKind uint8

const (
	filterTermText filterTermKind = iota
	filterTermTag
	filterTermMessage
	filterTermPackage
	filterTermLevel
	filterTermTagContains
	filterTermMessageContains
)

type compiledFilterQuery struct {
	root               *compiledFilterNode
	needsLowerMessage  bool
	needsLowerTag      bool
}

type compiledFilterNode struct {
	kind  filterNodeKind
	term  compiledFilterTerm
	left  *compiledFilterNode
	right *compiledFilterNode
}

type compiledFilterTerm struct {
	kind    filterTermKind
	value   string
	negated bool
}

func compileFilterQuery(query string) (compiledFilterQuery, error) {
	parser := newFilterParser(query)
	root, err := parser.parse()
	if err != nil {
		return compiledFilterQuery{}, err
	}
	if root == nil {
		return compiledFilterQuery{}, nil
	}

	compiled := compiledFilterQuery{root: root}
	compiled.inspectNeeds(root)
	return compiled, nil
}

func (q compiledFilterQuery) matchAll() bool {
	return q.root == nil
}

func (q compiledFilterQuery) matches(entry logcat.LogEntry, packageName string) bool {
	if q.matchAll() {
		return true
	}

	lowerMessage := ""
	if q.needsLowerMessage {
		lowerMessage = strings.ToLower(entry.Message)
	}

	lowerTag := ""
	if q.needsLowerTag {
		lowerTag = strings.ToLower(entry.Tag)
	}

	return q.root.matches(entry, packageName, lowerMessage, lowerTag)
}

func (q *compiledFilterQuery) inspectNeeds(node *compiledFilterNode) {
	if node == nil {
		return
	}
	if node.kind == filterNodeTerm {
		switch node.term.kind {
		case filterTermText, filterTermMessage, filterTermMessageContains:
			q.needsLowerMessage = true
		case filterTermTagContains:
			q.needsLowerTag = true
		}
		return
	}
	q.inspectNeeds(node.left)
	q.inspectNeeds(node.right)
}

func (n *compiledFilterNode) matches(
	entry logcat.LogEntry,
	packageName string,
	lowerMessage string,
	lowerTag string,
) bool {
	switch n.kind {
	case filterNodeAnd:
		return n.left.matches(entry, packageName, lowerMessage, lowerTag) &&
			n.right.matches(entry, packageName, lowerMessage, lowerTag)
	case filterNodeOr:
		return n.left.matches(entry, packageName, lowerMessage, lowerTag) ||
			n.right.matches(entry, packageName, lowerMessage, lowerTag)
	default:
		return n.term.matches(entry, packageName, lowerMessage, lowerTag)
	}
}

func (t compiledFilterTerm) matches(
	entry logcat.LogEntry,
	packageName string,
	lowerMessage string,
	lowerTag string,
) bool {
	matched := true
	switch t.kind {
	case filterTermTag:
		matched = entry.Tag == "" || strings.EqualFold(entry.Tag, t.value)
	case filterTermTagContains:
		matched = strings.Contains(lowerTag, t.value)
	case filterTermMessage:
		matched = strings.Contains(lowerMessage, t.value)
	case filterTermMessageContains:
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
	case strings.HasPrefix(normalized, "tag~:"):
		return compiledFilterTerm{
			kind:    filterTermTagContains,
			value:   strings.ToLower(trimQueryValue(strings.TrimPrefix(normalized, "tag~:"))),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "tag:"):
		return compiledFilterTerm{
			kind:    filterTermTag,
			value:   trimQueryValue(strings.TrimPrefix(normalized, "tag:")),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "message~:"):
		return compiledFilterTerm{
			kind:    filterTermMessageContains,
			value:   strings.ToLower(trimQueryValue(strings.TrimPrefix(normalized, "message~:"))),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "message:"):
		return compiledFilterTerm{
			kind:    filterTermMessage,
			value:   strings.ToLower(trimQueryValue(strings.TrimPrefix(normalized, "message:"))),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "package:"):
		return compiledFilterTerm{
			kind:    filterTermPackage,
			value:   trimQueryValue(strings.TrimPrefix(normalized, "package:")),
			negated: negated,
		}
	case strings.HasPrefix(normalized, "level:"):
		return compiledFilterTerm{
			kind:    filterTermLevel,
			value:   trimQueryValue(strings.TrimPrefix(normalized, "level:")),
			negated: negated,
		}
	default:
		return compiledFilterTerm{
			kind:    filterTermText,
			value:   strings.ToLower(trimQueryValue(normalized)),
			negated: negated,
		}
	}
}

func trimQueryValue(value string) string {
	trimmed := strings.TrimSpace(value)
	return strings.Trim(trimmed, "\"")
}

func normalizeFilterTerm(term string) (string, bool) {
	trimmed := strings.TrimSpace(term)
	if !strings.HasPrefix(trimmed, "-") {
		return trimmed, false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "-")), true
}

func (c *Controller) setAppliedFilterLocked(query string) {
	compiled, err := compileFilterQuery(query)
	if err != nil {
		c.setCompiledFilterLocked(query, compiledFilterQuery{})
		return
	}
	c.setCompiledFilterLocked(query, compiled)
}

func (c *Controller) setCompiledFilterLocked(query string, compiled compiledFilterQuery) {
	c.model.Filter.Applied = query
	c.compiledFilter = compiled
}

func (c *Controller) matchesAppliedFilterLocked(entry logcat.LogEntry) bool {
	return c.compiledFilter.matches(entry, c.model.SelectedPackage)
}
