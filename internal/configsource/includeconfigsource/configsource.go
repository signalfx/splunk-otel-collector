// Copyright Splunk, Inc.
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

package includeconfigsource

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

type includeConfigSource struct {
	Config
}

var _ configsource.ConfigSource = (*includeConfigSource)(nil)

func (i *includeConfigSource) NewSession(context.Context) (configsource.Session, error) {
	return newSession(i.Config)
}

func newConfigSource(_ *zap.Logger, config *Config) (*includeConfigSource, error) {
	if config.DeleteFiles && config.WatchFiles {
		return nil, errors.New("cannot be configured to delete and watch file at the same time")
	}

	return &includeConfigSource{*config}, nil
}
