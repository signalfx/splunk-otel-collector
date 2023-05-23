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

import glob
import os
import time

from pathlib import Path

import pytest

from tests.helpers.util import (
    copy_file_into_container,
    run_container_cmd,
    run_distro_container,
    wait_for,
    REPO_DIR,
    TESTS_DIR,
)


IMAGES_DIR = Path(__file__).parent.resolve() / "images"
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "rpm" / "Dockerfile.*"))]
OTELCOL_BIN_DIR = REPO_DIR / "bin"
COLLECTOR_CONFIG_PATH = TESTS_DIR / "instrumentation" / "config.yaml"
DEFAULT_CONF_PATH = "/etc/systemd/system.conf.d/splunk-otel-javaagent.conf"
DEFAULT_PROPERTIES_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties"
CUSTOM_CONF_PATH = TESTS_DIR / "instrumentation" / "splunk-otel-javaagent.conf"
CUSTOM_PROPERTIES_PATH = TESTS_DIR / "instrumentation" / "splunk-otel-javaagent.properties"
PKG_NAME = "splunk-otel-auto-instrumentation"
PKG_DIR = REPO_DIR / "instrumentation" / "dist"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"


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


def verify_tomcat_instrumentation(container, distro, config, otelcol_path=None):
    tomcat_service = "tomcat9"
    if distro in RPM_DISTROS and distro != "amazonlinux-2023":
        tomcat_service = "tomcat"

    # overwrite the default config installed by the package with the custom test config
    if config == "env_vars":
        copy_file_into_container(container, CUSTOM_CONF_PATH, DEFAULT_CONF_PATH)
        run_container_cmd(container, "systemctl daemon-reload")
    elif config == "properties_file":
        copy_file_into_container(container, CUSTOM_PROPERTIES_PATH, DEFAULT_PROPERTIES_PATH)

    # restart tomcat to pick up the env vars and ensure it is running
    run_container_cmd(container, f"systemctl restart {tomcat_service}", timeout="1m")
    time.sleep(5)
    run_container_cmd(container, f"systemctl status {tomcat_service}")

    # check tomcat logs to ensure the java agent was picked up
    _, logs = run_container_cmd(container, f"journalctl -u {tomcat_service} --no-pager")
    assert f"-javaagent:{JAVA_AGENT_PATH}" in logs.decode("utf-8"), \
        f"'{JAVA_AGENT_PATH}' not found in tomcat logs"

    # expected dimensions defined in the custom test config
    service_name = f"service_name_from_{config}"
    service_name_found = False
    deployment_environment = f"deployment_environment_from_{config}"
    deployment_environment_found = False

    # start the collector and check the output/logs for datapoints with the custom dimensions
    if otelcol_path:
        _, stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True)
    else:
        _, stream = container.exec_run("journalctl -f -u splunk-otel-collector", stream=True)

    start_time = time.time()
    for data in stream:
        output = data.decode("utf-8")
        print(output)
        if f"service: Str({service_name})" in output:
            service_name_found = True
        if f"deployment_environment: Str({deployment_environment})" in output:
            deployment_environment_found = True
        if service_name_found and deployment_environment_found:
            break
        assert (time.time() - start_time) < 300, \
            f"timed out waiting for '{service_name}' and '{deployment_environment}'"


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("config", ["env_vars", "properties_file"])
def test_package_install(distro, arch, config):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=dockerfile, arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        # install the instrumentation package
        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -i /test/{pkg_base}")
        else:
            run_container_cmd(container, f"rpm -i /test/{pkg_base}")

        # verify files were installed
        run_container_cmd(container, f"test -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_CONF_PATH}")

        verify_tomcat_instrumentation(container, distro, config, otelcol_path=f"/test/{otelcol_bin}")


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("config", ["env_vars", "properties_file"])
def test_package_upgrade(distro, arch, config):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    install_cmd = "sh /test/install.sh -- testing123 --realm test --without-fluentd " \
                  "--collector-config /test/config.yaml --with-instrumentation --instrumentation-version 0.76.0"

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=dockerfile, arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        # install the collector and an older version of the instrumentation package
        run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout="10m")

        # verify /etc/ld.so.preload and libsplunk.so were installed
        run_container_cmd(container, "test -f /etc/ld.so.preload")
        run_container_cmd(container, "test -f /usr/lib/splunk-instrumentation/libsplunk.so")

        # upgrade the instrumentation package
        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -i /test/{pkg_base}")
        elif distro in RPM_DISTROS:
            run_container_cmd(container, f"rpm -U /test/{pkg_base}")

        # verify /etc/ld.so.preload and libsplunk.so were deleted after upgrade
        run_container_cmd(container, "test ! -f /etc/ld.so.preload")
        run_container_cmd(container, "test ! -f /usr/lib/splunk-instrumentation/libsplunk.so")

        # verify files were installed after upgrade
        run_container_cmd(container, f"test -f {JAVA_AGENT_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_PROPERTIES_PATH}")
        run_container_cmd(container, f"test -f {DEFAULT_CONF_PATH}")

        verify_tomcat_instrumentation(container, distro, config)
