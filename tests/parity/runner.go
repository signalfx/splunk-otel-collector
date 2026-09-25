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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// tokens (BASE_DIR, AGENT_DIR, HEC_ENDPOINT, HEC_TOKEN, INDEX) and writes
	// each into the run's configDir; Adapter.Prepare installs them.
	ConfigFiles map[string]string
	// Index is the Splunk index this agent forwards to.
	Index string
	// Search is the SPL that reads this agent's events back. INDEX is
	// interpolated. Defaults to "search index=INDEX".
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

// RunAgent drives one agent through one case and returns the events it landed in
// Splunk. It owns a fresh sandbox: render configs, run setup, start the agent,
// run the script, poll the backend for this agent's index until the event count
// settles, then tear down.
func RunAgent(ctx context.Context, c *Case, run AgentRun, backend Backend, opts RunOptions) ([]Record, error) {
	opts = opts.withDefaults()

	baseDir, err := os.MkdirTemp("", "parity-"+sanitize(c.Name)+"-")
	if err != nil {
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	defer os.RemoveAll(baseDir)

	hec := backend.HEC()
	tokens := Tokens{
		BaseDir:     baseDir,
		AgentDir:    run.Adapter.InstallDir(),
		HECEndpoint: hec.Endpoint,
		HECToken:    hec.Token,
		Index:       run.Index,
	}

	configDir := filepath.Join(baseDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}
	for name, tmpl := range run.ConfigFiles {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(tokens.apply(tmpl)), 0o600); err != nil {
			return nil, err
		}
	}

	if err := run.Adapter.Prepare(configDir); err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	defer run.Adapter.Cleanup()

	if c.Setup != "" {
		if err := runShell(ctx, opts.Shell, tokens.apply(c.Setup), baseDir); err != nil {
			return nil, fmt.Errorf("setup: %w", err)
		}
	}

	if err := run.Adapter.Start(ctx); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	defer run.Adapter.Stop(ctx)

	if c.Script != "" {
		if err := runShell(ctx, opts.Shell, tokens.apply(c.Script), baseDir); err != nil {
			return nil, fmt.Errorf("script: %w", err)
		}
	}

	spl := run.Search
	if spl == "" {
		spl = "search index=INDEX"
	}
	spl = tokens.apply(spl)
	return waitForEvents(ctx, backend, spl, opts)
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
