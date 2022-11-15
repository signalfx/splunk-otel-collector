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

package envvarconfigsource

import (
	"context"

	"go.opentelemetry.io/collector/component"
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	// The "type" of environment variable config sources in configuration.
	typeStr = "env"
)

type envVarFactory struct{}

func (e *envVarFactory) Type() component.Type {
	return typeStr
}

func (e *envVarFactory) CreateDefaultConfig() expcfg.Source {
	return &Config{
		SourceSettings: expcfg.NewSourceSettings(component.NewID(typeStr)),
	}
}

func (e *envVarFactory) CreateConfigSource(_ context.Context, params configprovider.CreateParams, cfg expcfg.Source) (configsource.ConfigSource, error) {
	return newConfigSource(params, cfg.(*Config)), nil
}

// NewFactory creates a factory for Vault ConfigSource objects.
func NewFactory() configprovider.Factory {
	return &envVarFactory{}
}
