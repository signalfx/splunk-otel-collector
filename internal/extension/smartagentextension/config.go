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

package smartagentextension

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/signalfx/defaults"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/config/configmodels"
	"gopkg.in/yaml.v2"
)

// SmartAgentConfigProvider exposes config fields to other packages.
// This is needed since fields such as fields such as bundleDir and
// collectdConfig are mapped to camel case fields in the config and
// hence are not exposed.
type SmartAgentConfigProvider interface {
	BundleDir() string
	CollectdConfig() config.CollectdConfig
}

var _ SmartAgentConfigProvider = (*Config)(nil)

type Config struct {
	configmodels.ExtensionSettings `mapstructure:",squash"`
	bundleDir                      string
	collectdConfig                 config.CollectdConfig
}

func (c Config) BundleDir() string {
	return c.bundleDir
}

func (c Config) CollectdConfig() config.CollectdConfig {
	return c.collectdConfig
}

func customUnmarshaller(componentViperSection *viper.Viper, intoCfg interface{}) error {
	allSettings := componentViperSection.AllSettings()
	extensionCfg := intoCfg.(*Config)

	if bundleDir, ok := allSettings["bundledir"]; ok {
		extensionCfg.bundleDir = fmt.Sprintf("%s", bundleDir)
		delete(allSettings, "bundledir")
	}

	var collectdSettings map[string]interface{}
	var ok bool
	if collectdSettings, ok = allSettings["collectd"].(map[string]interface{}); !ok {
		// Nothing to do if user specified collectd settings do not exist.
		// Defaults will be picked up.
		return nil
	}

	var collectdConfig config.CollectdConfig
	yamlTags := yamlTagsFromStruct(reflect.TypeOf(collectdConfig))

	for key, val := range collectdSettings {
		updatedKey := yamlTags[key]
		if updatedKey != "" {
			delete(collectdSettings, key)
			collectdSettings[updatedKey] = val
		}
	}

	asBytes, err := yaml.Marshal(collectdSettings)
	if err != nil {
		return fmt.Errorf("failed constructing raw collectd config block: %w", err)
	}

	err = yaml.UnmarshalStrict(asBytes, &collectdConfig)
	if err != nil {
		return fmt.Errorf("failed creating collectd config: %w", err)
	}

	err = defaults.Set(&collectdConfig)
	if err != nil {
		return fmt.Errorf("failed setting collectd config defaults: %w", err)
	}

	extensionCfg.collectdConfig = collectdConfig
	return nil
}

// Copied from smartagent receiver.
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
