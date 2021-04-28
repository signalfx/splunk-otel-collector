// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/signalfx/defaults"
	_ "github.com/signalfx/signalfx-agent/pkg/core" // required to invoke monitor registration via init() calls
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"go.opentelemetry.io/collector/config"
	"gopkg.in/yaml.v2"
)

const defaultIntervalSeconds = 10

var _ config.CustomUnmarshable = (*Config)(nil)

var errDimensionClientValue = fmt.Errorf("dimensionClients must be an array of compatible exporter names")

type Config struct {
	monitorConfig           saconfig.MonitorCustomConfig
	config.ReceiverSettings `mapstructure:",squash"`
	// Generally an observer/receivercreator-set value via Endpoint.Target.
	// Will expand to MonitorCustomConfig Host and Port values if unset.
	Endpoint         string   `mapstructure:"endpoint"`
	DimensionClients []string `mapstructure:"dimensionclients"`
}

func (cfg *Config) validate() error {
	if cfg.monitorConfig == nil {
		return fmt.Errorf("you must supply a valid Smart Agent Monitor config")
	}

	monitorConfigCore := cfg.monitorConfig.MonitorConfigCore()
	if monitorConfigCore.IntervalSeconds == 0 {
		monitorConfigCore.IntervalSeconds = defaultIntervalSeconds
	} else if monitorConfigCore.IntervalSeconds < 0 {
		return fmt.Errorf("intervalSeconds must be greater than 0s (%d provided)", monitorConfigCore.IntervalSeconds)
	}

	if err := validation.ValidateStruct(cfg.monitorConfig); err != nil {
		return err
	}
	return validation.ValidateCustomConfig(cfg.monitorConfig)
}

// Unmarshal dynamically creates the desired Smart Agent monitor config
// from the provided receiver config content.
func (cfg *Config) Unmarshal(componentParser *config.Parser) error {
	// AllSettings() is the user provided config and intoCfg is the default Config instance we populate to
	// form the final desired version. To do so, we manually obtain all Config items, leaving only Smart Agent
	// monitor config settings to be unmarshalled to their respective custom monitor config types.
	allSettings := componentParser.Viper().AllSettings()
	monitorType, ok := allSettings["type"].(string)
	if !ok || monitorType == "" {
		return fmt.Errorf("you must specify a \"type\" for a smartagent receiver")
	}

	var endpoint interface{}
	if endpoint, ok = allSettings["endpoint"]; ok {
		cfg.Endpoint = fmt.Sprintf("%s", endpoint)
		delete(allSettings, "endpoint")
	}

	var err error
	cfg.DimensionClients, err = getStringSliceFromAllSettings(allSettings, "dimensionclients", errDimensionClientValue)
	if err != nil {
		return err
	}

	// monitors.ConfigTemplates is a map that all monitors use to register their custom configs in the Smart Agent.
	// The values are always pointers to an actual custom config.
	var customMonitorConfig saconfig.MonitorCustomConfig
	if customMonitorConfig, ok = monitors.ConfigTemplates[monitorType]; !ok {
		return fmt.Errorf("no known monitor type %q", monitorType)
	}
	monitorConfigType := reflect.TypeOf(customMonitorConfig).Elem()
	monitorConfig := reflect.New(monitorConfigType).Interface()

	// Viper is case insensitive and doesn't preserve a record of actual yaml map key cases from the provided config,
	// which is a problem when unmarshalling custom agent monitor configs.  Here we use a map of lowercase to supported
	// case tag key names and update the keys where applicable.
	yamlTags := yamlTagsFromStruct(monitorConfigType)
	recursivelyCapitalizeConfigKeys(allSettings, yamlTags)

	asBytes, err := yaml.Marshal(allSettings)
	if err != nil {
		return fmt.Errorf("failed constructing raw Smart Agent Monitor config block: %w", err)
	}

	err = yaml.UnmarshalStrict(asBytes, monitorConfig)
	if err != nil {
		return fmt.Errorf("failed creating Smart Agent Monitor custom config: %w", err)
	}

	err = defaults.Set(monitorConfig)
	if err != nil {
		return fmt.Errorf("failed setting Smart Agent Monitor config defaults: %w", err)
	}

	err = setHostAndPortViaEndpoint(cfg.Endpoint, monitorConfig)
	if err != nil {
		return err
	}

	cfg.monitorConfig = monitorConfig.(saconfig.MonitorCustomConfig)
	return nil
}

func recursivelyCapitalizeConfigKeys(settings map[string]interface{}, yamlTags map[string]string) {
	for key, val := range settings {
		updatedKey := yamlTags[key]
		if updatedKey != "" {
			delete(settings, key)
			settings[updatedKey] = val
			if m, ok := val.(map[string]interface{}); ok {
				recursivelyCapitalizeConfigKeys(m, yamlTags)
			}
		}
	}
}

func getStringSliceFromAllSettings(allSettings map[string]interface{}, key string, errToReturn error) ([]string, error) {
	var items []string
	if value, ok := allSettings[key]; ok {
		items = []string{}
		if valueAsSlice, isSlice := value.([]interface{}); isSlice {
			for _, c := range valueAsSlice {
				if client, isString := c.(string); isString {
					items = append(items, client)
				} else {
					return nil, errToReturn
				}
			}
		} else {
			return nil, errToReturn
		}
		delete(allSettings, key)
	}
	return items, nil
}

// Walks through a custom monitor config struct type, creating a map of
// lowercase to supported yaml struct tag name cases.
func yamlTagsFromStruct(s reflect.Type) map[string]string {
	yamlTags := map[string]string{}
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag := field.Tag
		yamlTag := strings.Split(tag.Get("yaml"), ",")[0]
		lowerTag := strings.ToLower(yamlTag)
		if yamlTag != lowerTag {
			yamlTags[lowerTag] = yamlTag
		}

		fieldType := field.Type
		switch fieldType.Kind() {
		case reflect.Struct:
			otherFields := yamlTagsFromStruct(fieldType)
			for k, v := range otherFields {
				yamlTags[k] = v
			}
		case reflect.Ptr:
			fieldTypeElem := fieldType.Elem()
			if fieldTypeElem.Kind() == reflect.Struct {
				otherFields := yamlTagsFromStruct(fieldTypeElem)
				for k, v := range otherFields {
					yamlTags[k] = v
				}
			}
		}
	}

	return yamlTags
}

// If using the receivercreator, observer-provided endpoints should be used to set
// the Host and Port fields of monitor config structs.  This can only be done by reflection without
// making type assertions over all possible monitor types.
func setHostAndPortViaEndpoint(endpoint string, monitorConfig interface{}) error {
	if endpoint == "" {
		return nil
	}

	var host string
	var port uint16
	splat := strings.Split(endpoint, ":")
	host = splat[0]
	if len(splat) == 2 {
		portStr := splat[1]
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return fmt.Errorf("cannot determine port via Endpoint: %w", err)
		}
		port = uint16(port64)
	}

	if host != "" {
		_, err := SetStructFieldIfZeroValue(monitorConfig, "Host", host)
		if err != nil {
			return fmt.Errorf("unable to set monitor Host field using Endpoint-derived value of %s: %w", host, err)
		}
	}

	if port != 0 {
		_, err := SetStructFieldIfZeroValue(monitorConfig, "Port", port)
		if err != nil {
			return fmt.Errorf("unable to set monitor Port field using Endpoint-derived value of %d: %w", port, err)
		}
	}

	return nil
}
