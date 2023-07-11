// Copyright Splunk, Inc.
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

package scriptedinputreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/scriptedinputsreceiver/procpipe"
)

const (
	typeStr   = "scripted_inputs"
	stability = component.StabilityLevelDevelopment
)

func NewFactory() receiver.Factory {
	return adapter.NewFactory(ReceiverType{}, stability)
}

// ReceiverType implements stanza.LogReceiverType
// to create a file tailing receiver
type ReceiverType struct{}

// Type is the receiver type
func (f ReceiverType) Type() component.Type {
	return typeStr
}

// CreateDefaultConfig creates a config with type and version
func (f ReceiverType) CreateDefaultConfig() component.Config {
	return &ScriptConfig{
		BaseConfig: adapter.BaseConfig{
			Operators: []operator.Config{},
		},
		InputConfig: *procpipe.NewConfig(),
	}
}

// BaseConfig gets the base config from config, for now
func (f ReceiverType) BaseConfig(cfg component.Config) adapter.BaseConfig {
	return cfg.(*ScriptConfig).BaseConfig
}

// ScriptConfig defines configuration for the filelog receiver
type ScriptConfig struct {
	InputConfig        procpipe.Config `mapstructure:",squash"`
	adapter.BaseConfig `mapstructure:",squash"`
}

// InputConfig unmarshals the input operator
func (f ReceiverType) InputConfig(cfg component.Config) operator.Config {
	return operator.NewConfig(&cfg.(*ScriptConfig).InputConfig)
}
