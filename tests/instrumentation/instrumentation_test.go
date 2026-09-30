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

package instrumentation

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	packageName = "splunk-otel-auto-instrumentation"
	libDir      = "/usr/lib/splunk-instrumentation"

	libOtelInjectPath = libDir + "/libotelinject.so"
	preloadPath       = "/etc/ld.so.preload"

	javaAgentPath       = libDir + "/splunk-otel-javaagent.jar"
	nodeAgentPath       = libDir + "/splunk-otel-js.tgz"
	dotnetAgentPathBase = libDir + "/splunk-otel-dotnet/glibc/linux-"

	injectorConfigPath     = "/etc/opentelemetry/injector/injector.conf"
	injectorDefaultEnvPath = "/etc/opentelemetry/injector/default_env.conf"

	legacyVersion    = "0.159.0"
	legacyReleaseURL = "https://github.com/signalfx/splunk-otel-collector/releases/download/v" + legacyVersion
	libsplunkPath    = libDir + "/libsplunk.so"
	zeroconfigDir    = "/etc/splunk/zeroconfig"
	legacyConfigDir  = libDir + "/legacy-zeroconfig"

	tomcatPIDFile  = "/usr/local/tomcat/temp/tomcat.pid"
	expressPIDFile = "/opt/express/express.pid"
	dotnetPIDFile  = "/opt/dotnet/dotnet.pid"

	collectorLogFile = "/test/otelcol.log"
	collectorPIDFile = "/test/otelcol.pid"

	preservedPreloadLine = "# This line should be preserved"
	upgradeWarning       = "WARNING: Upgrading " + packageName + " from a version using libsplunk.so. " +
		"Auto-instrumentation is switching from libsplunk.so to libotelinject.so, " +
		"and configuration files have moved from /etc/splunk/zeroconfig/ to " +
		"/etc/opentelemetry/injector/. See the release notes for details."
)

var (
	installedFiles = []string{
		javaAgentPath,
		nodeAgentPath,
		libOtelInjectPath,
		injectorConfigPath,
		injectorDefaultEnvPath,
	}
	tomcatEnv = map[string]string{
		"JAVA_HOME":     "/opt/java/openjdk",
		"CATALINA_PID":  tomcatPIDFile,
		"CATALINA_HOME": "/usr/local/tomcat",
		"CATALINA_BASE": "/usr/local/tomcat",
		"CATALINA_OPTS": "-Xms512M -Xmx1024M -server -XX:+UseParallelGC",
		"JAVA_OPTS":     "-Djava.awt.headless=true",
	}
	dotnetEnv = map[string]string{
		"OTEL_INJECTOR_LOG_LEVEL": "debug",
		"OTEL_LOG_LEVEL":          "debug",
	}
)

func TestTomcatInstrumentation(t *testing.T) {
	for _, distro := range selectedDistros(t) {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				container, otelcol := prepareInstrumentationContainer(t, distro, arch, "")
				installInstrumentationPackage(t, container, distro, arch)

				require.NoError(t, appendPreload(t, container))
				verifyAppInstrumentation(t, container, otelcol, "tomcat", map[string]string{
					`telemetry\.sdk\.language`: `Str\(java\)`,
					`service\.name`:            `Str\(Hello, World Application\)`,
				}, nil)

				require.NoError(t, copyFile(container, fixturePath(t, "test-java-env.conf"), injectorDefaultEnvPath))
				attributes := map[string]string{
					`telemetry\.sdk\.language`:      `Str\(java\)`,
					`service\.name`:                 `Str\(service_name_from_java\)`,
					`deployment\.environment\.name`: `Str\(deployment_environment_from_java\)`,
					`com\.splunk\.sourcetype`:       `Str\(otel\.profiling\)`,
				}
				verifyAppInstrumentation(t, container, otelcol, "tomcat", attributes, nil)
				verifyAppInstrumentation(t, container, otelcol, "tomcat", attributes, map[string]string{
					"OTEL_SERVICE_NAME": "service_name_from_app",
				})
			})
		}
	}
}

func TestExpressInstrumentation(t *testing.T) {
	for _, distro := range selectedDistros(t) {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				container, otelcol := prepareInstrumentationContainer(t, distro, arch, "NODE_VERSION=v18")
				installInstrumentationPackage(t, container, distro, arch)

				assertExec(t, container, 5*time.Minute, "mkdir -p "+libDir+"/splunk-otel-js")
				assertExec(t, container, 10*time.Minute, "bash -l -c 'cd "+libDir+"/splunk-otel-js && npm install "+nodeAgentPath+"'")
				require.NoError(t, appendPreload(t, container))

				verifyAppInstrumentation(t, container, otelcol, "express", map[string]string{
					`telemetry\.sdk\.language`: `Str\(nodejs\)`,
					`service\.name`:            `Str\(unnamed-node-service\)`,
				}, nil)

				require.NoError(t, copyFile(container, fixturePath(t, "test-node-env.conf"), injectorDefaultEnvPath))
				verifyAppInstrumentation(t, container, otelcol, "express", map[string]string{
					`telemetry\.sdk\.language`:      `Str\(nodejs\)`,
					`service\.name`:                 `Str\(service_name_from_node\)`,
					`deployment\.environment\.name`: `Str\(deployment_environment_from_node\)`,
					`com\.splunk\.sourcetype`:       `Str\(otel\.profiling\)`,
				}, nil)
			})
		}
	}
}

func TestDotnetInstrumentation(t *testing.T) {
	for _, distro := range selectedDistros(t) {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				container, otelcol := prepareInstrumentationContainer(t, distro, arch, "")
				installInstrumentationPackage(t, container, distro, arch)

				require.NoError(t, appendPreload(t, container))
				verifyAppInstrumentation(t, container, otelcol, "dotnet", map[string]string{
					`telemetry\.sdk\.language`: `Str\(dotnet\)`,
					`service\.name`:            `Str\(myWebApp\)`,
				}, nil)

				require.NoError(t, copyFile(container, fixturePath(t, "test-dotnet-env.conf"), injectorDefaultEnvPath))
				verifyAppInstrumentation(t, container, otelcol, "dotnet", map[string]string{
					`telemetry\.sdk\.language`:      `Str\(dotnet\)`,
					`service\.name`:                 `Str\(service_name_from_dotnet\)`,
					`deployment\.environment\.name`: `Str\(deployment_environment_from_dotnet\)`,
					`com\.splunk\.sourcetype`:       `Str\(otel\.profiling\)`,
				}, nil)
			})
		}
	}
}

func TestPackageUninstall(t *testing.T) {
	for _, distro := range selectedDistros(t) {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				container, _ := preparePackageContainer(t, distro, arch, "")
				assertExec(t, container, time.Minute, "echo "+shellQuote(preservedPreloadLine)+" >> "+preloadPath)
				installInstrumentationPackage(t, container, distro, arch)

				preuninstallPath := filepath.Join(repoRoot(t), "instrumentation", "packaging", "fpm", "preuninstall.sh")
				require.NoError(t, copyFile(container, preuninstallPath, "/test/preuninstall.sh"))
				assertExec(t, container, time.Minute, "mkdir -p "+legacyConfigDir)
				assertExec(t, container, time.Minute, "touch "+legacyConfigDir+"/java.conf")
				assertExec(t, container, time.Minute, "sh /test/preuninstall.sh --remove-legacy-config")
				assertFileAbsent(t, container, legacyConfigDir)

				assertExec(t, container, time.Minute, "mkdir -p "+legacyConfigDir)
				assertExec(t, container, time.Minute, "touch "+legacyConfigDir+"/java.conf")
				assertExec(t, container, time.Minute, "REMOVE_LEGACY_CONFIG=true sh /test/preuninstall.sh")
				assertFileAbsent(t, container, legacyConfigDir)

				assertExec(t, container, time.Minute, "mkdir -p "+legacyConfigDir)
				assertExec(t, container, time.Minute, "touch "+legacyConfigDir+"/java.conf")
				assertExec(t, container, time.Minute, "REMOVE_LEGACY_CONFIG=true sh /test/preuninstall.sh upgrade")
				assertFileExists(t, container, legacyConfigDir)
				assertExec(t, container, time.Minute, "sh /test/preuninstall.sh --remove-legacy-config")
				assertFileAbsent(t, container, legacyConfigDir)

				verifyPreload(t, container, preservedPreloadLine, true)
				verifyPreload(t, container, libOtelInjectPath, false)
				for _, action := range []string{"upgrade", "failed-upgrade", "1", "2"} {
					require.NoError(t, appendPreload(t, container))
					assertExec(t, container, time.Minute, "sh /test/preuninstall.sh "+action)
					verifyPreload(t, container, libOtelInjectPath, true)
					assertExec(t, container, time.Minute, "sed -i -e 's|"+libOtelInjectPath+"||' "+preloadPath)
				}
				for _, action := range []string{"remove", "0", ""} {
					require.NoError(t, appendPreload(t, container))
					assertExec(t, container, time.Minute, "sh /test/preuninstall.sh "+action)
					verifyPreload(t, container, libOtelInjectPath, false)
				}
				verifyPreload(t, container, preservedPreloadLine, true)

				require.NoError(t, appendPreload(t, container))
				uninstallPackage(t, container, distro)
				if distroIsDeb(t, distro) {
					assertExecNonZero(t, container, time.Minute, "dpkg -s "+packageName)
				} else {
					assertExecNonZero(t, container, time.Minute, "rpm -q "+packageName)
				}
				for _, path := range installedFiles {
					assertFileAbsent(t, container, path)
				}
				verifyPreload(t, container, libOtelInjectPath, false)
				verifyPreload(t, container, preservedPreloadLine, true)
			})
		}
	}
}

func TestPackageUpgradeFromLibsplunk(t *testing.T) {
	for _, distro := range selectedDistros(t) {
		for _, arch := range selectedArches(t) {
			t.Run(distro+"/"+arch, func(t *testing.T) {
				container, pkgPath := preparePackageContainer(t, distro, arch, "")
				legacyPkg := legacyPackageName(t, distro, arch)
				legacyURL := legacyReleaseURL + "/" + legacyPkg

				assertExec(t, container, time.Minute, "if command -v curl >/dev/null; then curl -sfL "+legacyURL+" -o /test/"+legacyPkg+"; else wget -q "+legacyURL+" -O /test/"+legacyPkg+"; fi")
				installLegacyPackage(t, container, distro, legacyPkg)

				expectedLegacyFiles := legacyFiles(arch)
				legacyConfigContents := map[string]string{}
				for _, path := range expectedLegacyFiles {
					assertFileExists(t, container, path)
					if path != libsplunkPath {
						legacyConfigContents[filepath.Base(path)] = assertExec(t, container, time.Minute, "cat "+path)
					}
				}

				assertExec(t, container, time.Minute, "echo "+shellQuote(preservedPreloadLine)+" >> "+preloadPath)
				assertExec(t, container, time.Minute, "echo "+libsplunkPath+" >> "+preloadPath)
				output := upgradeInstrumentationPackage(t, container, distro, "/test/"+filepath.Base(pkgPath))
				require.Contains(t, output, upgradeWarning)

				for _, path := range expectedLegacyFiles {
					assertFileAbsent(t, container, path)
				}
				expectedMigratedFiles := []string{legacyConfigDir + "/java.conf", legacyConfigDir + "/node.conf"}
				if arch == "amd64" {
					expectedMigratedFiles = append(expectedMigratedFiles, legacyConfigDir+"/dotnet.conf")
				}
				for _, path := range expectedMigratedFiles {
					assertFileExists(t, container, path)
					require.Equal(t, legacyConfigContents[filepath.Base(path)], assertExec(t, container, time.Minute, "cat "+path))
				}
				for _, path := range installedFiles {
					assertFileExists(t, container, path)
				}
				verifyPreload(t, container, libsplunkPath, false)
				verifyPreload(t, container, preservedPreloadLine, true)
				verifyPreload(t, container, libOtelInjectPath, false)
			})
		}
	}
}

func prepareInstrumentationContainer(t *testing.T, distro, arch, buildArg string) (*testutils.Container, string) {
	t.Helper()
	container, _ := preparePackageContainer(t, distro, arch, buildArg)
	binPath := filepath.Join(repoRoot(t), "bin", "otelcol_linux_"+arch)
	require.FileExists(t, binPath)
	require.NoError(t, copyFile(container, filepath.Join(repoRoot(t), "packaging", "tests", "instrumentation", "config.yaml"), "/test/config.yaml"))
	require.NoError(t, copyFile(container, binPath, "/test/otelcol_linux_"+arch))
	assertExec(t, container, time.Minute, "chmod a+x /test/otelcol_linux_"+arch)
	return container, "/test/otelcol_linux_" + arch
}

func preparePackageContainer(t *testing.T, distro, arch, buildArg string) (*testutils.Container, string) {
	t.Helper()
	pkgPath := requirePackage(t, distro, arch)
	root := repoRoot(t)
	packageType := packageType(t, distro)
	relDockerfile, err := filepath.Rel(filepath.Join(root, "packaging", "tests"), filepath.Join(root, "packaging", "tests", "instrumentation", "images", packageType, "Dockerfile."+distro))
	require.NoError(t, err)
	targetArch := arch
	platform := "linux/" + arch
	parsedPlatform, err := platforms.Parse(platform)
	require.NoError(t, err)
	builder := testutils.NewContainer().
		WithContext(filepath.Join(root, "packaging", "tests")).
		WithDockerfile(relDockerfile).
		WithBuildArgs(map[string]*string{"TARGETARCH": &targetArch}).
		WithDockerfileBuildOptionsModifier(func(options *dockerClient.ImageBuildOptions) {
			options.PullParent = true
			options.Platforms = append(options.Platforms, parsedPlatform)
		}).
		WithImagePlatform(platform).
		WithPrivileged(true).
		WithBinds("/sys/fs/cgroup:/sys/fs/cgroup:rw").
		WithHostConfigModifier(func(hostConfig *dockerContainer.HostConfig) {
			hostConfig.CgroupnsMode = dockerContainer.CgroupnsModeHost
		}).
		WithStartupTimeout(5 * time.Minute)
	if buildArg != "" {
		builder = builder.WithBuildArgs(map[string]*string{"TARGETARCH": &targetArch, "NODE_VERSION": stringPtr(strings.TrimPrefix(buildArg, "NODE_VERSION="))})
	}
	builder.WaitingFor = append(builder.WaitingFor, wait.ForNop(func(context.Context, wait.StrategyTarget) error { return nil }))
	container := builder.Build()
	require.NoError(t, container.Start(context.Background()))
	t.Cleanup(func() { require.NoError(t, container.Terminate(context.Background())) })
	waitForSystemd(t, container, arch)
	require.NoError(t, copyFile(container, pkgPath, "/test/"+filepath.Base(pkgPath)))
	return container, pkgPath
}

func verifyAppInstrumentation(t *testing.T, container *testutils.Container, otelcol, app string, attributes, appEnv map[string]string) {
	t.Helper()
	stopApp(t, container, app)
	stopCollector(t, container)
	startCollector(t, container, otelcol)
	startApp(t, container, app, appEnv)

	deadline := time.Now().Add(5 * time.Minute)
	found := make(map[string]bool, len(attributes))
	for key := range attributes {
		found[key] = false
	}
	for time.Now().Before(deadline) {
		output := commandOutput(t, container, "cat "+collectorLogFile)
		for key, value := range attributes {
			if !found[key] {
				found[key] = regexp.MustCompile(key + ": " + value).MatchString(output)
			}
		}
		allFound := true
		for _, present := range found {
			allFound = allFound && present
		}
		if allFound {
			return
		}
		time.Sleep(time.Second)
	}
	collectorOutput := commandOutput(t, container, "cat "+collectorLogFile)
	t.Log("=== collector output ===\n" + collectorOutput)
	switch app {
	case "tomcat":
		t.Log("=== Tomcat catalina.out ===\n" + commandOutput(t, container, "cat /usr/local/tomcat/logs/catalina.out"))
	case "dotnet":
		t.Log("=== dotnet auto-instrumentation logs ===\n" + commandOutput(t, container, "cat /var/log/opentelemetry/dotnet/*"))
	}
	for key, value := range attributes {
		require.True(t, found[key], "timed out waiting for %q: %s", key, value)
	}
}

func startCollector(t *testing.T, container *testutils.Container, otelcol string) {
	t.Helper()
	assertExec(t, container, time.Minute, "rm -f "+collectorLogFile+" "+collectorPIDFile+"; "+otelcol+" --config=/test/config.yaml >"+collectorLogFile+" 2>&1 & echo $! >"+collectorPIDFile)
}

func stopCollector(t *testing.T, container *testutils.Container) {
	t.Helper()
	execCode(t, container, time.Minute, "if test -f "+collectorPIDFile+"; then kill -TERM $(cat "+collectorPIDFile+") 2>/dev/null || true; rm -f "+collectorPIDFile+"; fi")
}

func startApp(t *testing.T, container *testutils.Container, app string, appEnv map[string]string) {
	t.Helper()
	env := mergeEnv(appEnv, map[string]string{})
	switch app {
	case "tomcat":
		env = mergeEnv(tomcatEnv, env)
		runCommand(t, container, time.Minute, "bash -c /usr/local/tomcat/bin/startup.sh", env, "tomcat:tomcat", "")
	case "express":
		runCommand(t, container, time.Minute, "bash -l -c 'node /opt/express/app.js & echo $! > "+expressPIDFile+"'", env, "express:express", "")
	case "dotnet":
		env = mergeEnv(dotnetEnv, env)
		runCommand(t, container, time.Minute, "bash -c '/opt/dotnet-sdk/dotnet /opt/dotnet/myWebApp.dll & echo $! > "+dotnetPIDFile+"'", env, "dotnet:dotnet", "/opt/dotnet")
	default:
		t.Fatalf("unsupported app %q", app)
	}
	port := map[string]string{"tomcat": "8080", "express": "3000", "dotnet": "5000"}[app]
	path := map[string]string{"tomcat": "/sample", "express": "", "dotnet": ""}[app]
	require.Eventually(t, func() bool {
		return execCode(t, container, time.Minute, "curl -sS -fL http://127.0.0.1:"+port+path) == 0
	}, 5*time.Minute, time.Second)
}

func stopApp(t *testing.T, container *testutils.Container, app string) {
	t.Helper()
	execCode(t, container, time.Minute, "systemctl stop "+app)
	pidFile := map[string]string{"tomcat": tomcatPIDFile, "express": expressPIDFile, "dotnet": dotnetPIDFile}[app]
	if pidFile == "" {
		return
	}
	if app == "tomcat" {
		runCommand(t, container, time.Minute, "bash -c /usr/local/tomcat/bin/shutdown.sh", tomcatEnv, "", "")
	} else {
		execCode(t, container, time.Minute, "if test -f "+pidFile+"; then kill -TERM $(cat "+pidFile+") 2>/dev/null || true; rm -f "+pidFile+"; fi")
	}
}

func installInstrumentationPackage(t *testing.T, container *testutils.Container, distro, arch string) {
	t.Helper()
	pkgPath := requirePackage(t, distro, arch)
	installPackage(t, container, distro, "/test/"+filepath.Base(pkgPath))
	for _, path := range installedFiles {
		assertFileExists(t, container, path)
	}
	dotnetArch := "x64"
	if arch == "arm64" {
		dotnetArch = "arm64"
	}
	assertFileExists(t, container, dotnetAgentPathBase+dotnetArch+"/OpenTelemetry.AutoInstrumentation.Native.so")
}

func installPackage(t *testing.T, container *testutils.Container, distro, path string) {
	t.Helper()
	if distroIsDeb(t, distro) {
		assertExec(t, container, 5*time.Minute, "dpkg -i "+path)
	} else {
		assertExec(t, container, 5*time.Minute, "rpm -ivh "+path)
	}
}

func installLegacyPackage(t *testing.T, container *testutils.Container, distro, packageFile string) {
	t.Helper()
	if distroIsDeb(t, distro) {
		assertExec(t, container, 5*time.Minute, "dpkg -i /test/"+packageFile)
	} else {
		assertExec(t, container, 5*time.Minute, "rpm -ivh /test/"+packageFile)
	}
}

func upgradeInstrumentationPackage(t *testing.T, container *testutils.Container, distro, path string) string {
	t.Helper()
	if distroIsDeb(t, distro) {
		return assertExec(t, container, 5*time.Minute, "dpkg -i "+path)
	}
	return assertExec(t, container, 5*time.Minute, "rpm -Uvh "+path)
}

func uninstallPackage(t *testing.T, container *testutils.Container, distro string) {
	t.Helper()
	if distroIsDeb(t, distro) {
		assertExec(t, container, 5*time.Minute, "dpkg -P "+packageName)
	} else {
		assertExec(t, container, 5*time.Minute, "rpm -e "+packageName)
	}
}

func waitForSystemd(t *testing.T, container *testutils.Container, arch string) {
	t.Helper()
	timeout := 10 * time.Second
	if arch == "arm64" {
		timeout = 30 * time.Second
	}
	require.Eventually(t, func() bool {
		return execCode(t, container, 5*time.Second, "systemctl show-environment") == 0
	}, timeout, time.Second)
}

func copyFile(container *testutils.Container, source, target string) error {
	targetDir := filepath.Dir(target)
	if code, _, err := container.Exec(context.Background(), []string{"sh", "-c", "mkdir -p " + targetDir}); err != nil {
		return err
	} else if code != 0 {
		return fmt.Errorf("creating target directory %s exited with code %d", targetDir, code)
	}
	if err := container.CopyFileToContainer(context.Background(), source, target, 0o644); err != nil {
		return err
	}
	if code, _, err := container.Exec(context.Background(), []string{"sh", "-c", "test -f " + target}); err != nil {
		return err
	} else if code != 0 {
		return fmt.Errorf("copied %s to %s, but target file was not found", source, target)
	}
	return nil
}

func appendPreload(t *testing.T, container *testutils.Container) error {
	t.Helper()
	return execError(t, container, time.Minute, "echo "+libOtelInjectPath+" >> "+preloadPath)
}

func verifyPreload(t *testing.T, container *testutils.Container, line string, exists bool) {
	t.Helper()
	content := assertExec(t, container, time.Minute, "cat "+preloadPath)
	present := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(line) + `$`).MatchString(content)
	if exists {
		require.True(t, present, "%q not found in %s", line, preloadPath)
	} else {
		require.False(t, present, "%q found in %s", line, preloadPath)
	}
}

func assertFileExists(t *testing.T, container *testutils.Container, path string) {
	t.Helper()
	assertExec(t, container, time.Minute, "test -e "+path)
}

func assertFileAbsent(t *testing.T, container *testutils.Container, path string) {
	t.Helper()
	assertExecNonZero(t, container, time.Minute, "test -e "+path)
}

func assertExec(t *testing.T, container *testutils.Container, timeout time.Duration, command string) string {
	t.Helper()
	code, stdout, stderr := runCommand(t, container, timeout, command, nil, "", "")
	require.Equalf(t, 0, code, "command %q failed\nstdout:\n%s\nstderr:\n%s", command, stdout, stderr)
	return stdout
}

func assertExecNonZero(t *testing.T, container *testutils.Container, timeout time.Duration, command string) {
	t.Helper()
	code, stdout, stderr := runCommand(t, container, timeout, command, nil, "", "")
	require.NotEqualf(t, 0, code, "command %q unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", command, stdout, stderr)
}

func commandOutput(t *testing.T, container *testutils.Container, command string) string {
	t.Helper()
	_, stdout, _ := runCommand(t, container, time.Minute, command, nil, "", "")
	return stdout
}

func execCode(t *testing.T, container *testutils.Container, timeout time.Duration, command string) int {
	t.Helper()
	code, _, _ := runCommand(t, container, timeout, command, nil, "", "")
	return code
}

func execError(t *testing.T, container *testutils.Container, timeout time.Duration, command string) error {
	t.Helper()
	code, _, stderr := runCommand(t, container, timeout, command, nil, "", "")
	if code != 0 {
		return fmt.Errorf("command %q failed: %s", command, stderr)
	}
	return nil
}

func runCommand(t *testing.T, container *testutils.Container, timeout time.Duration, command string, env map[string]string, user, workdir string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	options := []exec.ProcessOption{exec.Multiplexed()}
	if user != "" {
		options = append(options, exec.WithUser(user))
	}
	if workdir != "" {
		options = append(options, exec.WithWorkingDir(workdir))
	}
	if len(env) != 0 {
		variables := make([]string, 0, len(env))
		for key, value := range env {
			variables = append(variables, key+"="+value)
		}
		slices.Sort(variables)
		options = append(options, exec.WithEnv(variables))
	}
	code, reader, err := container.Exec(ctx, []string{"sh", "-c", command}, options...)
	require.NoError(t, err)
	var output bytes.Buffer
	_, err = io.Copy(&output, reader)
	require.NoError(t, err)
	return code, output.String(), ""
}

func selectedDistros(t *testing.T) []string {
	t.Helper()
	packageType := selectedPackageType(t)
	matches, err := filepath.Glob(filepath.Join(repoRoot(t), "packaging", "tests", "instrumentation", "images", packageType, "Dockerfile.*"))
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	distros := make([]string, 0, len(matches))
	for _, match := range matches {
		distros = append(distros, strings.TrimPrefix(filepath.Base(match), "Dockerfile."))
	}
	slices.Sort(distros)
	if selected := os.Getenv("PACKAGE_TEST_DISTRO"); selected != "" {
		require.Contains(t, distros, selected)
		return []string{selected}
	}
	return distros
}

func selectedArches(t *testing.T) []string {
	t.Helper()
	if arch := os.Getenv("PACKAGE_TEST_ARCH"); arch != "" {
		require.Contains(t, []string{"amd64", "arm64"}, arch)
		return []string{arch}
	}
	return []string{"amd64", "arm64"}
}

func selectedPackageType(t *testing.T) string {
	t.Helper()
	selected := os.Getenv("PACKAGE_TEST_TYPE")
	if selected == "" {
		selected = os.Getenv("SYS_PACKAGE")
	}
	if selected == "" {
		t.Skip("PACKAGE_TEST_TYPE or SYS_PACKAGE must select deb or rpm")
	}
	require.Contains(t, []string{"deb", "rpm"}, selected, "PACKAGE_TEST_TYPE or SYS_PACKAGE must select deb or rpm")
	return selected
}

func packageType(t *testing.T, distro string) string {
	t.Helper()
	for _, packageType := range []string{"deb", "rpm"} {
		if _, err := os.Stat(filepath.Join(repoRoot(t), "packaging", "tests", "instrumentation", "images", packageType, "Dockerfile."+distro)); err == nil {
			return packageType
		}
	}
	t.Fatalf("unknown instrumentation distro %q", distro)
	return ""
}

func distroIsDeb(t *testing.T, distro string) bool {
	t.Helper()
	return packageType(t, distro) == "deb"
}

func requirePackage(t *testing.T, distro, arch string) string {
	t.Helper()
	packageArch := arch
	if !distroIsDeb(t, distro) {
		packageArch = map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[arch]
	}
	pattern := filepath.Join(repoRoot(t), "instrumentation", "dist", packageName+"*"+packageArch+"."+packageType(t, distro))
	matches, err := filepath.Glob(pattern)
	require.NoError(t, err)
	require.NotEmpty(t, matches, "%s %s package not found in instrumentation/dist", packageName, arch)
	slices.Sort(matches)
	return matches[len(matches)-1]
}

func legacyPackageName(t *testing.T, distro, arch string) string {
	t.Helper()
	if distroIsDeb(t, distro) {
		return fmt.Sprintf("%s_%s_%s.deb", packageName, legacyVersion, arch)
	}
	rpmArch := "x86_64"
	if arch == "arm64" {
		rpmArch = "aarch64"
	}
	return fmt.Sprintf("%s-%s-1.%s.rpm", packageName, legacyVersion, rpmArch)
}

func legacyFiles(arch string) []string {
	files := []string{libsplunkPath, zeroconfigDir + "/java.conf", zeroconfigDir + "/node.conf"}
	if arch != "arm64" {
		files = append(files, zeroconfigDir+"/dotnet.conf")
	}
	return files
}

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "packaging", "tests", "instrumentation", name)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func mergeEnv(base, overrides map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(overrides))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

func stringPtr(value string) *string { return &value }

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
