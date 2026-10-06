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

type Container = testutils.Container

func Ptr(s string) *string { return &s }
func Env(k, fallback string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return fallback
}

func Root(t *testing.T) string {
	t.Helper()
	wd, e := os.Getwd()
	require.NoError(t, e)
	for {
		if _, e = os.Stat(filepath.Join(wd, "packaging/tests/images")); e == nil {
			return wd
		}
		next := filepath.Dir(wd)
		if next == wd {
			t.Fatal("repository root not found")
		}
		wd = next
	}
}
func Quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
func Run(t *testing.T, c *Container, cmd string) string {
	t.Helper()
	rc, out, err := c.AssertExec(t, 10*time.Minute, "sh", "-c", cmd)
	require.Equalf(t, 0, rc, "%s\n%s\n%s", cmd, out, err)
	return out
}

func Try(t *testing.T, c *Container, cmd string) (int, string) {
	t.Helper()
	rc, out, err := c.AssertExec(t, 10*time.Minute, "sh", "-c", cmd)
	return rc, out + err
}

func Copy(t *testing.T, c *Container, src, dst string) {
	t.Helper()
	Run(t, c, "mkdir -p "+Quote(filepath.Dir(dst)))
	require.NoError(t, c.CopyFileToContainer(context.Background(), src, dst, 0o644))
	File(t, c, dst, true)
}

func CopyText(t *testing.T, c *Container, content, dst string) {
	t.Helper()
	Run(t, c, "mkdir -p "+Quote(filepath.Dir(dst)))
	require.NoError(t, c.CopyToContainer(context.Background(), []byte(content), dst, 0o644))
	File(t, c, dst, true)
}

func File(t *testing.T, c *Container, path string, want bool) {
	t.Helper()
	rc, _ := Try(t, c, "test -f "+Quote(path))
	require.Equalf(t, want, rc == 0, "file %s existence", path)
}

func Config(t *testing.T, c *Container, path, pattern string, want bool) {
	t.Helper()
	if !want {
		rc, _ := Try(t, c, "test -f "+Quote(path))
		if rc != 0 {
			return
		}
	} else {
		File(t, c, path, true)
	}
	out := Run(t, c, "cat "+Quote(path))
	if path == SystemdConfig {
		pattern = "DefaultEnvironment=\"" + pattern + "\""
	}
	matched, e := regexp.MatchString("(?m)^"+pattern+"$", out)
	require.NoError(t, e)
	require.Equalf(t, want, matched, "pattern %q in %s:\n%s", pattern, path, out)
}

func KV(t *testing.T, c *Container, path, key, value string, want bool) {
	t.Helper()
	Config(t, c, path, key+"="+value, want)
}

func EnvFile(t *testing.T, c *Container, api, ingest, hec string) {
	t.Helper()
	for k, v := range map[string]string{"SPLUNK_ACCESS_TOKEN": Token, "SPLUNK_REALM": Realm, "SPLUNK_API_URL": api, "SPLUNK_INGEST_URL": ingest, "SPLUNK_HEC_URL": ingest + "/v1/log", "SPLUNK_HEC_TOKEN": hec} {
		KV(t, c, EnvPath, k, v, true)
	}
}

func PackageVersion(t *testing.T, c *Container, pkg, version string) {
	t.Helper()
	rc, _ := Try(t, c, "command -v dpkg-query")
	var got string
	if rc == 0 {
		got = Run(t, c, "dpkg-query --showformat='${Version}' --show "+pkg)
	} else {
		got = Run(t, c, "rpm -q --queryformat '%{VERSION}' "+pkg)
	}
	got = strings.TrimSpace(got)
	require.NotEmpty(t, got)
	if version != "latest" {
		require.Equal(t, version, got)
	}
}

func ServiceRunning(t *testing.T, c *Container, owner, process string) {
	t.Helper()
	require.Eventually(t, func() bool {
		a, _ := Try(t, c, "systemctl status "+ServiceName)
		b, _ := Try(t, c, "pgrep -a -u "+owner+" -f "+Quote(process))
		return a == 0 && b == 0
	}, 30*time.Second, time.Second)
}

func NodeInstalled(t *testing.T, c *Container) {
	t.Helper()
	rc, _ := Try(t, c, "cd "+NodePrefix+" && npm ls --global=false @splunk/otel")
	require.Zero(t, rc)
}

func VersionAtLeast(version, minimum string) bool {
	if version == "latest" {
		return true
	}
	var a, b, c, x, y, z int
	_, e := fmt.Sscanf(version, "%d.%d.%d", &a, &b, &c)
	if e != nil {
		return false
	}
	_, e = fmt.Sscanf(minimum, "%d.%d.%d", &x, &y, &z)
	if e != nil {
		return false
	}
	return a > x || a == x && (b > y || b == y && c >= z)
}

func UsesInjector(version string) bool {
	return version == "latest" || VersionAtLeast(version, "0.159.1") || (strings.HasPrefix(version, "0.159.0") && len(version) > len("0.159.0"))
}

func Distros(t *testing.T, subdir string) []string {
	t.Helper()
	paths, e := filepath.Glob(filepath.Join(Root(t), subdir, "Dockerfile.*"))
	require.NoError(t, e)
	out := []string{}
	for _, p := range paths {
		out = append(out, strings.TrimPrefix(filepath.Base(p), "Dockerfile."))
	}
	slices.Sort(out)
	return out
}

func Start(t *testing.T, dockerfile string, args map[string]*string, binds ...string) *Container {
	t.Helper()
	root := Root(t)
	args["TARGETARCH"] = Ptr("amd64")
	p, e := platforms.Parse("linux/amd64")
	require.NoError(t, e)
	b := testutils.NewContainer().WithContext(root).WithDockerfile(dockerfile).WithBuildArgs(args).WithDockerfileBuildOptionsModifier(func(o *dockerClient.ImageBuildOptions) { o.PullParent = true; o.Platforms = append(o.Platforms, p) }).WithImagePlatform("linux/amd64").WithPrivileged(true).WithBinds(append([]string{"/sys/fs/cgroup:/sys/fs/cgroup:rw"}, binds...)...).WithHostConfigModifier(func(h *dockerContainer.HostConfig) { h.CgroupnsMode = dockerContainer.CgroupnsModeHost }).WithStartupTimeout(5 * time.Minute)
	b.WaitingFor = append(b.WaitingFor, wait.ForNop(func(context.Context, wait.StrategyTarget) error { return nil }).WithStartupTimeout(5*time.Minute))
	c := b.Build()
	require.NoError(t, c.Start(context.Background()))
	t.Cleanup(func() {
		if t.Failed() {
			_, out, _ := c.AssertExec(t, time.Minute, "sh", "-c", "journalctl -u "+ServiceName+" --no-pager")
			t.Log(out)
		}
		require.NoError(t, c.Terminate(context.Background()))
	})
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rc, _, e := c.Exec(ctx, []string{"sh", "-c", "systemctl show-environment"})
		return e == nil && rc == 0
	}, 30*time.Second, time.Second)
	return c
}

type Instrumentation struct {
	Systemd, Custom     bool
	Version, Attributes string
}

func CheckInstrumentation(t *testing.T, c *Container, o Instrumentation) {
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
		NodeInstalled(t, c)
		checkInjector(t, c, o, attrs)
		if o.Custom {
			Config(t, c, "/etc/ld.so.preload", "# my extra library", true)
		}
		return
	}
	if o.Systemd {
		Config(t, c, "/etc/ld.so.preload", LibSplunk, false)
	} else {
		Config(t, c, "/etc/ld.so.preload", LibSplunk, true)
		File(t, c, SystemdConfig, false)
	}
	if o.Custom {
		Config(t, c, "/etc/ld.so.preload", "# my extra library", true)
	}
	if newStyle {
		NodeInstalled(t, c)
	}
	if o.Systemd {
		for _, p := range []string{JavaConfig, NodeConfig, DotnetConfig, InstrumentationConfig} {
			File(t, c, p, false)
		}
		checkLegacyEnv(t, c, SystemdConfig, o, attrs)
		KV(t, c, SystemdConfig, "NODE_OPTIONS", NodeOptions, newStyle)
		checkDotnet(t, c, SystemdConfig, newStyle && dotnet)
		return
	}
	if newStyle {
		File(t, c, SystemdConfig, false)
		File(t, c, InstrumentationConfig, false)
		KV(t, c, JavaConfig, "JAVA_TOOL_OPTIONS", "-javaagent:"+JavaAgent, true)
		KV(t, c, NodeConfig, "NODE_OPTIONS", NodeOptions, true)
		if dotnet {
			checkDotnet(t, c, DotnetConfig, true)
		} else {
			File(t, c, DotnetConfig, false)
		}
		paths := []string{JavaConfig, NodeConfig}
		if dotnet {
			paths = append(paths, DotnetConfig)
		}
		for _, p := range paths {
			checkLegacyEnv(t, c, p, o, attrs)
		}
		return
	}
	for _, p := range []string{JavaConfig, NodeConfig, DotnetConfig, SystemdConfig} {
		File(t, c, p, false)
	}
	KV(t, c, InstrumentationConfig, "java_agent_jar", JavaAgent, true)
	KV(t, c, InstrumentationConfig, "resource_attributes", attrs, true)
	settings := map[string]string{"service_name": "test", "generate_service_name": "false", "disable_telemetry": "true", "enable_profiler": "true", "enable_profiler_memory": "true", "enable_metrics": "true"}
	if !o.Custom {
		settings = map[string]string{"generate_service_name": "true", "disable_telemetry": "false", "enable_profiler": "false", "enable_profiler_memory": "false", "enable_metrics": "false"}
		KV(t, c, InstrumentationConfig, "service_name", ".*", false)
	}
	for k, v := range settings {
		KV(t, c, InstrumentationConfig, k, v, true)
	}
}

func checkLegacyEnv(t *testing.T, c *Container, path string, o Instrumentation, attrs string) {
	t.Helper()
	if path == SystemdConfig || path == JavaConfig {
		KV(t, c, path, "JAVA_TOOL_OPTIONS", "-javaagent:"+JavaAgent, true)
	}
	KV(t, c, path, "OTEL_RESOURCE_ATTRIBUTES", attrs, true)
	settings := map[string]string{"OTEL_SERVICE_NAME": "test", "SPLUNK_PROFILER_ENABLED": "true", "SPLUNK_PROFILER_MEMORY_ENABLED": "true", "SPLUNK_METRICS_ENABLED": "true", "OTEL_EXPORTER_OTLP_ENDPOINT": "http://0.0.0.0:4317", "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc", "OTEL_METRICS_EXPORTER": "none", "OTEL_LOGS_EXPORTER": "none"}
	if !o.Custom {
		for _, k := range []string{"OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_METRICS_EXPORTER", "OTEL_LOGS_EXPORTER"} {
			KV(t, c, path, k, ".*", false)
		}
		for _, k := range []string{"SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED"} {
			KV(t, c, path, k, "false", true)
		}
		return
	}
	for k, v := range settings {
		KV(t, c, path, k, v, true)
	}
}

func checkDotnet(t *testing.T, c *Container, path string, want bool) {
	t.Helper()
	for k, v := range DotnetVars {
		if !want {
			v = ".*"
		}
		KV(t, c, path, k, v, want)
	}
}

func checkInjector(t *testing.T, c *Container, o Instrumentation, attrs string) {
	t.Helper()
	for _, p := range []string{JavaConfig, NodeConfig, DotnetConfig, InstrumentationConfig} {
		File(t, c, p, false)
	}
	for k, v := range map[string]string{"jvm_auto_instrumentation_agent_path": JavaAgent, "nodejs_auto_instrumentation_agent_path": NodePrefix + "/node_modules/@splunk/otel/instrument.js", "dotnet_auto_instrumentation_agent_path_prefix": DotnetHome} {
		KV(t, c, InjectorConfig, k, v, true)
	}
	KV(t, c, InjectorConfig, "auto_instrumentation_disabled", ".*", false)
	if o.Systemd {
		Config(t, c, "/etc/ld.so.preload", LibInject, false)
		KV(t, c, SystemdConfig, "LD_PRELOAD", LibInject, true)
	} else {
		Config(t, c, "/etc/ld.so.preload", LibInject, true)
		File(t, c, SystemdConfig, false)
	}
	KV(t, c, InjectorDefault, "OTEL_DOTNET_AUTO_PLUGINS", DotnetVars["OTEL_DOTNET_AUTO_PLUGINS"], true)
	KV(t, c, InjectorDefault, "OTEL_RESOURCE_ATTRIBUTES", attrs, true)
	flags := "false"
	if o.Custom {
		flags = "true"
		KV(t, c, InjectorDefault, "OTEL_SERVICE_NAME", "test", true)
	} else {
		KV(t, c, InjectorDefault, "OTEL_SERVICE_NAME", ".*", false)
	}
	for _, k := range []string{"SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED"} {
		KV(t, c, InjectorDefault, k, flags, true)
	}
	for k, v := range map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://0.0.0.0:4317", "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc", "OTEL_METRICS_EXPORTER": "none", "OTEL_LOGS_EXPORTER": "none"} {
		if o.Custom {
			KV(t, c, InjectorDefault, k, v, true)
		} else {
			KV(t, c, InjectorDefault, k, ".*", false)
		}
	}
}
