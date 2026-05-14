// Copyright Splunk, Inc.
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

package configconverter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationAndDiscoveryIntegration_LegacyFormat(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions: [ext/one]
  extensions/splunk.discovery: [ext/two]
  pipelines:
    metrics:
      receivers: [recv/one]
      processors: [proc/one]
      exporters: [exp/one]
  receivers/splunk.discovery: [recv/two]
  telemetry:
    resource:
      service.name: my-collector
      host.name: prod-host
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two]
      processors: [proc/one]
      exporters: [exp/one]
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: prod-host
        - name: splunk_autodiscovery
          value: "true"
`)

	require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))
	require.NoError(t, SetupDiscovery(t.Context(), in))

	inMap := in.ToStringMap()
	expectedMap := expected.ToStringMap()

	inAttrs := extractAttributes(t, inMap)
	expectedAttrs := extractAttributes(t, expectedMap)

	require.ElementsMatch(t, expectedAttrs, inAttrs)

	inService := inMap["service"].(map[string]any)
	expectedService := expectedMap["service"].(map[string]any)
	inTelemetry := inService["telemetry"].(map[string]any)
	expectedTelemetry := expectedService["telemetry"].(map[string]any)
	inResource := inTelemetry["resource"].(map[string]any)
	expectedResource := expectedTelemetry["resource"].(map[string]any)
	delete(inResource, "attributes")
	delete(expectedResource, "attributes")

	// Compare full normalized config
	require.Equal(t, expectedMap, inMap)
}

func TestMigrationAndDiscoveryIntegration_DeclarativeFormat(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions: [ext/one]
  extensions/splunk.discovery: [ext/two]
  pipelines:
    metrics:
      receivers: [recv/one]
      processors: [proc/one]
      exporters: [exp/one]
  receivers/splunk.discovery: [recv/two]
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: prod-host
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two]
      processors: [proc/one]
      exporters: [exp/one]
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: prod-host
        - name: splunk_autodiscovery
          value: "true"
`)

	require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))
	require.NoError(t, SetupDiscovery(t.Context(), in))

	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestMigrationAndDiscoveryIntegration_NoTelemetry(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions/splunk.discovery: [ext/one]
  pipelines:
    metrics:
      receivers: [recv/one]
      processors: [proc/one]
      exporters: [exp/one]
  receivers/splunk.discovery: [recv/two]
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one]
  pipelines:
    metrics:
      receivers: [recv/one, recv/two]
      processors: [proc/one]
      exporters: [exp/one]
  telemetry:
    resource:
      attributes:
        - name: splunk_autodiscovery
          value: "true"
`)

	require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))
	require.NoError(t, SetupDiscovery(t.Context(), in))

	require.Equal(t, expected.ToStringMap(), in.ToStringMap())
}

func TestMigrationAndDiscoveryIntegration_UpdateExistingAttribute(t *testing.T) {
	in := confFromYaml(t, `service:
  extensions/splunk.discovery: [ext/one]
  receivers/splunk.discovery: [recv/one]
  pipelines:
    metrics:
      receivers: []
      exporters: [exp/one]
  telemetry:
    resource:
      splunk_autodiscovery: "false"
      service.name: my-service
`)

	expected := confFromYaml(t, `service:
  extensions: [ext/one]
  pipelines:
    metrics:
      receivers: [recv/one]
      exporters: [exp/one]
  telemetry:
    resource:
      attributes:
        - name: splunk_autodiscovery
          value: "true"
        - name: service.name
          value: my-service
`)

	require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))
	require.NoError(t, SetupDiscovery(t.Context(), in))

	inMap := in.ToStringMap()
	expectedMap := expected.ToStringMap()

	inAttrs := extractAttributes(t, inMap)
	expectedAttrs := extractAttributes(t, expectedMap)

	require.ElementsMatch(t, expectedAttrs, inAttrs)

	inService := inMap["service"].(map[string]any)
	expectedService := expectedMap["service"].(map[string]any)
	inTelemetry := inService["telemetry"].(map[string]any)
	expectedTelemetry := expectedService["telemetry"].(map[string]any)
	inResource := inTelemetry["resource"].(map[string]any)
	expectedResource := expectedTelemetry["resource"].(map[string]any)
	delete(inResource, "attributes")
	delete(expectedResource, "attributes")

	require.Equal(t, expectedMap, inMap)
}
