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
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/collector/confmap"
)

var validResourceV030Fields = map[string]struct{}{
	"attributes":      {},
	"attributes_list": {},
	"detectors":       {},
	"schema_url":      {},
}

func MigrateTelemetryResourceAttributes(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return nil
	}

	out := cfgMap.ToStringMap()

	service, _, err := getServiceExtensions(out)
	if err != nil {
		return err
	}

	var telemetry map[string]any
	if tel, hasTelemetry := service["telemetry"]; hasTelemetry && tel != nil {
		var ok bool
		if telemetry, ok = tel.(map[string]any); !ok {
			return fmt.Errorf("service.telemetry is of unexpected form (%T): %v", tel, tel)
		}
	} else {
		return nil
	}

	var resource map[string]any
	if res, hasResource := telemetry["resource"]; hasResource && res != nil {
		var ok bool
		if resource, ok = res.(map[string]any); !ok {
			return fmt.Errorf("service.telemetry.resource is of unexpected form (%T): %v", res, res)
		}
	} else {
		return nil
	}

	// Check if already using the new declarative format
	if attrs, hasAttributes := resource["attributes"]; hasAttributes {
		if _, ok := attrs.([]any); ok {
			return nil
		}
		log.Printf("WARNING: Found 'attributes' field with non-list value (%T) in service.telemetry.resource. This will be overwritten during migration. If this was intended as a resource attribute, please rename it.", attrs)
	}

	if detectors, hasDetectors := resource["detectors"]; hasDetectors {
		if _, ok := detectors.(map[string]any); !ok {
			log.Printf("WARNING: Found 'detectors' field with non-map value (%T) in service.telemetry.resource. This will not be migrated as an attribute. If this was intended as a resource attribute, please rename it.", detectors)
		}
	}

	var legacyKeys []string
	for key := range resource {
		if _, isKnownField := validResourceV030Fields[key]; !isKnownField {
			legacyKeys = append(legacyKeys, key)
		}
	}

	if len(legacyKeys) == 0 {
		return nil
	}

	attributes := make([]any, 0, len(legacyKeys))
	for _, name := range legacyKeys {
		attributes = append(attributes, map[string]any{
			"name":  name,
			"value": resource[name],
		})
		delete(resource, name)
	}

	resource["attributes"] = attributes
	log.Printf("WARNING: Deprecated service.telemetry.resource configuration format (v0.2.0) detected; auto-migrating to v0.3.0. Please update your configuration.")
	*cfgMap = *confmap.NewFromStringMap(out)
	return nil
}

func AddDeclarativeTelemetryResourceAttribute(service map[string]any, name string, value any) {
	telemetry := map[string]any{}
	if tel, ok := service["telemetry"]; ok && tel != nil {
		var telOk bool
		telemetry, telOk = tel.(map[string]any)
		if !telOk {
			log.Printf("WARNING: service.telemetry is of unexpected form (%T): %v. Skipping resource attribute", tel, tel)
			return
		}
	}
	service["telemetry"] = telemetry

	resource := map[string]any{}
	if res, ok := telemetry["resource"]; ok && res != nil {
		var resOk bool
		resource, resOk = res.(map[string]any)
		if !resOk {
			log.Printf("WARNING: service.telemetry.resource is of unexpected form (%T): %v. Skipping resource attribute", res, res)
			return
		}
	}
	telemetry["resource"] = resource

	var attributes []any
	if attrs, ok := resource["attributes"]; ok && attrs != nil {
		var attrsOk bool
		attributes, attrsOk = attrs.([]any)
		if !attrsOk {
			log.Printf("WARNING: service.telemetry.resource.attributes is of unexpected form (%T): %v. Skipping resource attribute", attrs, attrs)
			return
		}
	}

	found := false
	for _, attr := range attributes {
		if attrMap, ok := attr.(map[string]any); ok {
			if attrMap["name"] == name {
				attrMap["value"] = value
				found = true
				log.Printf("INFO: Updated existing resource attribute %q => %q", name, value)
				break
			}
		}
	}

	if !found {
		attributes = append(attributes, map[string]any{
			"name":  name,
			"value": value,
		})
	}

	resource["attributes"] = attributes
}
