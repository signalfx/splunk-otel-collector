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

package nutanixreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		port     int
	}{
		{
			name:     "host only",
			endpoint: "prism.example.com",
			port:     9440,
			want:     "https://prism.example.com:9440",
		},
		{
			name:     "scheme and port",
			endpoint: "http://127.0.0.1:8080",
			port:     9440,
			want:     "http://127.0.0.1:8080",
		},
		{
			name:     "strip path",
			endpoint: "https://prism.example.com:9440/foo?bar=baz",
			port:     9440,
			want:     "https://prism.example.com:9440",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEndpoint(tt.endpoint, tt.port)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.String())
		})
	}
}
