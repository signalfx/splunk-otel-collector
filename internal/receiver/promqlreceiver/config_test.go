// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promqlreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		cfg     Config
	}{
		{
			name:    "no queries",
			cfg:     Config{ClientConfig: confighttp.ClientConfig{Endpoint: "http://localhost:9090/api/v1/query"}},
			wantErr: "queries cannot be empty",
		},
		{
			name: "empty query string",
			cfg: Config{
				Queries:      []Query{{Query: ""}},
				ClientConfig: confighttp.ClientConfig{Endpoint: "http://localhost:9090/api/v1/query"},
			},
			wantErr: "query cannot be empty",
		},
		{
			name: "missing endpoint",
			cfg: Config{
				Queries: []Query{{Query: "up"}},
			},
			wantErr: "endpoint is required",
		},
		{
			name: "valid",
			cfg: Config{
				Queries:      []Query{{Query: "up"}},
				ClientConfig: confighttp.ClientConfig{Endpoint: "http://localhost:9090/api/v1/query"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}
