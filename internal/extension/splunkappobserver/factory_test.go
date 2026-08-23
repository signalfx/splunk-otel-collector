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

package splunkappobserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	require.NotNil(t, factory)
	assert.Equal(t, component.MustNewType(TypeStr), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()

	require.IsType(t, &Config{}, cfg)
	assert.Equal(t, defaultAppsPath, cfg.(*Config).AppsPath)
	assert.Equal(t, defaultRefreshInterval, cfg.(*Config).RefreshInterval)
}

func TestCreateExtension(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	ext, err := factory.Create(t.Context(), extensiontest.NewNopSettings(component.MustNewType(TypeStr)), cfg)

	require.NoError(t, err)
	require.NotNil(t, ext)
	assert.Implements(t, (*component.Component)(nil), ext)
}
