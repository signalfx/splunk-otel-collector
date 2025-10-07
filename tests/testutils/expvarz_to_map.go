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

package testutils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

func expvarzPageToMap(t testing.TB, body []byte, configType string) map[string]any {
	// Convert the full expvar with the equivalent of
	// cat <expvarz_page> | jq -r '.["splunk.config.initial"]'
	var expvarMap map[string]any
	err := json.Unmarshal(body, &expvarMap)
	require.NoError(t, err)
	configInJSON, ok := expvarMap["splunk.config."+configType]
	require.True(t, ok, "key 'splunk.config.%s' not found", configType)
	configStr, ok := configInJSON.(string)
	require.True(t, ok, "'splunk.config.%s' cannot be cast to string", configType)

	configMap := map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(configStr), &configMap))

	return confmap.NewFromStringMap(configMap).ToStringMap()
}
