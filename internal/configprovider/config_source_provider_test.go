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
	"errors"
	"fmt"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmapprovider"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

func TestConfigSourceConfigMapProvider(t *testing.T) {
	tests := []struct {
		parserProvider configmapprovider.Provider
		configLocation string
		wantErr        error
		name           string
		factories      []Factory
	}{
		{
			name: "success",
		},
		{
			name: "wrapped_parser_provider_get_error",
			parserProvider: &mockParserProvider{
				ErrOnGet: true,
			},
			wantErr: &errOnParserProviderGet{},
		},
		{
			name: "duplicated_factory_type",
			factories: []Factory{
				&mockCfgSrcFactory{},
				&mockCfgSrcFactory{},
			},
			wantErr: &errDuplicatedConfigSourceFactory{},
		},
		{
			name: "new_manager_builder_error",
			factories: []Factory{
				&mockCfgSrcFactory{
					ErrOnCreateConfigSource: errors.New("new_manager_builder_error forced error"),
				},
			},
			parserProvider: configmapprovider.NewFile(),
			configLocation: "file:" + path.Join("testdata", "basic_config.yaml"),
			wantErr:        &errConfigSourceCreation{},
		},
		{
			name:           "manager_resolve_error",
			parserProvider: configmapprovider.NewFile(),
			configLocation: "file:" + path.Join("testdata", "manager_resolve_error.yaml"),
			wantErr:        fmt.Errorf("error not wrappedProviders by specific error type: %w", configsource.ErrSessionClosed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factories := tt.factories
			if factories == nil {
				factories = []Factory{
					&mockCfgSrcFactory{},
				}
			}

			pp := NewConfigSourceConfigMapProvider(
				&mockParserProvider{},
				zap.NewNop(),
				component.NewDefaultBuildInfo(),
				factories...,
			)
			require.NotNil(t, pp)

			// Do not use the config.Default() to simplify the test setup.
			cspp := pp.(*configSourceConfigMapProvider)
			if tt.parserProvider != nil {
				cspp.wrappedProvider = tt.parserProvider
			}

			r, err := pp.Retrieve(context.Background(), tt.configLocation, nil)
			require.NoError(t, err)

			cp, err := r.Get(context.Background())
			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, cp)
			} else {
				assert.IsType(t, tt.wantErr, err)
				assert.Nil(t, cp)
				return
			}

			var watchForUpdatedError error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				watchForUpdatedError = cspp.WatchForUpdate()
			}()
			require.NotNil(t, cspp.csm)
			cspp.csm.WaitForWatcher()

			closeErr := cspp.Close(context.Background())
			assert.NoError(t, closeErr)

			wg.Wait()
			assert.Equal(t, configsource.ErrSessionClosed, watchForUpdatedError)
		})
	}
}

type mockParserProvider struct {
	ErrOnGet bool
}

var _ configmapprovider.Provider = (*mockParserProvider)(nil)

func (mpp *mockParserProvider) Retrieve(_ context.Context, _ string, _ configmapprovider.WatcherFunc) (configmapprovider.Retrieved, error) {
	return configmapprovider.NewRetrieved(mpp.Get)
}

func (mpp *mockParserProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (mpp *mockParserProvider) Get(context.Context) (*config.Map, error) {
	if mpp.ErrOnGet {
		return nil, &errOnParserProviderGet{errors.New("mockParserProvider.Get() forced test error")}
	}
	return config.NewMap(), nil
}

func (mpp *mockParserProvider) Close(context.Context) error {
	return nil
}

type errOnParserProviderGet struct{ error }
