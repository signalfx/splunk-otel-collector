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

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLogLevelToVerbosity(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/logging_loglevel.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	expectedCfgMap, err := confmaptest.LoadConf("testdata/logging_loglevel-after.yaml")
	require.NoError(t, err)
	require.NotNil(t, expectedCfgMap)

	err = LogLevelToVerbosity{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	require.Equal(t, expectedCfgMap, cfgMap)
}
