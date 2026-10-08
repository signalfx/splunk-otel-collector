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

package parity

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AgentRun bundles one agent's participation in a case: how to drive it, the
// agent-native config files it needs, and the Splunk index it forwards to. The
// two agents in a parity run deliberately use different config formats (UF
// .conf vs collector config.yaml) to reach the same result, so each supplies
// its own ConfigFiles. Keeping them on separate indexes lets both run without a
// clean-between step.
type AgentRun struct {
	Adapter Adapter
	// ConfigFiles maps a filename to its template. The runner interpolates the
	// tokens (${BASE_DIR}, ${HEC_ENDPOINT}, ${HEC_TOKEN}, ${INDEX}) and writes
	// each into the run's configDir; Adapter.Prepare installs them.
	ConfigFiles map[string]string
	// Index is the Splunk index this agent forwards to.
	Index string
	// Search is the SPL that reads this agent's events back. INDEX is
	// interpolated. Defaults to "search index=${INDEX}".
	Search string
}

// RunOptions tunes a run.
type RunOptions struct {
	Shell      string
	Quiescence time.Duration
	Timeout    time.Duration
	MinEvents  int
}

func (o RunOptions) withDefaults() RunOptions {
	if o.Quiescence == 0 {
		o.Quiescence = 3 * time.Second
	}
	if o.Timeout == 0 {
		o.Timeout = 90 * time.Second
	}
	if o.MinEvents == 0 {
		o.MinEvents = 1
	}
	if o.Shell == "" {
		o.Shell = "bash"
	}
	return o
}

// RunCase runs both agents through a case and validates candidate against
// oracle, comparing them on the fields the case selects. It is the direct
// UF-vs-candidate path, where both sides are captured in the same run rather
// than one being replayed from a golden.
func RunCase(ctx context.Context, c *Case, backend Backend, oracle, candidate AgentRun, v Validator, opts RunOptions) (Result, error) {
	candidateCapture, err := RunAgent(ctx, c, candidate, backend, opts)
	if err != nil {
		return Result{}, fmt.Errorf("candidate %s: %w", candidate.Adapter.Name(), err)
	}
	oracleCapture, err := RunAgent(ctx, c, oracle, backend, opts)
	if err != nil {
		return Result{}, fmt.Errorf("oracle %s: %w", oracle.Adapter.Name(), err)
	}
	// The reference is projected so only the selected fields take part; the
	// validator then ignores everything the candidate has beyond them.
	reference := oracleCapture.Records
	for i, r := range reference {
		reference[i] = project(r, c.Expected)
	}
	res := v.Validate(reference, candidateCapture.Records)

	// The validator only knows about records, so the observation is compared
	// here. Leaving it out would make a case's observe hook silently inert on
	// this path.
	if c.Observe != "" && candidateCapture.Observation != oracleCapture.Observation {
		res.Mismatches = append(res.Mismatches, Mismatch{
			Record:   -1,
			Field:    "observation",
			Expected: oracleCapture.Observation,
			Actual:   candidateCapture.Observation,
		})
		res.Match = false
	}
	return res, nil
}

// Capture is what one agent run produced: the events it landed in Splunk and,
// for a case with an observe hook, that hook's settled output.
type Capture struct {
	Records     []Record
	Observation string
}

// RunAgent drives one agent through one case and returns what it produced. It
// owns a fresh sandbox: render configs, run setup, start the agent, run the
// script, poll the backend for this agent's index until the event count settles,
// read the observe hook, then tear down.
func RunAgent(ctx context.Context, c *Case, run AgentRun, backend Backend, opts RunOptions) (Capture, error) {
	opts = opts.withDefaults()

	baseDir, err := os.MkdirTemp("", "parity-"+sanitize(c.Name)+"-")
	if err != nil {
		return Capture{}, fmt.Errorf("sandbox: %w", err)
	}
	defer os.RemoveAll(baseDir)

	hec := backend.HEC()
	tokens := Tokens{
		BaseDir:     baseDir,
		HECEndpoint: hec.Endpoint,
		HECToken:    hec.Token,
		Index:       run.Index,
	}

	configDir := filepath.Join(baseDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return Capture{}, err
	}
	for name, tmpl := range run.ConfigFiles {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(tokens.apply(tmpl)), 0o600); err != nil {
			return Capture{}, err
		}
	}

	if err := run.Adapter.Prepare(configDir); err != nil {
		return Capture{}, fmt.Errorf("prepare: %w", err)
	}
	defer run.Adapter.Cleanup()

	if c.Setup != "" {
		if err := runShell(ctx, opts.Shell, tokens.apply(c.Setup), baseDir); err != nil {
			return Capture{}, fmt.Errorf("setup: %w", err)
		}
	}

	if err := run.Adapter.Start(ctx); err != nil {
		return Capture{}, fmt.Errorf("start: %w", err)
	}
	defer run.Adapter.Stop(ctx)

	if c.Script != "" {
		if err := runShell(ctx, opts.Shell, tokens.apply(c.Script), baseDir); err != nil {
			return Capture{}, fmt.Errorf("script: %w", err)
		}
	}

	spl := run.Search
	if spl == "" {
		spl = "search index=${INDEX}"
	}
	spl = tokens.apply(spl)
	recs, err := waitForEvents(ctx, backend, spl, opts)
	capture := Capture{Records: recs}
	if err != nil {
		return capture, err
	}

	// The agent stays running: the hook reports what it did while reading, not
	// what it cleaned up on shutdown.
	if c.Observe != "" {
		obs, err := waitForObservation(ctx, tokens.apply(c.Observe), baseDir, opts)
		if err != nil {
			return capture, fmt.Errorf("observe: %w", err)
		}
		capture.Observation = obs
	}
	return capture, nil
}

// observePoll is how often an observe hook is re-run while waiting for its
// output to settle.
const observePoll = 500 * time.Millisecond

// waitForObservation runs the case's observe hook until its output is unchanged
// for Quiescence, or Timeout elapses, and returns that output with surrounding
// whitespace trimmed.
//
// It polls because an effect can trail the events that preceded it: fileconsumer
// emits a batched file's lines before it unlinks the file, so the events can go
// quiet while the deletion is still pending. Polling until the output settles is
// the same shape as the event capture loop, and it runs identically for both
// agents, so the oracle's reference and the candidate's value are produced the
// same way.
//
// The hook reports through stdout and must exit 0; a non-zero exit is a broken
// hook, not a failed assertion. Comparing the output against the oracle's is
// what makes the assertion, and that happens in the caller.
func waitForObservation(ctx context.Context, script, dir string, opts RunOptions) (string, error) {
	deadline := time.Now().Add(opts.Timeout)
	out, err := observeOnce(ctx, opts.Shell, script, dir)
	if err != nil {
		return "", err
	}
	stableSince := time.Now()
	for time.Since(stableSince) < opts.Quiescence && !time.Now().After(deadline) {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		time.Sleep(observePoll)
		next, err := observeOnce(ctx, opts.Shell, script, dir)
		if err != nil {
			return out, err
		}
		if next != out {
			out = next
			stableSince = time.Now()
		}
	}
	return out, nil
}

// observeOnce runs the hook once and returns its trimmed stdout. stderr is kept
// out of the observation so a warning on it cannot change the comparison.
func observeOnce(ctx context.Context, shell, script, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, shell, "-c", script)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// waitForEvents polls Search until the event count is >= MinEvents and stable
// for Quiescence, or Timeout elapses. It returns the last capture.
func waitForEvents(ctx context.Context, backend Backend, spl string, opts RunOptions) ([]Record, error) {
	deadline := time.Now().Add(opts.Timeout)
	poll := 1 * time.Second

	var last []Record
	lastCount := -1
	lastChange := time.Now()
	for {
		if ctx.Err() != nil {
			return last, ctx.Err()
		}
		recs, err := backend.Search(ctx, spl)
		if err == nil {
			last = recs
			if len(recs) != lastCount {
				lastCount = len(recs)
				lastChange = time.Now()
			}
			if lastCount >= opts.MinEvents && time.Since(lastChange) >= opts.Quiescence {
				return last, nil
			}
		}
		if time.Now().After(deadline) {
			return last, nil
		}
		time.Sleep(poll)
	}
}

func runShell(ctx context.Context, shell, script, dir string) error {
	cmd := exec.CommandContext(ctx, shell, "-c", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func sanitize(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '-')
		}
	}
	return string(b)
}
