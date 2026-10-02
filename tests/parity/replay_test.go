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
	"flag"
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

// Candidate modes: how the collector is told what to do. The golden is the same
// either way, since it records what the oracle landed for the case's input, not
// how the candidate was configured.
const (
	// collectorYAMLMode configures the candidate from the case's hand-written
	// collector.yaml. It asserts the golden is reachable with native collector
	// config, which is what tells a failure of the .conf path apart from a
	// golden no collector config can reach.
	collectorYAMLMode = "collector-yaml"
	// splunkInputsMode configures the candidate from the case's own conf/ via
	// the splunk_inputs receiver and splunk_outputs exporter, so a run covers
	// the .conf translation end to end instead of a hand-written equivalent.
	splunkInputsMode = "splunk-inputs"
)

// candidate selects the mode. Only one runs per invocation: the modes of a case
// share its index, which is what lets a case compare index like any other
// field, so a second mode in the same run would read back the first one's
// events. CI runs a job per mode.
var candidate = flag.String("candidate", collectorYAMLMode,
	"how to configure the otelcol candidate: "+collectorYAMLMode+" or "+splunkInputsMode)

// taRunnerGate registers splunk_inputs and splunk_outputs. It is alpha, so
// without it the components do not exist and the collector fails to start.
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
		if *candidate != collectorYAMLMode && *candidate != splunkInputsMode {
			t.Fatalf("unknown -candidate %q: want %s or %s", *candidate, collectorYAMLMode, splunkInputsMode)
		}
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
			for _, name := range []string{"collector.yaml", parity.GoldenFile} {
				if caseFile(t, path, name) == "" {
					t.Errorf("%s is empty", name)
				}
			}
		})
	}
}

// candidateRun builds the candidate's AgentRun for a case in the selected mode.
func candidateRun(t *testing.T, casePath, index string) parity.AgentRun {
	t.Helper()
	if *candidate == splunkInputsMode {
		return parity.AgentRun{
			Adapter:     otelcol.New("").EnableFeatureGates(taRunnerGate),
			ConfigFiles: splunkInputsFiles(caseConf(t, casePath)),
			Index:       index,
		}
	}
	// collector.yaml is the case-directory name; the adapter is handed it as
	// otelcol.ConfigFile.
	return parity.AgentRun{
		Adapter:     otelcol.New(""),
		ConfigFiles: map[string]string{otelcol.ConfigFile: caseFile(t, casePath, "collector.yaml")},
		Index:       index,
	}
}

// splunkConfRoot is the candidate's $SPLUNK_HOME, relative to the run's config
// directory: splunk_inputs and splunk_outputs search <root>/etc for conf files.
const splunkConfRoot = "splunkhome"

// systemLocalDir is the conf layer the case's files are installed into, the
// same one the UF adapter installs them into on the oracle side.
const systemLocalDir = splunkConfRoot + "/etc/system/local"

// outputsConf is the one conf file the candidate does not take from the case.
const outputsConf = "outputs.conf"

// splunkInputsConfig is the candidate's collector config for splunkInputsMode.
// It is owned here rather than per case because both components take a single
// base_dir and discover the stanzas themselves, so there is nothing a case
// could vary, and keeping it out of the case directory means a case cannot pin
// the translation it exists to test.
const splunkInputsConfig = `receivers:
  splunk_inputs:
    base_dir: ${CONFIG_DIR}/` + splunkConfRoot + `
exporters:
  splunk_outputs:
    base_dir: ${CONFIG_DIR}/` + splunkConfRoot + `
service:
  pipelines:
    logs:
      receivers: [splunk_inputs]
      exporters: [splunk_outputs]
`

// hecOutConf is the candidate's output side. The oracle ships events with
// outputs.conf [httpout], a kind splunk_outputs skips, so the candidate needs a
// [hecout] stanza the case's conf does not carry. Rendering it here keeps every
// case from repeating it, and it replaces the case's outputs.conf rather than
// merging with it so there is no question which stanza won.
const hecOutConf = `[hecout]
uri = ${HEC_ENDPOINT}
httpEventCollectorToken = ${HEC_TOKEN}
`

// splunkInputsFiles lays out the candidate's config for splunkInputsMode: the
// collector config, plus the case's conf/ materialized as a $SPLUNK_HOME tree
// the components can search.
func splunkInputsFiles(confFiles map[string]string) map[string]string {
	files := map[string]string{
		otelcol.ConfigFile:                 splunkInputsConfig,
		systemLocalDir + "/" + outputsConf: hecOutConf,
	}
	for name, body := range confFiles {
		if name == outputsConf {
			continue
		}
		files[systemLocalDir+"/"+name] = body
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

// caseConf reads the case's conf/ directory, the Splunk .conf structure the
// oracle is configured from and, in splunkInputsMode, the candidate too.
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
