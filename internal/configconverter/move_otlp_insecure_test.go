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
	"go.opentelemetry.io/collector/config/configtest"
)

func TestMoveOTLPInsecureKey(t *testing.T) {
	cfgMap, err := configtest.LoadConfigMap("testdata/otlp-insecure.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = MoveOTLPInsecureKey{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.False(t, cfgMap.IsSet("exporters::otlp::insecure"))
	assert.Equal(t, true, cfgMap.Get("exporters::otlp::tls::insecure"))
}

func TestMoveOTLPInsecureKey_Custom(t *testing.T) {
	cfgMap, err := configtest.LoadConfigMap("testdata/otlp-insecure-custom.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = MoveOTLPInsecureKey{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.False(t, cfgMap.IsSet("exporters::otlp/foo::insecure"))
	assert.Equal(t, true, cfgMap.Get("exporters::otlp/foo::tls::insecure"))
	assert.Equal(t, true, cfgMap.Get("exporters::otlp/foo::tls::insecure_skip_verify"))
}
