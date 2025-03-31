package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
			//assert.NoError(tt, os.MkdirAll(filepath.Join(tc.outDir, tc.testSchemaName), 0755))
			config, err := loadYaml(tc.sampleYamlPath, tc.testSchemaName)
			assert.NoError(tt, err)
			err = generateModinputConfig(config, tc.outDir)
			assert.NoError(tt, err)
			listPath(tc.outDir)
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
			//assert.NoError(tt, os.MkdirAll(filepath.Join(tc.outDir, tc.testSchemaName), 0755))
			config, err := loadYaml(tc.sampleYamlPath, tc.testSchemaName)
			assert.NoError(tt, err)
			err = generateTaModInputConfs(config, tc.sourceDir, tc.outDir)
			assert.NoError(tt, err)
			listPath(tc.outDir)
			assert.FileExists(tt, filepath.Join(tc.outDir, "default", "inputs.conf"))
			assert.FileExists(tt, filepath.Join(tc.outDir, "README", "inputs.conf.spec"))
		})
	}
}

func listPath(s string) {
	// List current directory, similar to basic "ls"
	entries, err := os.ReadDir(s)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print names only (like basic ls)
	for _, entry := range entries {
		fmt.Println(entry.Name())
	}

	// For ls -l style output with more details
	fmt.Println("\nDetailed listing:")
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		mode := info.Mode()
		size := info.Size()
		modTime := info.ModTime().Format("Jan _2 15:04")
		name := entry.Name()

		fmt.Printf("%s %8d %s %s\n", mode, size, modTime, name)
	}
}
