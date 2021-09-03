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

package zookeeperconfigsource

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	// The "type" of zookeeper config sources in configuration.
	typeStr = "zookeeper"

	defaultEndpoint = "localhost:2181"
	defaultTimeout  = time.Second * 10
)

// Private error type to help with testability.
type errMissingEndpoint struct{ error }

type zkFactory struct{}

func (v *zkFactory) Type() config.Type {
	return typeStr
}

func (v *zkFactory) CreateDefaultConfig() configprovider.ConfigSettings {
	return &Config{
		Settings:  configprovider.NewSettings(typeStr),
		Endpoints: []string{defaultEndpoint},
		Timeout:   defaultTimeout,
	}
}

func (v *zkFactory) CreateConfigSource(_ context.Context, params configprovider.CreateParams, cfg configprovider.ConfigSettings) (configsource.ConfigSource, error) {
	return newConfigSource(params, cfg.(*Config))
}

// NewFactory returns a new zookeekeper config source factory
func NewFactory() configprovider.Factory {
	return &zkFactory{}
}
