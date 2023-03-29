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

package configsource

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/envvarconfigsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/etcd2configsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/includeconfigsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/vaultconfigsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/zookeeperconfigsource"
)

var configSourceFactories = func() configsource.Factories {
	factories := make(configsource.Factories)
	for _, f := range []configsource.Factory{
		envvarconfigsource.NewFactory(),
		includeconfigsource.NewFactory(),
		vaultconfigsource.NewFactory(),
		zookeeperconfigsource.NewFactory(),
		etcd2configsource.NewFactory(),
	} {
		if _, ok := factories[f.Type()]; ok {
			panic(fmt.Sprintf("duplicate config source factory %q", f.Type()))
		}
		factories[f.Type()] = f
	}
	return factories
}()

type Hook interface {
	OnNew()
	OnRetrieve(scheme string, retrieved map[string]any)
	OnShutdown()
}

type Provider interface {
	Provider(provider confmap.Provider) confmap.Provider
}

func New(logger *zap.Logger, hooks []Hook) Provider {
	return &ProviderWrapper{
		hooks:         hooks,
		providers:     map[string]confmap.Provider{},
		providersLock: &sync.Mutex{},
		logger:        logger,
		factories:     configSourceFactories,
	}
}

type ProviderWrapper struct {
	providers     map[string]confmap.Provider
	providersLock *sync.Mutex
	logger        *zap.Logger
	factories     configsource.Factories
	hooks         []Hook
}

var _ confmap.Provider = (*wrappedProvider)(nil)

type wrappedProvider struct {
	wrapper       *ProviderWrapper
	provider      confmap.Provider
	lastRetrieved *confmap.Retrieved
}

func (w *wrappedProvider) Retrieve(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return w.wrapper.ResolveForWrapped(ctx, uri, watcher, w)
}

func (w *wrappedProvider) Scheme() string {
	return w.provider.Scheme()
}

func (w *wrappedProvider) Shutdown(ctx context.Context) error {
	w.wrapper.providersLock.Lock()
	defer w.wrapper.providersLock.Unlock()
	delete(w.wrapper.providers, w.Scheme())
	if len(w.wrapper.providers) == 0 {
		for _, h := range w.wrapper.hooks {
			h.OnShutdown()
		}
	}
	return w.provider.Shutdown(ctx)
}

func (c *ProviderWrapper) Provider(provider confmap.Provider) confmap.Provider {
	for _, h := range c.hooks {
		h.OnNew()
	}
	c.providersLock.Lock()
	defer c.providersLock.Unlock()
	c.providers[provider.Scheme()] = provider
	return &wrappedProvider{provider: provider, wrapper: c, lastRetrieved: &confmap.Retrieved{}}
}

// ResolveForWrapped will retrieve from the wrappedProvider provider and merge the result w/ a previous retrieved instance (if any) as latestConf.
// It will then "configsource.BuildConfigSourcesAndResolve(latestConf)" and return the resolved map as a confmap.Retrieved w/ resolve closer and the wrappedProvider provider closer.
func (c *ProviderWrapper) ResolveForWrapped(ctx context.Context, uri string, onChange confmap.WatcherFunc, w *wrappedProvider) (*confmap.Retrieved, error) {
	provider := w.provider
	retrieved := &confmap.Retrieved{}

	var tmpRetrieved *confmap.Retrieved
	var err error
	// Here we are getting the value directly from the provider, which
	// is what the core's resolver intends (invokes this parent method).
	if tmpRetrieved, err = provider.Retrieve(ctx, uri, onChange); err != nil {
		return nil, err
	} else if tmpRetrieved != nil {
		retrieved = tmpRetrieved
	}

	var previousConf *confmap.Conf
	if previousConf, err = w.lastRetrieved.AsConf(); err != nil {
		return nil, err
	} else if previousConf != nil {
		// Need to merge config maps that we've encountered so far
		if latestConf, e := retrieved.AsConf(); e != nil {
			return nil, fmt.Errorf("failed resolving wrappedProvider retrieve: %w", e)
		} else if e = latestConf.Merge(previousConf); e != nil {
			return nil, fmt.Errorf("failed merging previous wrappedProvider retrieve: %w", e)
		} else if tmpRetrieved, e = confmap.NewRetrieved(latestConf.ToStringMap()); e != nil {
			return nil, err
		} else if tmpRetrieved != nil {
			retrieved = tmpRetrieved
		}
	}

	w.lastRetrieved = retrieved

	latestConf, err := w.lastRetrieved.AsConf()
	if err != nil {
		return nil, err
	} else if latestConf == nil {
		return nil, fmt.Errorf("latest retrieved confmap.Conf is nil")
	}

	scheme, stringMap := provider.Scheme(), latestConf.ToStringMap()
	for _, h := range c.hooks {
		h.OnRetrieve(scheme, stringMap)
	}

	configSources, conf, err := configsource.BuildConfigSourcesAndSettings(ctx, latestConf, c.logger, c.factories)
	if err != nil {
		return nil, fmt.Errorf("failed resolving latestConf: %w", err)
	}

	resolved, closeFunc, err := configsource.ResolveWithConfigSources(ctx, configSources, conf, onChange)
	if err != nil {
		return nil, fmt.Errorf("failed resolving with config sources: %w", err)
	}

	return confmap.NewRetrieved(
		resolved.ToStringMap(), confmap.WithRetrievedClose(
			configsource.MergeCloseFuncs([]confmap.CloseFunc{closeFunc, w.lastRetrieved.Close}),
		),
	)
}
