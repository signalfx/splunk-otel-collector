// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestBasicSecretAccess relies on the presence of a vault container running alongside.
// You can run this container with:
// docker run --rm -it --name vault -p 8200:8200 -e VAULT_DEV_ROOT_TOKEN_ID=token -e VAULT_TOKEN=token -e VAULT_ADDR=http://0.0.0.0:8200 hashicorp/vault:latest
func TestBasicSecretAccess(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()

	cmd := exec.Command("/bin/sh", "-c", "docker exec vault vault kv put secret/kv k0=string.value k1=123.456")
	err := cmd.Run()
	require.NoError(t, err)

	collector, stop := tc.SplunkOtelCollector(
		"vault_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			collector = collector.WithEnv(map[string]string{
				"INSERT_ACTION":  "insert",
				"VAULT_HOSTNAME": "localhost",
			})
			return collector
		},
	)
	defer stop()

	effective := collector.EffectiveConfig(tc, 55554)
	if !testutils.CollectorImageIsSet() {
		// default collector process uses --set service.telemetry args
		delete(effective["service"].(map[string]any), "telemetry")
	}
	require.Equal(t, map[string]any{
		"exporters": map[string]any{"debug/noop": nil},
		"processors": map[string]any{
			"resource": map[string]any{
				"attributes": []any{
					map[string]any{"action": "insert", "key": "expands-vault-path-value", "value": "string.value"},
					map[string]any{"action": "insert", "key": "also-expands-vault-path-value", "value": 123.456}}}},
		"receivers": map[string]any{"otlp/noop": map[string]any{"protocols": map[string]any{"http": any(nil)}}},
		"service": map[string]any{
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"exporters":  []any{"debug/noop"},
					"processors": []any{"resource"},
					"receivers":  []any{"otlp/noop"}}}}}, effective)
}
