// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package scriptedinput implements the scripted input operator.
package scriptedinput

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

const operatorType = "scripted_input"

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// NewConfig creates a new input config with default values
func NewConfig() *Config {
	return NewConfigWithID(operatorType)
}

// NewConfigWithID creates a new input config with default values
func NewConfigWithID(operatorID string) *Config {
	return &Config{
		InputConfig: helper.NewInputConfig(operatorID, operatorType),
	}
}

// Config holds the configuration for the ScriptedInput operator.
type Config struct {
	BaseDir            string
	conf.Input         `mapstructure:"-"`
	helper.InputConfig `mapstructure:"-"`
}

// Build builds the Config into a ScriptedInput operator.
func (c Config) Build(set component.TelemetrySettings) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(set)
	if err != nil {
		return nil, err
	}

	input := &ScriptedInput{
		InputOperator: inputOperator,
		logger:        set.Logger,
		doneChan:      make(chan struct{}),
		cfg:           c,
	}

	return input, nil
}
