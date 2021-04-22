// Copyright 2020 Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envvarconfigsource

import (
	"context"

	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

type envVarConfigSource struct {
	defaults map[string]interface{}
}

var _ configsource.ConfigSource = (*envVarConfigSource)(nil)

func (e *envVarConfigSource) NewSession(context.Context) (configsource.Session, error) {
	return newSession(e.defaults)
}

func newConfigSource(_ *zap.Logger, cfg *Config) (*envVarConfigSource, error) {
	defaults := make(map[string]interface{})
	if cfg.Defaults != nil {
		defaults = cfg.Defaults
	}
	return &envVarConfigSource{
		defaults: defaults,
	}, nil
}
