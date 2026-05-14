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

	"github.com/signalfx/defaults"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
)

var _ confmap.Unmarshaler = (*Config)(nil)

type Config struct {
	// Agent uses yaml, which mapstructure doesn't support.
	// Custom unmarshaller required for yaml and SFx defaults usage.
	saconfig.Config `mapstructure:"-,squash"`
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	allSettings := componentParser.ToStringMap()

	config, err := smartAgentConfigFromSettingsMap(allSettings)
	if err != nil {
		return err
	}

	if config.BundleDir == "" {
		config.BundleDir = cfg.Config.BundleDir
	}

	cfg.Config = *config
	return nil
}

func smartAgentConfigFromSettingsMap(settings map[string]any) (*saconfig.Config, error) {
	var config saconfig.Config

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

	return &config, nil
}
