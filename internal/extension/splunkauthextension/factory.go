// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension"

	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/modularinput"
)

// NewFactory creates authenticators configured through collector YAML.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		createDefaultConfig,
		func(ctx context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
			config := cfg.(*Config)
			if err := config.Validate(); err != nil {
				return nil, err
			}
			source, err := newTokenSource(ctx, config)
			if err != nil {
				return nil, err
			}
			return newSplunkAuth(config, source, set.Logger), nil
		},
		metadata.ExtensionStability,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Endpoint:        os.Getenv(modularinput.EnvManagementURI),
		SessionKey:      configopaque.String(os.Getenv(modularinput.EnvSessionKey)),
		RefreshInterval: 3 * time.Minute,
		RefreshTimeout:  30 * time.Second,
		MaxStaleness:    10 * time.Minute,
	}
}
