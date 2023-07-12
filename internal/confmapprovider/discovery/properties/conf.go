// Copyright  Splunk, Inc.
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

package properties

import (
	"fmt"
	"sort"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/multierr"

	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/internal"
)

// File is loaded from properties.discovery.yaml
// and can consist of --set discovery property entries
// as well as a splunk.discovery: property mapping
type File struct {
	Mapping `mapstructure:"splunk.discovery"`
	Entries map[string]any `mapstructure:",remain"`
}

type Mapping struct {
	Receivers  map[string]Entry `mapstructure:"receivers"`
	Extensions map[string]Entry `mapstructure:"extensions"`
	Unknown    map[any]any      `mapstructure:",remain"`
}

type Entry struct {
	Config  *map[string]any `mapstructure:"config"`
	Enabled *bool           `mapstructure:"enabled"`
	Unknown map[any]any     `mapstructure:",remain"`
}

type propTuple struct {
	value any
	key   string
}

func LoadConf(raw map[string]any) (properties *confmap.Conf, warning, fatal error) {
	file := &File{}
	if err := confmap.NewFromStringMap(raw).Unmarshal(file); err != nil {
		return nil, nil, err
	}

	if len(file.Unknown) > 0 {
		var unknownMappings []string
		for unknown := range file.Unknown {
			unknownMappings = append(unknownMappings, fmt.Sprintf("%v", unknown))
		}
		// for testability
		sort.Strings(unknownMappings)
		for _, unknown := range unknownMappings {
			warning = multierr.Combine(warning, fmt.Errorf(`unknown discovery property mapping "splunk.discovery.%v"`, unknown))
		}
	}

	// all properties are loaded individually, so we maintain a slice of kv tuples
	var props []propTuple

	for _, entry := range []struct {
		entry map[string]Entry
		typ   string
	}{
		{typ: "extensions", entry: file.Extensions},
		{typ: "receivers", entry: file.Receivers},
	} {
		var unknownMappings []string
		for cid, ent := range entry.entry {
			prefix := fmt.Sprintf("splunk.discovery.%s.%s", entry.typ, cid)
			if ent.Enabled != nil {
				props = append(props, propTuple{key: fmt.Sprintf("%s.enabled", prefix), value: *ent.Enabled})
			}

			if ent.Config != nil {
				conf := confmap.NewFromStringMap(*ent.Config)
				for _, ck := range conf.AllKeys() {
					val := conf.Get(ck)
					props = append(props, propTuple{key: fmt.Sprintf("%s.config.%s", prefix, ck), value: val})
				}
			}

			for unknown := range ent.Unknown {
				unknownMappings = append(unknownMappings, fmt.Sprintf("%s.%v", prefix, unknown))
			}
		}

		sort.Strings(unknownMappings)
		for _, unknown := range unknownMappings {
			warning = multierr.Combine(warning, fmt.Errorf("unknown property %q", unknown))
		}
	}

	// --set property form has priority to splunk.discovery mappings (processed afterward and overwrites)
	entriesConf := confmap.NewFromStringMap(file.Entries)
	for _, k := range entriesConf.AllKeys() {
		props = append(props, propTuple{key: k, value: entriesConf.Get(k)})
	}

	confProperties := map[string]any{}

	for _, prop := range props {
		if p, err := NewProperty(prop.key, fmt.Sprintf("%v", prop.value)); err != nil {
			warning = multierr.Combine(warning, err)
		} else {
			if err = internal.MergeMaps(confProperties, p.ToStringMap()); err != nil {
				return nil, warning, fmt.Errorf("failed merging properties.discovery.yaml content: %w", err)
			}
		}
	}

	return confmap.NewFromStringMap(confProperties), warning, nil
}
