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

// Taken from https://github.com/open-telemetry/opentelemetry-collector/blob/v0.66.0/confmap/converter/overwritepropertiesconverter/properties.go
// to prevent breaking changes.
// "Deprecated: [v0.63.0] this converter will not be supported anymore because of dot separation limitation.
// See https://github.com/open-telemetry/opentelemetry-collector/issues/6294 for more details."
package configconverter

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/magiconair/properties"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
)

type converter struct {
	properties []string
}

func NewOverwritePropertiesConverter(properties []string) confmap.Converter {
	return &converter{properties: properties}
}

func (c *converter) Convert(_ context.Context, conf *confmap.Conf) error {
	if len(c.properties) == 0 {
		return nil
	}

	b := &bytes.Buffer{}
	for _, property := range c.properties {
		property = strings.TrimSpace(property)
		b.WriteString(property)
		b.WriteString("\n")
	}

	var props *properties.Properties
	var err error
	if props, err = properties.Load(b.Bytes(), properties.UTF8); err != nil {
		return err
	}

	// Create a map manually instead of using properties.Map() to not expand the env vars.
	parsed := make(map[string]interface{}, props.Len())
	for _, key := range props.Keys() {
		var value any
		if raw, ok := props.Get(key); ok {
			if err = yaml.Unmarshal([]byte(raw), &value); err != nil {
				return fmt.Errorf("error unmarshalling %q value: %w", key, err)
			}
		}
		parsed[key] = value
	}
	prop := maps.Unflatten(parsed, ".")
	return conf.Merge(confmap.NewFromStringMap(prop))
}
