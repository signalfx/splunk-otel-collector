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

package scriptedinputsreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr   = "scripted_inputs"
	stability = component.StabilityLevelDevelopment
)

func NewFactory() receiver.Factory {
	return adapter.NewFactory(scriptedInputsReceiver{}, stability)
}

var _ adapter.LogReceiverType = (*scriptedInputsReceiver)(nil)

type scriptedInputsReceiver struct{}

func (f scriptedInputsReceiver) Type() component.Type {
	return typeStr
}

func (f scriptedInputsReceiver) CreateDefaultConfig() component.Config {
	return createDefaultConfig()
}

func (f scriptedInputsReceiver) BaseConfig(component.Config) adapter.BaseConfig {
	// unused by this component so just satisfy the interface with empty defaults used
	return adapter.BaseConfig{}
}

func (f scriptedInputsReceiver) InputConfig(cfg component.Config) operator.Config {
	return operator.NewConfig(cfg.(*Config))
}
