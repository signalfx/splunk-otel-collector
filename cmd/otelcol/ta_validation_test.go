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

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validationModeXML is a minimal <items> XML document sent by Splunk on stdin
// when invoking the modular input with --validate-arguments.
const validationModeXML = `<items>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<item name="Splunk_TA_OTel_Collector://default">
		<param name="splunk_realm">us0</param>
		<param name="splunk_access_token">test_token</param>
	</item>
</items>`

// setupValidationModeTest sets up the package-level variables needed to run
// runFromCmdLine in TA validation mode during tests. It returns a *bytes.Buffer
// that captures anything written to stdoutWriter, and a cleanup function.
func setupValidationModeTest(t *testing.T, stdinData string) *bytes.Buffer {
	t.Helper()

	// Redirect stdin to the provided XML
	origStdin := stdinReader
	stdinReader = strings.NewReader(stdinData)
	t.Cleanup(func() { stdinReader = origStdin })

	// Capture stdout
	var buf bytes.Buffer
	origStdout := stdoutWriter
	stdoutWriter = &buf
	t.Cleanup(func() { stdoutWriter = origStdout })

	// Replace os.Exit so the test does not terminate the process
	origExit := exitFn
	exitFn = func(code int) {
		// Convert the os.Exit call into a panic that the test can catch.
		panic(exitCodePanic(code))
	}
	t.Cleanup(func() { exitFn = origExit })

	// Force TA modular-input mode without requiring a real splunkd parent process
	t.Setenv("SPLUNK_HOME", "/opt/splunk")
	t.Setenv("SPLUNK_TA_FORCE_MODULAR_INPUT", "1")

	return &buf
}

// exitCodePanic is used to distinguish an intercepted os.Exit call from other panics.
type exitCodePanic int

// runValidationModeExpectingExit calls runFromCmdLine and expects it to call
// exitFn(1). It returns the exit code that was passed to exitFn.
func runValidationModeExpectingExit(t *testing.T, args []string) {
	t.Helper()
	require.PanicsWithValue(t, exitCodePanic(1), func() {
		runFromCmdLine(args)
	}, "expected runFromCmdLine to call exitFn(1) in validation mode")
}

// TestRunFromCmdLine_ValidationMode_SettingsNewError verifies that when running
// as a TA in --validate-arguments mode, an error from settings.New is written
// as a Splunk XML validation error to stdout (not logged to stderr).
func TestRunFromCmdLine_ValidationMode_SettingsNewError(t *testing.T) {
	buf := setupValidationModeTest(t, validationModeXML)

	// Do NOT set SPLUNK_CONFIG or SPLUNK_CONFIG_YAML, and clear any env vars
	// that would let settings.New find a default config.
	t.Setenv("SPLUNK_CONFIG", "")
	t.Setenv("SPLUNK_CONFIG_YAML", "")
	t.Setenv("SPLUNK_REALM", "")
	t.Setenv("SPLUNK_ACCESS_TOKEN", "")

	// validateTAArguments sets args[1] = "validate", so settings.New receives
	// ["validate"] with no config source — it must fail.
	args := []string{"otelcol", "--validate-arguments"}
	runValidationModeExpectingExit(t, args)

	output := buf.String()
	assert.Contains(t, output, "<error>", "expected XML error element in stdout")
	assert.Contains(t, output, "<message>", "expected XML message element in stdout")
	// Ensure the error is not an empty message
	assert.NotContains(t, output, "<message></message>", "error message must not be empty")
}

// TestRunFromCmdLine_ValidationMode_RunError verifies that when running as a TA
// in --validate-arguments mode, an error from run(serviceSettings) is written
// as a Splunk XML validation error to stdout (not logged to stderr).
func TestRunFromCmdLine_ValidationMode_RunError(t *testing.T) {
	// Provide a SPLUNK_CONFIG_YAML with structurally invalid collector config so
	// that settings.New succeeds (a YAML value is present) but the collector's
	// "validate" sub-command rejects it.
	const invalidCollectorConfig = `
service:
  pipelines:
    traces:
      receivers: [nonexistent_receiver]
      exporters: [nonexistent_exporter]
`
	buf := setupValidationModeTest(t, validationModeXML)
	t.Setenv("SPLUNK_CONFIG_YAML", invalidCollectorConfig)
	t.Setenv("SPLUNK_REALM", "us0")
	t.Setenv("SPLUNK_ACCESS_TOKEN", "test_token")

	args := []string{"otelcol", "--validate-arguments"}
	runValidationModeExpectingExit(t, args)

	output := buf.String()
	assert.Contains(t, output, "<error>", "expected XML error element in stdout")
	assert.Contains(t, output, "<message>", "expected XML message element in stdout")
	assert.NotContains(t, output, "<message></message>", "error message must not be empty")
}
