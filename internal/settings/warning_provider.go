// Copyright  Splunk, Inc.
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

package settings

import (
	"context"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

// warningProviderFactory is a wrapper for a confmap.ProviderFactory that logs warnings for particular URIs.
type warningProviderFactory struct {
	confmap.ProviderFactory
	warnings map[string]string
}

func (wpf *warningProviderFactory) Create(set confmap.ProviderSettings) confmap.Provider {
	return &warningProvider{
		Provider: wpf.ProviderFactory.Create(set),
		warnings: wpf.warnings,
		logger:   set.Logger,
	}
}

type warningProvider struct {
	confmap.Provider
	warnings map[string]string
	logger   *zap.Logger
}

func (wp *warningProvider) Retrieve(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	val := uri[len(wp.Provider.Scheme())+1:]
	if warning, ok := wp.warnings[val]; ok {
		wp.logger.Warn(warning)
	}
	return wp.Provider.Retrieve(ctx, uri, watcher)
}
