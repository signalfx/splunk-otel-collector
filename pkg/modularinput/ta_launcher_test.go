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
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLaunchAsTA_NotModularInput(t *testing.T) {
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
	err := HandleLaunchAsTA(args, nil, "test-stanza")
	assert.NoError(t, err, "Expected no error when not in modular input mode")
}

func TestHandleLaunchAsTA_QueryModeScheme(t *testing.T) {
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
	err := HandleLaunchAsTA(args, nil, "test-stanza")
	assert.ErrorIs(t, err, ErrQueryMode, "Expected ErrQueryMode for --scheme argument")
}

func TestHandleLaunchAsTA_QueryModeValidate(t *testing.T) {
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
	err := HandleLaunchAsTA(args, nil, "test-stanza")
	assert.ErrorIs(t, err, ErrQueryMode, "Expected ErrQueryMode for --validate-arguments argument")
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

	err := HandleLaunchAsTA(args, invalidXML, "test-stanza")
	require.Error(t, err, "Expected error for invalid XML")
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
			<param name="api_key">secret123</param>
			<param name="endpoint">https://api.example.com</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	err := HandleLaunchAsTA(args, reader, "test-stanza")
	require.NoError(t, err, "Expected no error")

	assert.Equal(t, "secret123", envVars["API_KEY"], "Expected API_KEY to be set")
	assert.Equal(t, "https://api.example.com", envVars["ENDPOINT"], "Expected ENDPOINT to be set")
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
			<param name="config_value">$TEST_VAR</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	err := HandleLaunchAsTA(args, reader, "test-stanza")
	require.NoError(t, err, "Expected no error")

	assert.Equal(t, "expanded_value", envVars["CONFIG_VALUE"], "Expected CONFIG_VALUE to be expanded")
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
			<param name="api_key">secret123</param>
		</stanza>
	</configuration>
</input>`

	args := []string{"program"}
	reader := strings.NewReader(xmlData)

	err := HandleLaunchAsTA(args, reader, "test-stanza")
	require.Error(t, err, "Expected error")
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

	err := HandleLaunchAsTA(args, reader, "test-stanza")
	require.NoError(t, err, "Expected no error when stanza not found")
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

	err := HandleLaunchAsTA(args, reader, "test-stanza")
	require.NoError(t, err, "Expected no error with empty stanza")
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
	err := HandleLaunchAsTA(args, errorReader, "test-stanza")
	require.Error(t, err, "Expected error for read failure")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.False(t, isModularInput, "Expected not modular input mode without SPLUNK_HOME")
	assert.False(t, isQueryMode, "Expected not query mode without SPLUNK_HOME")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.False(t, isModularInput, "Expected not modular input mode when TA v1 is being launched")
	assert.False(t, isQueryMode, "Expected not query mode when TA v1 is being launched")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.False(t, isModularInput, "Expected not modular input mode when parent is not splunkd")
	assert.False(t, isQueryMode, "Expected not query mode when parent is not splunkd")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.True(t, isModularInput, "Expected modular input mode")
	assert.False(t, isQueryMode, "Expected not query mode")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.True(t, isModularInput, "Expected modular input mode")
	assert.True(t, isQueryMode, "Expected query mode for --scheme")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.True(t, isModularInput, "Expected modular input mode")
	assert.True(t, isQueryMode, "Expected query mode for --validate-arguments")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.True(t, isModularInput, "Expected modular input mode")
	assert.False(t, isQueryMode, "Expected not query mode for other arguments")
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
	isModularInput, isQueryMode := isModularInputMode(args)
	assert.True(t, isModularInput, "Expected modular input mode")
	assert.False(t, isQueryMode, "Expected not query mode for more than 2 arguments")
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct {
	err error
}

func (er *errorReader) Read(_ []byte) (n int, err error) {
	return 0, er.err
}

var _ io.Reader = (*errorReader)(nil)
