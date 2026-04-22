// Copyright Splunk, Inc.
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

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
)

// InjectConfigSourceTelemetryExtension is a confmap.Converter that automatically injects
// the configsource_telemetry extension into the service config.
// This ensures the extension always starts with the service lifecycle so the MeterProvider is always injected into the TelemetryHook.
func InjectConfigSourceTelemetryExtension(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := in.ToStringMap()

	// Ensure extensions section exists and contains our extension
	extensions, _ := out["extensions"].(map[string]any)
	if extensions == nil {
		extensions = map[string]any{}
	}
	// Add the extension with an empty config if not already present
	if _, exists := extensions[configsourcetelemetryextension.TypeStr]; !exists {
		extensions[configsourcetelemetryextension.TypeStr] = map[string]any{}
	}
	out["extensions"] = extensions

	// Ensure service section exists
	service, _ := out["service"].(map[string]any)
	if service == nil {
		service = map[string]any{}
	}

	// Ensure service.extensions list exists and contains our extension
	serviceExtensions := toStringSlice(service["extensions"])
	if !containsExtension(serviceExtensions, configsourcetelemetryextension.TypeStr) {
		serviceExtensions = append(serviceExtensions, configsourcetelemetryextension.TypeStr)
	}
	service["extensions"] = serviceExtensions
	out["service"] = service

	*in = *confmap.NewFromStringMap(out)
	return nil
}

// containsExtension returns true if the extension name is already in the list.
func containsExtension(extensions []string, name string) bool {
	for _, e := range extensions {
		if e == name {
			return true
		}
	}
	return false
}

// toStringSlice converts an interface{} to []string safely.
func toStringSlice(v any) []string {
	if v == nil {
		return []string{}
	}
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}
