package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/google/go-cmp/cmp"
	"github.com/splunk/otel-technical-addon/internal/modularinput"
	"github.com/splunk/otel-technical-addon/internal/packaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

	buildDir := modularinput.GetBuildDir()
	require.NotEmpty(t, buildDir)
	err := packaging.PackageAddon(filepath.Join(buildDir, "Sample_Addon"), addonPath)
	require.NoError(t, err)
	tc := startSplunk(t, addonPath)

	// Check Schema
	code, output, err := tc.Exec(ctx, []string{"sudo", "/opt/splunk/bin/splunk", "btool", "check", "--debug"})
	assert.NoError(t, err)
	// current issue is that it can't read... maybe sudo it?
	assert.Zero(t, code) // TODO may want to check context as well for "Invalid Key in Stanza" log, not sure about zero code
	read, err := io.ReadAll(output)
	assert.NoError(t, err)
	t.Logf("read: %s", read)

	// Check log output
	expectedJson := `'{"Flags":["--test-flag","$SPLUNK_OTEL_TA_HOME/local/access_token","--test-flag"],"EnvVars":["UNARY_FLAG_WITH_EVERYTHING_SET=$SPLUNK_OTEL_TA_HOME/local/access_token","EVERYTHING_SET=$SPLUNK_OTEL_TA_HOME/local/access_token"],"Platform":"linux"}'`
	code, output, err = tc.Exec(ctx, []string{"sudo", "grep", "-qi", expectedJson, "/opt/splunk/var/log/splunk/Sample_Addon.log"})
	assert.NoError(t, err)
	assert.Zero(t, code)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	t.Logf("read: %s", string(read))

	code, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/opt/splunk/var/log/splunk/Sample_Addon.log"})
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	t.Logf("read: %s", read)

	assert.NoError(t, tc.Terminate(ctx))
}

func TestRunnerConfigGeneration(t *testing.T) {

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

func startSplunk(t *testing.T, taPath string) testcontainers.Container {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	conContext := context.Background()
	addonLocation := fmt.Sprintf("/tmp/local-tas/%v", filepath.Base(taPath))

	req := testcontainers.ContainerRequest{
		Image: "splunk/splunk:9.1.2",
		HostConfigModifier: func(c *container.HostConfig) {
			c.NetworkMode = "host"
			c.Mounts = append(c.Mounts, mount.Mount{
				Source: filepath.Dir(taPath),
				Target: filepath.Dir(addonLocation),
				Type:   mount.TypeBind,
			})
		},
		Env: map[string]string{
			"SPLUNK_START_ARGS": "--accept-license",
			"SPLUNK_PASSWORD":   "Chang3d!",
			"SPLUNK_APPS_URL":   addonLocation,
		},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/en-US/account/login").WithPort("8000"),
			wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/Sample_Addon.log"}),
		).WithDeadline(4*time.Minute + 20*time.Second).WithStartupTimeoutDefault(4 * time.Minute),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&testLogConsumer{t: t}},
		},
	}

	tc, err := testcontainers.GenericContainer(conContext, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		logger.Info("Error while creating container")
		panic(err)
	}
	return tc
}

// testLogConsumer is a simple implementation of LogConsumer that logs to the test output.
type testLogConsumer struct {
	t *testing.T
}

func (l *testLogConsumer) Accept(log testcontainers.Log) {
	l.t.Log(log.LogType + ": " + strings.TrimSpace(string(log.Content)))
}
