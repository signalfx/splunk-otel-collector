package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/splunk/splunk-technical-addon/internal/packaging"
	"github.com/splunk/splunk-technical-addon/internal/testaddon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestZeroConfig(t *testing.T) {
	expectedJar := filepath.Join(packaging.GetBuildDir(), "Splunk_TA_otel_linux_autoinstrumentation", "linux_x86_64", "bin", "splunk-otel-javaagent.jar")
	require.FileExists(t, expectedJar)

	tests := []struct {
		testname  string
		testDir   string
		modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs
	}{
		{
			testname: "happypath",
			testDir:  t.TempDir(),
			modInputs: &SplunkTAOtelLinuxAutoinstrumentationModularInputs{
				SplunkOtelJavaAutoinstrumentationJarPath: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-javaagent.jar",
					Name:  "splunk_otel_java_autoinstrumentation_jar_path",
				},
				AutoinstrumentationPath: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/libsplunk_amd64.so",
					Name:  "autoinstrumentation_path",
				},
				AutoinstrumentationPreloadPath: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "ld.preload",
					Name:  "autoinstrumentation_preload_path",
				},
				ZeroconfigPath: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "zero.conf",
					Name:  "zeroconfig_path",
				},
				ResourceAttributes: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "asdasd",
					Name:  "resource_attributes",
				},
				ProfilerEnabled: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "false",
					Name:  "profiler_enabled",
				},
				ProfilerMemoryEnabled: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "false",
					Name:  "profiler_memory_enabled",
				},
				MetricsEnabled: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "false",
					Name:  "metrics_enabled",
				},
				LogsEnabled: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "false",
					Name:  "logs_enabled",
				},
				Remove: SplunkTAOtelLinuxAutoinstrumentationModInput{
					Value: "false",
					Name:  "remove",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			tt.modInputs.ZeroconfigPath.Value = filepath.Join(tt.testDir, tt.modInputs.ZeroconfigPath.Value)
			tt.modInputs.AutoinstrumentationPath.Value = filepath.Join(tt.testDir, tt.modInputs.AutoinstrumentationPath.Value)
			tt.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value = filepath.Join(tt.testDir, tt.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)

			require.NoError(t, CreateZeroConfigJava(tt.modInputs))

			assert.FileExists(t, tt.modInputs.ZeroconfigPath.Value)
			//assert.FileExists(t, tt.modInputs.AutoinstrumentationPath.Value)
			//assert.FileExists(t, tt.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)
			//testaddon.AssertFileShasEqual(t, expectedJar, tt.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)
		})
	}
}

func TestHappyPath(t *testing.T) {
	defaultModInputs := GetDefaultSplunkTAOtelLinuxAutoinstrumentationModularInputs()

	sourcedir, err := packaging.GetSourceDir()
	require.NoError(t, err)

	addonFunc := func(t *testing.T, addonPath string) error {
		// Copies  "local/inputs.conf"
		err2 := os.CopyFS(addonPath, os.DirFS(filepath.Join(
			sourcedir,
			"pkg",
			"splunk_ta_otel_linux_autoinstrumentation",
			"runner",
			"internal",
			"testdata",
			"happypath",
		)))
		return err2
	}
	zcAddonPath := testaddon.PackAddon(t, &defaultModInputs, addonFunc)
	otelAddonPath := filepath.Join(packaging.GetBuildDir(), "out", "distribution", "Splunk_TA_otel.tgz")
	repackedOtelAddon := testaddon.RepackAddon(t, otelAddonPath, func(tt *testing.T, addonDir string) error {
		// TODO copy over a debug output config
		_, err = fileutils.CopyFile("internal/testdata/happypath/local/ta-agent-config.yaml", filepath.Join(addonDir, "Splunk_TA_otellk", "configs", "ta-agent-config.yaml"))
		require.NoError(tt, err)
		return nil
	})
	tc := testaddon.StartSplunk(t, testaddon.SplunkStartOpts{AddonPaths: []string{zcAddonPath, repackedOtelAddon}, WaitStrategy: wait.ForAll(
		wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/Splunk_TA_otel_linux_autoinstrumentation.log"}),
		wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/Splunk_TA_otel.log"}),
	)})

	// Check Schema
	ctx := context.Background()
	code, output, err := tc.Exec(ctx, []string{"sudo", "/opt/splunk/bin/splunk", "btool", "check", "--debug"})
	assert.NoError(t, err)
	assert.LessOrEqual(t, code, 1)    // Other stanzas may be missing and thus have this be 0 or 1
	assert.GreaterOrEqual(t, code, 0) // bound to [0,1]
	read, err := io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotContains(t, string(read), "Invalid Key in Stanza")

	// check log output
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/opt/splunk/var/log/splunk/Splunk_TA_otel_linux_autoinstrumentation.log"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	//	assert.NotEmpty(t, read)

	// Check zeroconfig value
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/opt/splunk/etc/apps/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/config/zero.conf"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotEmpty(t, read)
	fmt.Println(string(read))
	// TODO check written value of zeroconfig, existence of ldpreload etc

	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/etc/ld.preload"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotEmpty(t, read)
	fmt.Println(string(read))

	_, output, err = tc.Exec(ctx, []string{"sudo", "sha256sum", "/opt/splunk/etc/apps/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/libsplunk_amd64.so"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotEmpty(t, read)
	fmt.Println(string(read))

	//time.Sleep(1 * time.Hour)

	assert.NoError(t, tc.Terminate(ctx))
}
