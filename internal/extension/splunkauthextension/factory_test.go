// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestFactoryCreate(t *testing.T) {
	factory := NewFactory(sourceFunc(func(_ context.Context) ([]HecTokenConfig, error) { return nil, nil }))
	cfg := factory.CreateDefaultConfig().(*Config)
	require.Equal(t, 3*time.Minute, cfg.RefreshInterval)
	require.Equal(t, 30*time.Second, cfg.RefreshTimeout)
	require.Equal(t, 10*time.Minute, cfg.MaxStaleness)
	require.NoError(t, cfg.Validate())
	ext, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
	_, err = NewFactory(nil).Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.EqualError(t, err, "token source is required")
	for _, field := range []string{"interval", "timeout", "staleness"} {
		for _, value := range []time.Duration{0, -time.Second} {
			invalid := *cfg
			switch field {
			case "interval":
				invalid.RefreshInterval = value
			case "timeout":
				invalid.RefreshTimeout = value
			case "staleness":
				invalid.MaxStaleness = value
			}
			require.Error(t, invalid.Validate())
			_, err = factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), &invalid)
			require.Error(t, err)
		}
	}
}
