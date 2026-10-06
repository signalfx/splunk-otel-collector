// Copyright Splunk Inc.
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

//go:build puppet_integration && windows

package puppet

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
)

const winService = "splunk-otel-collector"
const winConfig = `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`
const winModule = `C:\ProgramData\PuppetLabs\code\environments\production\modules\splunk_otel_collector`

func command(t *testing.T, cmd string, ok ...int) string {
	t.Helper()
	out, e := exec.Command("powershell", "-NoProfile", "-Command", cmd).CombinedOutput()
	rc := 0
	if e != nil {
		if ee, yes := e.(*exec.ExitError); yes {
			rc = ee.ExitCode()
		} else {
			require.NoError(t, e)
		}
	}
	if len(ok) == 0 {
		ok = []int{0}
	}
	require.Containsf(t, ok, rc, "%s:\n%s", cmd, out)
	return string(out)
}
func setup(t *testing.T) {
	t.Helper()
	command(t, "choco --version")
	release := env("PUPPET_RELEASE", "latest")
	cmd := "choco upgrade -y -f puppet-agent"
	if release != "latest" {
		cmd += " --version " + release
	}
	command(t, cmd)
	root := root(t)
	_ = os.RemoveAll(winModule)
	command(t, fmt.Sprintf("Copy-Item -Recurse -Force '%s' '%s'", filepath.Join(root, "deployments", "puppet"), winModule))
	for _, module := range []string{"puppet-archive", "puppetlabs-powershell", "puppetlabs-registry"} {
		command(t, "& 'C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat' module install "+module)
	}
}
func root(t *testing.T) string {
	t.Helper()
	wd, e := os.Getwd()
	require.NoError(t, e)
	for {
		if _, e = os.Stat(filepath.Join(wd, "deployments", "puppet")); e == nil {
			return wd
		}
		next := filepath.Dir(wd)
		if next == wd {
			t.Fatal("repository root not found")
		}
		wd = next
	}
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func regEnvValue(t *testing.T, name string) (string, bool) {
	t.Helper()
	k, e := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\splunk-otel-collector`, registry.QUERY_VALUE)
	require.NoError(t, e)
	defer k.Close()
	vars, _, e := k.GetStringsValue("Environment")
	require.NoError(t, e)
	for _, line := range vars {
		if strings.HasPrefix(line, name+"=") {
			return strings.TrimPrefix(line, name+"="), true
		}
	}
	return "", false
}
func regEnv(t *testing.T, name string) string {
	t.Helper()
	value, found := regEnvValue(t, name)
	require.True(t, found, name)
	return value
}
func service(t *testing.T) string {
	t.Helper()
	status := command(t, "sc.exe query "+winService)
	require.Contains(t, status, "RUNNING")
	return command(t, "sc.exe qc "+winService)
}
func supportsArgs() bool {
	v := env("WIN_COLLECTOR_VERSION", "123.456.789")
	if v == "latest" {
		return true
	}
	m := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`).FindStringSubmatch(v)
	if m == nil {
		return false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major > 0 || minor >= 127
}
func apply(t *testing.T, config string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "agent.pp")
	require.NoError(t, os.WriteFile(p, []byte(config), 0o644))
	command(t, fmt.Sprintf("& 'C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat' apply '%s'; if ($LASTEXITCODE -eq 2) { exit 0 } else { exit $LASTEXITCODE }", p))
}
func verify(t *testing.T, api, ingest, hec, listen string) {
	t.Helper()
	for k, v := range map[string]string{"SPLUNK_REALM": "test", "SPLUNK_ACCESS_TOKEN": "testing123", "SPLUNK_API_URL": api, "SPLUNK_INGEST_URL": ingest, "SPLUNK_HEC_URL": ingest + "/v1/log", "SPLUNK_HEC_TOKEN": hec} {
		require.Equal(t, v, regEnv(t, k), k)
	}
	if listen == "" {
		_, found := regEnvValue(t, "SPLUNK_LISTEN_INTERFACE")
		require.False(t, found)
	} else {
		require.Equal(t, listen, regEnv(t, "SPLUNK_LISTEN_INTERFACE"))
	}
	details := service(t)
	if supportsArgs() {
		require.Contains(t, details, `--config "`+winConfig+`"`)
		_, found := regEnvValue(t, "SPLUNK_CONFIG")
		require.False(t, found)
	} else {
		require.Equal(t, winConfig, regEnv(t, "SPLUNK_CONFIG"))
	}
}
func TestWindowsDefault(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_CASE") == "custom_vars" {
		t.Skip("default case selected separately")
	}
	setup(t)
	config := fmt.Sprintf("class { splunk_otel_collector: splunk_access_token => 'testing123', splunk_realm => 'test', collector_version => '%s', win_repo_url => '%s', }", env("WIN_COLLECTOR_VERSION", "123.456.789"), env("LOCAL_MSI_SERVER", "https://dl.observability.splunkcloud.com/splunk-otel-collector/msi/release"))
	apply(t, config)
	verify(t, "https://api.test.observability.splunkcloud.com", "https://ingest.test.observability.splunkcloud.com", "testing123", "")
}
func TestWindowsCustom(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_CASE") == "default" {
		t.Skip("custom case selected separately")
	}
	setup(t)
	config := fmt.Sprintf("class { splunk_otel_collector: splunk_access_token => 'testing123', splunk_realm => 'test', splunk_api_url => 'https://fake-splunk-api.com', splunk_ingest_url => 'https://fake-splunk-ingest.com', splunk_hec_token => 'fake-hec-token', splunk_listen_interface => '0.0.0.0', collector_version => '%s', win_repo_url => '%s', collector_command_line_args => '--discovery --set=processors.batch.timeout=10s', collector_additional_env_vars => { 'MY_CUSTOM_VAR1' => 'value1', 'MY_CUSTOM_VAR2' => 'value2' }, }", env("WIN_COLLECTOR_VERSION", "123.456.789"), env("LOCAL_MSI_SERVER", "https://dl.observability.splunkcloud.com/splunk-otel-collector/msi/release"))
	apply(t, config)
	verify(t, "https://fake-splunk-api.com", "https://fake-splunk-ingest.com", "fake-hec-token", "0.0.0.0")
	require.Equal(t, "value1", regEnv(t, "MY_CUSTOM_VAR1"))
	require.Equal(t, "value2", regEnv(t, "MY_CUSTOM_VAR2"))
	if supportsArgs() {
		require.Contains(t, service(t), "--discovery --set=processors.batch.timeout=10s")
	}
}
