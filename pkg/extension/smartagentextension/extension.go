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
	"context"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// SmartAgentConfigProvider exposes global saconfig.Config to other components
type SmartAgentConfigProvider interface {
	SmartAgentConfig() *saconfig.Config
}

type smartAgentConfigExtension struct {
	saCfg *saconfig.Config
}

var _ SmartAgentConfigProvider = (*smartAgentConfigExtension)(nil)

func (sae *smartAgentConfigExtension) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (sae *smartAgentConfigExtension) Shutdown(_ context.Context) error {
	return nil
}

func (sae *smartAgentConfigExtension) SmartAgentConfig() *saconfig.Config {
	return sae.saCfg
}

func newSmartAgentConfigExtension(cfg *Config) (extension.Extension, error) {
	return &smartAgentConfigExtension{saCfg: &cfg.Config}, nil
}
