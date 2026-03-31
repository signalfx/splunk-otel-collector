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

package modularinput

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLaunchAsTA_NotTARunMode(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return false
	isParentProcessSplunkdFn = func() bool {
		return false
	}

	// Ensure SPLUNK_HOME is not set
	os.Unsetenv("SPLUNK_HOME")

	args := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(args, nil, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error when not in modular input mode")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged when not in modular input mode")
}

func TestHandleLaunchAsTA_QueryModeScheme(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalStdout := stdoutWriter
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		stdoutWriter = originalStdout
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Capture stdout
	var buf bytes.Buffer
	stdoutWriter = &buf

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	const schemeContent = "<scheme><title>Test</title></scheme>"
	args := []string{"program", "--scheme"}
	resultArgs, mode, err := HandleLaunchAsTA(args, nil, "test-stanza", schemeContent, nil)
	require.NoError(t, err)
	assert.Equal(t, IntrospectionTARunMode, mode, "Expected IntrospectionTARunMode for --scheme argument")
	assert.Nil(t, resultArgs, "Expected nil args in introspection mode")
	assert.Contains(t, buf.String(), schemeContent, "Expected scheme content written to stdout in introspection mode")
}

func TestHandleLaunchAsTA_QueryModeValidate(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	// When validator is nil, validation mode skips validation and proceeds normally.
	args := []string{"program", "--validate-arguments"}
	resultArgs, mode, err := HandleLaunchAsTA(args, strings.NewReader(""), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error when validator is nil in validation mode")
	assert.Equal(t, ValidationTARunMode, mode, "Expected ValidationTARunMode for --validate-arguments argument")
	assert.Equal(t, args, resultArgs, "Expected args returned unchanged when validator is nil")
}

func TestHandleLaunchAsTA_InvalidXML(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program"}
	invalidXML := strings.NewReader("<input><invalid>")

	resultArgs, _, err := HandleLaunchAsTA(args, invalidXML, "test-stanza", "<scheme></scheme>", nil)
	require.Error(t, err, "Expected error for invalid XML")
	assert.Nil(t, resultArgs, "Expected nil args on error")
	assert.Contains(t, err.Error(), "launch as TA failed to read modular input XML from stdin")
}

func TestHandleLaunchAsTA_Success(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_api_key">secret123</param>
			<param name="splunk_endpoint">https://api.example.com</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	assert.Equal(t, "secret123", envVars["SPLUNK_API_KEY"], "Expected SPLUNK_API_KEY to be set")
	assert.Equal(t, "https://api.example.com", envVars["SPLUNK_ENDPOINT"], "Expected SPLUNK_ENDPOINT to be set")
}

func TestHandleLaunchAsTA_SuccessWithEnvExpansion(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME and a test env var
	t.Setenv("SPLUNK_HOME", "/opt/splunk")
	t.Setenv("TEST_VAR", "expanded_value")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_config_value">$TEST_VAR</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	assert.Equal(t, "expanded_value", envVars["SPLUNK_CONFIG_VALUE"], "Expected SPLUNK_CONFIG_VALUE to be expanded")
}

func TestHandleLaunchAsTA_SetEnvError(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Mock setEnv to return an error
	expectedErr := errors.New("setenv failed")
	setEnvFn = func(_, _ string) error {
		return expectedErr
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_api_key">secret123</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.Error(t, err, "Expected error")
	assert.Nil(t, resultArgs, "Expected nil args on error")
	assert.Contains(t, err.Error(), "launch as TA failed to set environment variable")
	assert.ErrorIs(t, err, expectedErr, "Expected wrapped error")
}

func TestHandleLaunchAsTA_StanzaNotFound(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track that setEnv is never called
	setEnvCalled := false
	setEnvFn = func(_, _ string) error {
		setEnvCalled = true
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="other-stanza" app="test-app">
			<param name="api_key">secret123</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error when stanza not found")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged when stanza not found")
	assert.False(t, setEnvCalled, "Expected setEnv to not be called when stanza not found")
}

func TestHandleLaunchAsTA_EmptyStanza(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track that setEnv is never called
	setEnvCalled := false
	setEnvFn = func(_, _ string) error {
		setEnvCalled = true
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error with empty stanza")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged with empty stanza")
	assert.False(t, setEnvCalled, "Expected setEnv to not be called with empty stanza")
}

func TestHandleLaunchAsTA_ReadError(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	// Create a reader that always returns an error
	errorReader := &errorReader{err: errors.New("read error")}

	args := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(args, errorReader, "test-stanza", "<scheme></scheme>", nil)
	require.Error(t, err, "Expected error for read failure")
	assert.Nil(t, resultArgs, "Expected nil args on error")
	assert.Contains(t, err.Error(), "launch as TA failed to read modular input XML from stdin")
}

func TestIsModularInputMode_NoSplunkHome(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Ensure SPLUNK_HOME is not set
	os.Unsetenv("SPLUNK_HOME")

	args := []string{"program"}
	mode := detectTARunMode(args)
	assert.Equal(t, NotTARunMode, mode, "Expected NotTARunMode mode without SPLUNK_HOME")
}

func TestIsModularInputMode_TAv1Launch(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")
	t.Setenv("SPLUNK_OTEL_TA_HOME", "/opt/splunk/etc/apps/Splunk_TA_otel")

	args := []string{"program"}
	mode := detectTARunMode(args)
	assert.Equal(t, NotTARunMode, mode, "Expected NotTARunMode mode when TA v1 is being launched")
}

func TestIsModularInputMode_ParentNotSplunkd(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return false
	isParentProcessSplunkdFn = func() bool {
		return false
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program"}
	mode := detectTARunMode(args)
	assert.Equal(t, NotTARunMode, mode, "Expected NotTARunMode mode when parent is not splunkd")
}

func TestIsModularInputMode_Normal(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program"}
	mode := detectTARunMode(args)
	assert.Equal(t, ExecutionTARunMode, mode, "Expected ExecutionTARunMode")
}

func TestIsModularInputMode_Scheme(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program", "--scheme"}
	mode := detectTARunMode(args)
	assert.Equal(t, IntrospectionTARunMode, mode, "Expected IntrospectionTARunMode for --scheme")
}

func TestIsModularInputMode_ValidateArguments(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program", "--validate-arguments"}
	mode := detectTARunMode(args)
	assert.Equal(t, ValidationTARunMode, mode, "Expected ValidationTARunMode for --validate-arguments")
}

func TestIsModularInputMode_OtherArguments(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program", "--other-flag"}
	mode := detectTARunMode(args)
	assert.Equal(t, ExecutionTARunMode, mode, "Expected ExecutionTARunMode for other arguments")
}

func TestIsModularInputMode_MultipleArguments(t *testing.T) {
	// Save original function and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	args := []string{"program", "--scheme", "--other-flag"}
	mode := detectTARunMode(args)
	assert.Equal(t, ExecutionTARunMode, mode, "Expected ExecutionTARunMode for more than 2 arguments")
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct {
	err error
}

func (er *errorReader) Read(_ []byte) (n int, err error) {
	return 0, er.err
}

var _ io.Reader = (*errorReader)(nil)

func TestHandleLaunchAsTA_TwoPassFiltering(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_realm">us0</param>
			<param name="splunk_access_token">secret123</param>
			<param name="other_param">should_not_be_set</param>
			<param name="api_key">also_not_set</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// Only splunk_ prefixed parameters should be set
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"], "Expected SPLUNK_REALM to be set")
	assert.Equal(t, "secret123", envVars["SPLUNK_ACCESS_TOKEN"], "Expected SPLUNK_ACCESS_TOKEN to be set")
	assert.NotContains(t, envVars, "OTHER_PARAM", "Expected OTHER_PARAM to not be set")
	assert.NotContains(t, envVars, "API_KEY", "Expected API_KEY to not be set")
}

func TestHandleLaunchAsTA_DependencyOrdering(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set and their order
	var setOrder []string
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		setOrder = append(setOrder, key)
		envVars[key] = value
		// Simulate setting in the actual environment for expansion
		t.Setenv(key, value)
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_ingest_url">https://ingest.${SPLUNK_REALM}.splunkcloud.com</param>
			<param name="splunk_realm">us0</param>
			<param name="splunk_api_url">https://api.${SPLUNK_REALM}.splunkcloud.com</param>
			<param name="splunk_access_token">secret123</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// Verify all variables are set
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"], "Expected SPLUNK_REALM to be set")
	assert.Equal(t, "secret123", envVars["SPLUNK_ACCESS_TOKEN"], "Expected SPLUNK_ACCESS_TOKEN to be set")
	assert.Equal(t, "https://ingest.us0.splunkcloud.com", envVars["SPLUNK_INGEST_URL"], "Expected SPLUNK_INGEST_URL to be expanded")
	assert.Equal(t, "https://api.us0.splunkcloud.com", envVars["SPLUNK_API_URL"], "Expected SPLUNK_API_URL to be expanded")

	// Verify ordering: variables without dependencies should be set first
	// Find positions in setOrder
	realmPos := -1
	tokenPos := -1
	ingestPos := -1
	apiPos := -1
	for i, name := range setOrder {
		switch name {
		case "SPLUNK_REALM":
			realmPos = i
		case "SPLUNK_ACCESS_TOKEN":
			tokenPos = i
		case "SPLUNK_INGEST_URL":
			ingestPos = i
		case "SPLUNK_API_URL":
			apiPos = i
		}
	}

	// Variables without dependencies (SPLUNK_REALM, SPLUNK_ACCESS_TOKEN) should come before
	// variables with dependencies (SPLUNK_INGEST_URL, SPLUNK_API_URL)
	assert.Less(t, realmPos, ingestPos, "SPLUNK_REALM should be set before SPLUNK_INGEST_URL")
	assert.Less(t, realmPos, apiPos, "SPLUNK_REALM should be set before SPLUNK_API_URL")
	assert.Less(t, tokenPos, ingestPos, "SPLUNK_ACCESS_TOKEN should be set before SPLUNK_INGEST_URL")
	assert.Less(t, tokenPos, apiPos, "SPLUNK_ACCESS_TOKEN should be set before SPLUNK_API_URL")
}

func TestHandleLaunchAsTA_MixedCaseSplunkPrefix(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="SPLUNK_Realm">us0</param>
			<param name="Splunk_Access_Token">secret123</param>
			<param name="splunk_trace_url">http://localhost:9411</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// All should be converted to uppercase
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"], "Expected SPLUNK_REALM to be set")
	assert.Equal(t, "secret123", envVars["SPLUNK_ACCESS_TOKEN"], "Expected SPLUNK_ACCESS_TOKEN to be set")
	assert.Equal(t, "http://localhost:9411", envVars["SPLUNK_TRACE_URL"], "Expected SPLUNK_TRACE_URL to be set")
}

func TestHandleLaunchAsTA_ComplexDependencies(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		// Simulate setting in the actual environment for expansion
		t.Setenv(key, value)
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_url">${SPLUNK_PROTOCOL}://${SPLUNK_HOST}:${SPLUNK_PORT}</param>
			<param name="splunk_protocol">https</param>
			<param name="splunk_host">api.example.com</param>
			<param name="splunk_port">8088</param>
			<param name="splunk_token">abc123</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	resultArgs, _, err := HandleLaunchAsTA(args, reader, "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// Verify all variables are set correctly
	assert.Equal(t, "https", envVars["SPLUNK_PROTOCOL"], "Expected SPLUNK_PROTOCOL to be set")
	assert.Equal(t, "api.example.com", envVars["SPLUNK_HOST"], "Expected SPLUNK_HOST to be set")
	assert.Equal(t, "8088", envVars["SPLUNK_PORT"], "Expected SPLUNK_PORT to be set")
	assert.Equal(t, "abc123", envVars["SPLUNK_TOKEN"], "Expected SPLUNK_TOKEN to be set")
	assert.Equal(t, "https://api.example.com:8088", envVars["SPLUNK_URL"], "Expected SPLUNK_URL to be fully expanded")
}

func TestSetEnvVarsInOrder_NoDependencies(t *testing.T) {
	envVars := map[string]string{
		"SPLUNK_REALM":        "us0",
		"SPLUNK_ACCESS_TOKEN": "secret123",
		"SPLUNK_TRACE_URL":    "http://localhost:9411",
	}

	// Track environment variables set
	setVars := make(map[string]string)
	mockSetEnv := func(key, value string) error {
		setVars[key] = value
		return nil
	}

	err := setEnvVarsInOrder(envVars, mockSetEnv)
	require.NoError(t, err, "Expected no error")

	assert.Equal(t, "us0", setVars["SPLUNK_REALM"])
	assert.Equal(t, "secret123", setVars["SPLUNK_ACCESS_TOKEN"])
	assert.Equal(t, "http://localhost:9411", setVars["SPLUNK_TRACE_URL"])
}

func TestSetEnvVarsInOrder_WithDependencies(t *testing.T) {
	envVars := map[string]string{
		"SPLUNK_REALM":      "us0",
		"SPLUNK_INGEST_URL": "https://ingest.${SPLUNK_REALM}.splunkcloud.com",
	}

	// Track environment variables set and their order
	var setOrder []string
	setVars := make(map[string]string)
	mockSetEnv := func(key, value string) error {
		setOrder = append(setOrder, key)
		setVars[key] = value
		// Simulate setting in the actual environment for expansion
		t.Setenv(key, value)
		return nil
	}

	err := setEnvVarsInOrder(envVars, mockSetEnv)
	require.NoError(t, err, "Expected no error")

	// Verify realm is set first
	assert.Equal(t, "SPLUNK_REALM", setOrder[0], "SPLUNK_REALM should be set first")
	assert.Equal(t, "SPLUNK_INGEST_URL", setOrder[1], "SPLUNK_INGEST_URL should be set second")

	// Verify values
	assert.Equal(t, "us0", setVars["SPLUNK_REALM"])
	assert.Equal(t, "https://ingest.us0.splunkcloud.com", setVars["SPLUNK_INGEST_URL"])
}

func TestSetEnvVarsInOrder_SetEnvError(t *testing.T) {
	envVars := map[string]string{
		"SPLUNK_REALM": "us0",
	}

	expectedErr := errors.New("setenv failed")
	mockSetEnv := func(_, _ string) error {
		return expectedErr
	}

	err := setEnvVarsInOrder(envVars, mockSetEnv)
	require.Error(t, err, "Expected error")
	assert.Contains(t, err.Error(), "launch as TA failed to set environment variable")
	assert.ErrorIs(t, err, expectedErr)
}

func TestSetEnvVarsInOrder_EmptyMap(t *testing.T) {
	envVars := make(map[string]string)

	called := false
	mockSetEnv := func(_, _ string) error {
		called = true
		return nil
	}

	err := setEnvVarsInOrder(envVars, mockSetEnv)
	require.NoError(t, err, "Expected no error")
	assert.False(t, called, "setEnv should not be called for empty map")
}

func TestHandleLaunchAsTA_StanzaPrefixMatch(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="otel://my-instance-123" app="test-app">
			<param name="splunk_realm">us0</param>
			<param name="splunk_access_token">secret123</param>
		</stanza>
		<stanza name="other://instance" app="test-app">
			<param name="splunk_realm">eu0</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	// Use prefix "otel://" to match only the first stanza
	resultArgs, _, err := HandleLaunchAsTA(args, reader, "otel://", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// Should match the first stanza only
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"], "Expected SPLUNK_REALM from first stanza")
	assert.Equal(t, "secret123", envVars["SPLUNK_ACCESS_TOKEN"], "Expected SPLUNK_ACCESS_TOKEN from first stanza")
}

func TestHandleLaunchAsTA_StanzaPrefixNoMatch(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track that setEnv is never called
	setEnvCalled := false
	setEnvFn = func(_, _ string) error {
		setEnvCalled = true
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="otel://instance1" app="test-app">
			<param name="splunk_realm">us0</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	// Use prefix that doesn't match
	resultArgs, _, err := HandleLaunchAsTA(args, reader, "nonexistent://", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error when prefix doesn't match")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged when prefix doesn't match")
	assert.False(t, setEnvCalled, "Expected setEnv to not be called when prefix doesn't match")
}

func TestParseEnvVarPairs_Empty(t *testing.T) {
	result, err := parseEnvVarPairs("")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseEnvVarPairs_SinglePair(t *testing.T) {
	result, err := parseEnvVarPairs("MY_KEY=my_value")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"MY_KEY": "my_value"}, result)
}

func TestParseEnvVarPairs_MultiplePairs(t *testing.T) {
	result, err := parseEnvVarPairs("KEY1=value1,KEY2=value2,KEY3=value3")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}, result)
}

func TestParseEnvVarPairs_KeyCasePreserved(t *testing.T) {
	result, err := parseEnvVarPairs("my_key=value,Mixed_Key=other")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"my_key":    "value",
		"Mixed_Key": "other",
	}, result)
}

func TestParseEnvVarPairs_PercentEncodedEquals(t *testing.T) {
	// '=' in a value may be percent-encoded as %3D and will be decoded correctly; literal '=' is also allowed
	result, err := parseEnvVarPairs("KEY=val%3Dwithin")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY": "val=within"}, result)
}

func TestParseEnvVarPairs_PercentEncodedComma(t *testing.T) {
	// ',' in a value must be percent-encoded as %2C
	result, err := parseEnvVarPairs("KEY=a%2Cb")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY": "a,b"}, result)
}

func TestParseEnvVarPairs_PercentEncodedInKey(t *testing.T) {
	// ',' percent-encoded in a key
	result, err := parseEnvVarPairs("KEY%2CSUFFIX=value")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY,SUFFIX": "value"}, result)
}

func TestParseEnvVarPairs_NonASCIIValue(t *testing.T) {
	// Non-ASCII characters may be percent-encoded
	result, err := parseEnvVarPairs("KEY=%C3%A9l%C3%A8ve")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY": "élève"}, result)
}

func TestParseEnvVarPairs_EmptyValue(t *testing.T) {
	result, err := parseEnvVarPairs("KEY=")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY": ""}, result)
}

func TestParseEnvVarPairs_MissingEquals(t *testing.T) {
	_, err := parseEnvVarPairs("NOEQUALS")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing '='")
}

func TestParseEnvVarPairs_EmptyKey(t *testing.T) {
	_, err := parseEnvVarPairs("=value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key must not be empty")
}

func TestParseEnvVarPairs_InvalidPercentEncoding(t *testing.T) {
	_, err := parseEnvVarPairs("KEY=val%ZZue")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid percent-encoding")
}

func TestParseEnvVarPairs_MultipleOTelResourceAttributes(t *testing.T) {
	envVars, err := parseEnvVarPairs("OTEL_LOG_LEVEL=debug,OTEL_RESOURCE_ATTRIBUTES=service.name=svc%2Ccustom_attr=00")
	require.NoError(t, err)
	assert.Equal(t, "debug", envVars["OTEL_LOG_LEVEL"])
	assert.Equal(t, "service.name=svc,custom_attr=00", envVars["OTEL_RESOURCE_ATTRIBUTES"])

	envVars, err = parseEnvVarPairs("OTEL_RESOURCE_ATTRIBUTES=service.name=svc%2Ccustom_attr=00,OTEL_LOG_LEVEL=debug")
	require.NoError(t, err)
	assert.Equal(t, "debug", envVars["OTEL_LOG_LEVEL"])
	assert.Equal(t, "service.name=svc,custom_attr=00", envVars["OTEL_RESOURCE_ATTRIBUTES"])
}

func TestHandleLaunchAsTA_EnvVarsSuffix(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }

	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_realm">us0</param>
			<param name="splunk_custom_env_vars">OTEL_LOG_LEVEL=debug,OTEL_RESOURCE_ATTRIBUTES=service.name=myapp</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, inputArgs, resultArgs, "Expected args to be returned unchanged")

	assert.Equal(t, "us0", envVars["SPLUNK_REALM"])
	assert.Equal(t, "debug", envVars["OTEL_LOG_LEVEL"])
	assert.Equal(t, "service.name=myapp", envVars["OTEL_RESOURCE_ATTRIBUTES"])
	// The _env_vars parameter itself must not appear as an env var
	assert.NotContains(t, envVars, "SPLUNK_CUSTOM_ENV_VARS")
}

func TestHandleLaunchAsTA_EnvVarsSuffix_PercentEncodedSeparators(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }

	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	// Value contains a literal ',' (%2C) and a literal '=' (%3D)
	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_extra_env_vars">K1=a%2Cb,K2=x%3Dy</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, inputArgs, resultArgs, "Expected args to be returned unchanged")

	assert.Equal(t, "a,b", envVars["K1"])
	assert.Equal(t, "x=y", envVars["K2"])
}

func TestHandleLaunchAsTA_EnvVarsSuffix_InvalidValue(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_bad_env_vars">NOEQUALS</param>
		</stanza>
	</configuration>
</input>`

	resultArgs, _, err := HandleLaunchAsTA([]string{"program"}, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.Error(t, err)
	assert.Nil(t, resultArgs, "Expected nil args on error")
	assert.Contains(t, err.Error(), "launch as TA failed to parse env vars from parameter 'splunk_bad_env_vars'")
}

func TestHandleLaunchAsTA_StanzaPrefixFirstMatch(t *testing.T) {
	// Save original functions and restore after test
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	// Mock parent process check to return true
	isParentProcessSplunkdFn = func() bool {
		return true
	}

	// Track environment variables set
	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	// Set SPLUNK_HOME
	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<configuration>
		<stanza name="otel://instance1" app="test-app">
			<param name="splunk_realm">us0</param>
		</stanza>
		<stanza name="otel://instance2" app="test-app">
			<param name="splunk_realm">eu0</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	// When multiple stanzas match the prefix, only the first one should be used
	resultArgs, _, err := HandleLaunchAsTA(args, reader, "otel://", "<scheme></scheme>", nil)
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, args, resultArgs, "Expected args to be returned unchanged")

	// Should use the first matching stanza
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"], "Expected SPLUNK_REALM from first matching stanza")
}

func TestHandleLaunchAsTA_CmdArgsSuffix(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_realm">us0</param>
			<param name="splunk_collector_cmd_args">--config=/etc/otel/config.yaml --feature-gates=foo</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"program", "--config=/etc/otel/config.yaml", "--feature-gates=foo"}, resultArgs)
}

func TestHandleLaunchAsTA_CmdArgsSuffix_QuotedArgs(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_collector_cmd_args">--config=/etc/otel/config.yaml "--some-arg=value with spaces"</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"program", "--config=/etc/otel/config.yaml", "--some-arg=value with spaces"}, resultArgs)
}

func TestHandleLaunchAsTA_CmdArgsSuffix_MultipleCmdArgParams(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_extra_cmd_args">--arg1=val1</param>
			<param name="splunk_collector_cmd_args">--arg2=val2</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"program", "--arg1=val1", "--arg2=val2"}, resultArgs)
}

func TestHandleLaunchAsTA_CmdArgsSuffix_EmptyValue(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_collector_cmd_args"></param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, inputArgs, resultArgs, "Expected args to be returned unchanged for empty cmd args value")
}

func TestHandleLaunchAsTA_CmdArgsSuffix_InvalidValue(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	setEnvFn = func(_, _ string) error { return nil }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_collector_cmd_args">--arg='unterminated</param>
		</stanza>
	</configuration>
</input>`

	resultArgs, _, err := HandleLaunchAsTA([]string{"program"}, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.Error(t, err)
	assert.Nil(t, resultArgs, "Expected nil args on error")
	assert.Contains(t, err.Error(), "launch as TA failed to parse cmd args from parameter 'splunk_collector_cmd_args'")
}

func TestHandleLaunchAsTA_CmdArgsSuffix_NotSetAsEnvVar(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalSetEnvFn := setEnvFn
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		setEnvFn = originalSetEnvFn
	}()

	isParentProcessSplunkdFn = func() bool { return true }

	envVars := make(map[string]string)
	setEnvFn = func(key, value string) error {
		envVars[key] = value
		return nil
	}

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<input>
	<configuration>
		<stanza name="test-stanza" app="test-app">
			<param name="splunk_realm">us0</param>
			<param name="splunk_collector_cmd_args">--config=/etc/otel/config.yaml</param>
		</stanza>
	</configuration>
</input>`

	inputArgs := []string{"program"}
	resultArgs, _, err := HandleLaunchAsTA(inputArgs, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"program", "--config=/etc/otel/config.yaml"}, resultArgs)
	// The _cmd_args parameter must not appear as an env var
	assert.NotContains(t, envVars, "SPLUNK_COLLECTOR_CMD_ARGS")
	assert.Equal(t, "us0", envVars["SPLUNK_REALM"])
}

func TestHandleLaunchAsTA_ValidationMode_ValidParams(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalStdout := stdoutWriter
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		stdoutWriter = originalStdout
	}()

	isParentProcessSplunkdFn = func() bool { return true }

	var buf bytes.Buffer
	stdoutWriter = &buf

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<items>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<item name="test-stanza://default">
		<param name="splunk_realm">us0</param>
		<param name="splunk_access_token">secret123</param>
	</item>
</items>`

	validatorCalled := false
	validator := func(_ *ValidationItems, args []string) ([]string, error) {
		validatorCalled = true
		args[1] = "validate"
		return args, nil
	}

	args := []string{"program", "--validate-arguments"}
	resultArgs, _, err := HandleLaunchAsTA(args, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", validator)
	require.NoError(t, err, "Expected no error after successful validation")
	assert.True(t, validatorCalled, "Expected validator to be called")
	assert.Equal(t, "validate", resultArgs[1], "Expected args[1] to be set to 'validate'")
	assert.Empty(t, buf.String(), "Expected no output written to stdout on success")
}

func TestHandleLaunchAsTA_ValidationMode_InvalidParams(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalStdout := stdoutWriter
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		stdoutWriter = originalStdout
	}()

	isParentProcessSplunkdFn = func() bool { return true }

	var buf bytes.Buffer
	stdoutWriter = &buf

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<items>
	<server_host>testhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<item name="test-stanza://default">
		<param name="splunk_realm">us0</param>
	</item>
</items>`

	validator := func(_ *ValidationItems, _ []string) ([]string, error) {
		return nil, errors.New("splunk_access_token is required")
	}

	args := []string{"program", "--validate-arguments"}
	_, _, err := HandleLaunchAsTA(args, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", validator)
	require.Error(t, err, "Expected error when validation fails")
	assert.Contains(t, buf.String(), "splunk_access_token is required", "Expected error message in stdout XML")
	assert.Contains(t, buf.String(), "<error>", "Expected XML error element in stdout")
}

func TestHandleLaunchAsTA_ValidationMode_XMLReadError(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	defer func() { isParentProcessSplunkdFn = originalIsParentFn }()

	isParentProcessSplunkdFn = func() bool { return true }

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	validator := func(_ *ValidationItems, args []string) ([]string, error) { return args, nil }

	args := []string{"program", "--validate-arguments"}
	_, _, err := HandleLaunchAsTA(args, &errorReader{err: errors.New("read error")}, "test-stanza", "<scheme></scheme>", validator)
	require.Error(t, err, "Expected error when stdin read fails in validation mode")
	assert.Contains(t, err.Error(), "validation mode failed to read XML from stdin")
}

func TestHandleLaunchAsTA_ValidationMode_ValidatorReceivesItems(t *testing.T) {
	originalIsParentFn := isParentProcessSplunkdFn
	originalStdout := stdoutWriter
	defer func() {
		isParentProcessSplunkdFn = originalIsParentFn
		stdoutWriter = originalStdout
	}()

	isParentProcessSplunkdFn = func() bool { return true }
	stdoutWriter = &bytes.Buffer{}

	t.Setenv("SPLUNK_HOME", "/opt/splunk")

	xmlData := `<items>
	<server_host>myhost</server_host>
	<server_uri>https://localhost:8089</server_uri>
	<session_key>test_key</session_key>
	<checkpoint_dir>/tmp/checkpoint</checkpoint_dir>
	<item name="test-stanza://default">
		<param name="splunk_realm">eu0</param>
		<param name="splunk_access_token">tok456</param>
	</item>
</items>`

	var receivedItems *ValidationItems
	validator := func(items *ValidationItems, args []string) ([]string, error) {
		receivedItems = items
		args[1] = "validate"
		return args, nil
	}

	args := []string{"program", "--validate-arguments"}
	_, _, err := HandleLaunchAsTA(args, strings.NewReader(xmlData), "test-stanza", "<scheme></scheme>", validator)
	require.NoError(t, err)
	require.NotNil(t, receivedItems)
	assert.Equal(t, "myhost", receivedItems.ServerHost)
	require.Len(t, receivedItems.Item, 1)
	assert.Equal(t, "test-stanza://default", receivedItems.Item[0].Name)
	require.Len(t, receivedItems.Item[0].Param, 2)
}
