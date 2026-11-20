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

// Package deploymenttest provides shared helpers for the puppet and salt
// deployment integration tests: building/starting the distro containers and
// asserting on the files, env vars and auto-instrumentation config they end
// up with.
package deploymenttest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	ConfigDir             = "/etc/otel/collector"
	EnvPath               = ConfigDir + "/splunk-otel-collector.conf"
	ServiceName           = "splunk-otel-collector"
	ServiceOwner          = ServiceName
	Token                 = "testing123"
	Realm                 = "test"
	APIURL                = "https://api.test.observability.splunkcloud.com"
	IngestURL             = "https://ingest.test.observability.splunkcloud.com"
	LibSplunk             = "/usr/lib/splunk-instrumentation/libsplunk.so"
	LibInject             = "/usr/lib/splunk-instrumentation/libotelinject.so"
	JavaAgent             = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
	InstrumentationConfig = "/usr/lib/splunk-instrumentation/instrumentation.conf"
	SystemdConfig         = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
	InjectorConfig        = "/etc/opentelemetry/injector/injector.conf"
	InjectorDefault       = "/etc/opentelemetry/injector/default_env.conf"
	JavaConfig            = "/etc/splunk/zeroconfig/java.conf"
	NodeConfig            = "/etc/splunk/zeroconfig/node.conf"
	DotnetConfig          = "/etc/splunk/zeroconfig/dotnet.conf"
	NodePrefix            = "/usr/lib/splunk-instrumentation/splunk-otel-js"
	NodeOptions           = "-r " + NodePrefix + "/node_modules/@splunk/otel/instrument"
	DotnetHome            = "/usr/lib/splunk-instrumentation/splunk-otel-dotnet"
)

// DotnetVars are the auto-instrumentation env vars that only apply once the
// .NET agent is bundled with the auto-instrumentation package (>= 0.99.0).
var DotnetVars = map[string]string{
	"CORECLR_ENABLE_PROFILING": "1",
	"CORECLR_PROFILER":         "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
	"CORECLR_PROFILER_PATH":    DotnetHome + "/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so",
	"DOTNET_ADDITIONAL_DEPS":   DotnetHome + "/AdditionalDeps",
	"DOTNET_SHARED_STORE":      DotnetHome + "/store",
	"DOTNET_STARTUP_HOOKS":     DotnetHome + "/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll",
	"OTEL_DOTNET_AUTO_HOME":    DotnetHome,
	"OTEL_DOTNET_AUTO_PLUGINS": "Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation",
}

// Container is the testcontainers wrapper used to build and run the distro
// images under test.
type Container = testutils.Container

// Ptr returns a pointer to s, for the *string-valued docker build args.
func Ptr(s string) *string { return &s }

// Env returns the value of the given environment variable, or fallback if it
// is unset or empty.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// FindRoot walks up from the current working directory until it finds a
// directory containing marker, and returns that directory. It is used to
// locate the repository root regardless of which package's test binary is
// running.
func FindRoot(t *testing.T, marker string) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err = os.Stat(filepath.Join(wd, marker)); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("repository root not found")
		}
		wd = parent
	}
}

// Root returns the repository root, located via the packaging/tests/images
// directory that only exists there.
func Root(t *testing.T) string {
	t.Helper()
	return FindRoot(t, "packaging/tests/images")
}

// Quote single-quotes s for safe interpolation into a shell command.
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// Run executes cmd in container via "sh -c" and requires that it succeeds,
// returning its stdout.
func Run(t *testing.T, container *Container, cmd string) string {
	t.Helper()
	rc, out, errOut := container.AssertExec(t, 10*time.Minute, "sh", "-c", cmd)
	require.Equalf(t, 0, rc, "%s\n%s\n%s", cmd, out, errOut)
	return out
}

// Exec executes cmd in container via "sh -c" and returns its exit code and
// combined stdout/stderr, without asserting on the result.
func Exec(t *testing.T, container *Container, cmd string) (int, string) {
	t.Helper()
	rc, out, errOut := container.AssertExec(t, 10*time.Minute, "sh", "-c", cmd)
	return rc, out + errOut
}

// Copy copies the local file at src into the container at dst, creating any
// missing parent directories first.
func Copy(t *testing.T, container *Container, src, dst string) {
	t.Helper()
	Run(t, container, "mkdir -p "+Quote(filepath.Dir(dst)))
	require.NoError(t, container.CopyFileToContainer(context.Background(), src, dst, 0o644))
	File(t, container, dst, true)
}

// CopyText writes content into the container at dst, creating any missing
// parent directories first.
func CopyText(t *testing.T, container *Container, content, dst string) {
	t.Helper()
	Run(t, container, "mkdir -p "+Quote(filepath.Dir(dst)))
	require.NoError(t, container.CopyToContainer(context.Background(), []byte(content), dst, 0o644))
	File(t, container, dst, true)
}

// File asserts that path exists in the container iff want is true.
func File(t *testing.T, container *Container, path string, want bool) {
	t.Helper()
	rc, _ := Exec(t, container, "test -f "+Quote(path))
	require.Equalf(t, want, rc == 0, "file %s existence", path)
}

// Config asserts that path exists and contains a line matching pattern iff
// want is true; if want is false and the file is missing entirely, that also
// counts as a pass.
func Config(t *testing.T, container *Container, path, pattern string, want bool) {
	t.Helper()
	if !want {
		rc, _ := Exec(t, container, "test -f "+Quote(path))
		if rc != 0 {
			return
		}
	} else {
		File(t, container, path, true)
	}
	out := Run(t, container, "cat "+Quote(path))
	if path == SystemdConfig {
		// systemd env files use DefaultEnvironment="KEY=VALUE" rather than
		// raw KEY=VALUE lines.
		pattern = `DefaultEnvironment="` + pattern + `"`
	}
	matched, err := regexp.MatchString("(?m)^"+pattern+"$", out)
	require.NoError(t, err)
	require.Equalf(t, want, matched, "pattern %q in %s:\n%s", pattern, path, out)
}

// KV asserts that path contains (or lacks) a "key=value" line.
func KV(t *testing.T, container *Container, path, key, value string, want bool) {
	t.Helper()
	Config(t, container, path, key+"="+value, want)
}

// EnvFile asserts that the collector's env file was rendered with the given
// access token, realm, API/ingest URLs and HEC token.
func EnvFile(t *testing.T, container *Container, api, ingest, hec string) {
	t.Helper()
	expected := map[string]string{
		"SPLUNK_ACCESS_TOKEN": Token,
		"SPLUNK_REALM":        Realm,
		"SPLUNK_API_URL":      api,
		"SPLUNK_INGEST_URL":   ingest,
		"SPLUNK_HEC_URL":      ingest + "/v1/log",
		"SPLUNK_HEC_TOKEN":    hec,
	}
	for key, value := range expected {
		KV(t, container, EnvPath, key, value, true)
	}
}

// PackageVersion asserts that pkg is installed, and that its version matches
// version (unless version is "latest", in which case any installed version
// passes).
func PackageVersion(t *testing.T, container *Container, pkg, version string) {
	t.Helper()
	rc, _ := Exec(t, container, "command -v dpkg-query")
	var got string
	if rc == 0 {
		got = Run(t, container, "dpkg-query --showformat='${Version}' --show "+pkg)
	} else {
		got = Run(t, container, "rpm -q --queryformat '%{VERSION}' "+pkg)
	}
	got = strings.TrimSpace(got)
	require.NotEmpty(t, got)
	if version != "latest" {
		require.Equal(t, version, got)
	}
}

// ServiceRunning waits for the systemd service and its process (running as
// owner) to both be up.
func ServiceRunning(t *testing.T, container *Container, owner, process string) {
	t.Helper()
	require.Eventually(t, func() bool {
		systemctlRC, _ := Exec(t, container, "systemctl status "+ServiceName)
		pgrepRC, _ := Exec(t, container, "pgrep -a -u "+owner+" -f "+Quote(process))
		return systemctlRC == 0 && pgrepRC == 0
	}, 30*time.Second, time.Second)
}

// NodeInstalled asserts that the @splunk/otel node module was installed
// under NodePrefix.
func NodeInstalled(t *testing.T, container *Container) {
	t.Helper()
	rc, _ := Exec(t, container, "cd "+NodePrefix+" && npm ls --global=false @splunk/otel")
	require.Zero(t, rc)
}

// VersionAtLeast reports whether version >= minimum, both given as
// "major.minor.patch" (optionally with a trailing suffix on version).
// "latest" is always considered at least any minimum.
func VersionAtLeast(version, minimum string) bool {
	if version == "latest" {
		return true
	}
	var major, minor, patch, minMajor, minMinor, minPatch int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(minimum, "%d.%d.%d", &minMajor, &minMinor, &minPatch); err != nil {
		return false
	}
	if major != minMajor {
		return major > minMajor
	}
	if minor != minMinor {
		return minor > minMinor
	}
	return patch >= minPatch
}

// UsesInjector reports whether version's auto-instrumentation package uses
// the shared otel injector (libotelinject.so) rather than the legacy
// libsplunk.so, which it does starting after 0.159.0 (including any
// pre-release build of 0.159.0 itself, e.g. "0.159.0-1").
func UsesInjector(version string) bool {
	if version == "latest" || VersionAtLeast(version, "0.159.1") {
		return true
	}
	return strings.HasPrefix(version, "0.159.0") && version != "0.159.0"
}

// Distros returns the sorted distro names found as Dockerfile.<name> under
// subdir.
func Distros(t *testing.T, subdir string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(Root(t), subdir, "Dockerfile.*"))
	require.NoError(t, err)
	distros := []string{}
	for _, path := range paths {
		distros = append(distros, strings.TrimPrefix(filepath.Base(path), "Dockerfile."))
	}
	slices.Sort(distros)
	return distros
}

// Start builds dockerfile (relative to the repository root) with args and
// binds, starts the resulting container privileged with cgroups mounted so
// systemd can run inside it, and registers cleanup to terminate it (logging
// the service's journal first if the test failed).
func Start(t *testing.T, dockerfile string, args map[string]*string, binds ...string) *Container {
	t.Helper()
	root := Root(t)
	args["TARGETARCH"] = Ptr("amd64")

	platform, err := platforms.Parse("linux/amd64")
	require.NoError(t, err)

	allBinds := append([]string{"/sys/fs/cgroup:/sys/fs/cgroup:rw"}, binds...)
	builder := testutils.NewContainer().
		WithContext(root).
		WithDockerfile(dockerfile).
		WithBuildArgs(args).
		WithDockerfileBuildOptionsModifier(func(o *dockerClient.ImageBuildOptions) {
			o.PullParent = true
			o.Platforms = append(o.Platforms, platform)
		}).
		WithImagePlatform("linux/amd64").
		WithPrivileged(true).
		WithBinds(allBinds...).
		WithHostConfigModifier(func(h *dockerContainer.HostConfig) {
			h.CgroupnsMode = dockerContainer.CgroupnsModeHost
		}).
		WithStartupTimeout(5 * time.Minute)
	builder.WaitingFor = append(builder.WaitingFor, wait.ForNop(func(context.Context, wait.StrategyTarget) error {
		return nil
	}).WithStartupTimeout(5*time.Minute))

	container := builder.Build()
	require.NoError(t, container.Start(context.Background()))
	t.Cleanup(func() {
		if t.Failed() {
			_, out, _ := container.AssertExec(t, time.Minute, "sh", "-c", "journalctl -u "+ServiceName+" --no-pager")
			t.Log(out)
		}
		require.NoError(t, container.Terminate(context.Background()))
	})

	// qemu is slow, so wait for systemd to be ready before returning.
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rc, _, err := container.Exec(ctx, []string{"sh", "-c", "systemctl show-environment"})
		return err == nil && rc == 0
	}, 30*time.Second, time.Second)
	return container
}

// Instrumentation describes the auto-instrumentation deployment under test,
// used by CheckInstrumentation to pick the expected config layout.
type Instrumentation struct {
	Version    string
	Attributes string
	Systemd    bool
	Custom     bool
}

// CheckInstrumentation asserts that the container's auto-instrumentation
// config matches what the installed package version and o are expected to
// produce. Package behavior falls into three generations, oldest first:
//   - pre-0.87.0: a single instrumentation.conf (libsplunk.so config format).
//   - 0.87.0 up to the injector cutover: per-language zeroconfig files
//     (java.conf, node.conf, dotnet.conf once >= 0.99.0), still loaded via
//     libsplunk.so.
//   - past the injector cutover (see UsesInjector): injector.conf plus
//     default_env.conf, loaded via libotelinject.so.
func CheckInstrumentation(t *testing.T, container *Container, o Instrumentation) {
	t.Helper()
	newStyle := VersionAtLeast(o.Version, "0.87.0")
	dotnet := VersionAtLeast(o.Version, "0.99.0")
	attrs := o.Attributes
	if o.Systemd {
		attrs += "-systemd"
	}
	if o.Custom {
		attrs += ",deployment.environment.name=test"
	}

	if UsesInjector(o.Version) {
		NodeInstalled(t, container)
		checkInjector(t, container, o, attrs)
		if o.Custom {
			Config(t, container, "/etc/ld.so.preload", "# my extra library", true)
		}
		return
	}

	if o.Systemd {
		Config(t, container, "/etc/ld.so.preload", LibSplunk, false)
	} else {
		Config(t, container, "/etc/ld.so.preload", LibSplunk, true)
		File(t, container, SystemdConfig, false)
	}
	if o.Custom {
		Config(t, container, "/etc/ld.so.preload", "# my extra library", true)
	}
	if newStyle {
		NodeInstalled(t, container)
	}

	if o.Systemd {
		for _, path := range []string{JavaConfig, NodeConfig, DotnetConfig, InstrumentationConfig} {
			File(t, container, path, false)
		}
		checkLegacyEnv(t, container, SystemdConfig, o, attrs)
		KV(t, container, SystemdConfig, "NODE_OPTIONS", NodeOptions, newStyle)
		checkDotnet(t, container, SystemdConfig, newStyle && dotnet)
		return
	}

	if newStyle {
		File(t, container, SystemdConfig, false)
		File(t, container, InstrumentationConfig, false)
		KV(t, container, JavaConfig, "JAVA_TOOL_OPTIONS", "-javaagent:"+JavaAgent, true)
		KV(t, container, NodeConfig, "NODE_OPTIONS", NodeOptions, true)
		if dotnet {
			checkDotnet(t, container, DotnetConfig, true)
		} else {
			File(t, container, DotnetConfig, false)
		}
		paths := []string{JavaConfig, NodeConfig}
		if dotnet {
			paths = append(paths, DotnetConfig)
		}
		for _, path := range paths {
			checkLegacyEnv(t, container, path, o, attrs)
		}
		return
	}

	for _, path := range []string{JavaConfig, NodeConfig, DotnetConfig, SystemdConfig} {
		File(t, container, path, false)
	}
	KV(t, container, InstrumentationConfig, "java_agent_jar", JavaAgent, true)
	KV(t, container, InstrumentationConfig, "resource_attributes", attrs, true)
	settings := map[string]string{
		"service_name":           "test",
		"generate_service_name":  "false",
		"disable_telemetry":      "true",
		"enable_profiler":        "true",
		"enable_profiler_memory": "true",
		"enable_metrics":         "true",
	}
	if !o.Custom {
		settings = map[string]string{
			"generate_service_name":  "true",
			"disable_telemetry":      "false",
			"enable_profiler":        "false",
			"enable_profiler_memory": "false",
			"enable_metrics":         "false",
		}
		KV(t, container, InstrumentationConfig, "service_name", ".*", false)
	}
	for key, value := range settings {
		KV(t, container, InstrumentationConfig, key, value, true)
	}
}

// checkLegacyEnv asserts the OTEL_* / SPLUNK_* env vars written into one of
// the pre-injector config files (SystemdConfig, JavaConfig, NodeConfig or
// DotnetConfig).
func checkLegacyEnv(t *testing.T, container *Container, path string, o Instrumentation, attrs string) {
	t.Helper()
	if path == SystemdConfig || path == JavaConfig {
		KV(t, container, path, "JAVA_TOOL_OPTIONS", "-javaagent:"+JavaAgent, true)
	}
	KV(t, container, path, "OTEL_RESOURCE_ATTRIBUTES", attrs, true)

	if !o.Custom {
		for _, key := range []string{"OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_METRICS_EXPORTER", "OTEL_LOGS_EXPORTER"} {
			KV(t, container, path, key, ".*", false)
		}
		for _, key := range []string{"SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED"} {
			KV(t, container, path, key, "false", true)
		}
		return
	}

	settings := map[string]string{
		"OTEL_SERVICE_NAME":              "test",
		"SPLUNK_PROFILER_ENABLED":        "true",
		"SPLUNK_PROFILER_MEMORY_ENABLED": "true",
		"SPLUNK_METRICS_ENABLED":         "true",
		"OTEL_EXPORTER_OTLP_ENDPOINT":    "http://0.0.0.0:4317",
		"OTEL_EXPORTER_OTLP_PROTOCOL":    "grpc",
		"OTEL_METRICS_EXPORTER":          "none",
		"OTEL_LOGS_EXPORTER":             "none",
	}
	for key, value := range settings {
		KV(t, container, path, key, value, true)
	}
}

// checkDotnet asserts the presence (or absence) of DotnetVars in path.
func checkDotnet(t *testing.T, container *Container, path string, want bool) {
	t.Helper()
	for key, value := range DotnetVars {
		if !want {
			value = ".*"
		}
		KV(t, container, path, key, value, want)
	}
}

// checkInjector asserts the injector.conf and default_env.conf layout used
// once the package is past UsesInjector's cutover.
func checkInjector(t *testing.T, container *Container, o Instrumentation, attrs string) {
	t.Helper()
	for _, path := range []string{JavaConfig, NodeConfig, DotnetConfig, InstrumentationConfig} {
		File(t, container, path, false)
	}

	agentPaths := map[string]string{
		"jvm_auto_instrumentation_agent_path":           JavaAgent,
		"nodejs_auto_instrumentation_agent_path":        NodePrefix + "/node_modules/@splunk/otel/instrument.js",
		"dotnet_auto_instrumentation_agent_path_prefix": DotnetHome,
	}
	for key, value := range agentPaths {
		KV(t, container, InjectorConfig, key, value, true)
	}
	KV(t, container, InjectorConfig, "auto_instrumentation_disabled", ".*", false)

	if o.Systemd {
		Config(t, container, "/etc/ld.so.preload", LibInject, false)
		KV(t, container, SystemdConfig, "LD_PRELOAD", LibInject, true)
	} else {
		Config(t, container, "/etc/ld.so.preload", LibInject, true)
		File(t, container, SystemdConfig, false)
	}

	KV(t, container, InjectorDefault, "OTEL_DOTNET_AUTO_PLUGINS", DotnetVars["OTEL_DOTNET_AUTO_PLUGINS"], true)
	KV(t, container, InjectorDefault, "OTEL_RESOURCE_ATTRIBUTES", attrs, true)

	flags := "false"
	if o.Custom {
		flags = "true"
		KV(t, container, InjectorDefault, "OTEL_SERVICE_NAME", "test", true)
	} else {
		KV(t, container, InjectorDefault, "OTEL_SERVICE_NAME", ".*", false)
	}
	for _, key := range []string{"SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED"} {
		KV(t, container, InjectorDefault, key, flags, true)
	}

	otelExporter := map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://0.0.0.0:4317",
		"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
		"OTEL_METRICS_EXPORTER":       "none",
		"OTEL_LOGS_EXPORTER":          "none",
	}
	for key, value := range otelExporter {
		KV(t, container, InjectorDefault, key, value, o.Custom)
	}
}
