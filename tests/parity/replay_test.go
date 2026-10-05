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
	"errors"
	"flag"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/tests/parity"
	"github.com/signalfx/splunk-otel-collector/tests/parity/adapter/otelcol"
	"github.com/signalfx/splunk-otel-collector/tests/parity/adapter/uf"
	splunkbackend "github.com/signalfx/splunk-otel-collector/tests/parity/backend/splunk"
)

// update needs a UF install (PARITY_UF_DIR); a normal run does not. Prefer
// `make update-goldens`, which installs the pinned UF and sets it.
var update = flag.Bool("update", false, "regenerate golden files by running the UF oracle")

// taRunnerGate registers splunk_inputs and splunk_outputs.
const taRunnerGate = "enableTARunner"

// hecToken must be a valid, non-zero GUID: UF httpout rejects tokens shorter
// than 36 chars and rejects the all-zeros GUID as "not in supported format"
// (splcore TokenEncryptionDecryptionHelper::isEmptyGuid).
const hecToken = "11111111-1111-1111-1111-111111111111"

// confDir is the case subdirectory holding the Splunk .conf structure.
const confDir = "conf"

// indexFor names a case's index after its directory. Cases share one Splunk
// container and the readback is scoped by index, so each needs its own or it
// would capture the events of every case that ran before it. Directory names are
// unique already and, being lowercase and dash-separated, are valid index names.
func indexFor(casePath string) string {
	return "parity_" + filepath.Base(filepath.Dir(casePath))
}

// TestParity runs the otelcol candidate for each case into a real Splunk and
// asserts the events it lands match the case's golden on the asserted fields.
// With -update it instead regenerates the goldens from the UF oracle and stops.
func TestParity(t *testing.T) {
	if *update {
		ufAdapter := uf.New("")
		if _, err := os.Stat(ufAdapter.InstallDir()); err != nil {
			t.Skipf("-update needs a UF install; not found at %s: %v", ufAdapter.InstallDir(), err)
		}
	} else {
		ocAdapter := otelcol.New("")
		if _, err := os.Stat(ocAdapter.InstallDir()); err != nil {
			t.Skipf("otelcol binary dir not found at %s (run `make otelcol`): %v", ocAdapter.InstallDir(), err)
		}
	}

	paths, err := filepath.Glob("tests/*/test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no cases found under tests/")
	}

	// The indexes have to exist before any agent forwards, so they are created at
	// backend startup.
	indexes := make([]string, len(paths))
	for i, path := range paths {
		indexes[i] = indexFor(path)
	}

	ctx := context.Background()
	backend := splunkbackend.New(splunkbackend.Config{
		// PARITY_SPLUNK_IMAGE overrides the image (e.g. to reuse a locally cached
		// tag and skip a slow pull). Empty falls back to the backend default.
		Image:   os.Getenv("PARITY_SPLUNK_IMAGE"),
		Token:   hecToken,
		Indexes: indexes,
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

	for i, path := range paths {
		c, err := parity.LoadCase(path)
		if err != nil {
			t.Fatalf("load %s: %v", path, err)
		}
		index := indexes[i]
		t.Run(c.Name, func(t *testing.T) {
			if !supportsOS(c) {
				t.Skipf("case does not target %s", currentOS())
			}

			goldenPath := filepath.Join(filepath.Dir(path), parity.GoldenFile)

			// Generating is not a verdict on the candidate: write the golden, stop,
			// and let the reviewed golden be replayed on a normal run.
			if *update {
				regenerateGolden(ctx, t, c, caseConf(t, path), index, backend, goldenPath, opts)
				return
			}

			golden, err := parity.LoadGolden(goldenPath)
			if err != nil {
				t.Fatalf("load golden (run with -update to generate): %v", err)
			}

			ucRecs, err := parity.RunAgent(ctx, c, candidateRun(t, path, index), backend, opts)
			if err != nil {
				t.Fatalf("run otelcol candidate: %v", err)
			}

			// The golden holds only the asserted fields, so SubsetValidator
			// compares the candidate on exactly those fields and ignores the rest.
			assertMatch(t, parity.SubsetValidator{}.Validate(golden, ucRecs), golden, ucRecs)
		})
	}
}

// splunkIndexName is what Splunk accepts for an index: lowercase letters,
// digits, underscores and hyphens, not leading with an underscore or hyphen.
var splunkIndexName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// TestCases checks every case directory is complete without needing Docker or a
// UF, so a missing or renamed case file fails fast instead of after a Splunk
// container has booted.
func TestCases(t *testing.T) {
	cases, err := filepath.Glob("tests/*/test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found under tests/")
	}
	for _, path := range cases {
		dir := filepath.Base(filepath.Dir(path))
		t.Run(dir, func(t *testing.T) {
			// The directory names the case's index, so it has to be a legal one.
			if !splunkIndexName.MatchString(dir) {
				t.Errorf("directory %q is not usable as a Splunk index name", dir)
			}
			c, err := parity.LoadCase(path)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if c.Name == "" {
				t.Fatal("name is empty")
			}

			caseConf(t, path)
			if caseFile(t, path, parity.GoldenFile) == "" {
				t.Errorf("%s is empty", parity.GoldenFile)
			}
			// collector.yaml is optional, but present and empty is a mistake.
			if config, ok := optionalCaseFile(t, path, collectorConfigFile); ok && config == "" {
				t.Errorf("%s is empty", collectorConfigFile)
			}
		})
	}
}

// TestCandidateRun covers which files each kind of case hands the candidate: a
// case with a collector.yaml is configured from it alone, a case without one
// from its conf/, with the case's outputs.conf replaced by [hecout].
func TestCandidateRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, confDir), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"inputs.conf":  "[monitor://foo.txt]\n",
		"outputs.conf": "[httpout]\nuri = ${HEC_ENDPOINT}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, confDir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	casePath := filepath.Join(dir, "test.yaml")

	want := map[string]string{
		otelcol.ConfigFile: splunkInputsConfig,
		"inputs.conf":      "[monitor://foo.txt]\n",
		outputsConf:        hecOutConf,
	}
	if got := candidateRun(t, casePath, "parity_uc").ConfigFiles; !maps.Equal(got, want) {
		t.Errorf("conf-driven case files = %v, want %v", got, want)
	}

	if err := os.WriteFile(filepath.Join(dir, collectorConfigFile), []byte("receivers:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	want = map[string]string{otelcol.ConfigFile: "receivers:\n"}
	if got := candidateRun(t, casePath, "parity_uc").ConfigFiles; !maps.Equal(got, want) {
		t.Errorf("collector.yaml case files = %v, want %v", got, want)
	}
}

// collectorConfigFile is the case file that opts a case out of being configured
// from its conf/.
const collectorConfigFile = "collector.yaml"

// candidateRun builds the candidate's AgentRun: from the case's collector.yaml
// when it has one, otherwise from the same conf/ the oracle reads.
func candidateRun(t *testing.T, casePath, index string) parity.AgentRun {
	t.Helper()
	if config, ok := optionalCaseFile(t, casePath, collectorConfigFile); ok {
		return parity.AgentRun{
			Adapter:     otelcol.New(""),
			ConfigFiles: map[string]string{otelcol.ConfigFile: config},
			Index:       index,
		}
	}
	return parity.AgentRun{
		Adapter:     otelcol.New("").EnableFeatureGates(taRunnerGate),
		ConfigFiles: splunkInputsFiles(caseConf(t, casePath)),
		Index:       index,
	}
}

// outputsConf is the one conf file the candidate does not take from the case.
const outputsConf = "outputs.conf"

// splunkInputsConfig is the collector config for a conf-driven case: the
// components discover every stanza themselves, so no case needs its own.
const splunkInputsConfig = `receivers:
  splunk_inputs:
exporters:
  splunk_outputs:
service:
  pipelines:
    logs:
      receivers: [splunk_inputs]
      exporters: [splunk_outputs]
`

// hecOutConf is the candidate's output side: splunk_outputs skips the oracle's
// [httpout], so it needs a [hecout] the case's conf does not carry.
const hecOutConf = `[hecout]
uri = ${HEC_ENDPOINT}
httpEventCollectorToken = ${HEC_TOKEN}
`

// splunkInputsFiles is a conf-driven case's candidate config: the collector
// config plus the case's own .conf files.
func splunkInputsFiles(confFiles map[string]string) map[string]string {
	files := map[string]string{
		otelcol.ConfigFile: splunkInputsConfig,
		outputsConf:        hecOutConf,
	}
	for name, body := range confFiles {
		if name == outputsConf {
			continue
		}
		files[name] = body
	}
	return files
}

// caseFile reads one of the case's agent config files. Both agents take their
// config from the case directory so a case can vary either side.
func caseFile(t *testing.T, casePath, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(filepath.Dir(casePath), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

// optionalCaseFile reads a case file that a case need not have, reporting
// whether it was there.
func optionalCaseFile(t *testing.T, casePath, name string) (string, bool) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(filepath.Dir(casePath), name))
	if errors.Is(err, os.ErrNotExist) {
		return "", false
	}
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b), true
}

// caseConf reads the case's conf/ directory, the Splunk .conf structure the
// oracle is configured from and, for a case without a collector.yaml, the
// candidate too.
func caseConf(t *testing.T, casePath string) map[string]string {
	t.Helper()
	dir := filepath.Join(filepath.Dir(casePath), confDir)
	paths, err := filepath.Glob(filepath.Join(dir, "*.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no *.conf files in %s", dir)
	}
	files := make(map[string]string, len(paths))
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if len(b) == 0 {
			t.Errorf("%s is empty", p)
		}
		files[filepath.Base(p)] = string(b)
	}
	return files
}

// regenerateGolden runs the UF oracle and writes its events to the golden,
// reduced to the fields the case selects.
func regenerateGolden(ctx context.Context, t *testing.T, c *parity.Case, ufConf map[string]string, index string, backend parity.Backend, goldenPath string, opts parity.RunOptions) {
	t.Helper()
	ufRun := parity.AgentRun{
		Adapter:     uf.New(""),
		ConfigFiles: ufConf,
		Index:       index,
	}
	ufRecs, err := parity.RunAgent(ctx, c, ufRun, backend, opts)
	if err != nil {
		t.Fatalf("run UF oracle: %v", err)
	}
	if len(ufRecs) == 0 {
		t.Fatalf("UF landed no events; not writing an empty %s", goldenPath)
	}
	if err := parity.WriteGolden(goldenPath, ufRecs, c.Expected); err != nil {
		t.Fatalf("write golden: %v", err)
	}
	t.Logf("regenerated %s from UF (%d record(s))", goldenPath, len(ufRecs))
}

func assertMatch(t *testing.T, res parity.Result, golden, candidate []parity.Record) {
	t.Helper()
	if res.Match {
		return
	}
	for _, m := range res.Mismatches {
		t.Errorf("record %d field %q: golden %q, got %q", m.Record, m.Field, m.Expected, m.Actual)
	}
	t.Logf("golden: %+v", golden)
	t.Logf("candidate landed %d record(s): %+v", len(candidate), candidate)
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
