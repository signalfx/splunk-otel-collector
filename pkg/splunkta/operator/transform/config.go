// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package transform

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	stanza_errors "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/stanzaerrors"
)

const operatorType = "transform"

// NewConfig creates a new Config for the given scope and transform.
func NewConfig(scope string, t conf.Transform) *Config {
	return &Config{
		Regex:        t.Regex,
		Replacement:  t.Format,
		ParserConfig: helper.NewParserConfig(fmt.Sprintf("transforms-%q-%q", scope, t.Name), operatorType),
	}
}

// Config is the configuration of a transform operator.
type Config struct {
	helper.ParserConfig `mapstructure:",squash"`

	Regex       string `mapstructure:"regex"`
	Replacement string `mapstructure:"replacement"`

	Cache struct {
		Size uint16 `mapstructure:"size"`
	} `mapstructure:"cache"`
}

// Build will build a transform operator.
func (c Config) Build(set component.TelemetrySettings) (operator.Operator, error) {
	parserOperator, err := c.ParserConfig.Build(set)
	if err != nil {
		return nil, err
	}

	if c.Regex == "" {
		return nil, errors.New("missing required field 'regex'")
	}

	r, err := regexp.Compile(c.Regex)
	if err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}

	namedCaptureGroups := 0
	for _, groupName := range r.SubexpNames() {
		if groupName != "" {
			namedCaptureGroups++
		}
	}
	if namedCaptureGroups == 0 && c.Replacement == "" {
		return nil, stanza_errors.NewError(
			"no named capture groups in regex pattern",
			"use named capture groups like '^(?P<my_key>.*)$' to specify the key name for the parsed field",
		)
	}

	op := &Parser{
		ParserOperator: parserOperator,
		regexp:         r,
		replacement:    c.Replacement,
		parseValues:    namedCaptureGroups > 0,
	}

	if c.Cache.Size > 0 {
		op.cache = newMemoryCache(c.Cache.Size, 0)
		set.Logger.Debug(
			"configured memory cache",
			zap.String("operator_id", op.ID()),
			zap.Uint16("size", op.cache.maxSize()),
		)
	}

	return op, nil
}
