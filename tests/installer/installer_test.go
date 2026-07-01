// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package installer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	dockerContainer "github.com/moby/moby/api/types/container"
	dockerClient "github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	splunkAccessToken = "testing123"
	splunkRealm       = "fake-realm"
	totalMemoryMiB    = "512"

	serviceName  = "splunk-otel-collector"
	serviceOwner = "splunk-otel-collector"
	otelcolBin   = "/usr/bin/otelcol"

	splunkEnvPath     = "/etc/otel/collector/splunk-otel-collector.conf"
	oldSplunkEnvPath  = "/etc/otel/collector/splunk_env"
	agentConfigPath   = "/etc/otel/collector/agent_config.yaml"
	gatewayConfigPath = "/etc/otel/collector/gateway_config.yaml"
	oldConfigPath     = "/etc/otel/collector/splunk_config_linux.yaml"

	libsplunkPath     = "/usr/lib/splunk-instrumentation/libsplunk.so"
	javaAgentPath     = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
	preloadPath       = "/etc/ld.so.preload"
	systemdConfigPath = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
	nodePackagePath   = "/usr/lib/splunk-instrumentation/splunk-otel-js.tgz"

	javaZeroconfigPath   = "/etc/splunk/zeroconfig/java.conf"
	nodeZeroconfigPath   = "/etc/splunk/zeroconfig/node.conf"
	dotnetZeroconfigPath = "/etc/splunk/zeroconfig/dotnet.conf"

	nodePrefix        = "/usr/lib/splunk-instrumentation/splunk-otel-js"
	dotnetHome        = "/usr/lib/splunk-instrumentation/splunk-otel-dotnet"
	obiInstallDir     = "/usr/local/bin"
	obiVersionDefault = "v0.6.0"

	installerTimeout = 30 * time.Minute

	splunkPlatformToken        = "test-hec-token"
	splunkPlatformURL          = "https://splunk.example.com:8088/services/collector"
	splunkPlatformLogsIndex    = "test-logs-index"
	splunkPlatformMetricsIndex = "test-metrics-index"
)

var (
	javaToolOptions = "-javaagent:" + javaAgentPath
	nodeOptions     = "-r " + nodePrefix + "/node_modules/@splunk/otel/instrument"
	dotnetAgentPath = dotnetHome + "/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
	obiBin          = filepath.Join(obiInstallDir, "obi")

	dotnetVars = map[string]string{
		"CORECLR_ENABLE_PROFILING": "1",
		"CORECLR_PROFILER":         "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
		"CORECLR_PROFILER_PATH":    dotnetAgentPath,
		"DOTNET_ADDITIONAL_DEPS":   dotnetHome + "/AdditionalDeps",
		"DOTNET_SHARED_STORE":      dotnetHome + "/store",
		"DOTNET_STARTUP_HOOKS":     dotnetHome + "/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll",
		"OTEL_DOTNET_AUTO_HOME":    dotnetHome,
		"OTEL_DOTNET_AUTO_PLUGINS": "Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation",
	}
)

func TestInstallerDefault(t *testing.T) {
	for _, distro := range packageDistros(t) {
		for _, arch := range []string{"amd64", "arm64"} {
			for _, mode := range []string{"agent", "gateway"} {
				distro, arch, mode := distro, arch, mode
				t.Run(fmt.Sprintf("%s/%s/%s", distro.name, arch, mode), func(t *testing.T) {
					container, shutdown := runDistroContainer(t, distro, arch, nil, nil)
					defer shutdown()
					defer journalctl(t, container)

					copyInstaller(t, container)

					cmd := installerCmd(t)
					if mode != "agent" {
						cmd += " --mode " + mode
					}
					runContainerCmd(t, container, cmd, runOptions{env: verifyTokenEnv(), timeout: installerTimeout})
					time.Sleep(5 * time.Second)

					require.False(t, packageInstalled(t, container, distro.kind, "splunk-otel-auto-instrumentation"))
					verifyEnvFile(t, container, envFileOptions{mode: mode})
					require.Eventually(t, func() bool {
						return serviceIsRunning(t, container, serviceOwner)
					}, 10*time.Second, time.Second)
					verifySupportBundle(t, container)
					verifyUninstall(t, container, distro.kind)
				})
			}
		}
	}
}

func TestInstallerCustom(t *testing.T) {
	const (
		collectorVersion = "0.126.0"
		customUser       = "test-user"
		customConfig     = "/etc/my-custom-config.yaml"
	)

	for _, distro := range packageDistros(t) {
		for _, arch := range []string{"amd64", "arm64"} {
			distro, arch := distro, arch
			t.Run(fmt.Sprintf("%s/%s", distro.name, arch), func(t *testing.T) {
				container, shutdown := runDistroContainer(t, distro, arch, nil, nil)
				defer shutdown()
				defer journalctl(t, container)

				require.NoError(t, container.CopyFileToContainer(context.Background(), filepath.Join(repoRoot(t), "packaging", "tests", "custom-config.yaml"), customConfig, 0o644))
				copyInstaller(t, container)

				cmd := strings.Join([]string{
					installerCmd(t),
					"--listen-interface 10.0.0.1",
					"--memory 256",
					"--service-user " + customUser,
					"--service-group " + customUser,
					"--collector-config " + customConfig,
					"--collector-version " + collectorVersion,
				}, " ")
				runContainerCmd(t, container, cmd, runOptions{env: verifyTokenEnv(), timeout: installerTimeout})
				time.Sleep(5 * time.Second)

				_, output := runContainerCmd(t, container, "otelcol --version", runOptions{})
				require.Equal(t, "otelcol version v"+collectorVersion, strings.TrimSpace(output))

				verifyEnvFile(t, container, envFileOptions{
					configPath: customConfig,
					memory:     "256",
					listenAddr: "10.0.0.1",
				})
				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container, customUser)
				}, 10*time.Second, time.Second)
				assert.NotZero(t, execStatus(t, container, "getent passwd "+serviceOwner))
				assert.NotZero(t, execStatus(t, container, "getent group "+serviceOwner))

				_, owner := runContainerCmd(t, container, "stat -c '%U:%G' /etc/otel", runOptions{})
				require.Equal(t, customUser+":"+customUser, strings.TrimSpace(owner))

				verifyUninstall(t, container, distro.kind)
			})
		}
	}
}

func TestInstallerWithInstrumentationDefault(t *testing.T) {
	for _, distro := range instrumentationDistros(t) {
		for _, arch := range []string{"amd64", "arm64"} {
			for _, method := range []string{"preload", "systemd"} {
				distro, arch, method := distro, arch, method
				t.Run(fmt.Sprintf("%s/%s/%s", distro.name, arch, method), func(t *testing.T) {
					nodeVersion := "v18"
					if arch == "arm64" && distro.name == "centos-7" {
						nodeVersion = "v14"
					}
					container, shutdown := runDistroContainer(t, distro, arch, map[string]string{"NODE_VERSION": nodeVersion}, nil)
					defer shutdown()

					copyInstaller(t, container)
					runContainerCmd(t, container, `sh -c 'echo "# This line should be preserved" >> `+preloadPath+`'`, runOptions{})
					copyLocalInstrumentationPackage(t, container)
					runContainerCmd(t, container, "sh -l -c 'npm config set global true'", runOptions{})

					cmd := installerCmd(t)
					if method == "systemd" {
						cmd += " --with-systemd-instrumentation"
					} else {
						cmd += " --with-instrumentation"
					}
					if os.Getenv("LOCAL_INSTRUMENTATION_PACKAGE") != "" {
						cmd += " --instrumentation-version /test/instrumentation.pkg"
					}
					runContainerCmd(t, container, cmd, runOptions{env: verifyTokenEnv(), timeout: installerTimeout})
					time.Sleep(5 * time.Second)

					version := installedInstrumentationPackageVersion(t, container, distro.kind)
					zcMethod := zcMethod(method, version)

					verifyEnvFile(t, container, envFileOptions{})
					verifyConfigFile(t, container, preloadPath, "# This line should be preserved", "", true)
					require.Eventually(t, func() bool {
						return serviceIsRunning(t, container, serviceOwner)
					}, 10*time.Second, time.Second)
					require.True(t, packageInstalled(t, container, distro.kind, "splunk-otel-auto-instrumentation"))
					require.True(t, nodePackageInstalled(t, container))
					require.NotZero(t, execStatus(t, container, "sh -l -c 'npm ls --global=true @splunk/otel'"), "splunk-otel-js installed globally")
					if arch == "amd64" {
						require.True(t, containerFileExists(t, container, dotnetAgentPath))
					}

					configAttributes := `splunk\.zc\.method=` + zcMethod
					if method == "preload" {
						verifyConfigFile(t, container, preloadPath, libsplunkPath, "", true)
						require.False(t, containerFileExists(t, container, systemdConfigPath))
						verifyDefaultPreloadInstrumentationConfig(t, container, arch, configAttributes)
					} else {
						verifyConfigFile(t, container, preloadPath, `.*`+regexp.QuoteMeta(libsplunkPath)+`.*`, "", false)
						verifyDefaultSystemdInstrumentationConfig(t, container, arch, configAttributes)
					}

					verifyUninstall(t, container, distro.kind)
					verifyConfigFile(t, container, preloadPath, "# This line should be preserved", "", true)
				})
			}
		}
	}
}

func TestInstallerWithInstrumentationCustom(t *testing.T) {
	for _, distro := range instrumentationDistros(t) {
		for _, arch := range []string{"amd64", "arm64"} {
			for _, method := range []string{"preload", "systemd"} {
				for _, sdk := range []string{"java", "node", "dotnet"} {
					distro, arch, method, sdk := distro, arch, method, sdk
					t.Run(fmt.Sprintf("%s/%s/%s/%s", distro.name, arch, method, sdk), func(t *testing.T) {
						nodeVersion := "v18"
						if arch == "arm64" && distro.name == "centos-7" {
							nodeVersion = "v14"
						}
						container, shutdown := runDistroContainer(t, distro, arch, map[string]string{"NODE_VERSION": nodeVersion}, nil)
						defer shutdown()

						copyInstaller(t, container)
						runContainerCmd(t, container, `sh -c 'echo "# This line should be preserved" >> `+preloadPath+`'`, runOptions{})
						copyLocalInstrumentationPackage(t, container)
						runContainerCmd(t, container, "sh -l -c 'npm config set global true'", runOptions{})

						serviceName := "service_name_from_" + method
						environment := "deployment_environment_from_" + method
						cmd := installerCmd(t)
						if method == "systemd" {
							cmd += " --with-systemd-instrumentation"
						} else {
							cmd += " --with-instrumentation"
						}
						cmd += " --with-instrumentation-sdk " + sdk
						cmd += " --deployment-environment " + environment
						cmd += " --service-name " + serviceName
						cmd += " --enable-profiler --enable-profiler-memory --enable-metrics"
						cmd += " --otlp-endpoint http://0.0.0.0:4318"
						cmd += " --otlp-endpoint-protocol http/protobuf"
						cmd += " --metrics-exporter none --logs-exporter none"
						if os.Getenv("LOCAL_INSTRUMENTATION_PACKAGE") != "" {
							cmd += " --instrumentation-version /test/instrumentation.pkg"
						}

						expectedExitCode := 0
						if sdk == "dotnet" && arch != "amd64" {
							expectedExitCode = 1
						}
						_, output := runContainerCmd(t, container, cmd, runOptions{env: verifyTokenEnv(), expectedExitCode: &expectedExitCode, timeout: installerTimeout})
						if sdk == "dotnet" && arch != "amd64" {
							verifyConfigFile(t, container, preloadPath, `.*`+regexp.QuoteMeta(libsplunkPath)+`.*`, "", false)
							verifyConfigFile(t, container, preloadPath, "# This line should be preserved", "", true)
							require.False(t, containerFileExists(t, container, systemdConfigPath))
							require.Contains(t, output, ".NET auto instrumentation is not currently supported")
							return
						}
						time.Sleep(5 * time.Second)

						version := installedInstrumentationPackageVersion(t, container, distro.kind)
						zcMethod := zcMethod(method, version)

						verifyEnvFile(t, container, envFileOptions{})
						verifyConfigFile(t, container, preloadPath, "# This line should be preserved", "", true)
						require.Eventually(t, func() bool {
							return serviceIsRunning(t, container, serviceOwner)
						}, 10*time.Second, time.Second)
						require.True(t, packageInstalled(t, container, distro.kind, "splunk-otel-auto-instrumentation"))
						if sdk == "node" {
							require.True(t, nodePackageInstalled(t, container))
						} else {
							require.False(t, nodePackageInstalled(t, container))
						}
						require.NotZero(t, execStatus(t, container, "sh -l -c 'npm ls --global=true @splunk/otel'"), "splunk-otel-js installed globally")
						if arch == "amd64" {
							require.True(t, containerFileExists(t, container, dotnetAgentPath))
						}

						configAttributes := strings.Join([]string{
							`splunk\.zc\.method=` + zcMethod,
							`deployment\.environment=` + environment,
						}, ",")
						if method == "preload" {
							verifyConfigFile(t, container, preloadPath, libsplunkPath, "", true)
							require.False(t, containerFileExists(t, container, systemdConfigPath))
							verifyCustomPreloadInstrumentationConfig(t, container, sdk, configAttributes, serviceName)
						} else {
							verifyConfigFile(t, container, preloadPath, `.*`+regexp.QuoteMeta(libsplunkPath)+`.*`, "", false)
							verifyCustomSystemdInstrumentationConfig(t, container, sdk, configAttributes, serviceName)
						}

						verifyUninstall(t, container, distro.kind)
						verifyConfigFile(t, container, preloadPath, "# This line should be preserved", "", true)
					})
				}
			}
		}
	}
}

func TestInstallerWithOBI(t *testing.T) {
	if !bpffsMountedOnHost() {
		t.Skip("bpffs not mounted on test host at /sys/fs/bpf; required for OBI")
	}

	for _, distro := range packageDistros(t) {
		for _, arch := range []string{"amd64", "arm64"} {
			distro, arch := distro, arch
			t.Run(fmt.Sprintf("%s/%s", distro.name, arch), func(t *testing.T) {
				container, shutdown := runDistroContainer(t, distro, arch, nil, []string{"/sys/fs/bpf:/sys/fs/bpf:rw"})
				defer shutdown()

				copyInstaller(t, container)
				runContainerCmd(t, container, installerCmd(t)+" --with-obi --obi-version "+envOrDefault("OBI_VERSION", obiVersionDefault), runOptions{env: verifyTokenEnv(), timeout: installerTimeout})
				time.Sleep(5 * time.Second)

				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container, serviceOwner)
				}, 10*time.Second, time.Second)
				require.True(t, containerFileExists(t, container, obiBin), "OBI binary not found at %s", obiBin)

				_, output := runContainerCmd(t, container, obiBin+" --version", runOptions{allowAnyExitCode: true})
				require.Regexp(t, regexp.MustCompile(`\b\d+\.\d+\.\d+\b`), output)

				runContainerCmd(t, container, uninstallCmd()+" --with-obi", runOptions{})
				require.False(t, containerFileExists(t, container, obiBin), "OBI binary was not removed from %s after uninstall", obiBin)
			})
		}
	}
}

func TestInstallerSplunkPlatformValidation(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "logs-gateway-mode",
			args: []string{
				"--splunk-platform-token " + splunkPlatformToken,
				"--splunk-platform-url " + splunkPlatformURL,
				"--splunk-platform-logs-index " + splunkPlatformLogsIndex,
				"--mode gateway",
			},
			expectedError: "not supported in gateway mode",
		},
		{
			name: "logs-missing-token",
			args: []string{
				"--splunk-platform-url " + splunkPlatformURL,
				"--splunk-platform-logs-index " + splunkPlatformLogsIndex,
			},
			expectedError: "--splunk-platform-token is required",
		},
		{
			name: "logs-missing-url",
			args: []string{
				"--splunk-platform-token " + splunkPlatformToken,
				"--splunk-platform-logs-index " + splunkPlatformLogsIndex,
			},
			expectedError: "--splunk-platform-url is required when --splunk-platform-token or --splunk-platform-logs-index or --splunk-platform-metrics-index is set",
		},
		{
			name: "metrics-missing-url",
			args: []string{
				"--splunk-platform-token " + splunkPlatformToken,
				"--splunk-platform-metrics-index " + splunkPlatformMetricsIndex,
			},
			expectedError: "--splunk-platform-url is required when --splunk-platform-token or --splunk-platform-logs-index or --splunk-platform-metrics-index is set",
		},
	}

	expectedExitCode := 1
	for _, tc := range tests {
		for _, distro := range packageDistros(t) {
			for _, arch := range []string{"amd64", "arm64"} {
				tc, distro, arch := tc, distro, arch
				t.Run(fmt.Sprintf("%s/%s/%s", tc.name, distro.name, arch), func(t *testing.T) {
					container, shutdown := runDistroContainer(t, distro, arch, nil, nil)
					defer shutdown()

					copyInstaller(t, container)
					copyLocalCollectorPackageForPlatformTest(t, container, distro.kind)

					cmd := strings.Join(append([]string{platformInstallerCmd(t)}, tc.args...), " ") + " 2>&1"
					_, output := runContainerCmd(t, container, cmd, runOptions{expectedExitCode: &expectedExitCode, timeout: installerTimeout})
					require.Contains(t, output, tc.expectedError)
				})
			}
		}
	}
}

type distroKind string

const (
	debDistro distroKind = "deb"
	rpmDistro distroKind = "rpm"
)

type distro struct {
	name           string
	kind           distroKind
	contextDir     string
	dockerfilePath string
}

type runOptions struct {
	env              map[string]string
	expectedExitCode *int
	allowAnyExitCode bool
	timeout          time.Duration
}

type envFileOptions struct {
	mode       string
	configPath string
	memory     string
	listenAddr string
}

func packageDistros(t *testing.T) []distro {
	return append(
		distrosFrom(t, filepath.Join(repoRoot(t), "packaging", "tests"), filepath.Join("images", "deb"), debDistro),
		distrosFrom(t, filepath.Join(repoRoot(t), "packaging", "tests"), filepath.Join("images", "rpm"), rpmDistro)...,
	)
}

func instrumentationDistros(t *testing.T) []distro {
	// The instrumentation Dockerfiles use `COPY instrumentation/setup-*.sh`,
	// so the build context must be packaging/tests (the parent), not packaging/tests/instrumentation.
	contextDir := filepath.Join(repoRoot(t), "packaging", "tests")
	return append(
		distrosFrom(t, contextDir, filepath.Join("instrumentation", "images", "deb"), debDistro),
		distrosFrom(t, contextDir, filepath.Join("instrumentation", "images", "rpm"), rpmDistro)...,
	)
}

func distrosFrom(t *testing.T, contextDir, dockerfileDir string, kind distroKind) []distro {
	matches, err := filepath.Glob(filepath.Join(contextDir, dockerfileDir, "Dockerfile.*"))
	require.NoError(t, err)
	sort.Strings(matches)

	distros := make([]distro, 0, len(matches))
	for _, match := range matches {
		name := strings.TrimPrefix(filepath.Base(match), "Dockerfile.")
		rel, err := filepath.Rel(contextDir, match)
		require.NoError(t, err)
		distros = append(distros, distro{
			name:           name,
			kind:           kind,
			contextDir:     contextDir,
			dockerfilePath: rel,
		})
	}
	require.NotEmpty(t, distros, "no distros found in %s", filepath.Join(contextDir, dockerfileDir))
	return distros
}

func runDistroContainer(t *testing.T, distro distro, arch string, buildArgs map[string]string, extraBinds []string) (*testutils.Container, func()) {
	t.Helper()

	args := map[string]*string{"TARGETARCH": ptr(arch)}
	for key, value := range buildArgs {
		args[key] = ptr(value)
	}

	startupTimeout := startupTimeoutForArch(arch)
	container := testutils.NewContainer().
		WithContext(distro.contextDir).
		WithDockerfile(distro.dockerfilePath).
		WithBuildArgs(args).
		WithDockerfileBuildOptionsModifier(func(opts *dockerClient.ImageBuildOptions) {
			opts.PullParent = true
			opts.Platforms = []ocispec.Platform{{OS: "linux", Architecture: arch}}
		}).
		WithImagePlatform("linux/" + arch).
		WithPrivileged(true).
		WithBinds(append([]string{"/sys/fs/cgroup:/sys/fs/cgroup:rw"}, extraBinds...)...).
		WithHostConfigModifier(func(hc *dockerContainer.HostConfig) {
			hc.CgroupnsMode = dockerContainer.CgroupnsModeHost
		})
	// testcontainers requires at least one wait strategy; use a no-op here
	// since the real readiness check is the require.Eventually loop below.
	container.WaitingFor = append(container.WaitingFor,
		wait.ForNop(func(context.Context, wait.StrategyTarget) error { return nil }).WithStartupTimeout(startupTimeout),
	)
	const maxBuildAttempts = 3
	var built *testutils.Container
	var startErr error
	for attempt := range maxBuildAttempts {
		built = container.Build()
		startErr = built.Start(context.Background())
		if startErr == nil {
			break
		}
		if attempt < maxBuildAttempts-1 {
			t.Logf("container start attempt %d failed: %v, retrying...", attempt+1, startErr)
			time.Sleep(5 * time.Second)
		}
	}
	require.NoError(t, startErr)

	require.Eventually(t, func() bool {
		return execStatus(t, built, "systemctl show-environment") == 0
	}, startupTimeout, time.Second)

	return built, func() {
		assert.NoError(t, built.Terminate(context.Background()))
	}
}

func startupTimeoutForArch(arch string) time.Duration {
	if arch == "amd64" {
		return 10 * time.Second
	}
	return 30 * time.Second
}

func installerCmd(t *testing.T) string {
	t.Helper()
	cmd := "sh -l " + debugFlag() + " /test/install.sh -- " + splunkAccessToken + " --realm " + splunkRealm
	if version := envOrDefault("VERSION", "latest"); version != "latest" {
		cmd += " --collector-version " + strings.TrimPrefix(version, "v")
	}
	if stage := envOrDefault("STAGE", "release"); stage != "release" {
		require.Contains(t, []string{"test", "beta"}, stage, "unsupported stage")
		cmd += " --" + stage
	}
	return strings.Join(strings.Fields(cmd), " ")
}

func platformInstallerCmd(t *testing.T) string {
	t.Helper()
	cmd := strings.Join(strings.Fields("sh -l "+debugFlag()+" /test/install.sh"), " ")
	if os.Getenv("LOCAL_COLLECTOR_PACKAGE") != "" {
		cmd += " --collector-version /test/collector.pkg --skip-collector-repo"
	} else if version := envOrDefault("VERSION", "latest"); version != "latest" {
		cmd += " --collector-version " + strings.TrimPrefix(version, "v")
	}
	if stage := envOrDefault("STAGE", "release"); stage != "release" {
		require.Contains(t, []string{"test", "beta"}, stage, "unsupported stage")
		cmd += " --" + stage
	}
	return cmd
}

func uninstallCmd() string {
	return strings.Join(strings.Fields("sh -l "+debugFlag()+" /test/install.sh --uninstall"), " ")
}

func debugFlag() string {
	if os.Getenv("DEBUG") == "yes" {
		return "-x"
	}
	return ""
}

func verifyTokenEnv() map[string]string {
	return map[string]string{"VERIFY_ACCESS_TOKEN": "false"}
}

func runContainerCmd(t *testing.T, container *testutils.Container, cmd string, opts runOptions) (int, string) {
	t.Helper()

	if opts.timeout > 0 {
		cmd = fmt.Sprintf("timeout %s %s", shellDuration(opts.timeout), cmd)
	}
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout+time.Minute)
	if opts.timeout == 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	}
	defer cancel()

	execOpts := []exec.ProcessOption{}
	if len(opts.env) > 0 {
		env := make([]string, 0, len(opts.env))
		for key, value := range opts.env {
			env = append(env, key+"="+value)
		}
		sort.Strings(env)
		execOpts = append(execOpts, exec.WithEnv(env))
	}

	rc, reader, err := container.Exec(ctx, []string{"sh", "-lc", cmd}, execOpts...)
	require.NoError(t, err)
	require.NotNil(t, reader)

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	require.NoError(t, err)
	output := stdout.String() + stderr.String()
	t.Logf("ran %q: exit=%d\n%s", cmd, rc, output)

	if !opts.allowAnyExitCode {
		expectedExitCode := 0
		if opts.expectedExitCode != nil {
			expectedExitCode = *opts.expectedExitCode
		}
		require.Equal(t, expectedExitCode, rc, output)
	}
	return rc, output
}

func execStatus(t *testing.T, container *testutils.Container, cmd string) int {
	t.Helper()
	status, _ := runContainerCmd(t, container, cmd, runOptions{allowAnyExitCode: true})
	return status
}

func containerFileExists(t *testing.T, container *testutils.Container, path string) bool {
	t.Helper()
	return execStatus(t, container, "test -f "+shellQuote(path)) == 0
}

func packageInstalled(t *testing.T, container *testutils.Container, kind distroKind, name string) bool {
	t.Helper()
	if kind == debDistro {
		return execStatus(t, container, "dpkg -s "+shellQuote(name)) == 0
	}
	return execStatus(t, container, "rpm -q "+shellQuote(name)) == 0
}

func verifyConfigFile(t *testing.T, container *testutils.Container, path, key, value string, exists bool) {
	t.Helper()
	if exists {
		require.True(t, containerFileExists(t, container, path), "%s does not exist", path)
	} else if !containerFileExists(t, container, path) {
		return
	}

	_, content := runContainerCmd(t, container, "cat "+shellQuote(path), runOptions{})
	line := key
	if value != "" {
		line = key + "=" + value
	}
	if path == systemdConfigPath {
		line = `DefaultEnvironment="` + line + `"`
	}

	matched, err := regexp.MatchString(`(?m)^`+line+`$`, content)
	require.NoError(t, err)
	if exists {
		require.True(t, matched, "%q not found in %s:\n%s", line, path, content)
	} else {
		require.False(t, matched, "%q found in %s:\n%s", line, path, content)
	}
}

func verifyEnvFile(t *testing.T, container *testutils.Container, opts envFileOptions) {
	t.Helper()
	mode := opts.mode
	if mode == "" {
		mode = "agent"
	}
	memory := opts.memory
	if memory == "" {
		memory = totalMemoryMiB
	}

	envPath := splunkEnvPath
	if containerFileExists(t, container, oldSplunkEnvPath) {
		envPath = oldSplunkEnvPath
	}

	configPath := opts.configPath
	if configPath == "" {
		if mode == "agent" {
			configPath = agentConfigPath
		} else {
			configPath = gatewayConfigPath
		}
		if containerFileExists(t, container, oldConfigPath) {
			configPath = oldConfigPath
		} else if mode == "gateway" && !containerFileExists(t, container, gatewayConfigPath) {
			configPath = agentConfigPath
		}
	}

	ingestURL := "https://ingest." + splunkRealm + ".observability.splunkcloud.com"
	apiURL := "https://api." + splunkRealm + ".observability.splunkcloud.com"

	verifyConfigFile(t, container, envPath, "SPLUNK_CONFIG", regexp.QuoteMeta(configPath), true)
	verifyConfigFile(t, container, envPath, "SPLUNK_ACCESS_TOKEN", splunkAccessToken, true)
	verifyConfigFile(t, container, envPath, "SPLUNK_REALM", splunkRealm, true)
	verifyConfigFile(t, container, envPath, "SPLUNK_API_URL", regexp.QuoteMeta(apiURL), true)
	verifyConfigFile(t, container, envPath, "SPLUNK_INGEST_URL", regexp.QuoteMeta(ingestURL), true)
	verifyConfigFile(t, container, envPath, "SPLUNK_HEC_URL", regexp.QuoteMeta(ingestURL+"/v1/log"), true)
	verifyConfigFile(t, container, envPath, "SPLUNK_HEC_TOKEN", splunkAccessToken, true)
	verifyConfigFile(t, container, envPath, "SPLUNK_MEMORY_TOTAL_MIB", memory, true)
	if opts.listenAddr != "" {
		verifyConfigFile(t, container, envPath, "SPLUNK_LISTEN_INTERFACE", opts.listenAddr, true)
	} else {
		verifyConfigFile(t, container, envPath, "SPLUNK_LISTEN_INTERFACE", ".*", false)
	}
}

func verifySupportBundle(t *testing.T, container *testutils.Container) {
	t.Helper()
	runContainerCmd(t, container, "/etc/otel/collector/splunk-support-bundle.sh -t /tmp/splunk-support-bundle", runOptions{})
	for _, path := range []string{
		"/tmp/splunk-support-bundle/config/agent_config.yaml",
		"/tmp/splunk-support-bundle/logs/splunk-otel-collector.log",
		"/tmp/splunk-support-bundle/logs/splunk-otel-collector.txt",
		"/tmp/splunk-support-bundle/metrics/collector-metrics.txt",
		"/tmp/splunk-support-bundle/metrics/df.txt",
		"/tmp/splunk-support-bundle/metrics/free.txt",
		"/tmp/splunk-support-bundle/metrics/top.txt",
		"/tmp/splunk-support-bundle/zpages/tracez.html",
		"/tmp/splunk-support-bundle.tar.gz",
	} {
		require.True(t, containerFileExists(t, container, path), "%s does not exist", path)
	}
}

func verifyUninstall(t *testing.T, container *testutils.Container, kind distroKind) {
	t.Helper()
	runContainerCmd(t, container, uninstallCmd(), runOptions{})
	for _, pkg := range []string{"splunk-otel-collector", "splunk-otel-auto-instrumentation"} {
		require.False(t, packageInstalled(t, container, kind, pkg), "%s was not uninstalled", pkg)
	}
	verifyConfigFile(t, container, preloadPath, `.*`+regexp.QuoteMeta(libsplunkPath)+`.*`, "", false)
	require.False(t, containerFileExists(t, container, systemdConfigPath))
	if containerFileExists(t, container, nodePackagePath) {
		require.False(t, nodePackageInstalled(t, container))
	}
}

func serviceIsRunning(t *testing.T, container *testutils.Container, owner string) bool {
	t.Helper()
	systemctl := execStatus(t, container, "systemctl status "+serviceName)
	pgrep := execStatus(t, container, "pgrep -a -u "+owner+" -f "+otelcolBin)
	return systemctl == 0 && pgrep == 0
}

func installedInstrumentationPackageVersion(t *testing.T, container *testutils.Container, kind distroKind) string {
	t.Helper()
	if kind == debDistro {
		_, output := runContainerCmd(t, container, "dpkg-query --showformat='${Version}' --show splunk-otel-auto-instrumentation", runOptions{})
		return strings.TrimSpace(output)
	}
	_, output := runContainerCmd(t, container, "rpm -q --queryformat='%{VERSION}' splunk-otel-auto-instrumentation", runOptions{})
	return strings.TrimSpace(output)
}

func zcMethod(method, version string) string {
	zcMethod := "splunk-otel-auto-instrumentation-" + strings.ReplaceAll(version, "~", "-")
	if method == "systemd" {
		zcMethod += "-systemd"
	}
	return zcMethod
}

func nodePackageInstalled(t *testing.T, container *testutils.Container) bool {
	t.Helper()
	return execStatus(t, container, "sh -l -c 'cd "+nodePrefix+" >/dev/null 2>&1 && npm ls --global=false @splunk/otel'") == 0
}

func verifyDotnetConfig(t *testing.T, container *testutils.Container, path string, exists bool) {
	t.Helper()
	for key, value := range dotnetVars {
		if !exists {
			value = ".*"
		}
		verifyConfigFile(t, container, path, key, regexp.QuoteMeta(value), exists)
	}
}

func verifyDefaultPreloadInstrumentationConfig(t *testing.T, container *testutils.Container, arch, configAttributes string) {
	t.Helper()
	verifyConfigFile(t, container, javaZeroconfigPath, "JAVA_TOOL_OPTIONS", regexp.QuoteMeta(javaToolOptions), true)
	verifyConfigFile(t, container, nodeZeroconfigPath, "NODE_OPTIONS", regexp.QuoteMeta(nodeOptions), true)
	configPaths := []string{javaZeroconfigPath, nodeZeroconfigPath}
	if arch == "amd64" {
		verifyDotnetConfig(t, container, dotnetZeroconfigPath, true)
		configPaths = append(configPaths, dotnetZeroconfigPath)
	} else {
		require.False(t, containerFileExists(t, container, dotnetZeroconfigPath))
	}
	for _, configPath := range configPaths {
		verifyCommonDefaultInstrumentationConfig(t, container, configPath, configAttributes)
	}
}

func verifyDefaultSystemdInstrumentationConfig(t *testing.T, container *testutils.Container, arch, configAttributes string) {
	t.Helper()
	verifyConfigFile(t, container, systemdConfigPath, "NODE_OPTIONS", regexp.QuoteMeta(nodeOptions), true)
	verifyConfigFile(t, container, systemdConfigPath, "JAVA_TOOL_OPTIONS", regexp.QuoteMeta(javaToolOptions), true)
	verifyCommonDefaultInstrumentationConfig(t, container, systemdConfigPath, configAttributes)
	verifyDotnetConfig(t, container, systemdConfigPath, arch == "amd64")
}

func verifyCommonDefaultInstrumentationConfig(t *testing.T, container *testutils.Container, configPath, configAttributes string) {
	t.Helper()
	verifyConfigFile(t, container, configPath, "OTEL_RESOURCE_ATTRIBUTES", configAttributes, true)
	verifyConfigFile(t, container, configPath, "SPLUNK_PROFILER_ENABLED", "false", true)
	verifyConfigFile(t, container, configPath, "SPLUNK_PROFILER_MEMORY_ENABLED", "false", true)
	verifyConfigFile(t, container, configPath, "SPLUNK_METRICS_ENABLED", "false", true)
	verifyConfigFile(t, container, configPath, "OTEL_EXPORTER_OTLP_ENDPOINT", ".*", false)
	verifyConfigFile(t, container, configPath, "OTEL_SERVICE_NAME", ".*", false)
	verifyConfigFile(t, container, configPath, "OTEL_METRICS_EXPORTER", ".*", false)
	verifyConfigFile(t, container, configPath, "OTEL_LOGS_EXPORTER", ".*", false)
	verifyConfigFile(t, container, configPath, "OTEL_EXPORTER_OTLP_PROTOCOL", ".*", false)
}

func verifyCustomPreloadInstrumentationConfig(t *testing.T, container *testutils.Container, sdk, configAttributes, serviceName string) {
	t.Helper()
	configPath := ""
	switch sdk {
	case "java":
		configPath = javaZeroconfigPath
		verifyConfigFile(t, container, configPath, "JAVA_TOOL_OPTIONS", regexp.QuoteMeta(javaToolOptions), true)
		require.False(t, containerFileExists(t, container, nodeZeroconfigPath))
		require.False(t, containerFileExists(t, container, dotnetZeroconfigPath))
	case "node":
		configPath = nodeZeroconfigPath
		verifyConfigFile(t, container, configPath, "NODE_OPTIONS", regexp.QuoteMeta(nodeOptions), true)
		require.False(t, containerFileExists(t, container, javaZeroconfigPath))
		require.False(t, containerFileExists(t, container, dotnetZeroconfigPath))
	case "dotnet":
		configPath = dotnetZeroconfigPath
		verifyDotnetConfig(t, container, configPath, true)
		require.False(t, containerFileExists(t, container, javaZeroconfigPath))
		require.False(t, containerFileExists(t, container, nodeZeroconfigPath))
	}
	verifyCommonCustomInstrumentationConfig(t, container, configPath, configAttributes, serviceName)
}

func verifyCustomSystemdInstrumentationConfig(t *testing.T, container *testutils.Container, sdk, configAttributes, serviceName string) {
	t.Helper()
	switch sdk {
	case "java":
		verifyConfigFile(t, container, systemdConfigPath, "JAVA_TOOL_OPTIONS", regexp.QuoteMeta(javaToolOptions), true)
		verifyConfigFile(t, container, systemdConfigPath, "NODE_OPTIONS", ".*", false)
		verifyDotnetConfig(t, container, systemdConfigPath, false)
	case "node":
		verifyConfigFile(t, container, systemdConfigPath, "NODE_OPTIONS", regexp.QuoteMeta(nodeOptions), true)
		verifyConfigFile(t, container, systemdConfigPath, "JAVA_TOOL_OPTIONS", ".*", false)
		verifyDotnetConfig(t, container, systemdConfigPath, false)
	case "dotnet":
		verifyDotnetConfig(t, container, systemdConfigPath, true)
		verifyConfigFile(t, container, systemdConfigPath, "JAVA_TOOL_OPTIONS", ".*", false)
		verifyConfigFile(t, container, systemdConfigPath, "NODE_OPTIONS", ".*", false)
	}
	verifyCommonCustomInstrumentationConfig(t, container, systemdConfigPath, configAttributes, serviceName)
}

func verifyCommonCustomInstrumentationConfig(t *testing.T, container *testutils.Container, configPath, configAttributes, serviceName string) {
	t.Helper()
	verifyConfigFile(t, container, configPath, "OTEL_RESOURCE_ATTRIBUTES", configAttributes, true)
	verifyConfigFile(t, container, configPath, "SPLUNK_PROFILER_ENABLED", "true", true)
	verifyConfigFile(t, container, configPath, "SPLUNK_PROFILER_MEMORY_ENABLED", "true", true)
	verifyConfigFile(t, container, configPath, "SPLUNK_METRICS_ENABLED", "true", true)
	verifyConfigFile(t, container, configPath, "OTEL_EXPORTER_OTLP_ENDPOINT", regexp.QuoteMeta("http://0.0.0.0:4318"), true)
	verifyConfigFile(t, container, configPath, "OTEL_SERVICE_NAME", serviceName, true)
	verifyConfigFile(t, container, configPath, "OTEL_METRICS_EXPORTER", "none", true)
	verifyConfigFile(t, container, configPath, "OTEL_LOGS_EXPORTER", "none", true)
	verifyConfigFile(t, container, configPath, "OTEL_EXPORTER_OTLP_PROTOCOL", regexp.QuoteMeta("http/protobuf"), true)
}

func bpffsMountedOnHost() bool {
	content, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == "/sys/fs/bpf" && parts[2] == "bpf" {
			return true
		}
	}
	return false
}

func copyInstaller(t *testing.T, container *testutils.Container) {
	t.Helper()
	require.NoError(t, container.CopyFileToContainer(context.Background(), filepath.Join(repoRoot(t), "packaging", "installer", "install.sh"), "/test/install.sh", 0o755))
}

func copyLocalInstrumentationPackage(t *testing.T, container *testutils.Container) {
	t.Helper()
	pkg := os.Getenv("LOCAL_INSTRUMENTATION_PACKAGE")
	if pkg == "" {
		return
	}
	require.NoError(t, container.CopyFileToContainer(context.Background(), pkg, "/test/instrumentation.pkg", 0o644))
}

func copyLocalCollectorPackageForPlatformTest(t *testing.T, container *testutils.Container, kind distroKind) {
	t.Helper()
	pkg := os.Getenv("LOCAL_COLLECTOR_PACKAGE")
	if pkg == "" {
		return
	}
	require.NoError(t, container.CopyFileToContainer(context.Background(), pkg, "/test/collector.pkg", 0o644))
	if kind == debDistro {
		runContainerCmd(t, container, "apt-get install -y libcap2-bin", runOptions{})
	}
}

func journalctl(t *testing.T, container *testutils.Container) {
	t.Helper()
	runContainerCmd(t, container, "journalctl -u "+serviceName+" --no-pager", runOptions{allowAnyExitCode: true})
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func shellDuration(duration time.Duration) string {
	if duration%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(duration/time.Minute))
	}
	return fmt.Sprintf("%ds", int(duration/time.Second))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func ptr(value string) *string {
	return &value
}
