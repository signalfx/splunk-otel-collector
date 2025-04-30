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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/splunk/splunk-technical-addon/internal/testaddon"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/splunk/splunk-technical-addon/internal/packaging"
	"github.com/splunk/splunk-technical-addon/internal/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ExampleOutput struct {
	Flags   []string
	EnvVars []string

	SplunkHome                 string
	TaHome                     string
	PlatformHome               string
	EverythingSet              string
	MinimalSet                 string
	MinimalSetRequired         string
	UnaryFlagWithEverythingSet string
	Platform                   string
}

func TestPascalization(t *testing.T) {
	tests := []struct {
		sample      string
		expected    string
		shouldError bool
	}{
		{
			sample:   "Splunk_Addon",
			expected: "SplunkAddon",
		},
		{
			sample:   "hello_world",
			expected: "HelloWorld",
		},
		{
			sample:   "NoBreaks",
			expected: "NoBreaks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.sample, func(t *testing.T) {
			actual := SnakeToPascal(tt.sample)
			if actual != tt.expected {
				t.Errorf("Expected %s but got %s", tt.expected, actual)
			}
		})
	}
}

func TestRunner(t *testing.T) {
	ctx := context.Background()
	addonPath := filepath.Join(t.TempDir(), "Sample_Addon.tgz")

	buildDir := packaging.GetBuildDir()
	require.NotEmpty(t, buildDir)
	err := packaging.PackageAddon(filepath.Join(buildDir, "Sample_Addon"), addonPath)
	require.NoError(t, err)
	tc := testaddon.StartSplunk(t, testaddon.SplunkStartOpts{
		AddonPaths:   []string{addonPath},
		WaitStrategy: wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/Sample_Addon.log"}).WithStartupTimeout(time.Minute * 4),
	})

	// Check Schema
	code, output, err := tc.Exec(ctx, []string{"sudo", "/opt/splunk/bin/splunk", "btool", "check", "--debug"})
	assert.NoError(t, err)
	assert.LessOrEqual(t, code, 1)    // Other stanzas may be missing and thus have this be 0 or 1
	assert.GreaterOrEqual(t, code, 0) // bound to [0,1]
	read, err := io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotContains(t, string(read), "Invalid Key in Stanza")

	// check log output
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/opt/splunk/var/log/splunk/Sample_Addon.log"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	expectedJSON := `{"Flags":["--test-flag","/opt/splunk/etc/apps/Sample_Addon/local/access_token","--test-flag"],"EnvVars":["EVERYTHING_SET=/opt/splunk/etc/apps/Sample_Addon/local/access_token","UNARY_FLAG_WITH_EVERYTHING_SET=/opt/splunk/etc/apps/Sample_Addon/local/access_token"], "SplunkHome":"/opt/splunk/etc", "TaHome":"/opt/splunk/etc/apps/Sample_Addon", "PlatformHome":"/opt/splunk/etc/apps/Sample_Addon/linux_x86_64", "EverythingSet":"/opt/splunk/etc/apps/Sample_Addon/local/access_token", "MinimalSet":"", "MinimalSetRequired":"", "UnaryFlagWithEverythingSet":"/opt/splunk/etc/apps/Sample_Addon/local/access_token","Platform":"linux"}`
	i := bytes.Index(read, []byte("Sample output:"))
	unmarshalled := &ExampleOutput{}
	dec := json.NewDecoder(bytes.NewReader(read[i+len("Sample output:"):]))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(unmarshalled))
	expected := &ExampleOutput{}
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), expected))
	assert.EqualValues(t, expected, unmarshalled)

	assert.NoError(t, tc.Terminate(ctx))
}

func TestRunnerConfigGeneration(t *testing.T) {
	sourceDir, err := packaging.GetSourceDir()
	require.NoError(t, err)
	sourceDir = filepath.Join(sourceDir, "cmd", "modinput_config_generator", "internal", "testdata")
	tests := []struct {
		testSchemaName string
		sampleYamlPath string
		outDir         string
		shouldError    bool
	}{
		{
			testSchemaName: "Sample_Addon",
			outDir:         t.TempDir(),
			sampleYamlPath: filepath.Join(sourceDir, "pkg/sample_addon/runner/modular-inputs.yaml"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.testSchemaName, func(tt *testing.T) {
			config, err := loadYaml(tc.sampleYamlPath, tc.testSchemaName)
			assert.NoError(tt, err)
			err = generateModinputConfig(config, tc.outDir)
			assert.NoError(tt, err)
			assert.FileExists(tt, filepath.Join(filepath.Dir(tc.sampleYamlPath), "modinput_config.go"))
		})
	}
}

func TestInputsConfGeneration(t *testing.T) {
	sourceDir, err := packaging.GetSourceDir()
	require.NoError(t, err)
	sourceDir = filepath.Join(sourceDir, "cmd", "modinput_config_generator", "internal", "testdata")
	tests := []struct {
		testSchemaName   string
		sampleYamlPath   string
		outDir           string
		addonSourceDir   string
		expectedSpecPath string
		shouldError      bool
	}{
		{
			testSchemaName: "Sample_Addon",
			outDir:         t.TempDir(),
			addonSourceDir: filepath.Join(sourceDir, "pkg", "sample_addon"),
			sampleYamlPath: filepath.Join(sourceDir, "pkg", "sample_addon", "runner", "modular-inputs.yaml"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.testSchemaName, func(tt *testing.T) {
			config, err := loadYaml(tc.sampleYamlPath, tc.testSchemaName)
			assert.NoError(tt, err)
			err = generateTaModInputConfs(config, tc.addonSourceDir, tc.outDir)
			assert.NoError(tt, err)
			testcommon.AssertFilesMatch(tt, filepath.Join("internal", "testdata", "pkg", "sample_addon", "expected", "inputs.conf"), filepath.Join(tc.outDir, tc.testSchemaName, "default", "inputs.conf"))
			testcommon.AssertFilesMatch(tt, filepath.Join("internal", "testdata", "pkg", "sample_addon", "expected", "inputs.conf.spec"), filepath.Join(tc.outDir, tc.testSchemaName, "README", "inputs.conf.spec"))
		})
	}
}
