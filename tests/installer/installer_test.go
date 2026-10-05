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

//go:build integration

package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/containerd/platforms"
	dockerContainer "github.com/moby/moby/api/types/container"
	dockerClient "github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	serviceName       = "splunk-otel-collector"
	serviceOwner      = "splunk-otel-collector"
	installScript     = "packaging/installer/install.sh"
	collectorEnv      = "/etc/otel/collector/splunk-otel-collector.conf"
	oldCollectorEnv   = "/etc/otel/collector/splunk_env"
	agentConfig       = "/etc/otel/collector/agent_config.yaml"
	gatewayConfig     = "/etc/otel/collector/gateway_config.yaml"
	oldConfig         = "/etc/otel/collector/splunk_config_linux.yaml"
	preloadPath       = "/etc/ld.so.preload"
	libInject         = "/usr/lib/splunk-instrumentation/libotelinject.so"
	systemdConfig     = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
	injectorConfig    = "/etc/opentelemetry/injector/injector.conf"
	injectorDefault   = "/etc/opentelemetry/injector/default_env.conf"
	nodePrefix        = "/usr/lib/splunk-instrumentation/splunk-otel-js"
	nodePackage       = "/usr/lib/splunk-instrumentation/splunk-otel-js.tgz"
	dotnetHome        = "/usr/lib/splunk-instrumentation/splunk-otel-dotnet"
	obiPath           = "/usr/local/bin/obi"
	installTimeout    = 30 * time.Minute
	defaultRealm      = "fake-realm"
	defaultToken      = "testing123"
	platformToken     = "test-hec-token"
	platformURL       = "https://splunk.example.com:8088/services/collector"
	platformLogsIndex = "test-logs-index"
	platformMetrics   = "test-metrics-index"
)

func TestInstallerDefault(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			for _, mode := range []string{"agent", "gateway"} {
				t.Run(fmt.Sprintf("%s/%s/%s", distro, arch, mode), func(t *testing.T) {
					if installerGroup() != "none" {
						t.Skip("base installer cases run in the none matrix")
					}
					c := installerContainer(t, distro, arch, false, nil, nil)
					installCmd := installerCommand(t, true)
					if mode != "agent" {
						installCmd += " --mode " + mode
					}
					run(t, c, installTimeout, "VERIFY_ACCESS_TOKEN=false "+installCmd)
					time.Sleep(5 * time.Second)
					assertNotInstalled(t, c, distro, "splunk-otel-auto-instrumentation")
					verifyEnv(t, c, mode, "512", "")
					require.Eventually(t, func() bool { return serviceRunning(t, c, serviceOwner) }, 30*time.Second, time.Second)
					verifySupportBundle(t, c)
					uninstall(t, c, distro)
				})
			}
		}
	}
}

func TestInstallerCustom(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				if installerGroup() != "none" {
					t.Skip("base installer cases run in the none matrix")
				}
				c := installerContainer(t, distro, arch, false, nil, nil)
				config := filepath.Join(repoRoot(t), "packaging/tests/custom-config.yaml")
				copyInto(t, c, config, "/etc/my-custom-config.yaml")
				cmd := installerCommand(t, true) + " --listen-interface 10.0.0.1 --memory 256 --service-user test-user --service-group test-user --collector-config /etc/my-custom-config.yaml"
				if os.Getenv("LOCAL_COLLECTOR_PACKAGE") == "" {
					cmd += " --collector-version 0.126.0"
				}
				run(t, c, installTimeout, "VERIFY_ACCESS_TOKEN=false "+cmd)
				time.Sleep(5 * time.Second)
				if os.Getenv("LOCAL_COLLECTOR_PACKAGE") == "" {
					require.Equal(t, "otelcol version v0.126.0", strings.TrimSpace(run(t, c, time.Minute, "otelcol --version")))
				}
				verifyEnv(t, c, "agent", "256", "10.0.0.1", "/etc/my-custom-config.yaml")
				require.Eventually(t, func() bool { return serviceRunning(t, c, "test-user") }, 30*time.Second, time.Second)
				assertFail(t, c, "getent passwd "+serviceOwner)
				assertFail(t, c, "getent group "+serviceOwner)
				require.Equal(t, "test-user:test-user", strings.TrimSpace(run(t, c, time.Minute, "stat -c '%U:%G' /etc/otel")))
				uninstall(t, c, distro)
			})
		}
	}
}

func TestInstallerInstrumentation(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			for _, method := range []string{"preload", "systemd"} {
				for _, sdk := range []string{"all", "java", "node", "dotnet"} {
					t.Run(fmt.Sprintf("%s/%s/%s/%s", distro, arch, method, sdk), func(t *testing.T) {
						if installerGroup() != method {
							t.Skip("instrumentation method is selected by the CI matrix")
						}
						nodeVersion := "v18"
						if arch == "arm64" && distro == "centos-7" {
							nodeVersion = "v14"
						}
						c := installerContainer(t, distro, arch, true, map[string]*string{"NODE_VERSION": ptr(nodeVersion)}, nil)
						keep := "# This line should be preserved"
						run(t, c, time.Minute, "echo "+shellQuote(keep)+" >> "+preloadPath)
						if e := os.Getenv("LOCAL_INSTRUMENTATION_PACKAGE"); e != "" {
							copyInto(t, c, e, "/test/instrumentation.pkg")
						}
						run(t, c, time.Minute, "npm config set global true")
						cmd := installerCommand(t, true)
						if method == "systemd" {
							cmd += " --with-systemd-instrumentation"
						} else {
							cmd += " --with-instrumentation"
						}
						if sdk != "all" {
							cmd += " --with-instrumentation-sdk " + sdk + " --deployment-environment deployment_environment_from_" + method + " --service-name service_name_from_" + method + " --enable-profiler --enable-profiler-memory --enable-metrics --otlp-endpoint http://0.0.0.0:4318 --otlp-endpoint-protocol http/protobuf --metrics-exporter none --logs-exporter none"
						}
						if os.Getenv("LOCAL_INSTRUMENTATION_PACKAGE") != "" {
							cmd += " --instrumentation-version /test/instrumentation.pkg"
						}
						run(t, c, installTimeout, "VERIFY_ACCESS_TOKEN=false "+cmd)
						time.Sleep(5 * time.Second)
						verifyEnv(t, c, "agent", "512", "")
						require.Eventually(t, func() bool { return serviceRunning(t, c, serviceOwner) }, 30*time.Second, time.Second)
						assertInstalled(t, c, distro, "splunk-otel-auto-instrumentation")
						if sdk == "all" || sdk == "node" {
							assertNodeInstalled(t, c, true)
						} else {
							assertNodeInstalled(t, c, false)
						}
						assertFail(t, c, "sh -l -c 'npm ls --global=true @splunk/otel'")
						if sdk == "all" || sdk == "dotnet" {
							dotnetArch := "x64"
							if arch == "arm64" {
								dotnetArch = "arm64"
							}
							assertFile(t, c, dotnetHome+"/glibc/linux-"+dotnetArch+"/OpenTelemetry.AutoInstrumentation.Native.so", true)
						}
						version := installedInstrumentationVersion(t, c, distro)
						zc := "splunk-otel-auto-instrumentation-" + strings.ReplaceAll(version, "~", "-")
						if method == "systemd" {
							zc += "-systemd"
						}
						resourceAttrs := "splunk\\.zc\\.method=" + regexp.QuoteMeta(zc)
						if sdk != "all" {
							resourceAttrs += ",deployment\\.environment\\.name=deployment_environment_from_" + method
						}
						if method == "preload" {
							assertConfig(t, c, preloadPath, libInject, true)
							assertFile(t, c, systemdConfig, false)
							disabled := map[string]string{"java": "auto_instrumentation_disabled=nodejs,dotnet", "node": "auto_instrumentation_disabled=jvm,dotnet", "dotnet": "auto_instrumentation_disabled=jvm,nodejs"}
							if sdk == "all" {
								assertConfig(t, c, injectorConfig, "auto_instrumentation_disabled=.*", false)
							} else {
								assertConfig(t, c, injectorConfig, disabled[sdk], true)
							}
							configPath := injectorDefault
							assertConfig(t, c, configPath, "OTEL_RESOURCE_ATTRIBUTES="+resourceAttrs, true)
							assertConfig(t, c, configPath, "SPLUNK_PROFILER_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							assertConfig(t, c, configPath, "SPLUNK_PROFILER_MEMORY_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							assertConfig(t, c, configPath, "SPLUNK_METRICS_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							if sdk == "all" {
								for _, key := range []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_SERVICE_NAME", "OTEL_METRICS_EXPORTER", "OTEL_LOGS_EXPORTER", "OTEL_EXPORTER_OTLP_PROTOCOL"} {
									assertConfig(t, c, configPath, key+"=.*", false)
								}
							}
						} else {
							assertConfig(t, c, preloadPath, libInject, false)
							assertConfig(t, c, systemdConfig, "NODE_OPTIONS=-r "+nodePrefix+"/node_modules/@splunk/otel/instrument", sdk == "all" || sdk == "node")
							assertConfig(t, c, systemdConfig, "JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", sdk == "all" || sdk == "java")
							if sdk == "all" || sdk == "dotnet" {
								verifyDotnetConfig(t, c, arch, true)
							} else {
								verifyDotnetConfig(t, c, arch, false)
							}
							assertConfig(t, c, systemdConfig, "OTEL_RESOURCE_ATTRIBUTES="+resourceAttrs, true)
							assertConfig(t, c, systemdConfig, "SPLUNK_PROFILER_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							assertConfig(t, c, systemdConfig, "SPLUNK_PROFILER_MEMORY_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							assertConfig(t, c, systemdConfig, "SPLUNK_METRICS_ENABLED="+map[bool]string{true: "true", false: "false"}[sdk != "all"], true)
							if sdk == "all" {
								for _, key := range []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_SERVICE_NAME", "OTEL_METRICS_EXPORTER", "OTEL_LOGS_EXPORTER", "OTEL_EXPORTER_OTLP_PROTOCOL"} {
									assertConfig(t, c, systemdConfig, key+"=.*", false)
								}
							}
						}
						if sdk != "all" {
							for _, pair := range [][2]string{{"SPLUNK_PROFILER_ENABLED", "true"}, {"SPLUNK_PROFILER_MEMORY_ENABLED", "true"}, {"SPLUNK_METRICS_ENABLED", "true"}, {"OTEL_EXPORTER_OTLP_ENDPOINT", "http://0.0.0.0:4318"}, {"OTEL_SERVICE_NAME", "service_name_from_" + method}, {"OTEL_METRICS_EXPORTER", "none"}, {"OTEL_LOGS_EXPORTER", "none"}, {"OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf"}} {
								p := injectorDefault
								if method == "systemd" {
									p = systemdConfig
								}
								assertConfig(t, c, p, pair[0]+"="+pair[1], true)
							}
						}
						uninstall(t, c, distro)
						assertConfig(t, c, preloadPath, keep, true)
					})
				}
			}
		}
	}
}

func TestInstallerOBI(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				if installerGroup() != "none" {
					t.Skip("OBI is included in the none matrix")
				}
				if !bpffsMounted() {
					t.Skip("bpffs is not mounted on the test host")
				}
				c := installerContainer(t, distro, arch, false, nil, map[string]string{"/sys/fs/bpf": "/sys/fs/bpf:rw"})
				cmd := installerCommand(t, true) + " --with-obi --obi-version " + envDefault("OBI_VERSION", "v0.6.0")
				run(t, c, installTimeout, "VERIFY_ACCESS_TOKEN=false "+cmd)
				time.Sleep(5 * time.Second)
				require.Eventually(t, func() bool { return serviceRunning(t, c, serviceOwner) }, 30*time.Second, time.Second)
				assertFile(t, c, obiPath, true)
				output := runAllowFailure(t, c, time.Minute, obiPath+" --version")
				require.Regexp(t, `\b\d+\.\d+\.\d+\b`, output)
				run(t, c, time.Minute, "sh /test/install.sh --uninstall --with-obi")
				assertFile(t, c, obiPath, false)
			})
		}
	}
}

func TestInstallerRejectsLegacyInstrumentation(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			for _, version := range []string{"0.150.0", "0.159.0"} {
				t.Run(fmt.Sprintf("%s/%s/%s", distro, arch, version), func(t *testing.T) {
					if installerGroup() != "none" {
						t.Skip("validation cases are included in the none matrix")
					}
					c := installerContainer(t, distro, arch, false, nil, nil)
					out := runExpectFailure(t, c, installTimeout, "VERIFY_ACCESS_TOKEN=false "+installerCommand(t, true)+" --with-instrumentation --instrumentation-version "+version)
					require.Contains(t, out, "requires the OpenTelemetry injector")
				})
			}
		}
	}
}

func TestInstallerSplunkPlatformValidation(t *testing.T) {
	for _, distro := range selectedDistros(t, "deb", "rpm") {
		for _, arch := range selectedArches(t) {
			for _, tc := range []struct{ name, args, want string }{
				{"gateway", "--splunk-platform-token " + envDefault("SPLUNK_PLATFORM_TOKEN", platformToken) + " --splunk-platform-url " + envDefault("SPLUNK_PLATFORM_URL", platformURL) + " --splunk-platform-logs-index " + platformLogsIndex + " --mode gateway", "not supported in gateway mode"},
				{"missing-token", "--splunk-platform-url " + envDefault("SPLUNK_PLATFORM_URL", platformURL) + " --splunk-platform-logs-index " + platformLogsIndex, "--splunk-platform-token is required"},
				{"missing-logs-url", "--splunk-platform-token " + envDefault("SPLUNK_PLATFORM_TOKEN", platformToken) + " --splunk-platform-logs-index " + platformLogsIndex, "--splunk-platform-url is required when --splunk-platform-token is set"},
				{"missing-metrics-url", "--splunk-platform-token " + envDefault("SPLUNK_PLATFORM_TOKEN", platformToken) + " --splunk-platform-metrics-index " + platformMetrics, "--splunk-platform-url is required when --splunk-platform-token is set"},
			} {
				t.Run(fmt.Sprintf("%s/%s/%s", distro, arch, tc.name), func(t *testing.T) {
					if installerGroup() != "none" {
						t.Skip("validation cases are included in the none matrix")
					}
					c := installerContainer(t, distro, arch, false, nil, nil)
					out := runExpectFailure(t, c, installTimeout, platformInstallerCommand(t)+" "+tc.args)
					require.Contains(t, out, tc.want)
				})
			}
		}
	}
}

func installerContainer(t *testing.T, distro, arch string, instrumentation bool, buildargs map[string]*string, binds map[string]string) *testutils.Container {
	t.Helper()
	root := repoRoot(t)
	kind := packageTypeForDistro(t, distro, instrumentation)
	imageDir := "images"
	if instrumentation {
		imageDir = filepath.Join("instrumentation", "images")
	}
	dockerfile, err := filepath.Rel(root, filepath.Join(root, "packaging", "tests", imageDir, kind, "Dockerfile."+distro))
	require.NoError(t, err)
	if buildargs == nil {
		buildargs = map[string]*string{}
	}
	target := arch
	buildargs["TARGETARCH"] = &target
	p, err := platforms.Parse("linux/" + arch)
	require.NoError(t, err)
	container := testutils.NewContainer().WithContext(root).WithDockerfile(dockerfile).WithBuildArgs(buildargs).WithDockerfileBuildOptionsModifier(func(o *dockerClient.ImageBuildOptions) { o.PullParent = true; o.Platforms = append(o.Platforms, p) }).WithImagePlatform("linux/" + arch).WithPrivileged(true).WithBinds("/sys/fs/cgroup:/sys/fs/cgroup:rw").WithHostConfigModifier(func(h *dockerContainer.HostConfig) { h.CgroupnsMode = dockerContainer.CgroupnsModeHost }).WithStartupTimeout(5 * time.Minute)
	for host, binding := range binds {
		container = container.WithBinds(host + ":" + binding)
	}
	container.WaitingFor = append(container.WaitingFor, wait.ForNop(func(context.Context, wait.StrategyTarget) error { return nil }).WithStartupTimeout(5*time.Minute))
	built := container.Build()
	require.NoError(t, built.Start(context.Background()))
	t.Cleanup(func() { require.NoError(t, built.Terminate(context.Background())) })
	t.Cleanup(func() {
		_, stdout, stderr := built.AssertExec(t, time.Minute, "sh", "-c", "journalctl -u "+serviceName+" --no-pager")
		if stdout != "" {
			t.Log(stdout)
		}
		if stderr != "" {
			t.Log(stderr)
		}
	})
	timeout := 10 * time.Second
	if arch != "amd64" {
		timeout = 30 * time.Second
	}
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rc, _, e := built.Exec(ctx, []string{"sh", "-c", "systemctl show-environment"})
		return e == nil && rc == 0
	}, timeout, time.Second)
	require.NoError(t, built.CopyFileToContainer(context.Background(), filepath.Join(root, installScript), "/test/install.sh", 0o644))
	if local := os.Getenv("LOCAL_COLLECTOR_PACKAGE"); local != "" {
		require.NoError(t, built.CopyFileToContainer(context.Background(), local, "/test/collector.pkg", 0o644))
		if kind == "deb" {
			run(t, built, 5*time.Minute, "apt-get update && apt-get install -y libcap2-bin")
		}
	}
	return built
}

func installerCommand(t *testing.T, withToken bool) string {
	t.Helper()
	cmd := "sh -l /test/install.sh"
	if os.Getenv("DEBUG") == "yes" {
		cmd = "sh -l -x /test/install.sh"
	}
	if withToken {
		cmd += " -- " + shellQuote(envDefault("SPLUNK_ACCESS_TOKEN", defaultToken)) + " --realm " + shellQuote(envDefault("SPLUNK_REALM", defaultRealm))
	}
	if local := os.Getenv("LOCAL_COLLECTOR_PACKAGE"); local != "" {
		cmd += " --collector-version /test/collector.pkg --skip-collector-repo"
	} else if v := os.Getenv("VERSION"); v != "" && v != "latest" {
		cmd += " --collector-version " + strings.TrimPrefix(v, "v")
	}
	stage := envDefault("STAGE", "release")
	if stage != "release" {
		require.Contains(t, []string{"test", "beta"}, stage)
		cmd += " --" + stage
	}
	return cmd
}

func platformInstallerCommand(t *testing.T) string {
	cmd := "sh -l /test/install.sh"
	if os.Getenv("DEBUG") == "yes" {
		cmd = "sh -l -x /test/install.sh"
	}
	if local := os.Getenv("LOCAL_COLLECTOR_PACKAGE"); local != "" {
		cmd += " --collector-version /test/collector.pkg --skip-collector-repo"
	} else if v := os.Getenv("VERSION"); v != "" && v != "latest" {
		cmd += " --collector-version " + strings.TrimPrefix(v, "v")
	}
	stage := envDefault("STAGE", "release")
	if stage != "release" {
		require.Contains(t, []string{"test", "beta"}, stage)
		cmd += " --" + stage
	}
	return cmd
}

func verifyEnv(t *testing.T, c *testutils.Container, mode, memory, listen string, custom ...string) {
	p := collectorEnv
	if fileExists(t, c, oldCollectorEnv) {
		p = oldCollectorEnv
	}
	cfg := agentConfig
	if mode == "gateway" {
		cfg = gatewayConfig
	}
	if len(custom) > 0 {
		cfg = custom[0]
	}
	if fileExists(t, c, oldConfig) {
		cfg = oldConfig
	} else if mode == "gateway" && !fileExists(t, c, gatewayConfig) {
		cfg = agentConfig
	}
	realm := envDefault("SPLUNK_REALM", defaultRealm)
	token := envDefault("SPLUNK_ACCESS_TOKEN", defaultToken)
	ingest := "https://ingest." + realm + ".observability.splunkcloud.com"
	for _, line := range []string{"SPLUNK_CONFIG=" + cfg, "SPLUNK_ACCESS_TOKEN=" + token, "SPLUNK_REALM=" + realm, "SPLUNK_API_URL=https://api." + realm + ".observability.splunkcloud.com", "SPLUNK_INGEST_URL=" + ingest, "SPLUNK_HEC_URL=" + ingest + "/v1/log", "SPLUNK_HEC_TOKEN=" + token, "SPLUNK_MEMORY_TOTAL_MIB=" + memory} {
		assertConfig(t, c, p, line, true)
	}
	if listen != "" {
		assertConfig(t, c, p, "SPLUNK_LISTEN_INTERFACE="+listen, true)
	} else {
		assertConfig(t, c, p, "SPLUNK_LISTEN_INTERFACE=", false)
	}
}

func verifySupportBundle(t *testing.T, c *testutils.Container) {
	run(t, c, 5*time.Minute, "/etc/otel/collector/splunk-support-bundle.sh -t /tmp/splunk-support-bundle")
	for _, p := range []string{"config/agent_config.yaml", "logs/splunk-otel-collector.log", "logs/splunk-otel-collector.txt", "metrics/collector-metrics.txt", "metrics/df.txt", "metrics/free.txt", "metrics/top.txt", "zpages/tracez.html"} {
		assertFile(t, c, "/tmp/splunk-support-bundle/"+p, true)
	}
	assertFile(t, c, "/tmp/splunk-support-bundle.tar.gz", true)
}

func uninstall(t *testing.T, c *testutils.Container, distro string) {
	run(t, c, time.Minute, "sh -l /test/install.sh --uninstall")
	assertNotInstalled(t, c, distro, "splunk-otel-collector")
	assertNotInstalled(t, c, distro, "splunk-otel-auto-instrumentation")
	assertConfig(t, c, preloadPath, libInject, false)
	assertFile(t, c, systemdConfig, false)
	if fileExists(t, c, nodePackage) {
		assertNodeInstalled(t, c, false)
	}
}

func installedInstrumentationVersion(t *testing.T, c *testutils.Container, distro string) string {
	command := "rpm -q --queryformat='%{VERSION}' splunk-otel-auto-instrumentation"
	if packageTypeForDistro(t, distro, false) == "deb" {
		command = "dpkg-query --showformat='${Version}' --show splunk-otel-auto-instrumentation"
	}
	return strings.TrimSpace(run(t, c, time.Minute, command))
}

func verifyDotnetConfig(t *testing.T, c *testutils.Container, arch string, want bool) {
	dotnetArch := "x64"
	if arch == "arm64" {
		dotnetArch = "arm64"
	}
	vars := map[string]string{
		"CORECLR_ENABLE_PROFILING": "1",
		"CORECLR_PROFILER":         "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
		"CORECLR_PROFILER_PATH":    dotnetHome + "/glibc/linux-" + dotnetArch + "/OpenTelemetry.AutoInstrumentation.Native.so",
		"DOTNET_ADDITIONAL_DEPS":   dotnetHome + "/glibc/AdditionalDeps",
		"DOTNET_SHARED_STORE":      dotnetHome + "/glibc/store",
		"DOTNET_STARTUP_HOOKS":     dotnetHome + "/glibc/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll",
		"OTEL_DOTNET_AUTO_HOME":    dotnetHome + "/glibc",
		"OTEL_DOTNET_AUTO_PLUGINS": "Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation",
	}
	for key, value := range vars {
		assertConfig(t, c, systemdConfig, key+"="+regexp.QuoteMeta(value), want)
	}
}

func assertInstalled(t *testing.T, c *testutils.Container, distro, pkg string) {
	q := "rpm -q " + pkg
	if packageTypeForDistro(t, distro, false) == "deb" {
		q = "dpkg -s " + pkg
	}
	run(t, c, time.Minute, q)
}

func assertNotInstalled(t *testing.T, c *testutils.Container, distro, pkg string) {
	q := "rpm -q " + pkg
	if packageTypeForDistro(t, distro, false) == "deb" {
		q = "dpkg -s " + pkg
	}
	assertFail(t, c, q)
}

func assertNodeInstalled(t *testing.T, c *testutils.Container, want bool) {
	cmd := "sh -l -c 'cd " + nodePrefix + " >/dev/null 2>&1 && npm ls --global=false @splunk/otel'"
	if want {
		run(t, c, time.Minute, cmd)
	} else {
		assertFail(t, c, cmd)
	}
}

func assertConfig(t *testing.T, c *testutils.Container, path, pattern string, want bool) {
	t.Helper()
	if !fileExists(t, c, path) {
		require.False(t, want, "%s is missing", path)
		return
	}
	content := run(t, c, time.Minute, "cat "+path)
	if path == systemdConfig {
		pattern = "DefaultEnvironment=\"" + pattern + "\""
	}
	matched, err := regexp.MatchString("(?m)^"+pattern+"$", content)
	require.NoError(t, err)
	require.Equalf(t, want, matched, "config line %q in %s:\n%s", pattern, path, content)
}

func fileExists(t *testing.T, c *testutils.Container, path string) bool {
	rc, _, _ := exec(t, c, time.Minute, "test -f "+path)
	return rc == 0
}

func assertFile(t *testing.T, c *testutils.Container, path string, want bool) {
	require.Equalf(t, want, fileExists(t, c, path), "file %s presence", path)
}

func serviceRunning(t *testing.T, c *testutils.Container, owner string) bool {
	rc, _, _ := exec(t, c, time.Minute, "systemctl status "+serviceName)
	if rc != 0 {
		return false
	}
	rc, _, _ = exec(t, c, time.Minute, "pgrep -a -u "+owner+" -x otelcol")
	return rc == 0
}

func run(t *testing.T, c *testutils.Container, timeout time.Duration, cmd string) string {
	t.Helper()
	rc, out, errout := exec(t, c, timeout, cmd)
	require.Equalf(t, 0, rc, "command %q failed\nstdout:\n%s\nstderr:\n%s", cmd, out, errout)
	return out
}

func runAllowFailure(t *testing.T, c *testutils.Container, timeout time.Duration, cmd string) string {
	_, out, errout := exec(t, c, timeout, cmd)
	return out + errout
}

func runExpectFailure(t *testing.T, c *testutils.Container, timeout time.Duration, cmd string) string {
	t.Helper()
	rc, out, errout := exec(t, c, timeout, cmd)
	require.NotZero(t, rc, "command unexpectedly passed: %s", cmd)
	return out + errout
}

func assertFail(t *testing.T, c *testutils.Container, cmd string) {
	t.Helper()
	rc, _, _ := exec(t, c, time.Minute, cmd)
	require.NotZero(t, rc, "command unexpectedly passed: %s", cmd)
}

func exec(t *testing.T, c *testutils.Container, timeout time.Duration, cmd string) (int, string, string) {
	t.Helper()
	return c.AssertExec(t, timeout, "sh", "-c", cmd)
}

func copyInto(t *testing.T, c *testutils.Container, src, dst string) {
	t.Helper()
	run(t, c, time.Minute, "mkdir -p "+filepath.Dir(dst))
	require.NoError(t, c.CopyFileToContainer(context.Background(), src, dst, 0o644))
}

func selectedDistros(t *testing.T, types ...string) []string {
	t.Helper()
	root := repoRoot(t)
	out := []string{}
	for _, typ := range types {
		dir := filepath.Join(root, "packaging/tests/images", typ)
		if installerGroup() != "none" {
			dir = filepath.Join(root, "packaging/tests/instrumentation/images", typ)
		}
		files, _ := filepath.Glob(filepath.Join(dir, "Dockerfile.*"))
		for _, f := range files {
			out = append(out, strings.TrimPrefix(filepath.Base(f), "Dockerfile."))
		}
	}
	slices.Sort(out)
	out = slices.Compact(out)
	if d := os.Getenv("INSTALLER_TEST_DISTRO"); d != "" {
		if !slices.Contains(out, d) && installerGroup() != "none" {
			t.Skipf("no instrumentation image is available for distro %s", d)
		}
		require.Contains(t, out, d)
		return []string{d}
	}
	return out
}

func selectedArches(t *testing.T) []string {
	t.Helper()
	if a := os.Getenv("INSTALLER_TEST_ARCH"); a != "" {
		require.Contains(t, []string{"amd64", "arm64"}, a)
		return []string{a}
	}
	return []string{"amd64", "arm64"}
}

func packageTypeForDistro(t *testing.T, distro string, instr bool) string {
	t.Helper()
	dir := "images"
	if instr {
		dir = filepath.Join("instrumentation", "images")
	}
	for _, typ := range []string{"deb", "rpm"} {
		if _, err := os.Stat(filepath.Join(repoRoot(t), "packaging/tests", dir, typ, "Dockerfile."+distro)); err == nil {
			return typ
		}
	}
	t.Fatalf("no package image for %s", distro)
	return ""
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err = os.Stat(filepath.Join(wd, "packaging/tests/images")); err == nil {
			return wd
		}
		p := filepath.Dir(wd)
		if p == wd {
			t.Fatal("repository root not found")
		}
		wd = p
	}
}
func installerGroup() string { return envDefault("INSTALLER_TEST_INSTRUMENTATION", "none") }
func envDefault(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}
func ptr(s string) *string       { return &s }
func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
func bpffsMounted() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	b, e := os.ReadFile("/proc/mounts")
	if e != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) > 2 && f[1] == "/sys/fs/bpf" && f[2] == "bpf" {
			return true
		}
	}
	return false
}
