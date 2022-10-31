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
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/settings"
)

var _ confmap.Provider = (*providerShim)(nil)

type Provider interface {
	ConfigDScheme() string
	ConfigDProvider() confmap.Provider
	DiscoveryModeScheme() string
	DiscoveryModeProvider() confmap.Provider
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
	if ll, ok := os.LookupEnv("SPLUNK_DISCOVERY_LOG_LEVEL"); ok {
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

func (m *mapProvider) retrieve(scheme string) func(context.Context, string, confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return func(ctx context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
		configDir, dryRun, err := configDirAndDryRun(uri, scheme)
		if err != nil {
			return nil, fmt.Errorf("uri failed validation: %w", err)
		}

		var cfg *Config
		var ok bool
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
			if dryRun {
				printYamlAndExit(discoveryCfg)
			}
			return confmap.NewRetrieved(discoveryCfg)
		}

		return nil, fmt.Errorf("unsupported %s scheme %q", scheme, uri)
	}
}

func configDirAndDryRun(uri, scheme string) (string, bool, error) {
	schemePrefix := fmt.Sprintf("%s:", scheme)
	if !strings.HasPrefix(uri, schemePrefix) {
		return "", false, fmt.Errorf("uri %q is not supported by %s provider", uri, scheme)
	}

	sepIdx := strings.IndexByte(uri, byte(rune(30)))
	if sepIdx == -1 {
		return "", false, fmt.Errorf("invalid uri missing record separator: %q", uri)
	}

	dryRunBoolStr := uri[len(schemePrefix):sepIdx]
	dryRun, err := strconv.ParseBool(dryRunBoolStr)
	if err != nil {
		return "", false, fmt.Errorf("invalid dry run arg %q from %q", dryRunBoolStr, uri)
	}

	configDir := uri[sepIdx+1:]
	return configDir, dryRun, nil
}

func printYamlAndExit(cfg map[string]any) {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		panic(fmt.Errorf("failed marshaling discovery config: %w", err))
	}
	fmt.Printf("%s", out)
	os.Exit(0)
}

func (m *mapProvider) ConfigDScheme() string {
	return settings.ConfigDScheme
}

func (m *mapProvider) DiscoveryModeScheme() string {
	return settings.DiscoveryModeScheme
}
