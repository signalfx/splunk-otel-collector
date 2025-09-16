// Copyright  Splunk, Inc.
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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestValidateReceiverMetadata(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
	}{
		{
			name: "valid_receiver",
		},
		{
			name:          "missing_service_type",
			expectedError: "service type cannot be empty for receiver a_receiver",
		},
		{
			name:          "missing_status",
			expectedError: "`status` must contain at least one `metrics` or `statements` list",
		},
		{
			name:          "invalid_status_types",
			expectedError: `"metrics" status match validation failed: invalid status "unsupported". must be one of [successful partial failed]`,
		},
		{
			name:          "multiple_status_match_types",
			expectedError: `"metrics" status match validation failed. Must provide one of [regexp strict expr] but received [strict regexp]`,
		},
		{
			name:          "missing_match_status",
			expectedError: `"metrics" status match validation failed: status cannot be empty`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("%s.yaml", test.name)))
			require.NoError(t, err)

			var metadata receiverMetadata
			err = yaml.Unmarshal(data, &metadata)
			require.NoError(t, err)

			err = validateReceiverMetadata(metadata)
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}
