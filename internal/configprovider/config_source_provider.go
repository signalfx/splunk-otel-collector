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

package configprovider

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmapprovider"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type errDuplicatedConfigSourceFactory struct{ error }

var (
	_ configmapprovider.Provider = (*configSourceConfigMapProvider)(nil)
)

type configSourceConfigMapProvider struct {
	logger            *zap.Logger
	csm               *Manager
	configServer      *configServer
	wrappedProviders  []configmapprovider.Provider
	wrappedRetrieveds []configmapprovider.Retrieved
	retrieved         configmapprovider.Retrieved
	buildInfo         component.BuildInfo
	factories         []Factory
}

// NewConfigSourceConfigMapProvider creates a ParserProvider that uses config sources.
func NewConfigSourceConfigMapProvider(wrapped []configmapprovider.Provider, logger *zap.Logger, buildInfo component.BuildInfo, factories ...Factory) configmapprovider.Provider {
	return &configSourceConfigMapProvider{
		wrappedProviders: wrapped,
		logger:           logger,
		factories:        factories,
		buildInfo:        buildInfo,
	}
}

func (c *configSourceConfigMapProvider) Retrieve(ctx context.Context, onChange func(*configmapprovider.ChangeEvent)) (configmapprovider.Retrieved, error) {
	for _, p := range c.wrappedProviders {
		retr, err := p.Retrieve(ctx, onChange)
		if err != nil {
			return nil, err
		}
		c.wrappedRetrieveds = append(c.wrappedRetrieveds, retr)
	}

	var err error
	c.retrieved, err = configmapprovider.NewRetrieved(
		c.Get,
		configmapprovider.WithClose(func(ctxF context.Context) error {
			var e error
			for _, ret := range c.wrappedRetrieveds {
				e = multierr.Append(e, ret.Close(ctxF))
			}
			return e
		}))

	return c.retrieved, err
}

func (c *configSourceConfigMapProvider) Shutdown(ctx context.Context) error {
	return c.mergedShutdown(ctx)
}

// Taken from https://github.com/open-telemetry/opentelemetry-collector/blob/40a7d72f9d77c749d2d24056a1b66a7757bc07e3/service/config_provider.go#L196
// Copyright The OpenTelemetry Authors
func (c *configSourceConfigMapProvider) mergedShutdown(ctx context.Context) error {
	var errs error
	for _, p := range c.wrappedProviders {
		errs = multierr.Append(errs, p.Shutdown(ctx))
	}
	return errs
}

// Get returns a config.Parser that wraps the config.Default() with a parser
// that can load and inject data from config sources. If there are no config sources
// in the configuration the returned parser behaves like the config.Default().
func (c *configSourceConfigMapProvider) Get(ctx context.Context) (*config.Map, error) {
	initialMap := config.NewMap()
	for _, r := range c.wrappedRetrieveds {
		cfgMap, err := r.Get(ctx)
		if err != nil {
			return nil, err
		}
		if err = initialMap.Merge(cfgMap); err != nil {
			return nil, err
		}
	}

	var err error
	factories, err := makeFactoryMap(c.factories)
	if err != nil {
		return nil, err
	}

	csm, err := NewManager(initialMap, c.logger, c.buildInfo, factories)
	if err != nil {
		return nil, err
	}

	effectiveMap, err := csm.Resolve(context.Background(), initialMap)
	if err != nil {
		return nil, err
	}

	c.configServer = newConfigServer(c.logger, initialMap.ToStringMap(), effectiveMap.ToStringMap())
	if err = c.configServer.start(); err != nil {
		return nil, err
	}

	c.csm = csm
	return effectiveMap, nil
}

// WatchForUpdate is used to monitor for updates on configuration values that
// were retrieved from config sources.
func (c *configSourceConfigMapProvider) WatchForUpdate() error {
	return c.csm.WatchForUpdate()
}

// Close ends the watch for updates and closes the parser provider and respective
// config sources.
func (c *configSourceConfigMapProvider) Close(ctx context.Context) error {
	if c.configServer != nil {
		_ = c.configServer.shutdown()
	}
	return multierr.Combine(c.csm.Close(ctx), c.retrieved.Close(ctx))
}

func makeFactoryMap(factories []Factory) (Factories, error) {
	fMap := make(Factories)
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return nil, &errDuplicatedConfigSourceFactory{fmt.Errorf("duplicate config source factory %q", f.Type())}
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}
