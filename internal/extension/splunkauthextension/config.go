// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"errors"
	"net/url"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
)

type HecTokenConfig struct {
	Token          configopaque.String
	DefaultIndex   string
	AllowedIndexes []string
	Disabled       bool
}

// Config controls token refresh and the lifetime of cached authorization data.
type Config struct {
	// Endpoint defaults to SPLUNK_MANAGEMENT_URI when using NewFactory.
	Endpoint string `mapstructure:"endpoint"`
	// SessionKey defaults to SPLUNK_SESSION_KEY when using NewFactory.
	SessionKey      configopaque.String    `mapstructure:"session_key"`
	TLS             configtls.ClientConfig `mapstructure:"tls"`
	RefreshInterval time.Duration          `mapstructure:"refresh_interval"`
	RefreshTimeout  time.Duration          `mapstructure:"refresh_timeout"`
	MaxStaleness    time.Duration          `mapstructure:"max_staleness"`
}

var _ component.Config = (*Config)(nil)

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if cfg.SessionKey == "" {
		return errors.New("session_key is required")
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("endpoint must be an HTTP or HTTPS URL without credentials, query, or fragment")
	}
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
