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

// Hook is a means of providing introspection to a confmap.Provider's lifecycle,
// useful for evaluating Retrieve()'ed content.
type Hook interface {
	OnNew()
	OnRetrieve(scheme string, retrieved map[string]any)
	OnShutdown()
}

// ProviderWrapper is the entrypoint for existing confmap.Providers to be provided with
// configsource.ConfigSource retrieval functionality. Once Wrap()'ed, their
// Retrieve() method as invoked by the service's confmap.Resolver will be
// further resolved by any applicable ConfigSource directives (including
// any initial `config_sources:` settings declarations).
type ProviderWrapper struct {
	providerFactories []confmap.ProviderFactory
	providers         map[string]confmap.Provider
	providersLock     *sync.Mutex
	logger            *zap.Logger
	factories         configsource.Factories
	hooks             []Hook
}

func New(logger *zap.Logger, hooks []Hook) *ProviderWrapper {
	return &ProviderWrapper{
		hooks:         hooks,
		providers:     map[string]confmap.Provider{},
		providersLock: &sync.Mutex{},
		logger:        logger,
		factories:     configSourceFactories,
	}
}

var (
	_ confmap.Provider        = (*wrappedProvider)(nil)
	_ confmap.ProviderFactory = (*wrappedProviderFactory)(nil)
)

type wrappedProviderFactory struct {
	wrapper         *ProviderWrapper
	providerFactory confmap.ProviderFactory
}

func (w *wrappedProviderFactory) Create(settings confmap.ProviderSettings) confmap.Provider {
	provider := w.providerFactory.Create(settings)
	w.wrapper.providersLock.Lock()
	defer w.wrapper.providersLock.Unlock()
	w.wrapper.providers[provider.Scheme()] = provider
	return &wrappedProvider{provider: provider, wrapper: w.wrapper}
}

type wrappedProvider struct {
	wrapper  *ProviderWrapper
	provider confmap.Provider
}

func (w *wrappedProvider) Retrieve(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return w.wrapper.ResolveForWrapped(ctx, uri, watcher, w)
}

func (w *wrappedProvider) Scheme() string {
	return w.provider.Scheme()
}

func (w *wrappedProvider) Shutdown(ctx context.Context) error {
	for _, h := range w.wrapper.hooks {
		h.OnShutdown()
	}
	w.wrapper.providersLock.Lock()
	delete(w.wrapper.providers, w.Scheme())
	w.wrapper.providersLock.Unlock()
	return w.provider.Shutdown(ctx)
}

// Wrap registers the provided confmap.ProviderWrapper in a provider store to proxy
// its methods and expose ConfigSource content resolution capabilities.
func (pw *ProviderWrapper) Wrap(provider confmap.ProviderFactory) confmap.ProviderFactory {
	for _, h := range pw.hooks {
		h.OnNew()
	}
	pw.providersLock.Lock()
	defer pw.providersLock.Unlock()
	pw.providerFactories = append(pw.providerFactories, provider)
	return &wrappedProviderFactory{providerFactory: provider, wrapper: pw}
}

// ResolveForWrapped will retrieve from the wrappedProvider and, if possible, resolve all config source directives with their resolved values.
// If the wrappedProvider's retrieved value is only valid AsRaw() (scalar/array) then that will be returned without further evaluation.
func (pw *ProviderWrapper) ResolveForWrapped(ctx context.Context, uri string, onChange confmap.WatcherFunc, w *wrappedProvider) (*confmap.Retrieved, error) {
	retrieved, err := w.provider.Retrieve(ctx, uri, onChange)
	if err != nil {
		return nil, fmt.Errorf("configsource provider failed retrieving: %w", err)
	}

	raw, arErr := retrieved.AsRaw()
	if arErr != nil {
		return nil, fmt.Errorf("configsource provider failed retrieving raw: %w", arErr)
	}

	// Scalar or array values should be returned as-is.
	if _, ok := raw.(map[string]any); !ok {
		return retrieved, nil
	}

	conf, acErr := retrieved.AsConf()
	if acErr != nil {
		return nil, fmt.Errorf("failed converting retrieved to conf: %w", acErr)
	}
	if conf == nil {
		return nil, fmt.Errorf("retrieved confmap.Conf is unexpectedly nil")
	}

	scheme, stringMap := w.provider.Scheme(), conf.ToStringMap()
	for _, h := range pw.hooks {
		h.OnRetrieve(scheme, stringMap)
	}

	// copy providers map for downstream resolution
	pw.providersLock.Lock()
	providers := map[string]confmap.Provider{}
	for s, p := range pw.providers {
		providers[s] = p
	}
	pw.providersLock.Unlock()
	configSources, confToResolve, err := configsource.BuildConfigSourcesFromConf(ctx, conf, pw.logger, pw.factories, providers)
	if err != nil {
		return nil, fmt.Errorf("failed resolving latestConf: %w", err)
	}

	resolved, closeFunc, err := configsource.ResolveWithConfigSources(ctx, configSources, nil, confToResolve, onChange)
	if err != nil {
		return nil, fmt.Errorf("failed resolving with config sources: %w", err)
	}

	return confmap.NewRetrieved(
		resolved.ToStringMap(), confmap.WithRetrievedClose(
			configsource.MergeCloseFuncs([]confmap.CloseFunc{closeFunc, retrieved.Close}),
		),
	)
}
