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

package smartagentextension

import (
	"context"
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/internal/components/componenttest"
)

func TestExtensionLifecycle(t *testing.T) {
	ctx := context.Background()
	createParams := component.ExtensionCreateSettings{}
	cfg := &Config{
		Config: config.Config{
			BundleDir: "/bundle/",
			Collectd: config.CollectdConfig{
				Timeout:   10,
				ConfigDir: "/config/",
			},
		},
	}

	f := NewFactory()
	fstExt, err := f.CreateExtension(ctx, createParams, cfg)
	require.NoError(t, err)
	require.NotNil(t, fstExt)

	mh := componenttest.NewAssertNoErrorHost(t)
	require.NoError(t, fstExt.Start(ctx, mh))
	require.NoError(t, fstExt.Shutdown(ctx))

	sndExt, err := f.CreateExtension(ctx, createParams, cfg)
	require.NoError(t, err)
	require.NotNil(t, sndExt)
	require.NoError(t, sndExt.Start(ctx, mh))
	require.NoError(t, sndExt.Shutdown(ctx))
}
