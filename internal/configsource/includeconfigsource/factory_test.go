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

package includeconfigsource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestIncludeConfigSourceFactory_CreateConfigSource(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, component.MustNewType("include"), factory.Type())

	tests := []struct {
		expected *includeConfigSource
		name     string
		config   Config
		wantErr  bool
	}{
		{
			name: "default",
			expected: &includeConfigSource{
				Config:       &Config{},
				watchedFiles: make(map[string]struct{}),
			},
		},
		{
			name:   "delete_files",
			config: Config{DeleteFiles: true},
			expected: &includeConfigSource{
				Config:       &Config{DeleteFiles: true},
				watchedFiles: make(map[string]struct{}),
			},
		},
		{
			name:   "watch_files",
			config: Config{WatchFiles: true},
			expected: &includeConfigSource{
				Config:       &Config{WatchFiles: true},
				watchedFiles: make(map[string]struct{}),
			},
		},
		{
			name: "err_on_delete_and_watch",
			config: Config{
				DeleteFiles: true,
				WatchFiles:  true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			actual, err := factory.CreateConfigSource(context.Background(), &tt.config, zap.NewNop())
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, actual)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
