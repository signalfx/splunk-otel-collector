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

package cyberarkconfigsource

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// outputDelimiter is a rare token used to join the requested -o fields on a single line so
// they can be split back apart positionally. It must not appear inside credential values.
const outputDelimiter = "@#@"

// defaultFolder is used when no folder is configured.
const defaultFolder = "Root"

// outputField couples a logical selector name (exposed to references, e.g. "UserName")
// with the CyberArk output attribute requested via the -o flag (e.g. "PassProps.UserName").
type outputField struct {
	// key is the logical name selected by a reference, e.g. ${cyberark:UserName}.
	key string
	// attr is the CyberArk attribute name passed to CLIPasswordSDK's -o flag.
	attr string
}

// outputFields is the fixed, ordered list of attributes requested from CLIPasswordSDK.
// The order is significant: the CLI emits values in this order joined by outputDelimiter,
// and parseOutput maps them back positionally.
//
// Parsing assumption: CLIPasswordSDK emits every requested -o field, in order, on a single
// line; attributes that are unset on the object still emit an empty positional value so the
// field count stays stable. A configurable field list is a possible future extension.
var outputFields = []outputField{
	{key: "Password", attr: "Password"},
	{key: "UserName", attr: "PassProps.UserName"},
	{key: "Address", attr: "PassProps.Address"},
	{key: "Database", attr: "PassProps.Database"},
	{key: "Port", attr: "PassProps.Port"},
	{key: "Name", attr: "Name"},
	{key: "Safe", attr: "Safe"},
	{key: "Folder", attr: "Folder"},
}

// commandRunner runs an external command and returns its stdout. It is injectable so tests
// can exercise the retriever without the real CLIPasswordSDK binary.
type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Error wrapper types to help with testability.
type (
	errCLIExec     struct{ error }
	errParseOutput struct{ error }
)

// cpRetriever fetches a CyberArk object via the local CLIPasswordSDK (Credential Provider).
type cpRetriever struct {
	runner     commandRunner
	binaryPath string
	appID      string
	safe       string
	folder     string
	object     string
}

func newCPRetriever(cfg *Config) *cpRetriever {
	return &cpRetriever{
		binaryPath: cfg.BinaryPath,
		appID:      cfg.AppID,
		safe:       cfg.Safe,
		folder:     cfg.Folder,
		object:     cfg.Object,
		runner:     execCommandRunner,
	}
}

// buildArgs constructs the CLIPasswordSDK GetPassword argument list.
func (r *cpRetriever) buildArgs() []string {
	folder := r.folder
	if folder == "" {
		folder = defaultFolder
	}

	attrs := make([]string, len(outputFields))
	for i, f := range outputFields {
		attrs[i] = f.attr
	}

	return []string{
		"GetPassword",
		"-p", "AppDescs.AppID=" + r.appID,
		"-p", fmt.Sprintf("Query=Safe=%s;Folder=%s;Object=%s", r.safe, folder, r.object),
		"-o", strings.Join(attrs, ","),
		"-d", outputDelimiter,
	}
}

func (r *cpRetriever) retrieve(ctx context.Context) (map[string]any, error) {
	out, err := r.runner(ctx, r.binaryPath, r.buildArgs()...)
	if err != nil {
		return nil, &errCLIExec{fmt.Errorf("CLIPasswordSDK invocation failed: %w", err)}
	}

	return parseOutput(out)
}

// parseOutput splits the delimiter-joined CLI output positionally and maps each value to its
// logical field key. A count mismatch is treated as an error rather than misassigning values.
func parseOutput(out []byte) (map[string]any, error) {
	line := strings.TrimRight(string(out), "\r\n")
	parts := strings.Split(line, outputDelimiter)
	if len(parts) != len(outputFields) {
		return nil, &errParseOutput{fmt.Errorf("expected %d fields in CLIPasswordSDK output, got %d", len(outputFields), len(parts))}
	}

	fields := make(map[string]any, len(outputFields))
	for i, f := range outputFields {
		fields[f.key] = parts[i]
	}
	return fields, nil
}

// execCommandRunner is the production commandRunner backed by os/exec.
func execCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	// #nosec G204 -- name and args are derived from validated, static collector config,
	// not from untrusted input.
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
