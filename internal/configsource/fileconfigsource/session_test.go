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

package fileconfigsource

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/experimental/configsource"
)

func TestFileConfigSource_Session(t *testing.T) {
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
			selector: "scalar_data_file",
			expected: []byte("42"),
		},
		{
			name: "invalid_param",
			params: map[string]interface{}{
				"unknown_params_field": true,
			},
			wantErr: &errInvalidRetrieveParams{},
		},
		{
			name:     "missing_file",
			selector: "not_to_be_found",
			wantErr:  &errMissingRequiredFile{},
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
			assert.Equal(t, tt.expected, r.Value())
		})
	}
}

func TestFileConfigSource_DeleteFile(t *testing.T) {
	s, err := newSession()
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()
	defer func() {
		assert.NoError(t, s.RetrieveEnd(ctx))
		assert.NoError(t, s.Close(ctx))
	}()

	// Copy test file
	src := path.Join("testdata", "scalar_data_file")
	contents, err := ioutil.ReadFile(src)
	require.NoError(t, err)
	dst := path.Join("testdata", "copy_scalar_data_file")
	require.NoError(t, ioutil.WriteFile(dst, contents, 0644))
	t.Cleanup(func() {
		// It should be removed prior to this so an error is expected.
		assert.Error(t, os.Remove(dst))
	})

	params := map[string]interface{}{
		"delete": true,
	}
	r, err := s.Retrieve(ctx, dst, params)

	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, []byte("42"), r.Value())

	currentTime := time.Now().Local()
	err = os.Chtimes(dst, currentTime, currentTime)
	assert.Error(t, err)
	assert.Equal(t, configsource.ErrWatcherNotSupported, r.WatchForUpdate())
}

func TestFileConfigSource_DeleteFileError(t *testing.T) {
	s, err := newSession()
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()
	defer func() {
		assert.NoError(t, s.RetrieveEnd(ctx))
		assert.NoError(t, s.Close(ctx))
	}()

	// Copy test file
	src := path.Join("testdata", "scalar_data_file")
	contents, err := ioutil.ReadFile(src)
	require.NoError(t, err)
	dst := path.Join("testdata", "copy_scalar_data_file")
	require.NoError(t, ioutil.WriteFile(dst, contents, 0644))
	f, err := os.OpenFile(dst, os.O_RDWR, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, os.Remove(dst))
	})

	params := map[string]interface{}{
		"delete": true,
	}
	r, err := s.Retrieve(ctx, dst, params)
	assert.IsType(t, &errFailedToDeleteFile{}, err)
	assert.Nil(t, r)
}

func TestFileConfigSource_DisableWatch(t *testing.T) {
	s, err := newSession()
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()
	defer func() {
		assert.NoError(t, s.RetrieveEnd(ctx))
		assert.NoError(t, s.Close(ctx))
	}()

	src := path.Join("testdata", "scalar_data_file")
	params := map[string]interface{}{
		"disable_watch": true,
	}
	r, err := s.Retrieve(ctx, src, params)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, []byte("42"), r.Value())

	assert.Equal(t, configsource.ErrWatcherNotSupported, r.WatchForUpdate())
}
