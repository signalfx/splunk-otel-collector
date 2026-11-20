// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	cm, err := confmaptest.LoadConf("testdata/config.yaml")
	require.NoError(t, err)
	sub, err := cm.Sub("splunk_modinput")
	require.NoError(t, err)
	cfg := NewFactory().CreateDefaultConfig()
	require.NoError(t, sub.Unmarshal(cfg))
	assert.Equal(t, &Config{Source: "test-source", Sourcetype: "test-sourcetype", Index: "test-index"}, cfg)
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	assert.Equal(t, &Config{}, NewFactory().CreateDefaultConfig())
}

func TestRejectUnknownConfig(t *testing.T) {
	t.Parallel()
	cfg := NewFactory().CreateDefaultConfig()
	err := confmap.NewFromStringMap(map[string]any{"unknown": true}).Unmarshal(cfg)
	require.ErrorContains(t, err, "unknown")
}
