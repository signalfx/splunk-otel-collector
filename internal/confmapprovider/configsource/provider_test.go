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
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

func TestConfigSourceConfigMapProvider(t *testing.T) {
	tests := []struct {
		providerFactory confmap.ProviderFactory
		factories       configsource.Factories
		wantErr         string
		name            string
		uris            []string
	}{
		{
			name: "success",
		},
		{
			name: "wrapped_parser_provider_get_error",
			providerFactory: &mockParserProviderFactory{
				ErrOnGet: true,
			},
			wantErr: "mockParserProvider.Get() forced test error",
		},
		{
			name: "new_manager_builder_error",
			factories: configsource.Factories{
				component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{
					ErrOnCreateConfigSource: errors.New("new_manager_builder_error forced error"),
				},
			},
			providerFactory: fileprovider.NewFactory(),
			uris:            []string{"file:" + path.Join("testdata", "basic_config.yaml")},
			wantErr:         "failed to create config source tstcfgsrc",
		},
		{
			name:            "manager_resolve_error",
			providerFactory: fileprovider.NewFactory(),
			uris:            []string{"file:" + path.Join("testdata", "manager_resolve_error.yaml")},
			wantErr:         "config source \"tstcfgsrc\" failed to retrieve value: no value for selector \"selector\"",
		},
		{
			name:            "multiple_config_success",
			providerFactory: fileprovider.NewFactory(),
			uris: []string{
				"file:" + path.Join("testdata", "arrays_and_maps_expected.yaml"),
				"file:" + path.Join("testdata", "yaml_injection_expected.yaml"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.factories == nil {
				tt.factories = configsource.Factories{
					component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{},
				}
			}

			hookOne := &mockHook{}
			hookTwo := &mockHook{}
			hooks := []*mockHook{hookOne, hookTwo}
			for _, h := range hooks {
				h.On("OnNew")
				h.On("OnRetrieve", mock.AnythingOfType("string"), mock.Anything)
				h.On("OnShutdown")
			}

			providerFactory := tt.providerFactory
			if providerFactory == nil {
				providerFactory = &mockParserProviderFactory{}
			}
			p := New(zap.NewNop(), []Hook{hookOne, hookTwo})
			require.NotNil(t, p)

			p.factories = tt.factories
			pp := p.Wrap(providerFactory).Create(confmap.ProviderSettings{})

			for _, h := range hooks {
				h.AssertCalled(t, "OnNew")
				h.AssertNotCalled(t, "OnRetrieve")
				h.AssertNotCalled(t, "OnShutdown")
			}

			provider := providerFactory.Create(confmap.ProviderSettings{})
			expectedScheme := provider.Scheme()

			i := 0
			for ok := true; ok; {
				var uri string
				if tt.uris != nil {
					uri = tt.uris[i]
				} else {
					uri = fmt.Sprintf("%s:", provider.Scheme())
				}

				r, err := pp.Retrieve(context.Background(), uri, nil)

				if tt.wantErr == "" {
					require.NoError(t, err)
					require.NotNil(t, r)
					rMap, errAsConf := r.AsConf()
					require.NoError(t, errAsConf)
					assert.NotNil(t, rMap)
					assert.NoError(t, r.Close(context.Background()))
				} else {
					assert.ErrorContains(t, err, tt.wantErr)
					assert.Nil(t, r)
					break
				}
				i++
				ok = i < len(tt.uris)
			}

			for _, h := range hooks {
				if tt.wantErr != "" {
					h.AssertNotCalled(t, "OnRetrieve")
				} else {
					h.AssertCalled(t, "OnRetrieve", expectedScheme, mock.Anything)
				}
				h.AssertNotCalled(t, "OnShutdown")
			}

			assert.NoError(t, pp.Shutdown(context.Background()))

			for _, h := range hooks {
				h.AssertCalled(t, "OnShutdown")
			}
		})
	}
}

type mockParserProviderFactory struct {
	ErrOnGet bool
}

func (mppf *mockParserProviderFactory) Create(_ confmap.ProviderSettings) confmap.Provider {
	return &mockParserProvider{ErrOnGet: mppf.ErrOnGet}
}

type mockParserProvider struct {
	ErrOnGet bool
}

var _ confmap.Provider = (*mockParserProvider)(nil)

func (mpp *mockParserProvider) Retrieve(context.Context, string, confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if mpp.ErrOnGet {
		return nil, errors.New("mockParserProvider.Get() forced test error")
	}
	return confmap.NewRetrieved(confmap.New().ToStringMap())
}

func (mpp *mockParserProvider) Shutdown(context.Context) error {
	return nil
}

func (mpp *mockParserProvider) Scheme() string {
	return "mock"
}

type mockHook struct {
	mock.Mock
}

var _ Hook = (*mockHook)(nil)

func (m *mockHook) OnNew() {
	m.Called()
}

func (m *mockHook) OnRetrieve(scheme string, _ map[string]any) {
	m.Called(scheme, mock.Anything)
}

func (m *mockHook) OnShutdown() {
	m.Called()
}

type MockCfgSrcFactory struct {
	ErrOnCreateConfigSource error
}

type MockCfgSrcSettings struct {
	configsource.SourceSettings
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}

var _ configsource.Settings = (*MockCfgSrcSettings)(nil)

var _ configsource.Factory = (*MockCfgSrcFactory)(nil)

func (m *MockCfgSrcFactory) Type() component.Type {
	return component.MustNewType("tstcfgsrc")
}

func (m *MockCfgSrcFactory) CreateDefaultConfig() configsource.Settings {
	return &MockCfgSrcSettings{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID("tstcfgsrc")),
		Endpoint:       "default_endpoint",
	}
}

func (m *MockCfgSrcFactory) CreateConfigSource(_ context.Context, cfg configsource.Settings, _ *zap.Logger) (configsource.ConfigSource, error) {
	if m.ErrOnCreateConfigSource != nil {
		return nil, m.ErrOnCreateConfigSource
	}
	return &TestConfigSource{
		ValueMap: map[string]valueEntry{
			cfg.ID().String(): {
				Value: cfg,
			},
		},
	}, nil
}

// TestConfigSource a ConfigSource to be used in tests.
type TestConfigSource struct {
	ValueMap map[string]valueEntry

	ErrOnRetrieve    error
	ErrOnRetrieveEnd error
	ErrOnClose       error

	OnRetrieve func(ctx context.Context, selector string, paramsConfigMap *confmap.Conf) error
}

type valueEntry struct {
	Value            any
	WatchForUpdateCh chan error
}

var _ configsource.ConfigSource = (*TestConfigSource)(nil)

func (t *TestConfigSource) Retrieve(ctx context.Context, selector string, paramsConfigMap *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if t.OnRetrieve != nil {
		if err := t.OnRetrieve(ctx, selector, paramsConfigMap); err != nil {
			return nil, err
		}
	}

	if t.ErrOnRetrieve != nil {
		return nil, t.ErrOnRetrieve
	}

	entry, ok := t.ValueMap[selector]
	if !ok {
		return nil, fmt.Errorf("no value for selector %q", selector)
	}

	if entry.WatchForUpdateCh != nil {
		doneCh := make(chan struct{})
		startWatch(entry.WatchForUpdateCh, doneCh, watcher)
		return confmap.NewRetrieved(entry.Value, confmap.WithRetrievedClose(func(_ context.Context) error {
			close(doneCh)
			return nil
		}))
	}

	return confmap.NewRetrieved(entry.Value)
}

func (t *TestConfigSource) Shutdown(context.Context) error {
	return t.ErrOnClose
}

func startWatch(watchForUpdateCh chan error, doneCh chan struct{}, watcher confmap.WatcherFunc) {
	go func() {
		select {
		case err := <-watchForUpdateCh:
			watcher(&confmap.ChangeEvent{Error: err})
			return
		case <-doneCh:
			return
		}
	}()
}
