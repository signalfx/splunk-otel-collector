// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
)

type HecTokenConfig struct {
	Token          configopaque.String
	DefaultIndex   string
	AllowedIndexes []string
	Disabled       bool
}

// Config controls token refresh and the lifetime of cached authorization data.
// The token source is supplied separately through NewFactory.
type Config struct {
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	RefreshTimeout  time.Duration `mapstructure:"refresh_timeout"`
	MaxStaleness    time.Duration `mapstructure:"max_staleness"`
}

var _ component.Config = (*Config)(nil)

// NewDefaultConfig returns the default refresh settings for programmatic callers.
func NewDefaultConfig() *Config {
	return &Config{
		RefreshInterval: 3 * time.Minute,
		RefreshTimeout:  30 * time.Second,
		MaxStaleness:    10 * time.Minute,
	}
}

func (cfg *Config) Validate() error {
	if cfg.RefreshInterval <= 0 {
		return errors.New("refresh_interval must be positive")
	}
	if cfg.RefreshTimeout <= 0 {
		return errors.New("refresh_timeout must be positive")
	}
	if cfg.MaxStaleness <= 0 {
		return errors.New("max_staleness must be positive")
	}
	// Compare by subtraction to avoid overflowing the sum of the durations.
	if cfg.RefreshInterval >= cfg.MaxStaleness || cfg.RefreshTimeout >= cfg.MaxStaleness-cfg.RefreshInterval {
		return errors.New("max_staleness must be greater than refresh_interval + refresh_timeout")
	}
	return nil
}
