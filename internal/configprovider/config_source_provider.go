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
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type errDuplicatedConfigSourceFactory struct{ error }

var (
	_ confmap.Provider = (*configSourceConfigMapProvider)(nil)
)

type configSourceConfigMapProvider struct {
	logger           *zap.Logger
	csm              *Manager
	configServer     *configServer
	wrappedProvider  confmap.Provider
	wrappedRetrieved confmap.Retrieved
	retrieved        confmap.Retrieved
	buildInfo        component.BuildInfo
	factories        []Factory
}

// NewConfigSourceConfigMapProvider creates a ParserProvider that uses config sources.
func NewConfigSourceConfigMapProvider(wrappedProvider confmap.Provider, logger *zap.Logger,
	buildInfo component.BuildInfo, factories ...Factory) confmap.Provider {
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
	onChange confmap.WatcherFunc,
) (confmap.Retrieved, error) {
	wr, err := c.wrappedProvider.Retrieve(ctx, location, onChange)
	if err != nil {
		return confmap.Retrieved{}, err
	}

	existingMap, err := c.wrappedRetrieved.AsConf()
	if err != nil {
		return confmap.Retrieved{}, err
	}

	// Need to merge config maps that we've encountered so far
	if existingMap != nil {
		wrMap, _ := wr.AsConf()
		wrMap.Merge(existingMap)
		c.wrappedRetrieved, err = confmap.NewRetrieved(wrMap.ToStringMap())
		if err != nil {
			return confmap.Retrieved{}, err
		}
	} else {
		c.wrappedRetrieved = wr
	}

	cfg, err := c.Get(ctx)
	if err != nil {
		return confmap.Retrieved{}, err
	}

	c.retrieved, err = confmap.NewRetrieved(
		cfg.ToStringMap(),
		confmap.WithRetrievedClose(wr.Close),
	)
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
func (c *configSourceConfigMapProvider) Get(context.Context) (*confmap.Conf, error) {
	factories, err := makeFactoryMap(c.factories)
	if err != nil {
		return nil, err
	}

	wrappedMap, err := c.wrappedRetrieved.AsConf()
	if err != nil {
		return nil, err
	}
	csm, err := NewManager(wrappedMap, c.logger, c.buildInfo, factories)
	if err != nil {
		return nil, err
	}

	effectiveMap, err := csm.Resolve(context.Background(), wrappedMap)
	if err != nil {
		return nil, err
	}

	// Only start config server if this is the first config source
	if c.configServer == nil {
		c.configServer = newConfigServer(c.logger, wrappedMap.ToStringMap(), effectiveMap.ToStringMap())
		if err = c.configServer.start(); err != nil {
			return nil, err
		}
	} else {
		// Update config server when getting different config sources
		c.configServer.setInitial(wrappedMap.ToStringMap())
		c.configServer.setEffective(effectiveMap.ToStringMap())
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
