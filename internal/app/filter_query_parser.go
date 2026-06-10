package app

import (
	"fmt"
	"strings"
)

type filterParser struct {
	query string
	index int
}

func newFilterParser(query string) *filterParser {
	return &filterParser{query: strings.TrimSpace(query)}
}

func (p *filterParser) parse() (*compiledFilterNode, error) {
	if p.query == "" {
		return nil, nil
	}

	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	p.skipSpaces()
	if p.index != len(p.query) {
		return nil, fmt.Errorf("filter_query_invalid: unexpected token near %q", p.query[p.index:])
	}
	return node, nil
}

func (p *filterParser) parseOr() (*compiledFilterNode, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for {
		p.skipSpaces()
		if !p.consumeOperator("||") && !p.consumeOperator("|") {
			return left, nil
		}

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &compiledFilterNode{
			kind:  filterNodeOr,
			left:  left,
			right: right,
		}
	}
}

func (p *filterParser) parseAnd() (*compiledFilterNode, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		p.skipSpaces()
		if !p.consumeOperator("&&") && !p.consumeOperator("&") {
			return left, nil
		}

		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &compiledFilterNode{
			kind:  filterNodeAnd,
			left:  left,
			right: right,
		}
	}
}

func (p *filterParser) parsePrimary() (*compiledFilterNode, error) {
	p.skipSpaces()
	if p.index >= len(p.query) {
		return nil, fmt.Errorf("filter_query_invalid: expected filter term")
	}

	if p.query[p.index] == '(' {
		p.index++
		node, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		p.skipSpaces()
		if p.index >= len(p.query) || p.query[p.index] != ')' {
			return nil, fmt.Errorf("filter_query_invalid: unmatched '('")
		}
		p.index++
		return node, nil
	}

	term, err := p.readTerm()
	if err != nil {
		return nil, err
	}
	return &compiledFilterNode{
		kind: filterNodeTerm,
		term: compileFilterTerm(term),
	}, nil
}

func (p *filterParser) readTerm() (string, error) {
	start := p.index
	inQuote := false

	for p.index < len(p.query) {
		ch := p.query[p.index]
		switch ch {
		case '"':
			inQuote = !inQuote
			p.index++
		case '(':
			if inQuote {
				p.index++
				continue
			}
			return "", fmt.Errorf("filter_query_invalid: unexpected '('")
		case ')':
			if inQuote {
				p.index++
				continue
			}
			term := strings.TrimSpace(p.query[start:p.index])
			if term == "" {
				return "", fmt.Errorf("filter_query_invalid: empty filter term")
			}
			return term, nil
		case '&', '|':
			if inQuote {
				p.index++
				continue
			}
			term := strings.TrimSpace(p.query[start:p.index])
			if term == "" {
				return "", fmt.Errorf("filter_query_invalid: empty filter term")
			}
			return term, nil
		default:
			p.index++
		}
	}

	if inQuote {
		return "", fmt.Errorf("filter_query_invalid: unmatched quote")
	}

	term := strings.TrimSpace(p.query[start:p.index])
	if term == "" {
		return "", fmt.Errorf("filter_query_invalid: empty filter term")
	}
	return term, nil
}

func (p *filterParser) skipSpaces() {
	for p.index < len(p.query) && p.query[p.index] == ' ' {
		p.index++
	}
}

func (p *filterParser) consumeOperator(operator string) bool {
	if !strings.HasPrefix(p.query[p.index:], operator) {
		return false
	}
	p.index += len(operator)
	return true
}
