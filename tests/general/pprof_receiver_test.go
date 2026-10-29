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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPprofReceiverExportsProfilesToSplunkHEC(t *testing.T) {
	received := make(chan string, 1)
	hecServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		select {
		case received <- string(body):
		default:
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"Success","code":0}`)
	}))
	defer hecServer.Close()

	tc := testutils.NewTestcase(t)
	defer tc.ShutdownOTLPReceiverSink()

	configPath := filepath.Join(t.TempDir(), "pprof-hec.yaml")
	config := `receivers:
  pprof:
    self:
      collection_interval: 1s
      initial_delay: 1s
exporters:
  splunk_hec:
    token: test-token
    endpoint: ${env:PPROF_HEC_ENDPOINT}
service:
  pipelines:
    profiles:
      receivers: [pprof]
      exporters: [splunk_hec]
`
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o600))

	_, shutdown := tc.SplunkOtelCollectorProcess("", func(collector testutils.Collector) testutils.Collector {
		return collector.WithArgs("--config", configPath, "--feature-gates=service.profilesSupport").WithEnv(map[string]string{
			"PPROF_HEC_ENDPOINT": hecServer.URL + "/services/collector",
		})
	})
	defer shutdown()

	select {
	case event := <-received:
		require.Contains(t, event, "profiling.data.format")
		require.Contains(t, event, "pprof-gzip-base64")
		require.Contains(t, event, "com.splunk.sourcetype")
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for the pprof profile to arrive at the HEC server")
	}
}
