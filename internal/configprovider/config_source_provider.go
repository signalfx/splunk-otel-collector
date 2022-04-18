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
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type errDuplicatedConfigSourceFactory struct{ error }

var (
	_ config.MapProvider = (*configSourceConfigMapProvider)(nil)
)

type configSourceConfigMapProvider struct {
	logger           *zap.Logger
	csm              *Manager
	configServer     *configServer
	wrappedProvider  config.MapProvider
	wrappedRetrieved config.Retrieved
	retrieved        config.Retrieved
	buildInfo        component.BuildInfo
	factories        []Factory
}

// NewConfigSourceConfigMapProvider creates a ParserProvider that uses config sources.
func NewConfigSourceConfigMapProvider(wrappedProvider config.MapProvider, logger *zap.Logger,
	buildInfo component.BuildInfo, factories ...Factory) config.MapProvider {
	return &configSourceConfigMapProvider{
		wrappedProvider: wrappedProvider,
		logger:          logger,
		factories:       factories,
		buildInfo:       buildInfo,
	}
}

func (c *configSourceConfigMapProvider) Retrieve(
	ctx context.Context,
	location string,
	onChange config.WatcherFunc,
) (config.Retrieved, error) {
	wr, err := c.wrappedProvider.Retrieve(ctx, location, onChange)
	if err != nil {
		return config.Retrieved{}, err
	}
	c.wrappedRetrieved = wr

	cfg, err := c.Get(ctx)
	if err != nil {
		return config.Retrieved{}, err
	}

	closeFunc := c.wrappedRetrieved.CloseFunc
	if closeFunc == nil {
		closeFunc = func(context.Context) error { return nil }
	}

	c.retrieved = config.Retrieved{
		Map:       cfg,
		CloseFunc: closeFunc,
	}

	return c.retrieved, err
}

func (c *configSourceConfigMapProvider) Scheme() string {
	return c.wrappedProvider.Scheme()
}

func (c *configSourceConfigMapProvider) Shutdown(ctx context.Context) error {
	return c.wrappedProvider.Shutdown(ctx)
}

// Get returns a config.Parser that wraps the config.Default() with a parser
// that can load and inject data from config sources. If there are no config sources
// in the configuration the returned parser behaves like the config.Default().
func (c *configSourceConfigMapProvider) Get(context.Context) (*config.Map, error) {
	factories, err := makeFactoryMap(c.factories)
	if err != nil {
		return nil, err
	}

	csm, err := NewManager(c.wrappedRetrieved.Map, c.logger, c.buildInfo, factories)
	if err != nil {
		return nil, err
	}

	effectiveMap, err := csm.Resolve(context.Background(), c.wrappedRetrieved.Map)
	if err != nil {
		return nil, err
	}

	c.configServer = newConfigServer(c.logger, c.wrappedRetrieved.Map.ToStringMap(), effectiveMap.ToStringMap())
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
	return multierr.Combine(c.csm.Close(ctx), c.retrieved.CloseFunc(ctx))
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
