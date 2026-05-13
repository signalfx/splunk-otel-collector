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

//go:build integration

package tests

import (
	"bytes"
	"os"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

// removeAllKeysOtherThan removes all bundled receivers from the discovery receiver to avoid having to update the
// expected config every time a new bundled receiver rule is added. It returns the number of bundled receivers.
func removeBundledReceivers(discReceiverCfg any) (removedCount int) {
	receiverToKeep := "prometheus_simple"
	discReceivers := discReceiverCfg.(map[string]any)["receivers"].(map[string]any)
	for k := range discReceivers {
		if k != receiverToKeep {
			delete(discReceivers, k)
			removedCount++
		}
	}
	return removedCount
}

func readConfigFromYamlTmplFile(t *testing.T, path string, ctxData map[string]any) map[string]any {
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	tml, err := template.New("tmpl").Parse(string(b))
	require.NoError(t, err)
	out := &bytes.Buffer{}
	require.NoError(t, tml.Execute(out, ctxData))
	require.NoError(t, err)
	cm, err := confmap.NewRetrievedFromYAML(out.Bytes())
	require.NoError(t, err)
	cmr, err := cm.AsRaw()
	require.NoError(t, err)
	return cmr.(map[string]any)
}
