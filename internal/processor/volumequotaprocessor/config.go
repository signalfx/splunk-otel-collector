// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package volumequotaprocessor

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	Limits       Limits        `mapstructure:"limits"`
	GlobalLimits GlobalLimits  `mapstructure:"global_limits"`
	Epoch        time.Duration `mapstructure:"epoch"`
	Lookback     int           `mapstructure:"lookback"`
}

func (c *Config) Validate() error {
	if c.GlobalLimits.Spans < 0 {
		return errors.New("spans global limit must be zero or positive")
	}
	if c.GlobalLimits.Traces < 0 {
		return errors.New("traces global limit must be zero or positive")
	}
	for k, v := range c.Limits.Spans {
		if v <= 0 {
			return fmt.Errorf("span limit for service %q must be positive", k)
		}
	}
	for k, v := range c.Limits.Traces {
		if v <= 0 {
			return fmt.Errorf("trace limit for service %q must be positive", k)
		}
	}
	if c.GlobalLimits.Spans <= 0 && c.GlobalLimits.Traces <= 0 && len(c.Limits.Traces) == 0 && len(c.Limits.Spans) == 0 {
		return errors.New("no limits set")
	}

	return nil
}

type GlobalLimits struct {
	Spans  int64 `mapstructure:"spans"`
	Traces int64 `mapstructure:"traces"`
}

type Limits struct {
	Spans  map[string]int64 `mapstructure:"spans"`
	Traces map[string]int64 `mapstructure:"traces"`
}
