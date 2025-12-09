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
  telemetry:
    resource:
      my-resource: test
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
  telemetry:
    resource:
      my-resource: test
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
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

	require.NoError(t, SetupDiscovery(context.Background(), in))
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
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
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
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
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
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
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
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestContinuousDiscoveryNoEntitiesPipeline(t *testing.T) {
	in := confFromYaml(t, `service:
  pipelines:
    metrics:
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
    logs/untouched:
      receivers: [recv/six]
      processors: [proc/six]
      exporters: [exp/six]
  receivers/splunk.discovery: [discovery/host_observer]
`)

	expected := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: [discovery/host_observer]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    metrics/untouched:
      receivers: [recv/six, recv/seven, recv/eight]
      processors: [proc/six, proc/seven, proc/eight]
      exporters: [exp/six, exp/seven, exp/eight]
    logs/untouched:
      receivers: [recv/six]
      processors: [proc/six]
      exporters: [exp/six]
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestContinuousDiscoveryWithEntitiesPipeline(t *testing.T) {
	in := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: [recv/one]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    logs/entities:
      receivers: [recv/one, recv/two]
      processors: [proc/one, proc/two]
      exporters: [exp/one, exp/two]
    logs/untouched:
      receivers: [recv/three]
      processors: [proc/three]
      exporters: [exp/three]
  receivers/splunk.discovery: [discovery/one, discovery/two]
`)

	expected := confFromYaml(t, `service:
  pipelines:
    metrics:
      receivers: [recv/one, discovery/one, discovery/two]
      processors: [proc/one, proc/two, proc/three]
      exporters: [exp/one, exp/two, exp/three]
    logs/entities:
      receivers: [recv/one, recv/two, discovery/one, discovery/two]
      processors: [proc/one, proc/two]
      exporters: [exp/one, exp/two]
    logs/untouched:
      receivers: [recv/three]
      processors: [proc/three]
      exporters: [exp/three]
  telemetry:
    resource:
      splunk_autodiscovery: "true"
`)

	require.NoError(t, SetupDiscovery(context.Background(), in))
	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func confFromYaml(tb testing.TB, content string) *confmap.Conf {
	var conf map[string]any
	if err := yaml.Unmarshal([]byte(content), &conf); err != nil {
		tb.Errorf("failed loading conf from yaml: %v", err)
	}
	return confmap.NewFromStringMap(conf)
}
