// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.uber.org/zap"
)

var errValueUpdated = errors.New("configuration must retrieve the updated value")

func BuildConfigSourcesAndResolve(ctx context.Context, confToFurtherResolve *confmap.Conf, logger *zap.Logger, factories Factories, watcher confmap.WatcherFunc) (*confmap.Conf, confmap.CloseFunc, error) {
	cfgSources, conf, err := BuildConfigSourcesFromConf(ctx, confToFurtherResolve, logger, factories, nil)
	if err != nil {
		return nil, nil, err
	}

	return ResolveWithConfigSources(ctx, cfgSources, nil, conf, watcher)
}

func TestConfigSourceManagerNewManager(t *testing.T) {
	tests := []struct {
		factories Factories
		wantErr   string
		name      string
		file      string
	}{
		{
			name: "basic_config",
			file: "basic_config",
			factories: Factories{
				component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{},
			},
		},
		{
			name:      "unknown_type",
			file:      "basic_config",
			factories: Factories{},
			wantErr:   "unknown config_sources type \"tstcfgsrc\"",
		},
		{
			name: "build_error",
			file: "basic_config",
			factories: Factories{
				component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{
					ErrOnCreateConfigSource: errors.New("forced test error"),
				},
			},
			wantErr: "failed to create config source tstcfgsrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := path.Join("testdata", tt.file+".yaml")
			parser, err := confmaptest.LoadConf(filename)
			require.NoError(t, err)

			_, _, err = BuildConfigSourcesAndResolve(context.Background(), parser, zap.NewNop(), tt.factories, nil)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigSourceResolved(t *testing.T) {
	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"test_selector": {Value: "test_value"},
			},
		},
	}

	originalCfg := map[string]any{
		"top0": map[string]any{
			"int":    1,
			"cfgsrc": "${tstcfgsrc:test_selector}",
		},
	}
	expectedCfg := map[string]any{
		"top0": map[string]any{
			"int":    1,
			"cfgsrc": "test_value",
		},
	}

	cp := confmap.NewFromStringMap(originalCfg)

	res, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, cp, func(*confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, res.ToStringMap())
	assert.NoError(t, closeFunc(context.Background()))
}

func TestConfigSourceManagerResolveRemoveConfigSourceSection(t *testing.T) {
	cfg := map[string]any{
		"another_section": map[string]any{
			"int": 42,
		},
	}

	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{},
	}

	res, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, confmap.NewFromStringMap(cfg), func(*confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	delete(cfg, "config_sources")
	assert.Equal(t, cfg, res.ToStringMap())
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestConfigSourceManagerResolveErrors(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		config          map[string]any
		configSourceMap map[string]ConfigSource
		name            string
	}{
		{
			name: "incorrect_cfgsrc_ref",
			config: map[string]any{
				"cfgsrc": "${tstcfgsrc:selector?{invalid}}",
			},
			configSourceMap: map[string]ConfigSource{
				"tstcfgsrc": &TestConfigSource{},
			},
		},
		{
			name: "error_on_retrieve",
			config: map[string]any{
				"cfgsrc": "${tstcfgsrc:selector}",
			},
			configSourceMap: map[string]ConfigSource{
				"tstcfgsrc": &TestConfigSource{ErrOnRetrieve: testErr},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, closeFunc, err := ResolveWithConfigSources(context.Background(), tt.configSourceMap, nil, confmap.NewFromStringMap(tt.config), func(*confmap.ChangeEvent) {
				panic("must not be called")
			})
			require.Error(t, err)
			require.Nil(t, res)
			assert.NoError(t, callClose(context.Background(), closeFunc))
		})
	}
}

func TestConfigSourceManagerYAMLInjection(t *testing.T) {
	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"valid_yaml_str": {Value: `
bool: true
int: 42
source: string
map:
  k0: v0
  k1: v1
`},
				"invalid_yaml_str": {Value: ":"},
			},
		},
	}

	file := path.Join("testdata", "yaml_injection.yaml")
	cp, err := confmaptest.LoadConf(file)
	require.NoError(t, err)

	expectedFile := path.Join("testdata", "yaml_injection_expected.yaml")
	expectedParser, err := confmaptest.LoadConf(expectedFile)
	require.NoError(t, err)
	expectedCfg := expectedParser.ToStringMap()

	res, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, cp, func(*confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, res.ToStringMap())
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestConfigSourceManagerArraysAndMaps(t *testing.T) {
	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"elem0": {Value: "elem0_value"},
				"elem1": {Value: "elem1_value"},
				"k0":    {Value: "k0_value"},
				"k1":    {Value: "k1_value"},
			},
		},
	}

	file := path.Join("testdata", "arrays_and_maps.yaml")
	cp, err := confmaptest.LoadConf(file)
	require.NoError(t, err)

	expectedFile := path.Join("testdata", "arrays_and_maps_expected.yaml")
	expectedParser, err := confmaptest.LoadConf(expectedFile)
	require.NoError(t, err)

	res, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, cp, func(*confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, expectedParser.ToStringMap(), res.ToStringMap())
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestConfigSourceManagerParamsHandling(t *testing.T) {
	tstCfgSrc := TestConfigSource{
		ValueMap: map[string]valueEntry{
			"elem0": {Value: nil},
			"elem1": {
				Value: map[string]any{
					"p0": true,
					"p1": "a string with spaces",
					"p3": 42,
				},
			},
			"k0": {Value: nil},
			"k1": {
				Value: map[string]any{
					"p0": true,
					"p1": "a string with spaces",
					"p2": map[string]any{
						"p2_0": "a nested map0",
						"p2_1": true,
					},
				},
			},
		},
	}

	// Set OnRetrieve to check if the parameters were parsed as expectedSettings.
	tstCfgSrc.OnRetrieve = func(_ context.Context, selector string, paramsConfigMap *confmap.Conf) error {
		var val any
		if paramsConfigMap != nil {
			val = paramsConfigMap.ToStringMap()
		}
		assert.Equal(t, tstCfgSrc.ValueMap[selector].Value, val)
		return nil
	}

	file := path.Join("testdata", "params_handling.yaml")
	cp, err := confmaptest.LoadConf(file)
	require.NoError(t, err)

	expectedFile := path.Join("testdata", "params_handling_expected.yaml")
	expectedParser, err := confmaptest.LoadConf(expectedFile)
	require.NoError(t, err)

	res, closeFunc, err := ResolveWithConfigSources(context.Background(), map[string]ConfigSource{"tstcfgsrc": &tstCfgSrc}, nil, cp, func(*confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, expectedParser.ToStringMap(), res.ToStringMap())
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestConfigSourceManagerWatchForUpdate(t *testing.T) {
	watchForUpdateCh := make(chan error, 1)

	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"test_selector": {
					Value:            "test_value",
					WatchForUpdateCh: watchForUpdateCh,
				},
			},
		},
	}

	originalCfg := map[string]any{
		"top0": map[string]any{
			"var0": "${tstcfgsrc:test_selector}",
		},
	}

	cp := confmap.NewFromStringMap(originalCfg)
	watchCh := make(chan *confmap.ChangeEvent)
	_, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, cp, func(event *confmap.ChangeEvent) {
		watchCh <- event
	})
	require.NoError(t, err)

	watchForUpdateCh <- nil

	ce := <-watchCh
	assert.NoError(t, ce.Error)
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestConfigSourceManagerMultipleWatchForUpdate(t *testing.T) {
	watchForUpdateCh := make(chan error, 2)
	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"test_selector": {
					Value:            "test_value",
					WatchForUpdateCh: watchForUpdateCh,
				},
			},
		},
	}

	originalCfg := map[string]any{
		"top0": map[string]any{
			"var0": "${tstcfgsrc:test_selector}",
			"var1": "${tstcfgsrc:test_selector}",
			"var2": "${tstcfgsrc:test_selector}",
			"var3": "${tstcfgsrc:test_selector}",
		},
	}

	cp := confmap.NewFromStringMap(originalCfg)
	watchCh := make(chan *confmap.ChangeEvent)
	_, closeFunc, err := ResolveWithConfigSources(context.Background(), cfgSources, nil, cp, func(event *confmap.ChangeEvent) {
		watchCh <- event
	})
	require.NoError(t, err)

	watchForUpdateCh <- errValueUpdated
	watchForUpdateCh <- errValueUpdated

	ce := <-watchCh
	require.ErrorIs(t, ce.Error, errValueUpdated)
	close(watchForUpdateCh)
	assert.NoError(t, callClose(context.Background(), closeFunc))
}

func TestManagerExpandString(t *testing.T) {
	ctx := context.Background()
	cfgSources := map[string]ConfigSource{
		"tstcfgsrc": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"str_key": {Value: "test_value"},
				"int_key": {Value: 1},
				"nil_key": {Value: nil},
			},
		},
		"tstcfgsrc/named": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"int_key": {Value: 42},
			},
		},
	}

	t.Setenv("envvar", "envvar_value")
	t.Setenv("envvar_str_key", "str_key")

	tests := []struct {
		want    any
		wantErr error
		name    string
		input   string
	}{
		{
			name:  "literal_string",
			input: "literal_string",
			want:  "literal_string",
		},
		{
			name:    "cfgsrc_int",
			input:   "$tstcfgsrc:int_key",
			wantErr: errors.New("invalid config source invocation $tstcfgsrc:int_key"),
		},
		{
			name:  "concatenate_cfgsrc_string",
			input: "prefix-${tstcfgsrc:str_key}",
			want:  "prefix-test_value",
		},
		{
			name:  "concatenate_cfgsrc_non_string",
			input: "prefix-${tstcfgsrc:int_key}",
			want:  "prefix-1",
		},
		{
			name:  "envvar",
			input: "${envvar}",
			want:  "envvar_value",
		},
		{
			name:  "prefixed_envvar",
			input: "prefix-${envvar}",
			want:  "prefix-envvar_value",
		},
		{
			name:    "envvar_treated_as_cfgsrc",
			input:   "${envvar:suffix}",
			wantErr: &errUnknownConfigSource{},
		},
		{
			name:    "cfgsrc_using_envvar",
			input:   "$tstcfgsrc:$envvar_str_key",
			wantErr: errors.New("invalid config source invocation $tstcfgsrc:$envvar_str_key"),
		},
		{
			name:    "envvar_cfgsrc_using_envvar",
			input:   "$envvar/$tstcfgsrc:$envvar_str_key",
			wantErr: errors.New("invalid config source invocation $envvar"),
		},
		{
			name:  "delimited_cfgsrc",
			input: "${tstcfgsrc:int_key}",
			want:  1,
		},
		{
			name:    "unknown_delimited_cfgsrc",
			input:   "${cfgsrc:int_key}",
			wantErr: &errUnknownConfigSource{},
		},
		{
			name:  "delimited_cfgsrc_with_spaces",
			input: "${ tstcfgsrc: int_key }",
			want:  1,
		},
		{
			name:  "interpolated_and_delimited_cfgsrc",
			input: "0/${ tstcfgsrc: $envvar_str_key }/2/${tstcfgsrc:int_key}",
			wantErr: fmt.Errorf(`failed to process selector for config source "tstcfgsrc" selector "$envvar_str_key: %w`,
				errors.New("invalid config source invocation $envvar_str_key")),
		},
		{
			name:    "named_config_src",
			input:   "$tstcfgsrc/named:int_key",
			wantErr: errors.New("invalid config source invocation $tstcfgsrc/named:int_key"),
		},
		{
			name:  "named_config_src_bracketed",
			input: "${tstcfgsrc/named:int_key}",
			want:  42,
		},
		{
			name:  "envvar_name_separator",
			input: "${envvar}/test/test",
			want:  "envvar_value/test/test",
		},
		{
			name:    "envvar_treated_as_cfgsrc",
			input:   "${envvar/test:test}",
			wantErr: &errUnknownConfigSource{},
		},
		{
			name:  "retrieved_nil",
			input: "${tstcfgsrc:nil_key}",
		},
		{
			name:  "retrieved_nil_on_string",
			input: "prefix-${tstcfgsrc:nil_key}-suffix",
			want:  "prefix--suffix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := map[string]confmap.Provider{
				"env": envprovider.NewFactory().Create(confmap.ProviderSettings{}),
			}
			got, closeFunc, err := resolveStringValue(ctx, cfgSources, providers, tt.input, func(_ *confmap.ChangeEvent) {
				panic("must not be called")
			})
			if tt.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, callClose(ctx, closeFunc))
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_parseCfgSrc(t *testing.T) {
	tests := []struct {
		params     any
		name       string
		str        string
		cfgSrcName string
		selector   string
		wantErr    bool
	}{
		{
			name:       "basic",
			str:        "cfgsrc:selector",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
		},
		{
			name:    "missing_selector",
			str:     "cfgsrc",
			wantErr: true,
		},
		{
			name:       "params",
			str:        "cfgsrc:selector?p0=1&p1=a_string&p2=true",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"p0": 1,
				"p1": "a_string",
				"p2": true,
			},
		},
		{
			name:       "query_pass_nil",
			str:        "cfgsrc:selector?p0&p1&p2",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"p0": nil,
				"p1": nil,
				"p2": nil,
			},
		},
		{
			name:       "array_in_params",
			str:        "cfgsrc:selector?p0=0&p0=1&p0=2&p1=done",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"p0": []any{0, 1, 2},
				"p1": "done",
			},
		},
		{
			name:       "empty_param",
			str:        "cfgsrc:selector?no_closing=",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"no_closing": any(nil),
			},
		},
		{
			name:       "use_url_encode",
			str:        "cfgsrc:selector?p0=contains+%3D+and+%26+too",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"p0": "contains = and & too",
			},
		},
		{
			name:       "parse_complex_values",
			str:        "cfgsrc:selector?p0=[1,2]&p1={\"k0\":\"v0\",\"k1\":\"v1\"}",
			cfgSrcName: "cfgsrc",
			selector:   "selector",
			params: map[string]any{
				"p0": []any{1, 2},
				"p1": map[string]any{
					"k0": "v0",
					"k1": "v1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgSrcName, selector, paramsConfigMap, err := parseCfgSrcInvocation(tt.str)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.cfgSrcName, cfgSrcName)
			assert.Equal(t, tt.selector, selector)
			var val any
			if paramsConfigMap != nil {
				val = paramsConfigMap.ToStringMap()
			}
			assert.Equal(t, tt.params, val)
		})
	}
}

func callClose(ctx context.Context, closeFunc confmap.CloseFunc) error {
	if closeFunc == nil {
		return nil
	}
	return closeFunc(ctx)
}

func TestConfigSourceBuild(t *testing.T) {
	ctx := context.Background()
	testFactories := Factories{
		component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{},
	}

	tests := []struct {
		configSettings     map[string]Settings
		factories          Factories
		expectedCfgSources map[string]ConfigSource
		wantErr            string
		name               string
	}{
		{
			name: "unknown_config_source",
			configSettings: map[string]Settings{
				"tstcfgsrc": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewIDWithName("unknown_config_source", "tstcfgsrc")),
				},
			},
			factories: testFactories,
			wantErr:   "unknown unknown_config_source config source type for tstcfgsrc",
		},
		{
			name: "creation_error",
			configSettings: map[string]Settings{
				"tstcfgsrc": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
				},
			},
			factories: Factories{
				component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{
					ErrOnCreateConfigSource: errors.New("forced test error"),
				},
			},
			wantErr: "failed to create config source tstcfgsrc: forced test error",
		},
		{
			name: "factory_return_nil",
			configSettings: map[string]Settings{
				"tstcfgsrc": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
				},
			},
			factories: Factories{
				component.MustNewType("tstcfgsrc"): &mockNilCfgSrcFactory{},
			},
			wantErr: "factory for \"tstcfgsrc\" produced a nil extension",
		},
		{
			name: "base_case",
			configSettings: map[string]Settings{
				"tstcfgsrc/named": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewIDWithName("tstcfgsrc", "named")),
					Endpoint:       "some_endpoint",
					Token:          "some_token",
				},
			},
			factories: testFactories,
			expectedCfgSources: map[string]ConfigSource{
				"tstcfgsrc/named": &TestConfigSource{
					ValueMap: map[string]valueEntry{
						"tstcfgsrc/named": {
							Value: &MockCfgSrcSettings{
								SourceSettings: NewSourceSettings(component.MustNewIDWithName("tstcfgsrc", "named")),
								Endpoint:       "some_endpoint",
								Token:          "some_token",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builtCfgSources, err := BuildConfigSources(ctx, tt.configSettings, zap.NewNop(), tt.factories)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedCfgSources, builtCfgSources)
		})
	}
}

type mockNilCfgSrcFactory struct{}

func (m *mockNilCfgSrcFactory) Type() component.Type {
	return component.MustNewType("tstcfgsrc")
}

var _ Factory = (*mockNilCfgSrcFactory)(nil)

func (m *mockNilCfgSrcFactory) CreateDefaultConfig() Settings {
	return &MockCfgSrcSettings{
		SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
		Endpoint:       "default_endpoint",
	}
}

func (m *mockNilCfgSrcFactory) CreateConfigSource(context.Context, Settings, *zap.Logger) (ConfigSource, error) {
	return nil, nil
}

type MockCfgSrcFactory struct {
	ErrOnCreateConfigSource error
}

type MockCfgSrcSettings struct {
	SourceSettings
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}

var _ Settings = (*MockCfgSrcSettings)(nil)

var _ Factory = (*MockCfgSrcFactory)(nil)

func (m *MockCfgSrcFactory) Type() component.Type {
	return component.MustNewType("tstcfgsrc")
}

func (m *MockCfgSrcFactory) CreateDefaultConfig() Settings {
	return &MockCfgSrcSettings{
		SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
		Endpoint:       "default_endpoint",
	}
}

func (m *MockCfgSrcFactory) CreateConfigSource(_ context.Context, cfg Settings, _ *zap.Logger) (ConfigSource, error) {
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

var _ ConfigSource = (*TestConfigSource)(nil)

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

var _ confmap.Provider = (*confmapProvider)(nil)

type confmapProvider struct {
	scheme      string
	shouldError bool
}

func (c confmapProvider) Scheme() string {
	return c.scheme
}

func (c confmapProvider) Retrieve(context.Context, string, confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if c.shouldError {
		return nil, errors.New("confmap provider error")
	}
	return confmap.NewRetrieved("value from confmap provider")
}

func (confmapProvider) Shutdown(context.Context) error {
	return nil
}

func TestConfigSourceConfmapProviderFallbackResolved(t *testing.T) {
	configSources := map[string]ConfigSource{
		"config_source": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"test_selector": {Value: "value from config source"},
			},
		},
		"conflicting_name": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"used_selector": {Value: "value from conflicting config source"},
			},
		},
	}

	confmapProviders := map[string]confmap.Provider{
		"confmap_provider": confmapProvider{scheme: "confmap_provider"},
		"conflicting_name": confmapProvider{scheme: "conflicting_name"},
	}

	originalCfg := map[string]any{
		"top0": map[string]any{
			"fromConfigSource":            "${config_source:test_selector}",
			"fromConflictingConfigSource": "${conflicting_name:used_selector}",
			"fromConfmapProvider":         "${confmap_provider:confmap_provider_selector}",
		},
	}
	expectedCfg := map[string]any{
		"top0": map[string]any{
			"fromConfigSource":            "value from config source",
			"fromConflictingConfigSource": "value from conflicting config source",
			"fromConfmapProvider":         "value from confmap provider",
		},
	}

	cp := confmap.NewFromStringMap(originalCfg)

	res, closeFunc, err := ResolveWithConfigSources(
		context.Background(), configSources, confmapProviders, cp,
		func(_ *confmap.ChangeEvent) {
			t.Fatal("shouldn't be called")
		},
	)
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, res.ToStringMap())
	assert.NoError(t, closeFunc(context.Background()))
}

func TestConfigSourceConfmapProviderFallbackWithError(t *testing.T) {
	configSources := map[string]ConfigSource{
		"config_source": &TestConfigSource{
			ValueMap: map[string]valueEntry{
				"test_selector": {Value: "value from config source"},
			},
		},
	}

	confmapProviders := map[string]confmap.Provider{
		"confmap_provider": confmapProvider{scheme: "confmap_provider", shouldError: true},
	}

	cfg := confmap.NewFromStringMap(map[string]any{
		"top0": map[string]any{
			"fromConfigSource":    "${config_source:test_selector}",
			"fromConfmapProvider": "${confmap_provider:confmap_provider_selector}",
		},
	})

	res, closeFunc, err := ResolveWithConfigSources(
		context.Background(), configSources, confmapProviders, cfg,
		func(_ *confmap.ChangeEvent) {
			t.Fatal("shouldn't be called")
		},
	)
	require.EqualError(t, err, `retrieve error from confmap provider "confmap_provider": confmap provider error`)
	assert.Nil(t, res)
	assert.Nil(t, closeFunc)
}
