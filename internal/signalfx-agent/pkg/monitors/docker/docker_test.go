//go:build dockerd

package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/signalfx/defaults"
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/event"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/trace"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

const (
	useDockerEngineDefault = "default"
)

func TestMinimumRequiredClientVersion(t *testing.T) {
	// Skip this test if not running on Linux GitHub runner
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-Linux OS")
	}
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Skip("Skipping test outside of GitHub Actions")
	}

	// Execute the dockerd upgrade script and fail the test if it fails.
	// This test is not rolling back the dockerd upgrade, so it can affect all subsequent tests.
	scriptPath := filepath.Join("testdata", "upgrade-dockerd-on-ubuntu.sh")
	scriptCmd := exec.Command("bash", scriptPath)
	scriptOut, err := scriptCmd.CombinedOutput()
	t.Logf("upgrade-dockerd-on-ubuntu.sh output:\n%s\n", string(scriptOut))
	require.NoError(t, err, "upgrade-dockerd-on-ubuntu.sh failed with exit code %d", scriptCmd.ProcessState.ExitCode())

	tt := []struct {
		minimumRequiredClientVersion string
	}{
		{
			minimumRequiredClientVersion: useDockerEngineDefault,
		},
		{
			minimumRequiredClientVersion: "1.24",
		},
	}

	for _, tc := range tt {
		t.Run(tc.minimumRequiredClientVersion, func(t *testing.T) {
			if tc.minimumRequiredClientVersion != useDockerEngineDefault {
				updateGHLinuxRunnerDockerDaemonMinClientVersion(t, tc.minimumRequiredClientVersion)
			}

			cleanupContainer := runDockerContainerToGenerateMetrics(t)
			// This needs to be in a defer so the container is removed before the docker daemon settings are reset.
			defer cleanupContainer()

			output := &fakeOutput{}
			monitor := &Monitor{
				Output: output,
			}
			config := &Config{
				MonitorConfig: config.MonitorConfig{
					IntervalSeconds: 1,
				},
			}
			defaults.Set(config)

			err := monitor.Configure(config)
			require.NoError(t, err, "Expected no error during monitor configuration")
			defer monitor.Shutdown()

			require.Eventually(t, func() bool {
				return output.HasDatapoints()
			}, 10*time.Second, 100*time.Millisecond, "Expected datapoints to be collected")
		})
	}
}

func updateGHLinuxRunnerDockerDaemonMinClientVersion(t *testing.T, minimumRequiredClientVersion string) {
	existingConfig, hasExistingConfig := readDockerDaemonConfig(t)
	daemonConfig := map[string]any{}
	if hasExistingConfig {
		require.NoError(t, json.Unmarshal(existingConfig, &daemonConfig), "Failed to unmarshal existing daemon.json")
	}

	daemonConfig["min-api-version"] = minimumRequiredClientVersion

	configJSON, err := json.MarshalIndent(daemonConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal daemon config")
	t.Logf("Docker daemon config JSON:\n%s", string(configJSON))

	writeDockerDaemonConfig(t, configJSON)

	restartDockerService(t)

	t.Cleanup(func() {
		if hasExistingConfig {
			writeDockerDaemonConfig(t, existingConfig)
		} else {
			runCommand(t, "sudo", "rm", "-f", "/etc/docker/daemon.json")
		}

		restartDockerService(t)

		requireDockerDaemonRunning(t)
	})

	requireDockerDaemonRunning(t)
}

func readDockerDaemonConfig(t *testing.T) ([]byte, bool) {
	t.Helper()
	configJSON, err := os.ReadFile("/etc/docker/daemon.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil, false
	}
	require.NoError(t, err, "Failed to read existing daemon.json")
	return configJSON, true
}

func writeDockerDaemonConfig(t *testing.T, configJSON []byte) {
	t.Helper()
	tempFileName := filepath.Join(t.TempDir(), "daemon.json")
	err := os.WriteFile(tempFileName, configJSON, 0o644)
	require.NoError(t, err, "Failed to write daemon.json")

	runCommand(t, "sudo", "mv", tempFileName, "/etc/docker/")
}

func restartDockerService(t *testing.T) {
	t.Helper()
	cmd := exec.Command("sudo", "service", "docker", "restart")
	output, err := cmd.CombinedOutput()
	// Ignore error since the docker daemon might automatically restart after
	// updating the config file.
	if err != nil {
		t.Logf("Docker daemon restart error: %s\nOutput:\n%s", err, string(output))
	} else if len(output) > 0 {
		t.Logf("Docker daemon restart output:\n%s", string(output))
	}
}

func runDockerContainerToGenerateMetrics(t *testing.T) func() {
	t.Helper()
	imageName := fmt.Sprintf("docker-client-test:%d", os.Getpid())
	containerName := fmt.Sprintf("docker-client-test-%d", os.Getpid())

	buildDockerTestImage(t, imageName)
	runCommand(t, "docker", "run", "-d", "--name", containerName, imageName)

	return func() {
		runCommand(t, "docker", "rm", "-f", containerName)
		runCommand(t, "docker", "rmi", "-f", imageName)
	}
}

func buildDockerTestImage(t *testing.T, imageName string) {
	t.Helper()
	tempDir := t.TempDir()
	sleeperSource := []byte(`package main

import "time"

func main() {
	time.Sleep(3 * time.Minute)
}
`)
	err := os.WriteFile(filepath.Join(tempDir, "main.go"), sleeperSource, 0o644)
	require.NoError(t, err, "Failed to write test container source")

	cmd := exec.Command("go", "build", "-trimpath", "-o", filepath.Join(tempDir, "sleeper"), filepath.Join(tempDir, "main.go"))
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "Failed to build test container binary\nOutput:\n%s", string(output))

	dockerfile := []byte("FROM scratch\nCOPY sleeper /sleeper\nENTRYPOINT [\"/sleeper\"]\n")
	err = os.WriteFile(filepath.Join(tempDir, "Dockerfile"), dockerfile, 0o644)
	require.NoError(t, err, "Failed to write test container Dockerfile")

	runCommand(t, "docker", "build", "--pull=false", "-t", imageName, tempDir)
}

func requireDockerDaemonRunning(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool {
		cmd := exec.Command("docker", "info")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("docker info failed: %s\nOutput:\n%s", err, string(output))
			return false
		}
		return true
	}, 30*time.Second, 1*time.Second)
}

func runCommand(t *testing.T, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "Command %q failed\nOutput:\n%s", strings.Join(cmd.Args, " "), string(output))
	return string(output)
}

type fakeOutput struct {
	datapoints []*datapoint.Datapoint
	mu         sync.Mutex
}

var _ types.FilteringOutput = (*fakeOutput)(nil)

func (fo *fakeOutput) AddDatapointExclusionFilter(_ dpfilters.DatapointFilter) {
	panic("unimplemented")
}

func (fo *fakeOutput) AddExtraDimension(_, _ string) {
	panic("unimplemented")
}

func (fo *fakeOutput) Copy() types.Output {
	panic("unimplemented")
}

func (fo *fakeOutput) EnabledMetrics() []string {
	return []string{}
}

func (fo *fakeOutput) HasAnyExtraMetrics() bool {
	panic("unimplemented")
}

func (fo *fakeOutput) HasEnabledMetricInGroup(_ string) bool {
	panic("unimplemented")
}

func (fo *fakeOutput) SendDimensionUpdate(_ *types.Dimension) {
	panic("unimplemented")
}

func (fo *fakeOutput) SendEvent(_ *event.Event) {
	panic("unimplemented")
}

func (fo *fakeOutput) SendMetrics(_ ...pmetric.Metric) {
	panic("unimplemented")
}

func (fo *fakeOutput) SendSpans(_ ...*trace.Span) {
	panic("unimplemented")
}

func (fo *fakeOutput) SendDatapoints(dps ...*datapoint.Datapoint) {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	fo.datapoints = append(fo.datapoints, dps...)
}

func (fo *fakeOutput) HasDatapoints() bool {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	return len(fo.datapoints) > 0
}
