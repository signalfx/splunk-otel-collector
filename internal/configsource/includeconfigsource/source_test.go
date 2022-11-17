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
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestIncludeConfigSource_Session(t *testing.T) {
	tests := []struct {
		defaults map[string]any
		params   map[string]any
		expected any
		wantErr  error
		name     string
		selector string
	}{
		{
			name:     "missing_file",
			selector: "not_to_be_found",
			wantErr:  &os.PathError{},
		},
		{
			name:     "scalar_data_file",
			selector: "scalar_data_file",
			expected: "42",
		},
		{
			name:     "no_params_template",
			selector: "no_params_template",
			expected: "bool_field: true",
		},
		{
			name:     "param_template",
			selector: "param_template",
			params: map[string]any{
				"glob_pattern": "myPattern",
			},
			expected: "logs_path: myPattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := newConfigSource(configprovider.CreateParams{}, &Config{})
			require.NoError(t, err)

			ctx := context.Background()
			file := path.Join("testdata", tt.selector)
			r, err := s.Retrieve(ctx, file, confmap.NewFromStringMap(tt.params), nil)
			if tt.wantErr != nil {
				assert.Nil(t, r)
				require.IsType(t, tt.wantErr, err)
				assert.NoError(t, s.Shutdown(ctx))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)

			val, err := r.AsRaw()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, val)
			require.NoError(t, r.Close(context.Background()))
			require.NoError(t, s.Shutdown(ctx))
		})
	}
}

func TestIncludeConfigSourceWatchFileClose(t *testing.T) {
	s, err := newConfigSource(configprovider.CreateParams{}, &Config{WatchFiles: true})
	require.NoError(t, err)
	require.NotNil(t, s)

	// Write out an initial test file
	f, err := os.CreateTemp("", "watch_file_test")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(f.Name()))
	}()
	_, err = f.Write([]byte("val1"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Perform initial retrieve
	ctx := context.Background()
	r, err := s.Retrieve(ctx, f.Name(), nil, func(event *confmap.ChangeEvent) {
		panic(event)
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	val, err := r.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, s.Shutdown(ctx))
}

func TestIncludeConfigSource_WatchFileUpdate(t *testing.T) {
	s, err := newConfigSource(configprovider.CreateParams{}, &Config{WatchFiles: true})
	require.NoError(t, err)
	require.NotNil(t, s)

	// Write out an initial test file
	dst := path.Join(t.TempDir(), "watch_file_test")
	require.NoError(t, os.WriteFile(dst, []byte("val1"), 0600))

	// Perform initial retrieve
	watchChannel := make(chan *confmap.ChangeEvent, 1)
	ctx := context.Background()
	r, err := s.Retrieve(ctx, dst, nil, func(event *confmap.ChangeEvent) {
		watchChannel <- event
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	val, err := r.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// Write update to file
	require.NoError(t, os.WriteFile(dst, []byte("val2"), 0600))

	ce := <-watchChannel
	assert.NoError(t, ce.Error)
	require.NoError(t, r.Close(context.Background()))

	// Check updated file after waiting for update
	r, err = s.Retrieve(ctx, dst, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, r)

	val, err = r.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "val2", val)
	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, s.Shutdown(ctx))
}

func TestIncludeConfigSourceDeleteFile(t *testing.T) {
	s, err := newConfigSource(configprovider.CreateParams{}, &Config{DeleteFiles: true})
	require.NoError(t, err)
	require.NotNil(t, s)

	// Copy test file
	contents, err := os.ReadFile(path.Join("testdata", "scalar_data_file"))
	require.NoError(t, err)
	dst := path.Join(t.TempDir(), "copy_scalar_data_file")
	require.NoError(t, os.WriteFile(dst, contents, 0600))

	ctx := context.Background()
	r, err := s.Retrieve(ctx, dst, nil, func(event *confmap.ChangeEvent) {
		panic(event)
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	val, err := r.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "42", val)

	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, s.Shutdown(ctx))
}

func TestIncludeConfigSource_DeleteFileError(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Locking the file is trivial on Windows, but not on *nix given the
		// golang API, run the test only on Windows.
		t.Skip("Windows only test")
	}

	s, err := newConfigSource(configprovider.CreateParams{}, &Config{DeleteFiles: true})
	require.NoError(t, err)

	// Copy test file
	contents, err := os.ReadFile(path.Join("testdata", "scalar_data_file"))
	require.NoError(t, err)
	dst := path.Join("testdata", "copy_scalar_data_file")
	require.NoError(t, os.WriteFile(dst, contents, 0600))
	f, err := os.OpenFile(dst, os.O_RDWR, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, os.Remove(dst))
	})

	ctx := context.Background()
	r, err := s.Retrieve(ctx, dst, nil, nil)
	assert.IsType(t, &errFailedToDeleteFile{}, err)
	assert.Nil(t, r)

	require.NoError(t, s.Shutdown(ctx))
}
