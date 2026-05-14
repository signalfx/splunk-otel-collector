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
	"bytes"
	"log"
	"testing"

	"go.opentelemetry.io/collector/confmap"

	"github.com/stretchr/testify/require"
)

func TestMigrateTelemetryResourceAttributes(t *testing.T) {
	tests := []struct {
		customAssertion  func(t *testing.T, in, expected *confmap.Conf)
		name             string
		input            string
		expected         string
		skipNilInput     bool
		useElementsMatch bool
	}{
		{
			name: "legacy_to_declarative",
			input: `service:
  telemetry:
    resource:
      service.name: my-collector
      host.name: collector-host
      deployment.environment: production
`,
			expected: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
        - name: deployment.environment
          value: production
`,
			useElementsMatch: true,
		},
		{
			name: "already_declarative",
			input: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
`,
			expected: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
`,
		},
		{
			name: "no_telemetry",
			input: `service:
  extensions: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one]
      exporters: [exp/one]
`,
			expected: `service:
  extensions: [ext/one, ext/two]
  pipelines:
    metrics:
      receivers: [recv/one]
      exporters: [exp/one]
`,
		},
		{
			name: "empty_resource",
			input: `service:
  telemetry:
    resource: {}
`,
			expected: `service:
  telemetry:
    resource: {}
`,
		},
		{
			name:         "nil_conf",
			skipNilInput: true,
		},
		{
			name: "no_service",
			input: `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
`,
			expected: `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
`,
		},
		{
			name: "mixed_values",
			input: `service:
  telemetry:
    resource:
      service.name: my-collector
      service.version: 1.0.0
      host.name: collector-host
      numeric.value: 42
      boolean.value: true
`,
			customAssertion: func(t *testing.T, in, _ *confmap.Conf) {
				inMap := in.ToStringMap()
				inAttrs := extractAttributes(t, inMap)

				require.Len(t, inAttrs, 5)

				actualAttrs := make(map[string]any, len(inAttrs))
				for _, attr := range inAttrs {
					attrMap, ok := attr.(map[string]any)
					require.True(t, ok)
					name, ok := attrMap["name"].(string)
					require.True(t, ok)
					actualAttrs[name] = attrMap["value"]
				}

				require.Equal(t, map[string]any{
					"service.name":    "my-collector",
					"service.version": "1.0.0",
					"host.name":       "collector-host",
					"numeric.value":   42,
					"boolean.value":   true,
				}, actualAttrs)
			},
		},
		{
			name: "v030_fields_not_migrated",
			input: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      schema_url: https://opentelemetry.io/schemas/1.6.1
      attributes_list: []
`,
			expected: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      schema_url: https://opentelemetry.io/schemas/1.6.1
      attributes_list: []
`,
		},
		{
			name: "legacy_and_v030_fields_mixed",
			input: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      schema_url: https://opentelemetry.io/schemas/1.6.1
      service.name: my-collector
      host.name: collector-host
`,
			expected: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      schema_url: https://opentelemetry.io/schemas/1.6.1
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
`,
			customAssertion: func(t *testing.T, in, expected *confmap.Conf) {
				inMap := in.ToStringMap()
				expectedMap := expected.ToStringMap()

				inResource := inMap["service"].(map[string]any)["telemetry"].(map[string]any)["resource"].(map[string]any)
				expectedResource := expectedMap["service"].(map[string]any)["telemetry"].(map[string]any)["resource"].(map[string]any)
				require.Equal(t, expectedResource["detectors"], inResource["detectors"])
				require.Equal(t, expectedResource["schema_url"], inResource["schema_url"])

				require.ElementsMatch(t,
					expectedResource["attributes"].([]any),
					inResource["attributes"].([]any),
				)
			},
		},
		{
			name: "name_attributes",
			input: `service:
  telemetry:
    resource:
      attributes: something
      service.name: my-collector
      host.name: collector-host
`,
			expected: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
`,
			useElementsMatch: true,
		},
		{
			name: "valid_attributes_untouched",
			input: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: prod-host
`,
			expected: `service:
  telemetry:
    resource:
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: prod-host
`,
		},
		{
			name: "scalar_attributes_field_warning",
			input: `service:
  telemetry:
    resource:
      attributes: somevalue
      host.name: collector-host
`,
			expected: `service:
  telemetry:
    resource:
      attributes:
        - name: host.name
          value: collector-host
`,
			useElementsMatch: true,
		},
		{
			name: "scalar_detectors_field_warning",
			input: `service:
  telemetry:
    resource:
      detectors: somevalue
      host.name: collector-host
`,
			customAssertion: func(t *testing.T, in, _ *confmap.Conf) {
				inMap := in.ToStringMap()
				resource := inMap["service"].(map[string]any)["telemetry"].(map[string]any)["resource"].(map[string]any)

				require.Equal(t, "somevalue", resource["detectors"])

				attrs, ok := resource["attributes"].([]any)
				require.True(t, ok, "attributes should have been created")
				require.Len(t, attrs, 1)

				attr := attrs[0].(map[string]any)
				require.Equal(t, "host.name", attr["name"])
				require.Equal(t, "collector-host", attr["value"])
			},
		},
		{
			name: "both_scalar_reserved_fields_with_legacy_attrs",
			input: `service:
  telemetry:
    resource:
      attributes: scalar1
      detectors: scalar2
      service.name: my-collector
      deployment.environment: production
`,
			customAssertion: func(t *testing.T, in, _ *confmap.Conf) {
				inMap := in.ToStringMap()
				resource := inMap["service"].(map[string]any)["telemetry"].(map[string]any)["resource"].(map[string]any)

				require.Equal(t, "scalar2", resource["detectors"])

				attrs, ok := resource["attributes"].([]any)
				require.True(t, ok, "attributes should be an array after migration")
				require.Len(t, attrs, 2)

				attrMap := make(map[string]any)
				for _, attr := range attrs {
					a := attr.(map[string]any)
					attrMap[a["name"].(string)] = a["value"]
				}
				require.Equal(t, "my-collector", attrMap["service.name"])
				require.Equal(t, "production", attrMap["deployment.environment"])
			},
		},
		{
			name: "valid_detectors_map_with_legacy_attrs",
			input: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      service.name: my-collector
      host.name: collector-host
`,
			expected: `service:
  telemetry:
    resource:
      detectors:
        attributes:
          included: [system, env]
      attributes:
        - name: service.name
          value: my-collector
        - name: host.name
          value: collector-host
`,
			useElementsMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipNilInput {
				require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), nil))
				return
			}

			in := confFromYaml(t, tt.input)
			var expected *confmap.Conf
			if tt.expected != "" {
				expected = confFromYaml(t, tt.expected)
			}

			require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))

			switch {
			case tt.customAssertion != nil:
				tt.customAssertion(t, in, expected)
			case tt.useElementsMatch:
				inMap := in.ToStringMap()
				expectedMap := expected.ToStringMap()
				inAttrs := extractAttributes(t, inMap)
				expectedAttrs := extractAttributes(t, expectedMap)
				require.ElementsMatch(t, expectedAttrs, inAttrs)
			default:
				require.Equal(t, expected.ToStringMap(), in.ToStringMap())
			}
		})
	}
}

func TestMigrateTelemetryResourceAttributesWarnings(t *testing.T) {
	tests := []struct {
		input           string
		name            string
		expectedWarning string
	}{
		{
			name: "scalar_attributes_warns",
			input: `service:
  telemetry:
    resource:
      attributes: somevalue
      host.name: collector-host
`,
			expectedWarning: "Found 'attributes' field with non-list value",
		},
		{
			name: "scalar_detectors_warns",
			input: `service:
  telemetry:
    resource:
      detectors: somevalue
      service.name: my-service
`,
			expectedWarning: "Found 'detectors' field with non-map value",
		},
		{
			name: "both_scalars_warn_about_attributes",
			input: `service:
  telemetry:
    resource:
      attributes: value1
      detectors: value2
      host.name: test-host
`,
			expectedWarning: "Found 'attributes' field with non-list value",
		},
		{
			name: "both_scalars_warn_about_detectors",
			input: `service:
  telemetry:
    resource:
      attributes: value1
      detectors: value2
      host.name: test-host
`,
			expectedWarning: "Found 'detectors' field with non-map value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			oldWriter := log.Writer()
			log.SetOutput(&logOutput)
			defer log.SetOutput(oldWriter)

			in := confFromYaml(t, tt.input)
			require.NoError(t, MigrateTelemetryResourceAttributes(t.Context(), in))

			logStr := logOutput.String()
			require.Contains(t, logStr, tt.expectedWarning, "Expected warning message not found in logs")
		})
	}
}

func TestAddDeclarativeTelemetryResourceAttribute(t *testing.T) {
	tests := []struct {
		attrValue any
		service   map[string]any
		validate  func(t *testing.T, service map[string]any)
		name      string
		attrName  string
	}{
		{
			name:      "new_attribute",
			service:   map[string]any{},
			attrName:  "test.attribute",
			attrValue: "test-value",
			validate: func(t *testing.T, service map[string]any) {
				telemetry, ok := service["telemetry"].(map[string]any)
				require.True(t, ok)

				resource, ok := telemetry["resource"].(map[string]any)
				require.True(t, ok)

				attributes, ok := resource["attributes"].([]any)
				require.True(t, ok)
				require.Len(t, attributes, 1)

				attr, ok := attributes[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "test.attribute", attr["name"])
				require.Equal(t, "test-value", attr["value"])
			},
		},
		{
			name: "update_existing",
			service: map[string]any{
				"telemetry": map[string]any{
					"resource": map[string]any{
						"attributes": []any{
							map[string]any{
								"name":  "test.attribute",
								"value": "old-value",
							},
						},
					},
				},
			},
			attrName:  "test.attribute",
			attrValue: "new-value",
			validate: func(t *testing.T, service map[string]any) {
				telemetry := service["telemetry"].(map[string]any)
				resource := telemetry["resource"].(map[string]any)
				attributes := resource["attributes"].([]any)

				require.Len(t, attributes, 1)
				attr := attributes[0].(map[string]any)
				require.Equal(t, "test.attribute", attr["name"])
				require.Equal(t, "new-value", attr["value"])
			},
		},
		{
			name: "multiple_attributes",
			service: map[string]any{
				"telemetry": map[string]any{
					"resource": map[string]any{
						"attributes": []any{
							map[string]any{
								"name":  "attribute.one",
								"value": "value-one",
							},
							map[string]any{
								"name":  "attribute.two",
								"value": "value-two",
							},
						},
					},
				},
			},
			attrName:  "attribute.three",
			attrValue: "value-three",
			validate: func(t *testing.T, service map[string]any) {
				telemetry := service["telemetry"].(map[string]any)
				resource := telemetry["resource"].(map[string]any)
				attributes := resource["attributes"].([]any)

				require.Len(t, attributes, 3)

				found := false
				for _, attr := range attributes {
					attrMap := attr.(map[string]any)
					if attrMap["name"] == "attribute.three" {
						require.Equal(t, "value-three", attrMap["value"])
						found = true
						break
					}
				}
				require.True(t, found, "attribute.three should be present")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddDeclarativeTelemetryResourceAttribute(tt.service, tt.attrName, tt.attrValue)
			tt.validate(t, tt.service)
		})
	}
}

func TestAddDeclarativeTelemetryResourceAttribute_MalformedData(t *testing.T) {
	tests := []struct {
		attrValue any
		service   map[string]any
		validate  func(t *testing.T, service map[string]any)
		name      string
		attrName  string
	}{
		{
			name: "malformed_telemetry",
			service: map[string]any{
				"telemetry": "invalid-string-not-map",
			},
			attrName:  "test.attribute",
			attrValue: "test-value",
			validate: func(t *testing.T, service map[string]any) {
				require.Equal(t, "invalid-string-not-map", service["telemetry"])
			},
		},
		{
			name: "malformed_resource",
			service: map[string]any{
				"telemetry": map[string]any{
					"resource": "invalid-string-not-map",
				},
			},
			attrName:  "test.attribute",
			attrValue: "test-value",
			validate: func(t *testing.T, service map[string]any) {
				telemetry := service["telemetry"].(map[string]any)
				require.Equal(t, "invalid-string-not-map", telemetry["resource"])
			},
		},
		{
			name: "malformed_attributes",
			service: map[string]any{
				"telemetry": map[string]any{
					"resource": map[string]any{
						"attributes": "invalid-string-not-array",
					},
				},
			},
			attrName:  "test.attribute",
			attrValue: "test-value",
			validate: func(t *testing.T, service map[string]any) {
				telemetry := service["telemetry"].(map[string]any)
				resource := telemetry["resource"].(map[string]any)
				require.Equal(t, "invalid-string-not-array", resource["attributes"])
			},
		},
		{
			name: "malformed_attribute_elements",
			service: map[string]any{
				"telemetry": map[string]any{
					"resource": map[string]any{
						"attributes": []any{
							"invalid-string",
							map[string]any{
								"name":  "valid.attribute",
								"value": "valid-value",
							},
							123,
						},
					},
				},
			},
			attrName:  "new.attribute",
			attrValue: "new-value",
			validate: func(t *testing.T, service map[string]any) {
				telemetry := service["telemetry"].(map[string]any)
				resource := telemetry["resource"].(map[string]any)
				attributes := resource["attributes"].([]any)

				require.Len(t, attributes, 4)

				found := false
				for _, attr := range attributes {
					if attrMap, ok := attr.(map[string]any); ok {
						if attrMap["name"] == "new.attribute" {
							require.Equal(t, "new-value", attrMap["value"])
							found = true
							break
						}
					}
				}
				require.True(t, found, "new.attribute should be present")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				AddDeclarativeTelemetryResourceAttribute(tt.service, tt.attrName, tt.attrValue)
			})
			tt.validate(t, tt.service)
		})
	}
}

func extractAttributes(t *testing.T, configMap map[string]any) []any {
	service, ok := configMap["service"].(map[string]any)
	require.True(t, ok)

	telemetry, ok := service["telemetry"].(map[string]any)
	require.True(t, ok)

	resource, ok := telemetry["resource"].(map[string]any)
	require.True(t, ok)

	attributes, ok := resource["attributes"].([]any)
	require.True(t, ok)

	return attributes
}
