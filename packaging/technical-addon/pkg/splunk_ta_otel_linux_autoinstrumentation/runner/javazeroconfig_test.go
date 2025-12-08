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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/splunk/splunk-technical-addon/internal/packaging"
	"github.com/splunk/splunk-technical-addon/internal/testaddon"
	"github.com/splunk/splunk-technical-addon/internal/testcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestZeroConfig(t *testing.T) {
	expectedJar := filepath.Join(packaging.GetBuildDir(), "Splunk_TA_otel_linux_autoinstrumentation", "linux_x86_64", "bin", "splunk-otel-javaagent.jar")
	require.FileExists(t, expectedJar)

	tests := []struct {
		testname       string
		testDir        string
		modInputs      *SplunkTAOtelLinuxAutoinstrumentationModularInputs
		expectedConfig string
		preload        bool
	}{
		{
			testname: "happypath-preload",
			testDir:  t.TempDir(),
			preload:  false,
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
					Value: "/etc/ld.so.preload",
					Name:  "autoinstrumentation_preload_path",
				},
				JavaZeroconfigPath: SplunkTAOtelLinuxAutoinstrumentationModInput{
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
			expectedConfig: strings.ReplaceAll(`JAVA_TOOL_OPTIONS=-javaagent:REPLACED_WITH_TESTDIR/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-javaagent.jar
OTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-JAVA_VERSION,asdasd
SPLUNK_PROFILER_ENABLED=false
SPLUNK_PROFILER_MEMORY_ENABLED=false
SPLUNK_METRICS_ENABLED=false
`, "JAVA_VERSION", strings.TrimSpace(javaVersion)),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testname, func(tt *testing.T) {
			require.NoError(tt, os.CopyFS(filepath.Join(tc.testDir, "Splunk_TA_otel_linux_autoinstrumentation"), os.DirFS(filepath.Join(packaging.GetBuildDir(), "Splunk_TA_otel_linux_autoinstrumentation"))))

			tc.modInputs.JavaZeroconfigPath.Value = filepath.Join(tc.testDir, tc.modInputs.JavaZeroconfigPath.Value)
			tc.modInputs.AutoinstrumentationPath.Value = filepath.Join(tc.testDir, tc.modInputs.AutoinstrumentationPath.Value)
			tc.modInputs.AutoinstrumentationPreloadPath.Value = filepath.Join(tc.testDir, tc.modInputs.AutoinstrumentationPreloadPath.Value)
			tc.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value = filepath.Join(tc.testDir, tc.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)

			require.NoError(tt, CreateZeroConfigJava(tc.modInputs))
			assert.FileExists(tt, tc.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)
			testaddon.AssertFileShasEqual(t, expectedJar, tc.modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value)
			assert.FileExists(tt, tc.modInputs.AutoinstrumentationPath.Value)
			if tc.preload {
				assert.FileExists(tt, tc.modInputs.AutoinstrumentationPreloadPath.Value)
			} else {
				assert.NoFileExists(tt, tc.modInputs.AutoinstrumentationPreloadPath.Value)
			}
			require.FileExists(tt, tc.modInputs.JavaZeroconfigPath.Value)
			expectedPath := filepath.Join(tc.testDir, "expected-zeroconfig.conf")
			assert.NoFileExists(tt, expectedPath)
			require.NoError(tt, os.WriteFile(expectedPath, []byte(strings.ReplaceAll(tc.expectedConfig, "REPLACED_WITH_TESTDIR", tc.testDir)), 0o600))
			testcommon.AssertFilesMatch(tt, expectedPath, tc.modInputs.JavaZeroconfigPath.Value)
		})
	}
}

func TestHappyPath(t *testing.T) {
	defaultModInputs := GetDefaultSplunkTAOtelLinuxAutoinstrumentationModularInputs()

	sourcedir, err := packaging.GetSourceDir()
	require.NoError(t, err)

	addonFunc := func(_ *testing.T, addonPath string) error {
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
	startupTimeout := 8 * time.Minute
	tc := testaddon.StartSplunk(t, testaddon.SplunkStartOpts{
		AddonPaths:  []string{zcAddonPath},
		SplunkUser:  "root",
		SplunkGroup: "root",
		Timeout:     startupTimeout,
		WaitStrategy: wait.ForAll(
			wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/splunkd.log"}).WithStartupTimeout(startupTimeout),
			wait.ForExec([]string{"sudo", "stat", "/opt/splunk/var/log/splunk/Splunk_TA_otel_linux_autoinstrumentation.log"}).WithStartupTimeout(startupTimeout),
		).WithDeadline(startupTimeout + 15*time.Second).WithStartupTimeoutDefault(startupTimeout),
	})

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
	assert.Contains(t, string(read), "Successfully generated java autoinstrumentation config at \"/etc/splunk/zeroconfig/java.conf\"")

	// Check zeroconfig value
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/etc/splunk/zeroconfig/java.conf"}, tcexec.Multiplexed())
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("JAVA_TOOL_OPTIONS=-javaagent:/opt/splunk/etc/apps/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-javaagent.jar\nOTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-%s\nSPLUNK_PROFILER_ENABLED=false\nSPLUNK_PROFILER_MEMORY_ENABLED=false\nSPLUNK_METRICS_ENABLED=false", strings.TrimSpace(javaVersion)), strings.TrimSpace(string(read)))

	// Check preload config
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/etc/ld.so.preload"}, tcexec.Multiplexed())
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotEmpty(t, read)
	assert.Equal(t, "/opt/splunk/etc/apps/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/libsplunk_amd64.so", strings.TrimSpace(string(read)))

	// check jar
	_, output, err = tc.Exec(ctx, []string{"sudo", "sha256sum", "/opt/splunk/etc/apps/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-javaagent.jar"}, tcexec.Multiplexed())
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.Contains(t, string(read), strings.TrimSpace(javaAgent256Sum))

	// check for errors
	_, output, err = tc.Exec(ctx, []string{"sudo", "cat", "/opt/splunk/var/log/splunk/Splunk_TA_otel_linux_autoinstrumentation.log"})
	require.NoError(t, err)
	read, err = io.ReadAll(output)
	assert.NoError(t, err)
	assert.NotRegexp(t, regexp.MustCompile(`(?i).*error.*`), string(read))
}
