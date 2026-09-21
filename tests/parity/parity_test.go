// Copyright Splunk, Inc.
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

package parity_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/tests/parity"
	"github.com/signalfx/splunk-otel-collector/tests/parity/adapter/otelcol"
	"github.com/signalfx/splunk-otel-collector/tests/parity/adapter/uf"
	splunkbackend "github.com/signalfx/splunk-otel-collector/tests/parity/backend/splunk"
)

// hecToken must be a valid, non-zero GUID: UF httpout rejects tokens shorter
// than 36 chars and rejects the all-zeros GUID as "not in supported format"
// (splcore TokenEncryptionDecryptionHelper::isEmptyGuid).
const hecToken = "11111111-1111-1111-1111-111111111111"

const (
	indexUF = "parity_uf"
	indexUC = "parity_uc"
)

// ufOutputs is the UF httpout outputs.conf. httpout posts cooked S2S to
// HEC_ENDPOINT/services/collector/s2s; batch settings are forced low so events
// flush immediately instead of waiting the 30s default.
const ufOutputs = `[httpout]
uri = HEC_ENDPOINT
httpEventCollectorToken = HEC_TOKEN
sslVerifyServerCert = false
batchTimeout = 1
batchSize = 1
`

// TestParity runs each case through the UF oracle and the otelcol candidate into
// one real Splunk, each on its own index, and asserts both land the case's
// authored event. Direct UF-vs-candidate field-by-field parity (source,
// sourcetype, index normalization) is the next milestone; this slice proves both
// agents produce the expected event via their own native configs.
func TestParity(t *testing.T) {
	ufAdapter := uf.New("")
	if _, err := os.Stat(ufAdapter.InstallDir()); err != nil {
		t.Skipf("UF install not found at %s: %v", ufAdapter.InstallDir(), err)
	}
	ocAdapter := otelcol.New("")
	if _, err := os.Stat(ocAdapter.InstallDir()); err != nil {
		t.Skipf("otelcol binary dir not found at %s (run `make otelcol`): %v", ocAdapter.InstallDir(), err)
	}

	cases, err := filepath.Glob("tests/*/test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found under tests/")
	}

	ctx := context.Background()
	backend := splunkbackend.New(splunkbackend.Config{
		// PARITY_SPLUNK_IMAGE overrides the image (e.g. to reuse a locally cached
		// tag and skip a slow pull). Empty falls back to the backend default.
		Image:   os.Getenv("PARITY_SPLUNK_IMAGE"),
		Token:   hecToken,
		Indexes: []string{indexUF, indexUC},
	})
	t.Log("starting Splunk backend (first run pulls the image; can take a few minutes)")
	if err := backend.Start(ctx); err != nil {
		t.Fatalf("start backend: %v", err)
	}
	t.Cleanup(func() { backend.Stop(context.Background()) })

	opts := parity.RunOptions{
		Quiescence: 3 * time.Second,
		Timeout:    90 * time.Second,
	}

	for _, path := range cases {
		c, err := parity.LoadCase(path)
		if err != nil {
			t.Fatalf("load %s: %v", path, err)
		}
		t.Run(c.Name, func(t *testing.T) {
			if !supportsOS(c) {
				t.Skipf("case does not target %s", currentOS())
			}

			collectorCfg, err := os.ReadFile(filepath.Join(filepath.Dir(path), "collector.yaml"))
			if err != nil {
				t.Fatalf("read collector.yaml: %v", err)
			}

			ufRun := parity.AgentRun{
				Adapter:     uf.New(""),
				ConfigFiles: map[string]string{"inputs.conf": c.Conf, "outputs.conf": ufOutputs},
				Index:       indexUF,
			}
			ucRun := parity.AgentRun{
				Adapter:     otelcol.New(""),
				ConfigFiles: map[string]string{otelcol.ConfigFile: string(collectorCfg)},
				Index:       indexUC,
			}

			reference := []parity.Record{c.Expected.AsReference()}
			v := parity.SubsetValidator{}

			ufRecs, err := parity.RunAgent(ctx, c, ufRun, backend, opts)
			if err != nil {
				t.Fatalf("run UF: %v", err)
			}
			assertMatch(t, "UF", v.Validate(reference, ufRecs), ufRecs)

			ucRecs, err := parity.RunAgent(ctx, c, ucRun, backend, opts)
			if err != nil {
				t.Fatalf("run otelcol: %v", err)
			}
			assertMatch(t, "otelcol", v.Validate(reference, ucRecs), ucRecs)
		})
	}
}

func assertMatch(t *testing.T, agent string, res parity.Result, records []parity.Record) {
	t.Helper()
	if res.Match {
		return
	}
	for _, m := range res.Mismatches {
		t.Errorf("%s: record %d field %q: expected %q, got %q", agent, m.Record, m.Field, m.Expected, m.Actual)
	}
	t.Logf("%s captured %d record(s): %+v", agent, len(records), records)
}

func currentOS() string { return runtime.GOOS }

func supportsOS(c *parity.Case) bool {
	if len(c.OS) == 0 {
		return true
	}
	for _, o := range c.OS {
		if o == currentOS() {
			return true
		}
	}
	return false
}
