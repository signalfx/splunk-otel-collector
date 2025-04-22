package main

import (
	"github.com/google/go-cmp/cmp"
	"github.com/splunk/otel-technical-addon/internal/packaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const testDataPrefix = "internal/testdata/"

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
	addonPath := filepath.Join(t.TempDir(), "Sample_Addon.tgz")
	err := packaging.PackageAddon("Sample_Addon", addonPath)
	require.NoError(t, err)
	// TODO add testcontainer run here
}

func TestRunnerConfigGeneration(t *testing.T) {
	// This is a smoketest, any actual functionality test should be tested via the
	// test addon's "runner" itself

	tests := []struct {
		testSchemaName string
		sampleYamlPath string
		outDir         string
		shouldError    bool
	}{
		{
			testSchemaName: "Sample_Addon",
			outDir:         t.TempDir(),
			sampleYamlPath: filepath.Join(os.Getenv("SOURCE_DIR"), "pkg/sample_addon/runner/modular-inputs.yaml"),
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

	tests := []struct {
		testSchemaName   string
		sampleYamlPath   string
		outDir           string
		sourceDir        string
		expectedSpecPath string
		shouldError      bool
	}{
		{
			testSchemaName: "Sample_Addon",
			outDir:         t.TempDir(),
			sourceDir:      filepath.Join(os.Getenv("SOURCE_DIR"), "pkg/sample_addon"),
			sampleYamlPath: filepath.Join(os.Getenv("SOURCE_DIR"), "pkg/sample_addon/runner/modular-inputs.yaml"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.testSchemaName, func(tt *testing.T) {
			config, err := loadYaml(tc.sampleYamlPath, tc.testSchemaName)
			assert.NoError(tt, err)
			err = generateTaModInputConfs(config, tc.sourceDir, tc.outDir)
			assert.NoError(tt, err)
			assertFilesMatch(tt, filepath.Join("internal", "testdata", "pkg", "sample_addon", "expected", "inputs.conf"), filepath.Join(tc.outDir, "default", "inputs.conf"))
			assertFilesMatch(tt, filepath.Join("internal", "testdata", "pkg", "sample_addon", "expected", "inputs.conf.spec"), filepath.Join(tc.outDir, "README", "inputs.conf.spec"))
		})
	}
}

func assertFilesMatch(tt *testing.T, expectedPath string, actualPath string) {
	require.FileExists(tt, actualPath)
	require.FileExists(tt, expectedPath)
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		tt.Fatalf("Failed to read expected file: %v", err)
	}

	actual, err := os.ReadFile(actualPath)
	if err != nil {
		tt.Fatalf("Failed to read actual file: %v", err)
	}

	if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
		tt.Errorf("File contents mismatch (-expected +actual)\npaths: (%s, %s):\n%s", expectedPath, actualPath, diff)
	}
}
