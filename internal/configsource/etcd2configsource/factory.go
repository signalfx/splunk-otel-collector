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

package etcd2configsource

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/component"
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	// The "type" of etcd2 config sources in configuration.
	typeStr = "etcd2"

	defaultEndpoints = "http://localhost:2379"
)

// Private error types to help with testability.
type (
	errMissingEndpoint struct{ error }
	errInvalidEndpoint struct{ error }
)

type etcd2Factory struct{}

func (v *etcd2Factory) Type() component.Type {
	return typeStr
}

func (v *etcd2Factory) CreateDefaultConfig() expcfg.Source {
	return &Config{
		SourceSettings: expcfg.NewSourceSettings(component.NewID(typeStr)),
		Endpoints:      []string{defaultEndpoints},
	}
}

func (v *etcd2Factory) CreateConfigSource(_ context.Context, params configprovider.CreateParams, cfg expcfg.Source) (configsource.ConfigSource, error) {
	etcd2Cfg := cfg.(*Config)

	if len(etcd2Cfg.Endpoints) == 0 {
		return nil, &errMissingEndpoint{errors.New("cannot connect to etcd2 without any endpoints")}
	}

	for _, uri := range etcd2Cfg.Endpoints {
		if _, err := url.ParseRequestURI(uri); err != nil {
			return nil, &errInvalidEndpoint{fmt.Errorf("invalid endpoint %q: %w", uri, err)}
		}
	}

	return newConfigSource(params, etcd2Cfg)
}

// NewFactory creates a new etcd2Factory instance
func NewFactory() configprovider.Factory {
	return &etcd2Factory{}
}
