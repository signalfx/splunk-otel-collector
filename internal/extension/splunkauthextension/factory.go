// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// NewFactory creates authenticators that refresh tokens using source.
// A source must return a complete listing on success, including an empty listing
// when no tokens remain. It must honor context cancellation.
func NewFactory(source TokenSource) extension.Factory {
	return extension.NewFactory(
		component.MustNewType("splunkauth"),
		createDefaultConfig,
		func(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
			if source == nil {
				return nil, errors.New("token source is required")
			}
			config := cfg.(*Config)
			if err := config.Validate(); err != nil {
				return nil, err
			}
			return newSplunkAuth(config, source, set.Logger), nil
		},
		component.StabilityLevelAlpha,
	)
}

func createDefaultConfig() component.Config {
	return NewDefaultConfig()
}
