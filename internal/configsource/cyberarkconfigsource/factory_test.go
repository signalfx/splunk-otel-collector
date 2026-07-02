// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cyberarkconfigsource

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestCyberArkFactory_CreateConfigSource(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, component.MustNewType("cyberark"), factory.Type())

	// baseCfg returns a minimal valid config that individual cases mutate.
	baseCfg := func() *Config {
		return &Config{
			RetrievalMode: retrievalModeCP,
			BinaryPath:    defaultBinaryPath,
			AppID:         "collector-app",
			Safe:          "DBSecrets",
			Object:        "prod-db",
			PollInterval:  defaultPollInterval,
		}
	}

	tests := []struct {
		mutate  func(*Config)
		wantErr error
		name    string
	}{
		{
			name:    "unsupported_mode",
			mutate:  func(c *Config) { c.RetrievalMode = "ccp" },
			wantErr: &errUnsupportedMode{},
		},
		{
			name:    "missing_app_id",
			mutate:  func(c *Config) { c.AppID = "" },
			wantErr: &errMissingAppID{},
		},
		{
			name:    "missing_safe",
			mutate:  func(c *Config) { c.Safe = "" },
			wantErr: &errMissingSafe{},
		},
		{
			name:    "missing_object",
			mutate:  func(c *Config) { c.Object = "" },
			wantErr: &errMissingObject{},
		},
		{
			name: "auto_refresh_without_poll_interval",
			mutate: func(c *Config) {
				c.AutoRefresh = true
				c.PollInterval = 0
			},
			wantErr: &errNonPositivePollInterval{},
		},
		{
			name:   "success_static",
			mutate: func(*Config) {},
		},
		{
			name: "success_auto_refresh",
			mutate: func(c *Config) {
				c.AutoRefresh = true
				c.PollInterval = 30 * time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseCfg()
			tt.mutate(cfg)

			actual, err := factory.CreateConfigSource(context.Background(), cfg, zap.NewNop())
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, actual)
				assert.IsType(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, actual)
			}
		})
	}
}

func TestCyberArkFactory_CreateDefaultConfig(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig().(*Config)
	assert.Equal(t, retrievalModeCP, cfg.RetrievalMode)
	assert.Equal(t, defaultBinaryPath, cfg.BinaryPath)
	assert.Equal(t, defaultPollInterval, cfg.PollInterval)
	assert.False(t, cfg.AutoRefresh)
}
