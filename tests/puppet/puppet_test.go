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

//go:build puppet_integration && !windows

package puppet

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	d "github.com/signalfx/splunk-otel-collector/tests/deploymenttest"
	"github.com/stretchr/testify/require"
)

func selectedDistros(t *testing.T) []string {
	t.Helper()
	base := "packaging/tests/deployments/puppet/images/"
	all := append(d.Distros(t, base+"deb"), d.Distros(t, base+"rpm")...)
	if selected := os.Getenv("DEPLOYMENT_TEST_DISTRO"); selected != "" {
		require.Contains(t, all, selected)
		return []string{selected}
	}
	return all
}
func releases() []string { return strings.Split(d.Env("PUPPET_RELEASE", "8"), ",") }
func container(t *testing.T, distro, release string) *d.Container {
	t.Helper()
	if distro == "opensuse-16" {
		t.Skip("Puppet does not support openSUSE 16")
	}
	n, _ := strconv.Atoi(release)
	if distro == "ubuntu-xenial" && n >= 8 {
		t.Skip("Puppet release unsupported on Xenial")
	}
	kind := "rpm"
	if strings.HasPrefix(distro, "debian-") || strings.HasPrefix(distro, "ubuntu-") {
		kind = "deb"
	}
	return d.Start(t, "packaging/tests/deployments/puppet/images/"+kind+"/Dockerfile."+distro, map[string]*string{"PUPPET_RELEASE": d.Ptr(release)})
}
func apply(t *testing.T, c *d.Container, config string) {
	t.Helper()
	d.CopyText(t, c, config, "/root/test.pp")
	rc, out := d.Try(t, c, "puppet apply --detailed-exitcodes /root/test.pp")
	require.Equalf(t, 2, rc, "puppet apply:\n%s", out)
}
func manifest(lines ...string) string {
	return "class { splunk_otel_collector:\n    splunk_access_token => '" + d.Token + "',\n    splunk_realm => '" + d.Realm + "',\n    " + strings.Join(lines, ",\n    ") + ",\n}\n"
}
func env(t *testing.T, c *d.Container) { t.Helper(); d.EnvFile(t, c, d.APIURL, d.IngestURL, d.Token) }
func TestDefault(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_GROUP") == "instrumentation" {
		t.Skip("base cases run in base matrix")
	}
	for _, distro := range selectedDistros(t) {
		for _, release := range releases() {
			t.Run(distro+"/"+release, func(t *testing.T) {
				c := container(t, distro, release)
				version := d.Env("VERSION", "0.0.1")
				apply(t, c, manifest("collector_version => '"+version+"'"))
				d.PackageVersion(t, c, d.ServiceName, version)
				env(t, c)
				d.KV(t, c, d.EnvPath, "SPLUNK_LISTEN_INTERFACE", ".*", false)
				d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
			})
		}
	}
}
func TestCustom(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_GROUP") == "instrumentation" {
		t.Skip("base cases run in base matrix")
	}
	for _, distro := range selectedDistros(t) {
		for _, release := range releases() {
			t.Run(distro+"/"+release, func(t *testing.T) {
				c := container(t, distro, release)
				version := d.Env("VERSION", "0.0.1")
				apply(t, c, manifest("collector_version => '"+version+"'", "splunk_api_url => 'https://fake-splunk-api.com'", "splunk_ingest_url => 'https://fake-splunk-ingest.com'", "splunk_hec_token => 'fake-hec-token'", "splunk_listen_interface => '0.0.0.0'", "collector_command_line_args => '--discovery --set=processors.batch.timeout=10s'", "collector_additional_env_vars => { 'MY_CUSTOM_VAR1' => 'value1', 'MY_CUSTOM_VAR2' => 'value2', 'SPLUNK_OPAMP_SUPERVISOR_ENABLED' => 'true' }", "service_user => 'custom-user'", "service_group => 'custom-group'"))
				d.PackageVersion(t, c, d.ServiceName, version)
				d.EnvFile(t, c, "https://fake-splunk-api.com", "https://fake-splunk-ingest.com", "fake-hec-token")
				for k, v := range map[string]string{"SPLUNK_LISTEN_INTERFACE": "0.0.0.0", "OTELCOL_OPTIONS": "--discovery --set=processors.batch.timeout=10s", "MY_CUSTOM_VAR1": "value1", "MY_CUSTOM_VAR2": "value2", "SPLUNK_OPAMP_SUPERVISOR_ENABLED": "true"} {
					d.KV(t, c, d.EnvPath, k, v, true)
				}
				d.ServiceRunning(t, c, "custom-user", "/usr/bin/otelcol")
				d.ServiceRunning(t, c, "custom-user", "opampsupervisor")
				for _, p := range []string{d.ConfigDir, "/var/lib/otelcol"} {
					require.Equal(t, "custom-user:custom-group:755", strings.TrimSpace(d.Run(t, c, "stat -c '%U:%G:%a' "+p)))
				}
				require.Zero(t, func() int { rc, _ := d.Try(t, c, "su -s /bin/sh -c 'test -w "+d.ConfigDir+"' custom-user"); return rc }())
				for _, p := range []string{d.ConfigDir + "/agent_config.yaml", d.ConfigDir + "/supervisor", d.ConfigDir + "/supervisor/supervisor_config.yaml", "/var/lib/otelcol/supervisor"} {
					require.Equal(t, "custom-user:custom-group", strings.TrimSpace(d.Run(t, c, "stat -c '%U:%G' "+p)))
				}
			})
		}
	}
}
func instrumentationManifest(version, collector string, systemd, custom bool) string {
	lines := []string{"collector_version => '" + collector + "'", "with_auto_instrumentation => true", "auto_instrumentation_version => '" + version + "'", fmt.Sprintf("auto_instrumentation_systemd => %t", systemd)}
	if custom {
		lines = append(lines, "auto_instrumentation_ld_so_preload => '# my extra library'", "auto_instrumentation_resource_attributes => 'deployment.environment.name=test'", "auto_instrumentation_generate_service_name => false", "auto_instrumentation_disable_telemetry => true", "auto_instrumentation_service_name => 'test'", "auto_instrumentation_enable_profiler => true", "auto_instrumentation_enable_profiler_memory => true", "auto_instrumentation_enable_metrics => true", "auto_instrumentation_otlp_endpoint => 'http://0.0.0.0:4317'", "auto_instrumentation_otlp_endpoint_protocol => 'grpc'", "auto_instrumentation_metrics_exporter => 'none'", "auto_instrumentation_logs_exporter => 'none'")
	}
	return manifest(lines...)
}
func TestInstrumentation(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_GROUP") == "base" {
		t.Skip("instrumentation cases run in the instrumentation matrix")
	}
	for _, distro := range selectedDistros(t) {
		for _, release := range releases() {
			for _, systemd := range []bool{true, false} {
				for _, custom := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/systemd=%t/custom=%t", distro, release, systemd, custom), func(t *testing.T) {
						c := container(t, distro, release)
						version := d.Env("AUTO_INSTRUMENTATION_VERSION", "latest")
						apply(t, c, instrumentationManifest(version, d.Env("VERSION", "0.0.1"), systemd, custom))
						d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", version)
						env(t, c)
						d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
						d.CheckInstrumentation(t, c, d.Instrumentation{Systemd: systemd, Custom: custom, Version: version, Attributes: "splunk.zc.method=splunk-otel-auto-instrumentation-.*"})
					})
				}
			}
		}
	}
}
func TestUpgradeFromLibsplunk(t *testing.T) {
	if os.Getenv("DEPLOYMENT_TEST_GROUP") == "base" {
		t.Skip("instrumentation cases run in the instrumentation matrix")
	}
	for _, distro := range selectedDistros(t) {
		for _, release := range releases() {
			for _, systemd := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/systemd=%t", distro, release, systemd), func(t *testing.T) {
					version := d.Env("AUTO_INSTRUMENTATION_VERSION", "latest")
					require.True(t, d.UsesInjector(version))
					c := container(t, distro, release)
					apply(t, c, instrumentationManifest("0.159.0", d.Env("VERSION", "0.0.1"), systemd, false))
					d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", "0.159.0")
					if systemd {
						d.File(t, c, d.SystemdConfig, true)
						for _, p := range []string{d.JavaConfig, d.NodeConfig} {
							d.File(t, c, p, false)
						}
					} else {
						for _, p := range []string{d.JavaConfig, d.NodeConfig} {
							d.File(t, c, p, true)
						}
						d.Config(t, c, "/etc/ld.so.preload", d.LibSplunk, true)
					}
					apply(t, c, instrumentationManifest(version, d.Env("VERSION", "0.0.1"), systemd, false))
					d.PackageVersion(t, c, "splunk-otel-auto-instrumentation", version)
					env(t, c)
					d.ServiceRunning(t, c, d.ServiceOwner, "/usr/bin/otelcol")
					d.File(t, c, d.LibSplunk, false)
					d.CheckInstrumentation(t, c, d.Instrumentation{Systemd: systemd, Version: version, Attributes: "splunk.zc.method=splunk-otel-auto-instrumentation-.*"})
					if !systemd {
						d.Config(t, c, "/etc/ld.so.preload", d.LibSplunk, false)
					}
				})
			}
		}
	}
}
