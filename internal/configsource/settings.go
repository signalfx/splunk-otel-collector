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

package configsource

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

// Settings is the configuration of a config source. Specific config sources must implement this
// interface and will typically embed SourceSettings struct or a struct that extends it.
type Settings interface {
	// ID returns the ID of the component that this configuration belongs to.
	ID() component.ID
	// SetIDName updates the name part of the ID for the component that this configuration belongs to.
	SetIDName(idName string)
}

// SourceSettings defines common settings of a Settings configuration.
// Specific config sources can embed this struct and extend it with more fields if needed.
// When embedded it must be with `mapstructure:",squash"` tag.
type SourceSettings struct {
	id component.ID `mapstructure:"-"`
}

// ID returns the ID of the component that this configuration belongs to.
func (s *SourceSettings) ID() component.ID {
	return s.id
}

// SetIDName updates the name part of the ID for the component that this configuration belongs to.
func (s *SourceSettings) SetIDName(idName string) {
	s.id = component.MustNewIDWithName(s.id.Type().String(), idName)
}

// NewSourceSettings return a new config.SourceSettings struct with the given ComponentID.
func NewSourceSettings(id component.ID) SourceSettings {
	return SourceSettings{id}
}

// SettingsFromConf reads the configuration for ConfigSource objects from the given confmap.Conf
// and returns a map of Settings and the remaining Conf without the `config_sources` mapping
func SettingsFromConf(ctx context.Context, conf *confmap.Conf, factories Factories, confmapProviders map[string]confmap.Provider) (map[string]Settings, *confmap.Conf, error) {
	settings, err := sourceSettings(ctx, conf, factories, confmapProviders)
	if err != nil {
		return nil, nil, fmt.Errorf("failed resolving config sources settings: %w", err)
	}

	splitMap := map[string]any{}
	for _, k := range conf.AllKeys() {
		if !strings.HasPrefix(k, configSourcesKey) {
			splitMap[k] = conf.Get(k)
		}
	}

	return settings, confmap.NewFromStringMap(splitMap), nil
}

func sourceSettings(ctx context.Context, v *confmap.Conf, factories Factories, confmapProviders map[string]confmap.Provider) (map[string]Settings, error) {
	splitMap := map[string]any{}
	for _, key := range v.AllKeys() {
		if strings.HasPrefix(key, configSourcesKey) {
			value, _, err := resolveConfigValue(ctx, make(map[string]ConfigSource), confmapProviders, v.Get(key), nil)
			if err != nil {
				return nil, err
			}
			splitMap[key] = value
		}
	}
	settingsMap, err := confmap.NewFromStringMap(splitMap).Sub(configSourcesKey)
	if err != nil {
		return nil, fmt.Errorf("invalid settings sub map content: %w", err)
	}

	return loadSettings(settingsMap.ToStringMap(), factories)
}

func loadSettings(settingsMap map[string]any, factories Factories) (map[string]Settings, error) {
	settings := make(map[string]Settings)

	for key, value := range settingsMap {
		settingsValue := confmap.NewFromStringMap(cast.ToStringMap(value))

		// Decode the key into type and fullName components.
		componentID := component.ID{}
		if err := componentID.UnmarshalText([]byte(key)); err != nil {
			return nil, fmt.Errorf("invalid %s type and name key %q: %w", configSourcesKey, key, err)
		}

		// Find the factory based on "type" that we read from config source.
		factory, ok := factories[componentID.Type()]
		if !ok || factory == nil {
			return nil, fmt.Errorf("unknown %s type %q for %q", configSourcesKey, componentID.Type(), componentID)
		}

		cfgSrcSettings := factory.CreateDefaultConfig()
		cfgSrcSettings.SetIDName(componentID.Name())

		// Now that the default settings struct is created we can Unmarshal into it
		// and it will apply user-defined config on top of the default.
		if err := settingsValue.Unmarshal(&cfgSrcSettings); err != nil {
			return nil, fmt.Errorf("error reading %s configuration for %q: %w", configSourcesKey, componentID, err)
		}

		fullName := componentID.String()
		if _, ok = settings[fullName]; ok {
			return nil, fmt.Errorf("duplicate %s name %s", configSourcesKey, fullName)
		}

		settings[fullName] = cfgSrcSettings
	}

	return settings, nil
}
