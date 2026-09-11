// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package transform

import (
	"context"
	"fmt"
	"regexp"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

// Parser is an operator that parses regex in an entry.
type Parser struct {
	helper.ParserOperator
	cache       cache
	regexp      *regexp.Regexp
	replacement string
	parseValues bool
}

func (p *Parser) Stop() error {
	if p.cache != nil {
		p.cache.stop()
	}
	return nil
}

func (p *Parser) ProcessBatch(ctx context.Context, entries []*entry.Entry) error {
	return p.ProcessBatchWith(ctx, entries, p.parse)
}

// Process will parse an entry for regex.
func (p *Parser) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.parse)
}

// parse will parse a value using the supplied regex.
func (p *Parser) parse(value any) (any, error) {
	var raw string
	switch m := value.(type) {
	case string:
		raw = m
	default:
		return nil, fmt.Errorf("type '%T' cannot be parsed as regex", value)
	}
	return p.match(raw)
}

func (p *Parser) match(value string) (any, error) {
	if p.cache != nil {
		if x := p.cache.get(value); x != nil {
			return x, nil
		}
	}

	parsedValues := make(map[string]any)
	if p.parseValues {
		var err error
		parsedValues, err = helper.MatchValues(value, p.regexp)
		if err != nil {
			return nil, err
		}

		if p.cache != nil {
			p.cache.add(value, parsedValues)
		}
	}
	if p.replacement != "" {
		parsedValues["body"] = p.regexp.ReplaceAllString(value, p.replacement)
		if p.cache != nil {
			p.cache.add(value, parsedValues)
		}
	}
	return parsedValues, nil
}
