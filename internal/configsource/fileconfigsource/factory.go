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

package fileconfigsource

import (
	"context"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	// The "type" of file config sources in configuration.
	typeStr = "file"
)

type fileFactory struct{}

func (f *fileFactory) Type() config.Type {
	return typeStr
}

func (f *fileFactory) CreateDefaultConfig() configprovider.ConfigSettings {
	return &Config{
		Settings: configprovider.NewSettings(typeStr),
	}
}

func (f *fileFactory) CreateConfigSource(_ context.Context, params configprovider.CreateParams, cfg configprovider.ConfigSettings) (configsource.ConfigSource, error) {
	return newConfigSource(params.Logger, cfg.(*Config))
}

// NewFactory creates a factory for Vault ConfigSource objects.
func NewFactory() configprovider.Factory {
	return &fileFactory{}
}
