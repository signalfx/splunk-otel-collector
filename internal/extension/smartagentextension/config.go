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
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/config"
	"gopkg.in/yaml.v2"
)

// SmartAgentConfigProvider exposes config fields to other packages.
// This is needed since fields such  as bundleDir and collectdConfig
// are mapped to camel case fields in the config and hence are not exposed.
type SmartAgentConfigProvider interface {
	BundleDir() string
	CollectdConfig() *saconfig.CollectdConfig
}

var _ SmartAgentConfigProvider = (*Config)(nil)

type Config struct {
	config.ExtensionSettings `mapstructure:",squash"`
	bundleDir                string
	collectdConfig           saconfig.CollectdConfig
}

func (c Config) BundleDir() string {
	return c.bundleDir
}

func (c Config) CollectdConfig() *saconfig.CollectdConfig {
	return &c.collectdConfig
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
		// We must set the BundleDir field on the resulting CollectdConfig
		// so we use an empty instance.  Defaults will be picked up.
		collectdSettings = map[string]interface{}{}
	}

	var collectdConfig saconfig.CollectdConfig
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

	// The default on CollectdConfig is 0, use the default if this is the case.
	if collectdConfig.IntervalSeconds == 0 {
		collectdConfig.IntervalSeconds = defaultIntervalSeconds
	}

	collectdConfig.BundleDir = extensionCfg.bundleDir
	extensionCfg.collectdConfig = collectdConfig

	return nil
}

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
	}

	return yamlTags
}
