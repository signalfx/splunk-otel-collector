// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

func TestDryRun(t *testing.T) {
	config := confmap.NewFromStringMap(map[string]interface{}{"one": "one"})
	dr := NewDryRun(false)
	require.NotPanics(t, func() {
		dr.Convert(context.Background(), config)
	})

	origStdOut := os.Stdout
	stdout, err := os.CreateTemp("", "stdout")
	require.NoError(t, err)
	require.NotNil(t, stdout)

	t.Cleanup(func() func() {
		os.Stdout = stdout
		return func() {
			os.Stdout = origStdOut
			require.NoError(t, stdout.Close())
			require.NoError(t, os.Remove(stdout.Name()))
		}
	}())

	dr = NewDryRun(true)
	require.Panics(t, func() {
		dr.Convert(context.Background(), config)
	})
	os.Stdout = origStdOut
	stdout.Seek(0, 0)
	out, err := io.ReadAll(stdout)
	require.NoError(t, err)
	actual := map[string]any{}
	yaml.Unmarshal(out, &actual)
	require.Equal(t, config.ToStringMap(), actual)
}
