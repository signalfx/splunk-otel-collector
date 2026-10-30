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

//go:build salt_integration

package salt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	d "github.com/signalfx/splunk-otel-collector/tests/deploymenttest"
)

const localArtifacts = "/opt/splunk-otel-local-artifacts"

// options maps a package kind ("deb" or "rpm") to its distros, each mapped
// to the "KEY=value" docker build args used to build that distro's image.
type options map[string]map[string][]string

func distros(t *testing.T) ([]string, options) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(d.Root(t), "packaging/tests/deployments/salt/images/distro_docker_opts.yaml"))
	require.NoError(t, err)
	var opts options
	require.NoError(t, yaml.Unmarshal(data, &opts))
	all := []string{}
	for _, group := range opts {
		for name := range group {
			all = append(all, name)
		}
	}
	if name := os.Getenv("DEPLOYMENT_TEST_DISTRO"); name != "" {
		require.Contains(t, all, name)
		return []string{name}, opts
	}
	return all, opts
}

func isDeb(opts options, distro string) bool { _, ok := opts["deb"][distro]; return ok }
func container(t *testing.T, distro string, opts options) *d.Container {
	t.Helper()
	kind := "rpm"
	if isDeb(opts, distro) {
		kind = "deb"
	}
	args := map[string]*string{}
	for _, item := range opts[kind][distro] {
		key, value, _ := strings.Cut(item, "=")
		args[key] = d.Ptr(value)
	}
	binds := []string{}
	if local() {
		binds = append(binds, filepath.Join(d.Root(t), "dist")+":"+localArtifacts+":ro")
	}
	return d.Start(t, "packaging/tests/deployments/salt/images/Dockerfile."+kind, args, binds...)
}
func local() bool { return strings.EqualFold(os.Getenv("LOCAL_ARTIFACT_TESTING_ENABLED"), "true") }
func source(t *testing.T, c *d.Container, distro string, opts options, pkg string) string {
	t.Helper()
	glob := ""
	if isDeb(opts, distro) {
		glob = pkg + "_*amd64.deb"
	} else {
		glob = pkg + "-*x86_64.rpm"
	}
	matches, err := filepath.Glob(filepath.Join(d.Root(t), "dist", glob))
	require.NoError(t, err)
	require.Len(t, matches, 1)
	path := localArtifacts + "/" + filepath.Base(matches[0])
	d.File(t, c, path, true)
	return path
}

func withArtifacts(t *testing.T, c *d.Container, distro string, opts options, config string, instrumentation bool) string {
	t.Helper()
	if !local() {
		return config
	}
	config += fmt.Sprintf(`
  local_artifact_testing_enabled: True
  collector_package_source: '%s'
`, source(t, c, distro, opts, d.ServiceName))
	if instrumentation {
		config += fmt.Sprintf("  auto_instrumentation_package_source: '%s'\n", source(t, c, distro, opts, "splunk-otel-auto-instrumentation"))
	}
	return config
}

func apply(t *testing.T, c *d.Container, config string) {
	t.Helper()
	d.CopyText(t, c, config, "/srv/pillar/splunk-otel-collector.sls")
	d.Run(t, c, "salt-call --local state.apply")
}

func verifyCollector(t *testing.T, c *d.Container) {
	t.Helper()
	if local() {
		d.PackageVersion(t, c, d.ServiceName, d.Env("VERSION", "latest"))
	}
}

func verifyEnv(t *testing.T, c *d.Container, api, ingest, hec, listen, args string) {
	t.Helper()
	d.EnvFile(t, c, api, ingest, hec)
	d.KV(t, c, d.EnvPath, "SPLUNK_CONFIG", d.ConfigDir+"/agent_config.yaml", true)
	d.KV(t, c, d.EnvPath, "SPLUNK_MEMORY_TOTAL_MIB", "512", true)
	if args != "" {
		d.KV(t, c, d.EnvPath, "OTELCOL_OPTIONS", args, true)
	} else {
		d.Config(t, c, d.EnvPath, "OTELCOL_OPTIONS=", true)
	}
	if listen != "" {
		d.KV(t, c, d.EnvPath, "SPLUNK_LISTEN_INTERFACE", listen, true)
	} else {
		d.Config(t, c, d.EnvPath, ".*SPLUNK_LISTEN_INTERFACE.*", false)
	}
}

func defaultConfig() string {
	return fmt.Sprintf(`splunk-otel-collector:
  splunk_access_token: '%s'
  splunk_realm: '%s'
`, d.Token, d.Realm)
}

func customConfig() string {
	return defaultConfig() + `  splunk_ingest_url: 'https://fake-ingest.com'
  splunk_api_url: 'https://fake-api.com'
  splunk_hec_token: 'fake-hec-token'
  collector_version: '0.126.0'
  splunk_service_user: 'test-user'
  splunk_service_group: 'test-user'
  splunk_listen_interface: '0.0.0.0'
  splunk_otel_collector_command_line_args: '--discovery --set=processors.batch.timeout=10s'
  collector_additional_env_vars:
    MY_CUSTOM_VAR1: value1
    MY_CUSTOM_VAR2: value2
    SPLUNK_OPAMP_SUPERVISOR_ENABLED: 'true'
`
}

func platformLogsConfig(accessToken, index string) string {
	config := "splunk-otel-collector:\n"
	if accessToken != "" {
		config += fmt.Sprintf("  splunk_access_token: '%s'\n  splunk_realm: '%s'\n", accessToken, d.Realm)
	}
	config += `  splunk_platform_url: 'https://splunk.example.com:8088/services/collector'
  splunk_platform_token: 'fake-platform-token'
`
	if index != "" {
		config += fmt.Sprintf("  splunk_platform_logs_index: '%s'\n", index)
	}
	return config
}

func applyExpectFailure(t *testing.T, c *d.Container, config, expectedMessage string) {
	t.Helper()
	d.CopyText(t, c, config, "/srv/pillar/splunk-otel-collector.sls")
	rc, output := d.Exec(t, c, "salt-call --local state.apply")
	require.NotZero(t, rc, output)
	require.Contains(t, output, expectedMessage)
}

func absentInstrumentation(t *testing.T, c *d.Container, deb bool) {
	t.Helper()
	cmd := "rpm -q splunk-otel-auto-instrumentation"
	if deb {
		cmd = "dpkg -s splunk-otel-auto-instrumentation"
	}
	rc, _ := d.Exec(t, c, cmd)
	require.NotZero(t, rc)
}

func TestDefault(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_CASE") == "custom" {
		t.Skip("custom case selected")
	}
	all, opts := distros(t)
	for _, distro := range all {
		t.Run(distro, func(t *testing.T) {
			c := container(t, distro, opts)
			apply(t, c, withArtifacts(t, c, distro, opts, defaultConfig(), false))
			verifyCollector(t, c)
			verifyEnv(t, c, d.APIURL, d.IngestURL, d.Token, "", "")
			d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
			absentInstrumentation(t, c, isDeb(opts, distro))
		})
	}
}

func TestCustom(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_CASE") == "default" {
		t.Skip("default case selected")
	}
	all, opts := distros(t)
	for _, distro := range all {
		t.Run(distro, func(t *testing.T) {
			c := container(t, distro, opts)
			apply(t, c, withArtifacts(t, c, distro, opts, customConfig(), false))
			verifyCollector(t, c)
			verifyEnv(t, c, "https://fake-api.com", "https://fake-ingest.com", "fake-hec-token", "0.0.0.0", "--discovery --set=processors.batch.timeout=10s")
			for k, v := range map[string]string{"MY_CUSTOM_VAR1": "value1", "MY_CUSTOM_VAR2": "value2", "SPLUNK_OPAMP_SUPERVISOR_ENABLED": "true"} {
				d.KV(t, c, d.EnvPath, k, v, true)
			}
			d.ServiceRunning(t, c, "test-user", "/usr/bin/otelcol")
			d.ServiceRunning(t, c, "test-user", "opampsupervisor")
			require.Equal(t, "test-user:test-user", strings.TrimSpace(d.Run(t, c, "stat -c '%U:%G' "+d.EnvPath)))
			for _, p := range []string{d.ConfigDir, "/var/lib/otelcol"} {
				require.Equal(t, "test-user:test-user:755", strings.TrimSpace(d.Run(t, c, "stat -c '%U:%G:%a' "+p)))
			}
			for _, p := range []string{d.ConfigDir + "/agent_config.yaml", d.ConfigDir + "/supervisor", d.ConfigDir + "/supervisor/supervisor_config.yaml", "/var/lib/otelcol/supervisor"} {
				require.Equal(t, "test-user:test-user", strings.TrimSpace(d.Run(t, c, "stat -c '%U:%G' "+p)))
			}
			absentInstrumentation(t, c, isDeb(opts, distro))
		})
	}
}

func TestPlatformLogsOnly(t *testing.T) {
	all, opts := distros(t)
	for _, distro := range all {
		t.Run(distro, func(t *testing.T) {
			c := container(t, distro, opts)
			config := withArtifacts(t, c, distro, opts, platformLogsConfig("", ""), false)
			apply(t, c, config)
			verifyCollector(t, c)
			d.KV(t, c, d.EnvPath, "SPLUNK_PLATFORM_URL", "https://splunk.example.com:8088/services/collector", true)
			d.KV(t, c, d.EnvPath, "SPLUNK_PLATFORM_TOKEN", "fake-platform-token", true)
			d.KV(t, c, d.EnvPath, "SPLUNK_FILE_STORAGE_EXTENSION_PATH", "/var/lib/otelcol/filelogs", true)
			d.KV(t, c, d.EnvPath, "OTELCOL_OPTIONS", "--config /etc/otel/collector/splunk_logs_config_linux.yaml", true)
			d.Config(t, c, d.EnvPath, `^SPLUNK_CONFIG=`, false)
			d.Config(t, c, d.EnvPath, `^SPLUNK_ACCESS_TOKEN=`, false)
			d.Config(t, c, d.EnvPath, `^SPLUNK_PLATFORM_LOGS_INDEX=`, false)
			d.KV(t, c, d.EnvPath, "SPLUNK_LISTEN_INTERFACE", "0.0.0.0", true)
			d.Config(t, c, d.ConfigDir+"/splunk_logs_config_linux.yaml", `[[:space:]]*index: "\$\{env:SPLUNK_PLATFORM_LOGS_INDEX:-main\}"`, true)
			require.Equal(t, "700", strings.TrimSpace(d.Run(t, c, "stat -c '%a' /var/lib/otelcol/filelogs")))
			d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
		})
	}
}

func TestPlatformLogsWithObservability(t *testing.T) {
	all, opts := distros(t)
	for _, distro := range all {
		t.Run(distro, func(t *testing.T) {
			c := container(t, distro, opts)
			config := withArtifacts(t, c, distro, opts, platformLogsConfig(d.Token, "custom-index"), false)
			apply(t, c, config)
			verifyCollector(t, c)
			d.KV(t, c, d.EnvPath, "SPLUNK_ACCESS_TOKEN", d.Token, true)
			d.KV(t, c, d.EnvPath, "SPLUNK_PLATFORM_URL", "https://splunk.example.com:8088/services/collector", true)
			d.KV(t, c, d.EnvPath, "SPLUNK_PLATFORM_TOKEN", "fake-platform-token", true)
			d.KV(t, c, d.EnvPath, "SPLUNK_PLATFORM_LOGS_INDEX", "custom-index", true)
			d.KV(t, c, d.EnvPath, "OTELCOL_OPTIONS", "--config /etc/otel/collector/agent_config.yaml --config /etc/otel/collector/splunk_logs_config_linux.yaml --feature-gates=confmap.enableMergeAppendOption", true)
			d.Config(t, c, d.EnvPath, `^SPLUNK_CONFIG=`, false)
			d.KV(t, c, d.EnvPath, "SPLUNK_FILE_STORAGE_EXTENSION_PATH", "/var/lib/otelcol/filelogs", true)
			d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
		})
	}
}

func TestPlatformLogsValidation(t *testing.T) {
	all, opts := distros(t)
	distro := all[0]
	c := container(t, distro, opts)
	for _, test := range []struct {
		name    string
		config  string
		message string
	}{
		{
			name:    "missing both credentials",
			config:  "splunk-otel-collector:\n",
			message: "Provide splunk_access_token or both splunk_platform_url and splunk_platform_token.",
		},
		{
			name: "Platform URL without token",
			config: `splunk-otel-collector:
  splunk_platform_url: 'https://splunk.example.com:8088/services/collector'
`,
			message: "splunk_platform_url and splunk_platform_token must be provided together.",
		},
		{
			name: "Platform token without URL",
			config: `splunk-otel-collector:
  splunk_platform_token: 'fake-platform-token'
`,
			message: "splunk_platform_url and splunk_platform_token must be provided together.",
		},
		{
			name:    "Platform logs index without URL",
			config:  fmt.Sprintf("splunk-otel-collector:\n  splunk_access_token: '%s'\n  splunk_realm: '%s'\n  splunk_platform_logs_index: 'custom-index'\n", d.Token, d.Realm),
			message: "splunk_platform_logs_index requires splunk_platform_url.",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			applyExpectFailure(t, c, test.config, test.message)
		})
	}
}

func instrumentationConfig(version, collector string, systemd, custom bool) string {
	config := defaultConfig()
	if custom {
		config += fmt.Sprintf("  collector_version: '%s'\n", collector)
	}
	config += fmt.Sprintf(`  install_auto_instrumentation: True
  auto_instrumentation_version: '%s'
  auto_instrumentation_systemd: %t
`, version, systemd)
	if custom {
		config += `  auto_instrumentation_ld_so_preload: '# my extra library'
  auto_instrumentation_resource_attributes: 'deployment.environment.name=test'
  auto_instrumentation_service_name: 'test'
  auto_instrumentation_generate_service_name: False
  auto_instrumentation_disable_telemetry: True
  auto_instrumentation_enable_profiler: True
  auto_instrumentation_enable_profiler_memory: True
  auto_instrumentation_enable_metrics: True
  auto_instrumentation_otlp_endpoint: 'http://0.0.0.0:4317'
  auto_instrumentation_otlp_endpoint_protocol: 'grpc'
  auto_instrumentation_metrics_exporter: 'none'
  auto_instrumentation_logs_exporter: 'none'
`
	}
	return config
}

func versions() []string {
	if local() {
		return []string{d.Env("AUTO_INSTRUMENTATION_VERSION", "latest")}
	}
	return []string{"0.86.0", "0.159.0", "latest"}
}

func TestInstrumentation(t *testing.T) {
	all, opts := distros(t)
	for _, distro := range all {
		for _, version := range versions() {
			for _, systemd := range []bool{true, false} {
				for _, custom := range []bool{false, true} {
					if custom && os.Getenv("DEPLOYMENT_TEST_CASE") == "default" || !custom && os.Getenv("DEPLOYMENT_TEST_CASE") == "custom" {
						continue
					}
					t.Run(fmt.Sprintf("%s/%s/systemd=%t/custom=%t", distro, version, systemd, custom), func(t *testing.T) {
						c := container(t, distro, opts)
						collectorVersion := version
						if local() {
							collectorVersion = d.Env("VERSION", "latest")
						}
						config := instrumentationConfig(version, collectorVersion, systemd, custom)
						apply(t, c, withArtifacts(t, c, distro, opts, config, true))
						verifyCollector(t, c)
						d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", version)
						verifyEnv(t, c, d.APIURL, d.IngestURL, d.Token, "", "")
						d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
						d.CheckInstrumentation(t, c, d.Instrumentation{Systemd: systemd, Custom: custom, Version: version, Attributes: "splunk.zc.method=splunk-otel-auto-instrumentation-" + version})
					})
				}
			}
		}
	}
}

func TestUpgradeFromLibsplunk(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_CASE") == "custom" {
		t.Skip("upgrade case runs in default matrix")
	}
	if local() {
		t.Skip("local artifacts contain only the current injector package")
	}
	all, opts := distros(t)
	for _, distro := range all {
		for _, systemd := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/systemd=%t", distro, systemd), func(t *testing.T) {
				version := d.Env("AUTO_INSTRUMENTATION_VERSION", "latest")
				require.True(t, d.UsesInjector(version))
				c := container(t, distro, opts)
				apply(t, c, instrumentationConfig("0.159.0", "", systemd, false))
				d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", "0.159.0")
				for _, p := range []string{d.JavaConfig, d.NodeConfig} {
					d.File(t, c, p, true)
				}
				if systemd {
					d.File(t, c, d.SystemdConfig, true)
				} else {
					d.Config(t, c, "/etc/ld.so.preload", d.LibSplunk, true)
				}
				apply(t, c, instrumentationConfig(version, "", systemd, false))
				verifyCollector(t, c)
				d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", version)
				verifyEnv(t, c, d.APIURL, d.IngestURL, d.Token, "", "")
				d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
				d.File(t, c, d.LibSplunk, false)
				d.CheckInstrumentation(t, c, d.Instrumentation{Systemd: systemd, Version: version, Attributes: "splunk.zc.method=splunk-otel-auto-instrumentation-" + version})
				if !systemd {
					d.Config(t, c, "/etc/ld.so.preload", d.LibSplunk, false)
				}
			})
		}
	}
}
