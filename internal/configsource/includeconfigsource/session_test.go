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
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
)

func TestIncludeConfigSource_Session(t *testing.T) {
	tests := []struct {
		defaults map[string]interface{}
		params   map[string]interface{}
		expected interface{}
		wantErr  error
		name     string
		selector string
	}{
		{
			name:     "simple",
			selector: "no_params_template",
			expected: map[string]interface{}{
				"bool_field": true,
			},
		},
		{
			name:     "missing_file",
			selector: "not_to_be_found",
			wantErr:  &os.PathError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := newSession()
			require.NoError(t, err)
			require.NotNil(t, s)

			ctx := context.Background()
			defer func() {
				assert.NoError(t, s.RetrieveEnd(ctx))
				assert.NoError(t, s.Close(ctx))
			}()

			file := path.Join("testdata", tt.selector)
			r, err := s.Retrieve(ctx, file, tt.params)
			if tt.wantErr != nil {
				assert.Nil(t, r)
				require.IsType(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, r)
			buf := bytes.NewBuffer(r.Value().([]byte))
			p, err := config.NewParserFromBuffer(buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, p.ToStringMap())
		})
	}
}
