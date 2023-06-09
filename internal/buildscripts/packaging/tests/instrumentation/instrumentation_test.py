# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import asyncio
import glob
import os

from pathlib import Path

import pytest

from tests.helpers.util import (
    copy_file_into_container,
    retry,
    run_container_cmd,
    run_distro_container,
    wait_for,
    wait_for_container_cmd,
    REPO_DIR,
    TESTS_DIR,
)


IMAGES_DIR = Path(__file__).parent.resolve() / "images"
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "rpm" / "Dockerfile.*"))]
OTELCOL_BIN_DIR = REPO_DIR / "bin"
COLLECTOR_CONFIG_PATH = TESTS_DIR / "instrumentation" / "config.yaml"
DEFAULT_CONF_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf"
DEFAULT_PROPERTIES_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties"
CUSTOM_CONF_PATH = TESTS_DIR / "instrumentation" / "01-my-env-vars.conf"
CUSTOM_CONF_INSTALL_PATH = os.path.join(os.path.dirname(DEFAULT_CONF_PATH), os.path.basename(CUSTOM_CONF_PATH))
CUSTOM_PROPERTIES_PATH = TESTS_DIR / "instrumentation" / "splunk-otel-javaagent.properties"
PKG_NAME = "splunk-otel-systemd-auto-instrumentation"
OLD_PKG_NAME = "splunk-otel-auto-instrumentation"
PKG_DIR = REPO_DIR / "instrumentation" / "dist"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"


def get_dockerfile(distro):
    if distro in DEB_DISTROS:
        return IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        return IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"


def get_package(distro, name, path, arch):
    pkg_paths = []
    if distro in DEB_DISTROS:
        pkg_paths = glob.glob(str(path / f"{name}*{arch}.deb"))
    elif distro in RPM_DISTROS:
        if arch == "amd64":
            arch = "x86_64"
        elif arch == "arm64":
            arch = "aarch64"
        pkg_paths = glob.glob(str(path / f"{name}*{arch}.rpm"))

    if pkg_paths:
        return sorted(pkg_paths)[-1]
    else:
        return None


def install_package(container, distro, path):
    if distro in DEB_DISTROS:
        run_container_cmd(container, f"apt-get install -y {path}")
    elif "opensuse" in distro:
        run_container_cmd(container, f"zypper --no-gpg-checks install -y {path}")
    elif container.exec_run("command -v yum").exit_code == 0:
        run_container_cmd(container, f"yum install -y {path}")
    elif container.exec_run("command -v dnf").exit_code == 0:
        run_container_cmd(container, f"dnf install -y {path}")



def verify_tomcat_instrumentation(container, config, otelcol_path=None):
    # overwrite the default config installed by the package with the custom test config
    if config == "env_vars":
        copy_file_into_container(container, CUSTOM_CONF_PATH, CUSTOM_CONF_INSTALL_PATH)
    elif config == "properties_file":
        copy_file_into_container(container, CUSTOM_PROPERTIES_PATH, DEFAULT_PROPERTIES_PATH)

    print("Restarting container ...")
    run_container_cmd(container, "systemctl stop tomcat")
    container.restart()

    # wait for systemd and verify env vars were configured
    _, env_vars = wait_for_container_cmd(container, "systemctl show-environment", timeout=30)
    assert f"JAVA_TOOL_OPTIONS=-javaagent:{JAVA_AGENT_PATH}" in str(env_vars)
    assert f"OTEL_JAVAAGENT_CONFIGURATION_FILE={DEFAULT_PROPERTIES_PATH}" in str(env_vars)

    # start the collector and get the output stream
    if otelcol_path:
        _, stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True)
    else:
        wait_for_container_cmd(container, "systemctl status splunk-otel-collector", timeout=30)
        _, stream = container.exec_run("journalctl -f -u splunk-otel-collector", stream=True)

    print("Waiting for tomcat ...")
    wait_for_container_cmd(container, "systemctl status tomcat", timeout=30)
    wait_for_container_cmd(container, "curl -sSL http://127.0.0.1:8080/sample", timeout=180)

    # check tomcat logs to ensure the java agent was picked up
    _, tomcat_logs = run_container_cmd(container, "cat /usr/local/tomcat/logs/catalina.out")
    assert f"-javaagent:{JAVA_AGENT_PATH}" in str(tomcat_logs), f"'{JAVA_AGENT_PATH}' not found in tomcat logs"

    # check the collector output for span and attributes
    async def check_output():
        span = "http.target: Str(/sample)"
        span_found = False
        service_name = f"service: Str(service_name_from_{config})"
        service_name_found = False
        deployment_environment = f"deployment_environment: Str(deployment_environment_from_{config})"
        deployment_environment_found = False
        profiling = "com.splunk.sourcetype: Str(otel.profiling)"
        profiling_found = False
        for data in stream:
            output = data.decode("utf-8")
            print(output.rstrip())
            if span in output:
                span_found = True
            if service_name in output:
                service_name_found = True
            if deployment_environment in output:
                deployment_environment_found = True
            if profiling in output:
                profiling_found = True
            if span_found and service_name_found and deployment_environment_found and profiling_found:
                return
            await asyncio.sleep(1)

    try:
        asyncio.run(asyncio.wait_for(check_output(), timeout=300))
    except asyncio.exceptions.TimeoutError:
        raise AssertionError("timed out waiting for span/attributes")


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("config", ["env_vars", "properties_file"])
def test_package_install(distro, arch, config):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        # install the instrumentation package
        install_package(container, distro, f"/test/{pkg_base}")

        # verify files were installed
        run_container_cmd(container, f"test -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_CONF_PATH}")

        verify_tomcat_instrumentation(container, config, otelcol_path=f"/test/{otelcol_bin}")


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("config", ["env_vars", "properties_file"])
def test_package_upgrade(distro, arch, config):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    install_cmd = "sh /test/install.sh -- testing123 --realm test --without-fluentd " \
                  "--collector-config /test/config.yaml --with-instrumentation --instrumentation-version 0.76.0"

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        # add a comment to /etc/ld.so.preload
        run_container_cmd(container, "sh -c 'echo \"# This line should be preserved\" >> /etc/ld.so.preload'")

        # install the collector and an older version of the instrumentation package
        run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout="10m")

        # verify the comment was preserved after install
        run_container_cmd(container, "grep '# This line should be preserved' /etc/ld.so.preload")

        # verify the libsplunk.so entry was added after install
        run_container_cmd(container, "grep '/usr/lib/splunk-instrumentation/libsplunk.so' /etc/ld.so.preload")

        # verify libsplunk.so was installed
        run_container_cmd(container, "test -f /usr/lib/splunk-instrumentation/libsplunk.so")

        # upgrade the instrumentation package
        install_package(container, distro, f"/test/{pkg_base}")

        # verify the comment was preserved after upgrade
        run_container_cmd(container, "grep '# This line should be preserved' /etc/ld.so.preload")

        # verify the libsplunk.so entry was removed after upgrade
        run_container_cmd(container, "grep -v '/usr/lib/splunk-instrumentation/libsplunk.so' /etc/ld.so.preload")

        # verify libsplunk.so was deleted after upgrade
        run_container_cmd(container, "test ! -f /usr/lib/splunk-instrumentation/libsplunk.so")

        # verify files were installed after upgrade
        run_container_cmd(container, f"test -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_CONF_PATH}")

        verify_tomcat_instrumentation(container, config)

@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_package_uninstall(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        # install the package
        install_package(container, distro, f"/test/{pkg_base}")

        # verify files were installed
        run_container_cmd(container, f"test -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_CONF_PATH}")

        # uninstall the package
        if distro in DEB_DISTROS:
            run_container_cmd(container, f"apt-get purge -y {PKG_NAME}")
        elif "opensuse" in distro:
            run_container_cmd(container, f"zypper remove -y {PKG_NAME}")
        elif container.exec_run("command -v yum").exit_code == 0:
            run_container_cmd(container, f"yum remove -y {PKG_NAME}")
        elif container.exec_run("command -v dnf").exit_code == 0:
            run_container_cmd(container, f"dnf remove -y {PKG_NAME}")

        # verify the package was uninstalled
        if distro in DEB_DISTROS:
            assert container.exec_run(f"dpkg -s {PKG_NAME}").exit_code != 0
            assert container.exec_run(f"dpkg -s {OLD_PKG_NAME}").exit_code != 0
        else:
            assert container.exec_run(f"rpm -q {PKG_NAME}").exit_code != 0
            assert container.exec_run(f"rpm -q {OLD_PKG_NAME}").exit_code != 0

        # verify files were uninstalled
        run_container_cmd(container, "test ! -f /etc/ld.so.preload")
        run_container_cmd(container, "test ! -f /usr/lib/splunk-instrumentation/libsplunk.so")
        run_container_cmd(container, f"test ! -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test ! -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test ! -f {DEFAULT_CONF_PATH}")
