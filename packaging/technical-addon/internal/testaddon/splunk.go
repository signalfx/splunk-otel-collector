package testaddon

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/splunk/splunk-technical-addon/internal/modularinput"
	"github.com/splunk/splunk-technical-addon/internal/packaging"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type SplunkStartOpts struct {
	AddonPaths   []string
	WaitStrategy wait.Strategy
}

func StartSplunk(t *testing.T, startOpts SplunkStartOpts) testcontainers.Container {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	conContext := context.Background()
	require.NotEmpty(t, startOpts.AddonPaths)
	localAddonsDir := t.TempDir()
	containerAddonsDir := "/tmp/local-tas"
	var addonNames []string
	for _, addonPath := range startOpts.AddonPaths {
		addonFileName := filepath.Base(addonPath)
		_, err = fileutils.CopyFile(addonPath, filepath.Join(localAddonsDir, addonFileName))
		require.NoError(t, err)
		addonNames = append(addonNames, filepath.Join(containerAddonsDir, addonFileName))
	}
	splunkStartURL := strings.Join(addonNames, ",")
	t.Logf("Splunk start url: %s", splunkStartURL)
	req := testcontainers.ContainerRequest{
		Image: "splunk/splunk:9.4.1",
		HostConfigModifier: func(c *container.HostConfig) {
			c.NetworkMode = "host"
			c.Mounts = append(c.Mounts, mount.Mount{
				Source: localAddonsDir,
				Target: containerAddonsDir,
				Type:   mount.TypeBind,
			})
			c.AutoRemove = false // change to false for debugging
		},
		Env: map[string]string{
			"SPLUNK_START_ARGS": "--accept-license",
			"SPLUNK_PASSWORD":   "Chang3d!",
			"SPLUNK_APPS_URL":   splunkStartURL,
		},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/en-US/account/login").WithPort("8000"),
			startOpts.WaitStrategy,
		).WithStartupTimeoutDefault(4 * time.Minute).WithDeadline(4*time.Minute + 20*time.Second),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&testLogConsumer{t: t}},
		},
	}

	tc, err := testcontainers.GenericContainer(conContext, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		logger.Error("error starting up splunk")
	}
	// Uncomment this line if you'd like to debug the container
	// docker container ls --all
	// Grab id of splunk container
	// docker exec -it $container_id bash
	// See README.md in this package for more info
	//time.Sleep(20 * time.Minute)
	require.NoError(t, err)
	return tc
}

type testLogConsumer struct {
	t *testing.T
}

func (l *testLogConsumer) Accept(log testcontainers.Log) {
	l.t.Log(log.LogType + ": " + strings.TrimSpace(string(log.Content)))
}

type RepackFunc func(t *testing.T, addonDir string) error

func PackAddon(t *testing.T, defaultModInputs *modularinput.GenericModularInput, repackFunc RepackFunc) string {
	packedDir := filepath.Join(t.TempDir(), defaultModInputs.SchemaName)
	buildDir := packaging.GetBuildDir()
	require.NotEmpty(t, buildDir)

	addonSource := filepath.Join(buildDir, defaultModInputs.SchemaName)
	err := os.CopyFS(packedDir, os.DirFS(addonSource))
	require.NoError(t, err)

	err = repackFunc(t, packedDir)
	require.NoError(t, err)

	addonPath := filepath.Join(t.TempDir(), fmt.Sprintf("%s.tgz", defaultModInputs.SchemaName))
	err = packaging.PackageAddon(packedDir, addonPath)
	require.NoError(t, err)

	return addonPath
}
func GetModInputsForConfig(t *testing.T, addonPath string) *modularinput.GenericModularInput {
	return nil
}

func AssertFileShasEqual(t *testing.T, expected string, actual string) {
	expectedSum, err := packaging.Sha256Sum(expected)
	require.NoError(t, err)
	actualSum, err := packaging.Sha256Sum(actual)
	require.NoError(t, err)
	assert.Equal(t, expectedSum, actualSum)
}

func RepackAddon(t *testing.T, sourceAddonPath string, repackFunc RepackFunc) string {
	repackedDir := t.TempDir()
	err := packaging.ExtractAddon(sourceAddonPath, repackedDir)
	require.NoError(t, err)
	require.NoError(t, repackFunc(t, repackedDir))
	repackedAddon := filepath.Join(t.TempDir(), filepath.Base(sourceAddonPath))
	require.NoError(t, packaging.PackageAddon(repackedDir, repackedAddon))
	return repackedAddon
}
