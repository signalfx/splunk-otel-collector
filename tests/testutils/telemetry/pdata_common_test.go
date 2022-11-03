// Copyright  Splunk, Inc.
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

package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"gopkg.in/yaml.v2"
)

func TestSanitizeAttributes(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(".", "testdata", "common", "attributes.yaml"))
	require.NoError(t, err)
	anyMap := map[string]any{}
	require.NoError(t, yaml.Unmarshal(b, &anyMap))
	pMap := pcommon.NewMap()
	pMap.PutInt("int", 1)
	pMap.PutDouble("float", 1.234)
	pMap.PutStr("string", "a\nlong\nstring\n")
	pMap.PutBool("bool", false)
	pMap.PutEmpty("empty")
	require.Equal(t, sanitizeAttributes(pMap.AsRaw()), sanitizeAttributes(anyMap))
}
