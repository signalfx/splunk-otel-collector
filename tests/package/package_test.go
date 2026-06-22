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

//go:build package_integration

package packagetest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	packageName       = "splunk-otel-collector"
	serviceName       = "splunk-otel-collector"
	serviceOwner      = "splunk-otel-collector"
	serviceProcess    = "otelcol"
	envPath           = "/etc/otel/collector/splunk-otel-collector.conf"
	agentConfigPath   = "/etc/otel/collector/agent_config.yaml"
	gatewayConfigPath = "/etc/otel/collector/gateway_config.yaml"
)

func TestTarCollectorPackageInstall(t *testing.T) {
	skipUnlessLinux(t)
	skipUnlessPackageType(t, "tar")

	for _, distro := range selectedDistros(t, "tar") {
		for _, arch := range selectedArches(t) {
			t.Run(fmt.Sprintf("%s/%s", distro, arch), func(t *testing.T) {
				pkgPath := requirePackage(t, "tar", arch)
				container := runDistroContainer(t, "tar", distro, arch)

				copyFileToContainer(t, container, pkgPath)
				assertExec(t, container, time.Minute, "tar xzf /test/"+filepath.Base(pkgPath)+" -C /tmp")

				bundleDir := "/tmp/splunk-otel-collector"
				assertExec(t, container, time.Minute, "test -d "+bundleDir+"/bin")
				assertExec(t, container, time.Minute, "test -f "+bundleDir+"/bin/otelcol")
				assertExec(t, container, time.Minute, "test -f "+bundleDir+"/opt/opentelemetry-java-contrib-jmx-metrics.jar")
				assertExec(t, container, time.Minute, "test -f "+bundleDir+"/config/agent_config.yaml")
				assertExec(t, container, time.Minute, "test -f "+bundleDir+"/config/gateway_config.yaml")
			})
		}
	}
}

func TestCollectorPackageInstall(t *testing.T) {
	skipUnlessLinux(t)
	packageType := skipUnlessPackageType(t, "deb", "rpm")

	for _, distro := range selectedDistros(t, packageType) {
		for _, arch := range selectedArches(t) {
			t.Run(fmt.Sprintf("%s/%s", distro, arch), func(t *testing.T) {
				pkgPath := requirePackage(t, packageType, arch)
				container := runDistroContainer(t, packageType, distro, arch)
				defer logJournal(t, container)

				installLibcap(t, container, packageType)
				copyFileToContainer(t, container, pkgPath)
				installPackage(t, container, packageType, "/test/"+filepath.Base(pkgPath))

				assertExec(t, container, time.Minute, "test -f "+agentConfigPath)
				assertExec(t, container, time.Minute, "test -f "+gatewayConfigPath)

				time.Sleep(5 * time.Second)
				require.False(t, serviceIsRunning(t, container), "service should not be running after package install without config")

				assertExec(t, container, time.Minute, "cp -f "+envPath+".example "+envPath)
				assertExec(t, container, time.Minute, "systemctl start "+serviceName)
				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container)
				}, 10*time.Second, time.Second)

				assertExec(t, container, time.Minute, "systemctl restart "+serviceName)
				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container)
				}, 10*time.Second, time.Second)

				assertExec(t, container, time.Minute, "systemctl stop "+serviceName)
				time.Sleep(5 * time.Second)
				require.False(t, serviceIsRunning(t, container), "service should stop cleanly")

				assertExec(t, container, time.Minute, "systemctl start "+serviceName)
				time.Sleep(5 * time.Second)

				uninstallPackage(t, container, packageType)
				time.Sleep(5 * time.Second)
				require.False(t, serviceIsRunning(t, container), "service should not be running after uninstall")
				assertExec(t, container, time.Minute, "test -f "+envPath)
			})
		}
	}
}

func TestCollectorPackageUpgrade(t *testing.T) {
	skipUnlessLinux(t)
	packageType := skipUnlessPackageType(t, "deb", "rpm")

	for _, distro := range selectedDistros(t, packageType) {
		for _, arch := range selectedArches(t) {
			t.Run(fmt.Sprintf("%s/%s", distro, arch), func(t *testing.T) {
				pkgPath := requirePackage(t, packageType, arch)
				container := runDistroContainer(t, packageType, distro, arch)
				defer logJournal(t, container)

				copyFileToContainer(t, container, filepath.Join(repoRoot(t), "packaging", "installer", "install.sh"))
				assertExec(t, container, 10*time.Minute, "VERIFY_ACCESS_TOKEN=false sh /test/install.sh -- testing123 --realm test --collector-version 0.35.0")

				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container)
				}, 20*time.Second, time.Second)

				copyFileToContainer(t, container, pkgPath)
				upgradePackage(t, container, packageType, "/test/"+filepath.Base(pkgPath))

				require.Eventually(t, func() bool {
					return serviceIsRunning(t, container)
				}, 20*time.Second, time.Second)
			})
		}
	}
}

func skipUnlessLinux(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("package integration tests require Linux systemd containers")
	}
}

func skipUnlessPackageType(t *testing.T, packageTypes ...string) string {
	t.Helper()
	selected := os.Getenv("PACKAGE_TEST_TYPE")
	if selected == "" {
		selected = os.Getenv("SYS_PACKAGE")
	}
	if selected == "" {
		if len(packageTypes) == 1 {
			return packageTypes[0]
		}
		t.Skip("PACKAGE_TEST_TYPE or SYS_PACKAGE must select one of: " + strings.Join(packageTypes, ", "))
	}
	if !slices.Contains(packageTypes, selected) {
		t.Skipf("package type %q is not covered by this test", selected)
	}
	return selected
}

func selectedArches(t *testing.T) []string {
	t.Helper()
	if arch := os.Getenv("PACKAGE_TEST_ARCH"); arch != "" {
		require.Contains(t, []string{"amd64", "arm64"}, arch)
		return []string{arch}
	}
	return []string{"amd64", "arm64"}
}

func selectedDistros(t *testing.T, packageType string) []string {
	t.Helper()
	distros := distrosForPackageType(t, packageType)
	if distro := os.Getenv("PACKAGE_TEST_DISTRO"); distro != "" {
		require.Contains(t, distros, distro)
		return []string{distro}
	}
	return distros
}

func distrosForPackageType(t *testing.T, packageType string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot(t), "packaging", "tests", "images", packageType, "Dockerfile.*"))
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no Dockerfiles found for package type %q", packageType)

	distros := make([]string, 0, len(matches))
	for _, match := range matches {
		distros = append(distros, strings.TrimPrefix(filepath.Base(match), "Dockerfile."))
	}
	slices.Sort(distros)
	return distros
}

func requirePackage(t *testing.T, packageType, arch string) string {
	t.Helper()
	packageArch := arch
	if packageType == "rpm" {
		switch arch {
		case "amd64":
			packageArch = "x86_64"
		case "arm64":
			packageArch = "aarch64"
		}
	}

	pattern := filepath.Join(repoRoot(t), "dist", packageName+"*"+packageArch+"."+packageType)
	if packageType == "tar" {
		pattern = filepath.Join(repoRoot(t), "dist", packageName+"*"+packageArch+".tar.gz")
	}
	matches, err := filepath.Glob(pattern)
	require.NoError(t, err)
	require.NotEmpty(t, matches, "%s %s package not found in dist", packageName, arch)
	slices.Sort(matches)
	return matches[len(matches)-1]
}

func runDistroContainer(t *testing.T, packageType, distro, arch string) *testutils.Container {
	t.Helper()
	root := repoRoot(t)
	dockerfile, err := filepath.Rel(root, filepath.Join(root, "packaging", "tests", "images", packageType, "Dockerfile."+distro))
	require.NoError(t, err)

	targetArch := arch
	platform := "linux/" + arch
	parsedPlatform, err := platforms.Parse(platform)
	require.NoError(t, err)
	startupTimeout := 5 * time.Minute

	container := testutils.NewContainer().
		WithContext(root).
		WithDockerfile(dockerfile).
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
		WithStartupTimeout(startupTimeout)

	// The Python package tests started the container without a Docker wait strategy,
	// then retried systemctl below. testcontainers requires at least one strategy
	// because the local wrapper builds wait.ForAll, so use a no-op strategy here
	// and keep the real readiness probe in waitForSystemd.
	container.WaitingFor = append(container.WaitingFor,
		wait.ForNop(func(context.Context, wait.StrategyTarget) error { return nil }).WithStartupTimeout(startupTimeout),
	)

	built := container.Build()
	require.NoError(t, built.Start(context.Background()))
	t.Cleanup(func() {
		require.NoError(t, built.Terminate(context.Background()))
	})

	waitForSystemd(t, built, arch)
	return built
}

func waitForSystemd(t *testing.T, container *testutils.Container, arch string) {
	t.Helper()
	timeout := 10 * time.Second
	if arch != "amd64" {
		timeout = 30 * time.Second
	}
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		rc, _, err := container.Exec(ctx, []string{"sh", "-c", "systemctl show-environment"})
		return err == nil && rc == 0
	}, timeout, time.Second)
}

func copyFileToContainer(t *testing.T, container *testutils.Container, sourcePath string) {
	t.Helper()
	assertExec(t, container, time.Minute, "mkdir -p /test")
	require.NoError(t, container.CopyFileToContainer(context.Background(), sourcePath, "/test/"+filepath.Base(sourcePath), 0o644))
	time.Sleep(2 * time.Second)
}

func installLibcap(t *testing.T, container *testutils.Container, packageType string) {
	t.Helper()
	if packageType == "deb" {
		assertExec(t, container, 5*time.Minute, "apt-get update")
		assertExec(t, container, 5*time.Minute, "apt-get install -y libcap2-bin")
		return
	}
	assertExec(t, container, 5*time.Minute, libcapCommand(t, container))
}

func libcapCommand(t *testing.T, container *testutils.Container) string {
	t.Helper()
	for _, candidate := range []struct {
		command string
		install string
	}{
		{"command -v yum", "yum install -y libcap"},
		{"command -v dnf", "dnf install -y libcap"},
	} {
		rc, _, _ := exec(t, container, time.Minute, candidate.command)
		if rc == 0 {
			return candidate.install
		}
	}
	return "zypper install -y libcap-progs"
}

func installPackage(t *testing.T, container *testutils.Container, packageType, packagePath string) {
	t.Helper()
	switch packageType {
	case "deb":
		assertExec(t, container, 5*time.Minute, "dpkg -i "+packagePath)
	case "rpm":
		assertExec(t, container, 5*time.Minute, "rpm -i "+packagePath)
	default:
		t.Fatalf("unsupported package type %q", packageType)
	}
}

func upgradePackage(t *testing.T, container *testutils.Container, packageType, packagePath string) {
	t.Helper()
	switch packageType {
	case "deb":
		assertExec(t, container, 5*time.Minute, "dpkg -i --force-confnew "+packagePath)
	case "rpm":
		assertExec(t, container, 5*time.Minute, "rpm -U "+packagePath)
	default:
		t.Fatalf("unsupported package type %q", packageType)
	}
}

func uninstallPackage(t *testing.T, container *testutils.Container, packageType string) {
	t.Helper()
	switch packageType {
	case "deb":
		assertExec(t, container, 5*time.Minute, "dpkg -P "+packageName)
	case "rpm":
		assertExec(t, container, 5*time.Minute, "rpm -e "+packageName)
	default:
		t.Fatalf("unsupported package type %q", packageType)
	}
}

func serviceIsRunning(t *testing.T, container *testutils.Container) bool {
	t.Helper()
	systemctlCode, _, _ := exec(t, container, time.Minute, "systemctl status "+serviceName)
	pgrepCode, _, _ := exec(t, container, time.Minute, "pgrep -a -u "+serviceOwner+" -f "+serviceProcess)
	return systemctlCode|pgrepCode == 0
}

func logJournal(t *testing.T, container *testutils.Container) {
	t.Helper()
	_, stdout, stderr := exec(t, container, time.Minute, "journalctl -u "+serviceName+" --no-pager")
	if stdout != "" {
		t.Log(stdout)
	}
	if stderr != "" {
		t.Log(stderr)
	}
}

func assertExec(t *testing.T, container *testutils.Container, timeout time.Duration, command string) string {
	t.Helper()
	rc, stdout, stderr := exec(t, container, timeout, command)
	require.Equalf(t, 0, rc, "command %q failed\nstdout:\n%s\nstderr:\n%s", command, stdout, stderr)
	return stdout
}

func exec(t *testing.T, container *testutils.Container, timeout time.Duration, command string) (int, string, string) {
	t.Helper()
	return container.AssertExec(t, timeout, "sh", "-c", command)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err = os.Stat(filepath.Join(wd, "packaging", "tests", "images")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not locate repository root")
		}
		wd = parent
	}
}
