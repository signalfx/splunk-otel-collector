// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
)

type HecTokenConfig struct {
	Token          configopaque.String
	DefaultIndex   string
	AllowedIndexes []string
}

// Config specifies how the Per-RPC bearer token based authentication data should be obtained.
type Config struct {
	_      struct{}
	Tokens []HecTokenConfig `mapstructure:"tokens,omitempty"`
}

var (
	_                  component.Config = (*Config)(nil)
	errNoTokenProvided                  = errors.New("no bearer token provided")
)

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Tokens) == 0 {
		return errNoTokenProvided
	}
	return nil
}
