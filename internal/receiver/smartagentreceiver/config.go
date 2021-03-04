// Copyright 2021, OpenTelemetry Authors
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
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/config/configmodels"
	"gopkg.in/yaml.v2"
)

const defaultIntervalSeconds = 10

var errDimensionClientValue = fmt.Errorf("dimensionClients must be an array of compatible exporter names")
var errEventClientValue = fmt.Errorf("eventClients must be an array of compatible exporter names")

type Config struct {
	configmodels.ReceiverSettings `mapstructure:",squash"`
	// Generally an observer/receivercreator-set value via Endpoint.Target.
	// Will expand to MonitorCustomConfig Host and Port values if unset.
	Endpoint         string    `mapstructure:"endpoint"`
	DimensionClients *[]string `mapstructure:"dimensionclients"`
	EventClients     *[]string `mapstructure:"eventclients"`
	monitorConfig    config.MonitorCustomConfig
}

func (rCfg *Config) validate() error {
	if rCfg.monitorConfig == nil {
		return fmt.Errorf("you must supply a valid Smart Agent Monitor config")
	}

	monitorConfigCore := rCfg.monitorConfig.MonitorConfigCore()
	if monitorConfigCore.IntervalSeconds == 0 {
		monitorConfigCore.IntervalSeconds = defaultIntervalSeconds
	} else if monitorConfigCore.IntervalSeconds < 0 {
		return fmt.Errorf("intervalSeconds must be greater than 0s (%d provided)", monitorConfigCore.IntervalSeconds)
	}

	if err := validation.ValidateStruct(rCfg.monitorConfig); err != nil {
		return err
	}
	return validation.ValidateCustomConfig(rCfg.monitorConfig)
}

// mergeConfigs is used as a custom unmarshaller to dynamically create the desired Smart Agent monitor config
// from the provided receiver config content.
func mergeConfigs(componentViperSection *viper.Viper, intoCfg interface{}) error {
	// AllSettings() will include anything not already unmarshalled in the Config instance (*intoCfg).
	// This includes all Smart Agent monitor config settings that can be unmarshalled to their
	// respective custom monitor config types.
	allSettings := componentViperSection.AllSettings()
	monitorType, ok := allSettings["type"].(string)
	if !ok || monitorType == "" {
		return fmt.Errorf("you must specify a \"type\" for a smartagent receiver")
	}

	receiverCfg := intoCfg.(*Config)
	var endpoint interface{}
	if endpoint, ok = allSettings["endpoint"]; ok {
		receiverCfg.Endpoint = fmt.Sprintf("%s", endpoint)
		delete(allSettings, "endpoint")
	}

	// Config.DimensionClients and Config.EventClients should end up as *[]string, and we
	// need to arrive at that from allSetting's interface{} value.
	vas := viperAllSettings(allSettings)
	var err error
	receiverCfg.DimensionClients, err = vas.getPointerToStringSliceFromYaml("dimensionclients", errDimensionClientValue)
	if err != nil {
		return err
	}
	receiverCfg.EventClients, err = vas.getPointerToStringSliceFromYaml("eventclients", errEventClientValue)
	if err != nil {
		return err
	}

	// monitors.ConfigTemplates is a map that all monitors use to register their custom configs in the Smart Agent.
	// The values are always pointers to an actual custom config.
	var customMonitorConfig config.MonitorCustomConfig
	if customMonitorConfig, ok = monitors.ConfigTemplates[monitorType]; !ok {
		return fmt.Errorf("no known monitor type %q", monitorType)
	}
	monitorConfigType := reflect.TypeOf(customMonitorConfig).Elem()
	monitorConfig := reflect.New(monitorConfigType).Interface()

	// Viper is case insensitive and doesn't preserve a record of actual yaml map key cases from the provided config,
	// which is a problem when unmarshalling custom agent monitor configs.  Here we use a map of lowercase to supported
	// case tag key names and update the keys where applicable.
	yamlTags := yamlTagsFromStruct(monitorConfigType)
	for key, val := range allSettings {
		updatedKey := yamlTags[key]
		if updatedKey != "" {
			delete(allSettings, key)
			allSettings[updatedKey] = val
		}
	}

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

	err = setHostAndPortViaEndpoint(receiverCfg.Endpoint, monitorConfig)
	if err != nil {
		return err
	}

	receiverCfg.monitorConfig = monitorConfig.(config.MonitorCustomConfig)
	return nil
}

type viperAllSettings map[string]interface{}

func (allSettings viperAllSettings) getPointerToStringSliceFromYaml(key string, errToReturn error) (*[]string, error) {
	var slicePtr *[]string
	if value, ok := allSettings[key]; ok {
		if valueAsSlice, isSlice := value.([]interface{}); isSlice {
			var items []string
			for _, c := range valueAsSlice {
				if client, isString := c.(string); isString {
					items = append(items, client)
				} else {
					return nil, errToReturn
				}
			}
			slicePtr = &items
		} else {
			return nil, errToReturn
		}
		delete(allSettings, key)
	}
	return slicePtr, nil
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
		if fieldType.Kind() == reflect.Struct {
			otherFields := yamlTagsFromStruct(fieldType)
			for k, v := range otherFields {
				yamlTags[k] = v
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
