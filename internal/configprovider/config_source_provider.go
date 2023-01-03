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
	"go.uber.org/zap"
)

var (
	_ confmap.Provider = (*configSourceConfigMapProvider)(nil)
)

type Hook interface {
	OnNew()
	OnRetrieve(scheme string, retrieved map[string]any)
	OnShutdown()
}

type configSourceConfigMapProvider struct {
	logger           *zap.Logger
	hooks            []Hook
	wrappedProvider  confmap.Provider
	wrappedRetrieved *confmap.Retrieved
	buildInfo        component.BuildInfo
	factories        []Factory
}

// NewConfigSourceConfigMapProvider creates a ParserProvider that uses config sources.
func NewConfigSourceConfigMapProvider(wrappedProvider confmap.Provider, logger *zap.Logger,
	buildInfo component.BuildInfo, hooks []Hook, factories ...Factory) confmap.Provider {
	for _, h := range hooks {
		h.OnNew()
	}
	return &configSourceConfigMapProvider{
		hooks:            hooks,
		wrappedProvider:  wrappedProvider,
		logger:           logger,
		factories:        factories,
		buildInfo:        buildInfo,
		wrappedRetrieved: &confmap.Retrieved{},
	}
}

func (c *configSourceConfigMapProvider) Retrieve(ctx context.Context, uri string, onChange confmap.WatcherFunc) (*confmap.Retrieved, error) {
	var tmpWR *confmap.Retrieved
	var err error
	newWrappedRetrieved := &confmap.Retrieved{}
	if tmpWR, err = c.wrappedProvider.Retrieve(ctx, uri, onChange); err != nil {
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

	factories, err := makeFactoryMap(c.factories)
	if err != nil {
		return nil, err
	}

	wrappedMap, err := c.wrappedRetrieved.AsConf()
	if err != nil {
		return nil, err
	}

	scheme, stringMap := c.Scheme(), wrappedMap.ToStringMap()
	for _, h := range c.hooks {
		h.OnRetrieve(scheme, stringMap)
	}

	retrieved, closeFunc, err := Resolve(ctx, wrappedMap, c.logger, c.buildInfo, factories, onChange)
	if err != nil {
		return nil, err
	}

	return confmap.NewRetrieved(retrieved, confmap.WithRetrievedClose(mergeCloseFuncs([]confmap.CloseFunc{closeFunc, c.wrappedRetrieved.Close})))
}

func (c *configSourceConfigMapProvider) Scheme() string {
	return c.wrappedProvider.Scheme()
}

func (c *configSourceConfigMapProvider) Shutdown(ctx context.Context) error {
	for _, h := range c.hooks {
		h.OnShutdown()
	}
	return c.wrappedProvider.Shutdown(ctx)
}

func makeFactoryMap(factories []Factory) (Factories, error) {
	fMap := make(Factories)
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return nil, fmt.Errorf("duplicate config source factory %q", f.Type())
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}
