// Copyright Splunk, Inc.
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

package discovery

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/properties"
	"github.com/signalfx/splunk-otel-collector/internal/settings"
)

var _ confmap.Provider = (*providerShim)(nil)

type Provider interface {
	ConfigDScheme() string
	ConfigDProvider() confmap.Provider
	DiscoveryModeScheme() string
	DiscoveryModeProvider() confmap.Provider
	PropertyScheme() string
	PropertyProvider() confmap.Provider
}

type providerShim struct {
	retrieve func(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error)
	scheme   string
}

func (p providerShim) Retrieve(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return p.retrieve(ctx, uri, watcher)
}

func (p providerShim) Scheme() string {
	return p.scheme
}

func (p providerShim) Shutdown(ctx context.Context) error {
	return nil // nolint:unparam
}

type mapProvider struct {
	logger     *zap.Logger
	configs    map[string]*Config
	discoverer *discoverer
}

func New() (Provider, error) {
	m := &mapProvider{configs: map[string]*Config{}}
	zapConfig := zap.NewProductionConfig()
	logLevel := zap.WarnLevel
	if ll, ok := os.LookupEnv(logLevelEnvVar); ok {
		if l, err := zapcore.ParseLevel(ll); err == nil {
			logLevel = l
		}
	}
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)
	var err error
	if m.logger, err = zapConfig.Build(); err != nil {
		return (*mapProvider)(nil), err
	}
	if m.discoverer, err = newDiscoverer(m.logger); err != nil {
		return (*mapProvider)(nil), err
	}

	return m, nil
}

func (m *mapProvider) ConfigDProvider() confmap.Provider {
	return providerShim{
		scheme:   m.ConfigDScheme(),
		retrieve: m.retrieve(m.ConfigDScheme()),
	}
}

func (m *mapProvider) DiscoveryModeProvider() confmap.Provider {
	return providerShim{
		scheme:   m.DiscoveryModeScheme(),
		retrieve: m.retrieve(m.DiscoveryModeScheme()),
	}
}

func (m *mapProvider) PropertyProvider() confmap.Provider {
	return providerShim{
		scheme:   m.PropertyScheme(),
		retrieve: m.retrieve(m.PropertyScheme()),
	}
}

func (m *mapProvider) retrieve(scheme string) func(context.Context, string, confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return func(ctx context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
		schemePrefix := fmt.Sprintf("%s:", scheme)
		if !strings.HasPrefix(uri, schemePrefix) {
			return nil, fmt.Errorf("uri %q is not supported by %s provider", uri, scheme)
		}

		uriVal := uri[len(schemePrefix):]
		if schemePrefix == fmt.Sprintf("%s:", settings.PropertyScheme) {
			return m.parsedProperty(uriVal)
		}

		var cfg *Config
		var ok bool
		configDir := uriVal
		if cfg, ok = m.configs[configDir]; !ok {
			cfg = NewConfig(m.logger)
			if err := cfg.Load(configDir); err != nil {
				return nil, err
			}
			m.configs[configDir] = cfg
		}

		if strings.HasPrefix(uri, settings.ConfigDScheme) {
			return confmap.NewRetrieved(cfg.toServiceConfig())
		}

		if strings.HasPrefix(uri, settings.DiscoveryModeScheme) {
			discoveryCfg, err := m.discoverer.discover(cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to successfully discover target services: %w", err)
			}
			return confmap.NewRetrieved(discoveryCfg)
		}

		return nil, fmt.Errorf("unsupported %s scheme %q", scheme, uri)
	}
}

func (m *mapProvider) ConfigDScheme() string {
	return settings.ConfigDScheme
}

func (m *mapProvider) DiscoveryModeScheme() string {
	return settings.DiscoveryModeScheme
}

func (m *mapProvider) PropertyScheme() string {
	return settings.PropertyScheme
}

func (m *mapProvider) parsedProperty(rawProperty string) (*confmap.Retrieved, error) {
	// split property from value
	equalsIdx := strings.Index(rawProperty, "=")
	if equalsIdx == -1 || len(rawProperty) <= equalsIdx+1 {
		return nil, fmt.Errorf("invalid discovery property %q not of form <property>=<value>", rawProperty)
	}
	prop, err := properties.NewProperty(rawProperty[:equalsIdx], rawProperty[equalsIdx+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid discovery property: %w", err)
	}
	m.discoverer.propertiesConf.Merge(confmap.NewFromStringMap(prop.ToStringMap()))
	// return nil confmap to satisfy signature
	return confmap.NewRetrieved(nil)
}
