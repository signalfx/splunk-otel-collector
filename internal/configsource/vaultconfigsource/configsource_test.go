// Copyright 2020 Splunk, Inc.
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

package vaultconfigsource

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestVaultNewConfigSource(t *testing.T) {
	tests := []struct {
		config  *Config
		name    string
		wantErr bool
	}{
		{
			name: "minimal",
			config: &Config{
				Endpoint: "https://some.server:1234/",
			},
		},
		{
			name: "invalid_endpoint",
			config: &Config{
				Endpoint: "some\bad_endpoint",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgSrc, err := newConfigSource(zap.NewNop(), tt.config)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfgSrc)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfgSrc)
		})
	}
}
