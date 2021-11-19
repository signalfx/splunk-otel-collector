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
	_ configmapprovider.Provider  = (*configSourceParserProvider)(nil)
	_ configmapprovider.Retrieved = (*configSourceParserProvider)(nil)
)

type configSourceParserProvider struct {
	logger       *zap.Logger
	csm          *Manager
	configServer *configServer
	pp           configmapprovider.Provider
	retrieved    configmapprovider.Retrieved
	buildInfo    component.BuildInfo
	factories    []Factory
}

// NewConfigSourceParserProvider creates a ParserProvider that uses config sources.
func NewConfigSourceParserProvider(pp configmapprovider.Provider, logger *zap.Logger, buildInfo component.BuildInfo, factories ...Factory) configmapprovider.Provider {
	return &configSourceParserProvider{
		pp:        pp,
		logger:    logger,
		factories: factories,
		buildInfo: buildInfo,
	}
}

func (c *configSourceParserProvider) Retrieve(ctx context.Context, onChange func(*configmapprovider.ChangeEvent)) (configmapprovider.Retrieved, error) {
	var err error
	c.retrieved, err = c.pp.Retrieve(ctx, onChange)
	return c, err
}

func (c *configSourceParserProvider) Shutdown(ctx context.Context) error {
	return c.pp.Shutdown(ctx)
}

// Get returns a config.Parser that wraps the config.Default() with a parser
// that can load and inject data from config sources. If there are no config sources
// in the configuration the returned parser behaves like the config.Default().
func (c *configSourceParserProvider) Get(ctx context.Context) (*config.Map, error) {
	initialMap, err := c.retrieved.Get(ctx)
	if err != nil {
		return nil, err
	}

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
func (c *configSourceParserProvider) WatchForUpdate() error {
	return c.csm.WatchForUpdate()
}

// Close ends the watch for updates and closes the parser provider and respective
// config sources.
func (c *configSourceParserProvider) Close(ctx context.Context) error {
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
