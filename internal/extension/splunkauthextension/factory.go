// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType("splunkauth"),
		createDefaultConfig,
		createExtension,
		component.StabilityLevelAlpha,
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createExtension(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	return newSplunkAuth(cfg.(*Config), set.Logger), nil
}
