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
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
)

func TestDiscovery(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions: [ext/one, ext/two, ext/three, ext/four]
  extensions/splunk.discovery: [ext/four, ext/five]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three, recv/four]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
  receivers/splunk.discovery: [recv/four, recv/five]
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one, ext/two, ext/three, ext/four, ext/five]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three, recv/four, recv/five]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestDiscoveryNotDetected(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions: [ext/one, ext/two, ext/three]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one, ext/two, ext/three]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestDiscoveryExtensionsOnly(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions/splunk.discovery: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestDiscoveryEmptyExtensions(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions/splunk.discovery: []
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
  receivers/splunk.discovery: [recv/four, recv/five]
`)

	expected := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: [recv/one, recv/two, recv/three, recv/four, recv/five]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestDiscoveryReceiversOnly(t *testing.T) {
	in := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: []
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
  receivers/splunk.discovery: [recv/one, recv/two]
`)

	expected := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: [recv/one, recv/two]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestDiscoveryEmptyReceivers(t *testing.T) {
	in := confFromYaml(t, `service:
  pipelines:
    metrics:
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
  receivers/splunk.discovery: []
`)

	expected := confFromYaml(t, `service:
  pipelines:
    metrics:
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
`)

	require.NoError(t, Discovery{}.Convert(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func confFromYaml(t testing.TB, content string) *confmap.Conf {
	var conf map[string]any
	if err := yaml.Unmarshal([]byte(content), &conf); err != nil {
		t.Errorf("failed loading conf from yaml: %v", err)
	}
	return confmap.NewFromStringMap(conf)
}
