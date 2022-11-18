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

package configprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	configSourcesKey = "config_sources"
)

// Load reads the configuration for ConfigSource objects from the given parser and returns a map
// from the full name of config sources to the respective ConfigSettings.
func Load(ctx context.Context, v *confmap.Conf, factories Factories) (map[string]Source, error) {
	processedParser, err := processParser(ctx, v)
	if err != nil {
		return nil, err
	}

	cfgSrcSettings, err := loadSettings(cast.ToStringMap(processedParser.Get(configSourcesKey)), factories)
	if err != nil {
		return nil, err
	}

	return cfgSrcSettings, nil
}

// processParser prepares a confmap.Conf to be used to load config source settings.
func processParser(ctx context.Context, v *confmap.Conf) (*confmap.Conf, error) {
	processedParser := map[string]any{}
	for _, key := range v.AllKeys() {
		if !strings.HasPrefix(key, configSourcesKey) {
			// In Load we only care about config sources, ignore everything else.
			continue
		}

		value, _, err := parseConfigValue(ctx, make(map[string]ConfigSource), v.Get(key), nil)
		if err != nil {
			return nil, err
		}
		processedParser[key] = value
	}

	return confmap.NewFromStringMap(processedParser), nil
}

func loadSettings(css map[string]any, factories Factories) (map[string]Source, error) {
	// Prepare resulting map.
	cfgSrcToSettings := make(map[string]Source)

	// Iterate over extensions and create a config for each.
	for key, value := range css {
		settingsParser := confmap.NewFromStringMap(cast.ToStringMap(value))

		// Decode the key into type and fullName components.
		componentID := component.ID{}
		if err := componentID.UnmarshalText([]byte(key)); err != nil {
			return nil, fmt.Errorf("invalid %s type and name key %q: %w", configSourcesKey, key, err)
		}

		// Find the factory based on "type" that we read from config source.
		factory := factories[componentID.Type()]
		if factory == nil {
			return nil, fmt.Errorf("unknown %s type %q for %q", configSourcesKey, componentID.Type(), componentID)
		}

		// Create the default config.
		cfgSrcSettings := factory.CreateDefaultConfig()
		cfgSrcSettings.SetIDName(componentID.Name())

		// Now that the default settings struct is created we can Unmarshal into it
		// and it will apply user-defined config on top of the default.
		if err := settingsParser.Unmarshal(&cfgSrcSettings, confmap.WithErrorUnused()); err != nil {
			return nil, fmt.Errorf("error reading %s configuration for %q: %w", configSourcesKey, componentID, err)
		}

		fullName := componentID.String()
		if cfgSrcToSettings[fullName] != nil {
			return nil, fmt.Errorf("duplicate %s name %s", configSourcesKey, fullName)
		}

		cfgSrcToSettings[fullName] = cfgSrcSettings
	}

	return cfgSrcToSettings, nil
}
