// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package smartagentreceiver

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/otelcol/otelcoltest"
)

func TestLoadUnsupportedCollectdMonitorOnWindows(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := otelcoltest.LoadConfig(
		path.Join(".", "testdata", "collectd_apache.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		`error reading receivers configuration for "smartagent/collectd/apache": smart agent monitor type "collectd/apache" is not supported on windows platforms`)
	require.Nil(t, cfg)
}
