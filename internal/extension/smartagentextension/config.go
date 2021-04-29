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

package smartagentextension

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/signalfx/defaults"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"go.opentelemetry.io/collector/config"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/utils"
)

// SmartAgentConfigProvider exposes global saconfig.Config to other components
type SmartAgentConfigProvider interface {
	SmartAgentConfig() *saconfig.Config
}

var _ SmartAgentConfigProvider = (*Config)(nil)
var _ config.CustomUnmarshable = (*Config)(nil)

type Config struct {
	config.ExtensionSettings `mapstructure:",squash"`
	// Agent uses yaml, which mapstructure doesn't support.
	// Custom unmarshaller required for yaml and SFx defaults usage.
	saconfig.Config `mapstructure:"-,squash"`
}

func (cfg Config) SmartAgentConfig() *saconfig.Config {
	return &cfg.Config
}

func (cfg *Config) Unmarshal(componentParser *config.Parser) error {
	allSettings := componentParser.Viper().AllSettings()

	configDirSet := false
	if collectd, ok := allSettings["collectd"]; ok {
		if collectdBlock, ok := collectd.(map[string]interface{}); ok {
			if _, ok := collectdBlock["configdir"]; ok {
				configDirSet = true
			}
		}
	}

	config, err := smartAgentConfigFromSettingsMap(allSettings)
	if err != nil {
		return err
	}

	if config.BundleDir == "" {
		config.BundleDir = cfg.Config.BundleDir
	}
	config.Collectd.BundleDir = config.BundleDir

	if !configDirSet {
		config.Collectd.ConfigDir = filepath.Join(config.Collectd.BundleDir, "run", "collectd")
	}

	cfg.Config = *config
	return nil
}

func smartAgentConfigFromSettingsMap(settings map[string]interface{}) (*saconfig.Config, error) {
	var config saconfig.Config
	utils.RespectYamlTagsInAllSettings(reflect.TypeOf(config), settings)

	var collectdSettings map[string]interface{}
	var ok bool
	if collectdSettings, ok = settings["collectd"].(map[string]interface{}); !ok {
		collectdSettings = map[string]interface{}{}
	}

	var collectdConfig saconfig.CollectdConfig
	utils.RespectYamlTagsInAllSettings(reflect.TypeOf(collectdConfig), collectdSettings)

	settings["collectd"] = collectdSettings

	asBytes, err := yaml.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed constructing raw Smart Agent config: %w", err)
	}

	err = yaml.UnmarshalStrict(asBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed creating Smart Agent config: %w", err)
	}

	err = defaults.Set(&config)
	if err != nil {
		return nil, fmt.Errorf("failed setting config defaults: %w", err)
	}

	// The default on CollectdConfig is 0, use the default if this is the case.
	if config.Collectd.IntervalSeconds == 0 {
		config.Collectd.IntervalSeconds = defaultIntervalSeconds
	}

	config.Collectd.BundleDir = config.BundleDir
	return &config, nil
}
