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
	"fmt"

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
)

// InjectConfigSourceTelemetryExtension is a confmap.Converter that automatically injects
// the configsource_telemetry extension into the service config.
func InjectConfigSourceTelemetryExtension(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := in.ToStringMap()

	service, serviceExtensions, err := getServiceExtensions(out)
	if err != nil {
		return err
	}

	extensions := map[string]any{}
	if raw, exists := out["extensions"]; exists && raw != nil {
		var ok bool
		if extensions, ok = raw.(map[string]any); !ok {
			return fmt.Errorf("extensions is of unexpected form (%T): %v", raw, raw)
		}
	}
	if _, exists := extensions[configsourcetelemetryextension.TypeStr]; !exists {
		extensions[configsourcetelemetryextension.TypeStr] = map[string]any{}
	}
	out["extensions"] = extensions

	// Append the extension to service.extensions, deduplicating with appendUnique.
	service["extensions"] = appendUnique(serviceExtensions, []any{configsourcetelemetryextension.TypeStr})
	out["service"] = service

	*in = *confmap.NewFromStringMap(out)
	return nil
}
