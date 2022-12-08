// Copyright Splunk, Inc.
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

package manifests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigMap(t *testing.T) {
	cm := ConfigMap{
		Name:      "some.config.map",
		Namespace: "some.namespace",
		Data: `config: |
  key.one: value one
  key.two:
    nested.key.one: nested.value.one
    nested.key.two: nested.value.two`,
	}

	manifest, err := cm.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: v1
kind: ConfigMap
metadata:
  name: some.config.map
  namespace: some.namespace
data:
  config: |
    key.one: value one
    key.two:
      nested.key.one: nested.value.one
      nested.key.two: nested.value.two
`, manifest)
}

func TestEmptyConfigMap(t *testing.T) {
	cm := ConfigMap{}
	manifest, err := cm.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: v1
kind: ConfigMap
metadata:
data:
`, manifest)
}
