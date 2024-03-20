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
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
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

func (v *etcd2Factory) CreateDefaultConfig() configsource.Settings {
	return &Config{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
		Endpoints:      []string{defaultEndpoints},
	}
}

func (v *etcd2Factory) CreateConfigSource(_ context.Context, settings configsource.Settings, logger *zap.Logger) (configsource.ConfigSource, error) {
	etcd2Cfg := settings.(*Config)

	if len(etcd2Cfg.Endpoints) == 0 {
		return nil, &errMissingEndpoint{errors.New("cannot connect to etcd2 without any endpoints")}
	}

	for _, uri := range etcd2Cfg.Endpoints {
		if _, err := url.ParseRequestURI(uri); err != nil {
			return nil, &errInvalidEndpoint{fmt.Errorf("invalid endpoint %q: %w", uri, err)}
		}
	}

	return newConfigSource(etcd2Cfg, logger)
}

// NewFactory creates a new etcd2Factory instance
func NewFactory() configsource.Factory {
	return &etcd2Factory{}
}
