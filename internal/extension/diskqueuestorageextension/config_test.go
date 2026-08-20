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

package diskqueuestorageextension

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr string
		cfg         Config
	}{
		{
			name: "valid config",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: 1024,
				SyncEvery:       1,
				SyncTimeout:     time.Second,
				CompactInterval: time.Minute,
			},
		},
		{
			name: "missing path",
			cfg: Config{
				Path:            "",
				MaxBytesPerFile: 1024,
				SyncEvery:       1,
				SyncTimeout:     time.Second,
				CompactInterval: time.Minute,
			},
			expectedErr: "path must be a valid folder",
		},
		{
			name: "invalid max_bytes_per_file",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: 0,
				SyncEvery:       1,
				SyncTimeout:     time.Second,
				CompactInterval: time.Minute,
			},
			expectedErr: "max_bytes_per_file must be a positive value",
		},
		{
			name: "negative max_bytes_per_file",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: -1,
				SyncEvery:       1,
				SyncTimeout:     time.Second,
				CompactInterval: time.Minute,
			},
			expectedErr: "max_bytes_per_file must be a positive value",
		},
		{
			name: "invalid sync_every",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: 1024,
				SyncEvery:       0,
				SyncTimeout:     time.Second,
				CompactInterval: time.Minute,
			},
			expectedErr: "sync_every must be a positive value",
		},
		{
			name: "invalid sync_timeout",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: 1024,
				SyncEvery:       1,
				SyncTimeout:     0,
				CompactInterval: time.Minute,
			},
			expectedErr: "sync_timeout must be a positive value",
		},
		{
			name: "invalid compact_interval",
			cfg: Config{
				Path:            "/tmp/queue",
				MaxBytesPerFile: 1024,
				SyncEvery:       1,
				SyncTimeout:     time.Second,
				CompactInterval: 0,
			},
			expectedErr: "compact_interval must be a positive value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
