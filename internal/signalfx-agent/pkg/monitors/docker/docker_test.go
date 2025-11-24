//go:build dockerd

package docker

import (
	"encoding/json"
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
	// Fail if there is already a daemon.json file
	if _, err := os.Stat("/etc/docker/daemon.json"); err == nil {
		t.Fatal("daemon.json already exists, cannot update minimum required client version")
	}

	daemonConfig := map[string]string{
		"min-api-version": minimumRequiredClientVersion,
	}

	configJSON, err := json.MarshalIndent(daemonConfig, "", "  ")
	require.NoError(t, err, "Failed to marshal daemon config")
	t.Logf("Docker daemon config JSON:\n%s", string(configJSON))

	// Create a temporary daemon.json file with the new configuration then
	// move it using sudo to the correct location.
	tempFileName := filepath.Join(t.TempDir(), "daemon.json")
	err = os.WriteFile(tempFileName, configJSON, 0o644)
	require.NoError(t, err, "Failed to write daemon.json")

	cmd := exec.Command("sudo", "mv", tempFileName, "/etc/docker/")
	err = cmd.Run()
	require.NoError(t, err, "Failed to move daemon.json")

	cmd = exec.Command("sudo", "service", "docker", "restart")
	// Ignore error since the docker daemon might automatically restart after
	// adding the config file
	err = cmd.Run()
	if err != nil {
		t.Logf("Docker daemon restart error: %s", err)
	}

	t.Cleanup(func() {
		cmd := exec.Command("sudo", "rm", "/etc/docker/daemon.json")
		err := cmd.Run()
		require.NoError(t, err, "Failed to remove daemon.json")

		cmd = exec.Command("sudo", "service", "docker", "restart")
		// Ignore error since the docker daemon might automatically restart after
		// removing the config file
		err = cmd.Run()
		if err != nil {
			t.Logf("Docker daemon restart error: %s", err)
		}

		requireDockerDaemonRunning(t)
	})

	requireDockerDaemonRunning(t)
}

func runDockerContainerToGenerateMetrics(t *testing.T) func() {
	cmd := exec.Command("docker", "run", "-d", "--name", "docker-client-test", "alpine", "sleep", "180")
	err := cmd.Run()
	require.NoError(t, err, "Failed to run docker container")
	return func() {
		cmd := exec.Command("docker", "rm", "-f", "docker-client-test")
		err := cmd.Run()
		require.NoError(t, err, "Failed to remove docker container")
	}
}

func requireDockerDaemonRunning(t *testing.T) {
	require.Eventually(t, func() bool {
		isRunning, err := isServiceRunning(t, "docker")
		require.NoError(t, err, "Failed to get docker service status")
		return isRunning
	}, 30*time.Second, 1*time.Second)
}

func isServiceRunning(t *testing.T, serviceName string) (bool, error) {
	cmd := exec.Command("service", serviceName, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// The 'service status' command often returns a non-zero exit code if the service is not running
		// or if there's an error. We still need to parse the output to determine the status.
		t.Logf("Error getting %q service status: %v", serviceName, err)
	}

	outputStr := string(output)

	if strings.Contains(outputStr, "active (running)") {
		return true, nil
	}

	return false, nil
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
