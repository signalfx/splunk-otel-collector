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

import pytest

from tests.helpers.util import (
    copy_file_into_container,
    run_container_cmd,
    run_distro_container,
    service_is_running,
    wait_for,
    DEB_DISTROS,
    REPO_DIR,
    RPM_DISTROS,
    TAR_DISTROS,
    TESTS_DIR,
)

INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"
PKG_NAME = "splunk-otel-collector"
PKG_DIR = REPO_DIR / "dist"
SERVICE_NAME = "splunk-otel-collector"
SERVICE_OWNER = "splunk-otel-collector"
SERVICE_PROC = "otelcol"
ENV_PATH = "/etc/otel/collector/splunk-otel-collector.conf"
AGENT_CONFIG_PATH = "/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_PATH = "/etc/otel/collector/gateway_config.yaml"
BUNDLE_DIR = "/usr/lib/splunk-otel-collector/agent-bundle"

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
    elif distro in TAR_DISTROS:
        pkg_paths = glob.glob(str(path / f"{name}*{arch}.tar.gz"))

    if pkg_paths:
        return sorted(pkg_paths)[-1]
    else:
        return None


def get_libcap_command(container):
    if container.exec_run("command -v yum").exit_code == 0:
        return "yum install -y libcap"
    elif container.exec_run("command -v dnf").exit_code == 0:
        return "dnf install -y libcap"
    else:
        return "zypper install -y libcap-progs"

@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.tar) for distro in TAR_DISTROS]
)
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_tar_collector_package_install(distro, arch):
    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} {arch} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, arch=arch) as container:

        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        run_container_cmd(container, f"tar xzf /test/{pkg_base} -C /tmp")
        bundle_dir = "/tmp/splunk-otel-collector"
        run_container_cmd(container, f"test -d {bundle_dir}/agent-bundle")
        run_container_cmd(container, f"test -d {bundle_dir}/agent-bundle/collectd-python")
        run_container_cmd(container, f"{bundle_dir}/agent-bundle/bin/patch-interpreter {bundle_dir}/agent-bundle")
        run_container_cmd(container, f"{bundle_dir}/agent-bundle/jre/bin/java -version")
        run_container_cmd(container, f"{bundle_dir}/agent-bundle/bin/python --version")
        run_container_cmd(container, f"test -d {bundle_dir}/bin")
        run_container_cmd(container, f"test -f {bundle_dir}/bin/otelcol")
        run_container_cmd(container, f"test -f {bundle_dir}/bin/translatesfx")
        run_container_cmd(container, f"test -f {bundle_dir}/config/agent_config.yaml")
        run_container_cmd(container, f"test -f {bundle_dir}/config/gateway_config.yaml")

@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_collector_package_install(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} {arch} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, arch) as container:
        # qemu is slow, so wait for systemd to be ready
        assert wait_for(lambda: run_container_cmd(container, "systemctl list-units --no-pager"), timeout=30)

        # install setcap dependency
        if distro in RPM_DISTROS:
            run_container_cmd(container, get_libcap_command(container))
        else:
            run_container_cmd(container, "apt-get update")
            run_container_cmd(container, "apt-get install -y libcap2-bin")

        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        try:
            # install package
            if distro in DEB_DISTROS:
                run_container_cmd(container, f"dpkg -i /test/{pkg_base}")
            elif distro in RPM_DISTROS:
                run_container_cmd(container, f"rpm -i /test/{pkg_base}")

            run_container_cmd(container, f"test -d {BUNDLE_DIR}")
            run_container_cmd(container, f"test -d {BUNDLE_DIR}/run/collectd")
            run_container_cmd(container, f"{BUNDLE_DIR}/jre/bin/java -version")
            run_container_cmd(container, f"{BUNDLE_DIR}/bin/python --version")
            run_container_cmd(container, f"test -d {BUNDLE_DIR}/collectd-python")

            run_container_cmd(container, f"test -f {AGENT_CONFIG_PATH}")
            run_container_cmd(container, f"test -f {GATEWAY_CONFIG_PATH}")

            # verify service is not running after install without config file
            time.sleep(5)
            assert not service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC)

            # verify service starts with config file
            run_container_cmd(container, f"cp -f {ENV_PATH}.example {ENV_PATH}")
            run_container_cmd(container, f"systemctl start {SERVICE_NAME}")
            time.sleep(5)
            assert wait_for(lambda: service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC))

            # verify service restart
            run_container_cmd(container, f"systemctl restart {SERVICE_NAME}")
            time.sleep(5)
            assert wait_for(lambda: service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC))

            # verify service stop
            run_container_cmd(container, f"systemctl stop {SERVICE_NAME}")
            time.sleep(5)
            assert not service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC)
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")

        # verify uninstall
        run_container_cmd(container, f"systemctl start {SERVICE_NAME}")

        time.sleep(5)

        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -P {PKG_NAME}")
        elif distro in RPM_DISTROS:
            run_container_cmd(container, f"rpm -e {PKG_NAME}")

        time.sleep(5)
        assert not service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC)

        # verify config file is not removed
        run_container_cmd(container, f"test -f {ENV_PATH}")


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_collector_package_upgrade(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    install_cmd = f"sh /test/install.sh -- testing123 --realm test --without-fluentd --collector-version 0.35.0"

    pkg_path = get_package(distro, PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, arch=arch) as container:
        # qemu is slow, so wait for systemd to be ready
        assert wait_for(lambda: run_container_cmd(container, "systemctl list-units --no-pager"), timeout=30)

        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # install an older version of the collector package
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})

            time.sleep(5)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            # change the config
            run_container_cmd(container, f"sh -c 'echo \"# This line should be preserved\" >> {AGENT_CONFIG_PATH}'")

            copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

            # upgrade package
            if distro in DEB_DISTROS:
                run_container_cmd(container, f"dpkg -i --force-confold /test/{pkg_base}")
            elif distro in RPM_DISTROS:
                run_container_cmd(container, f"rpm -U /test/{pkg_base}")

            time.sleep(5)

            run_container_cmd(container, f"test -d {BUNDLE_DIR}")
            run_container_cmd(container, f"test -d {BUNDLE_DIR}/run/collectd")
            run_container_cmd(container, f"{BUNDLE_DIR}/jre/bin/java -version")
            run_container_cmd(container, f"{BUNDLE_DIR}/bin/python --version")
            run_container_cmd(container, f"test -d {BUNDLE_DIR}/collectd-python")

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            # verify changed config was preserved after upgrade
            run_container_cmd(container, f"grep '# This line should be preserved' {AGENT_CONFIG_PATH}")

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
