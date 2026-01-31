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

package includeconfigsource

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

const (
	// The "type" of file config sources in configuration.
	typeStr = "include"
)

type includeFactory struct{}

func (f *includeFactory) Type() component.Type {
	return component.MustNewType(typeStr)
}

func (f *includeFactory) CreateDefaultConfig() configsource.Settings {
	return &Config{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
	}
}

func (f *includeFactory) CreateConfigSource(_ context.Context, settings configsource.Settings, logger *zap.Logger) (configsource.ConfigSource, error) {
	return newConfigSource(settings.(*Config), logger)
}

// NewFactory creates a factory for include ConfigSource objects.
func NewFactory() configsource.Factory {
	return &includeFactory{}
}
