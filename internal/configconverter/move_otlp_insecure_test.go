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

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configmapprovider"
)

func TestConverterProvider_Noop(t *testing.T) {
	pp := &converterProvider{
		wrapped: configmapprovider.NewFile("testdata/otlp-insecure.yaml"),
	}
	r, err := pp.Retrieve(context.Background(), nil)
	require.NoError(t, err)
	v, err := r.Get(context.Background())
	require.NoError(t, err)
	assert.True(t, v.IsSet("exporters::otlp::insecure"))
}

func TestMoveOTLPInsecureKey(t *testing.T) {
	pp := &converterProvider{
		wrapped:     configmapprovider.NewFile("testdata/otlp-insecure.yaml"),
		cfgMapFuncs: []CfgMapFunc{MoveOTLPInsecureKey},
	}
	r, err := pp.Retrieve(context.Background(), nil)
	require.NoError(t, err)
	v, err := r.Get(context.Background())
	require.NoError(t, err)
	assert.False(t, v.IsSet("exporters::otlp::insecure"))
	assert.Equal(t, true, v.Get("exporters::otlp::tls::insecure"))
}

func TestMoveOTLPInsecureKey_Custom(t *testing.T) {
	pp := &converterProvider{
		wrapped:     configmapprovider.NewFile("testdata/otlp-insecure-custom.yaml"),
		cfgMapFuncs: []CfgMapFunc{MoveOTLPInsecureKey},
	}
	r, err := pp.Retrieve(context.Background(), nil)
	require.NoError(t, err)
	v, err := r.Get(context.Background())
	require.NoError(t, err)
	assert.False(t, v.IsSet("exporters::otlp/foo::insecure"))
	assert.Equal(t, true, v.Get("exporters::otlp/foo::tls::insecure"))
	assert.Equal(t, true, v.Get("exporters::otlp/foo::tls::insecure_skip_verify"))
}
