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

	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
)

type errDuplicatedConfigSourceFactory struct{ error }

var (
	_ confmap.Provider = (*configSourceConfigMapProvider)(nil)
)

type configSourceConfigMapProvider struct {
	logger           *zap.Logger
	csm              *Manager
	configServer     *configconverter.ConfigServer
	wrappedProvider  confmap.Provider
	wrappedRetrieved *confmap.Retrieved
	retrieved        *confmap.Retrieved
	buildInfo        component.BuildInfo
	factories        []Factory
}

// NewConfigSourceConfigMapProvider creates a ParserProvider that uses config sources.
func NewConfigSourceConfigMapProvider(wrappedProvider confmap.Provider, logger *zap.Logger,
	buildInfo component.BuildInfo, configServer *configconverter.ConfigServer, factories ...Factory) confmap.Provider {
	configServer.Register()
	return &configSourceConfigMapProvider{
		configServer:     configServer,
		wrappedProvider:  wrappedProvider,
		logger:           logger,
		factories:        factories,
		buildInfo:        buildInfo,
		wrappedRetrieved: &confmap.Retrieved{},
		retrieved:        &confmap.Retrieved{},
	}
}

func (c *configSourceConfigMapProvider) Retrieve(
	ctx context.Context,
	location string,
	onChange confmap.WatcherFunc,
) (*confmap.Retrieved, error) {
	var tmpWR *confmap.Retrieved
	var err error
	newWrappedRetrieved := &confmap.Retrieved{}
	if tmpWR, err = c.wrappedProvider.Retrieve(ctx, location, onChange); err != nil {
		return nil, err
	} else if tmpWR != nil {
		newWrappedRetrieved = tmpWR
	}

	existingMap, err := c.wrappedRetrieved.AsConf()
	if err != nil {
		return nil, err
	}

	// Need to merge config maps that we've encountered so far
	if existingMap != nil {
		wrMap, _ := newWrappedRetrieved.AsConf()
		wrMap.Merge(existingMap)
		if tmpWR, err = confmap.NewRetrieved(wrMap.ToStringMap()); err != nil {
			return nil, err
		} else if tmpWR != nil {
			newWrappedRetrieved = tmpWR
		}
	}
	c.wrappedRetrieved = newWrappedRetrieved

	var cfg *confmap.Conf
	if cfg, err = c.Get(ctx, location); err != nil {
		return nil, err
	} else if cfg == nil {
		cfg = &confmap.Conf{}
	}

	c.retrieved, err = confmap.NewRetrieved(
		cfg.ToStringMap(),
		confmap.WithRetrievedClose(c.wrappedRetrieved.Close),
	)
	return c.retrieved, err
}

func (c *configSourceConfigMapProvider) Scheme() string {
	return c.wrappedProvider.Scheme()
}

func (c *configSourceConfigMapProvider) Shutdown(ctx context.Context) error {
	c.configServer.Unregister()
	return c.wrappedProvider.Shutdown(ctx)
}

// Get returns a config.Parser that wraps the config.Default() with a parser
// that can load and inject data from config sources. If there are no config sources
// in the configuration the returned parser behaves like the config.Default().
func (c *configSourceConfigMapProvider) Get(_ context.Context, uri string) (*confmap.Conf, error) {
	factories, err := makeFactoryMap(c.factories)
	if err != nil {
		return nil, err
	}

	wrappedMap, err := c.wrappedRetrieved.AsConf()
	if err != nil {
		return nil, err
	}
	c.configServer.SetForScheme(c.Scheme(), wrappedMap.ToStringMap())
	csm, err := NewManager(wrappedMap, c.logger, c.buildInfo, factories)
	if err != nil {
		return nil, err
	}

	effectiveMap, err := csm.Resolve(context.Background(), wrappedMap)
	if err != nil {
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
	c.configServer.Unregister()
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
