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
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

const (
	// The "type" of environment variable config sources in configuration.
	typeStr = "env"
)

type envVarFactory struct{}

func (e *envVarFactory) Type() component.Type {
	return typeStr
}

func (e *envVarFactory) CreateDefaultConfig() configsource.Settings {
	return &Config{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
	}
}

func (e *envVarFactory) CreateConfigSource(_ context.Context, settings configsource.Settings, _ *zap.Logger) (configsource.ConfigSource, error) {
	return newConfigSource(settings.(*Config)), nil
}

// NewFactory creates a factory for Vault ConfigSource objects.
func NewFactory() configsource.Factory {
	return &envVarFactory{}
}
