# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

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

PKG_NAME = "splunk-otel-script"
PKG_DIR = REPO_DIR / "dist"
SERVICE_NAME = "splunk-otel-script"
SERVICE_OWNER = "splunk-otel-collector"
SERVICE_PROC = "cron"
ENV_PATH = "/etc/otel/collector/splunk-otel-script.conf"
BUNDLE_DIR = "/usr/lib/splunk-otel-script"

def get_package(distro, packageType, name, path, arch):
    pkg_paths = []
    if packageType == "tar":
        pkg_paths = glob.glob(str(path / f"{name}*{arch}.tar.gz"))
    elif distro in DEB_DISTROS:
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


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.tar) for distro in TAR_DISTROS]
)
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_tar_script_package_install(distro, arch):
    pkg_path = get_package(distro, "tar", PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} {arch} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, arch=arch) as container:

        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        run_container_cmd(container, f"tar xzf /test/{pkg_base} -C /tmp")
        bundle_dir = "/tmp/splunk-otel-script"
        run_container_cmd(container, f"test -f {bundle_dir}/cron.py")

@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_script_package_install(distro, arch):
    pkg_path = get_package(distro, "", PKG_NAME, PKG_DIR, arch)
    assert pkg_path, f"{PKG_NAME} {arch} package not found in {PKG_DIR}"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, arch) as container:
        # install python dependency
        if distro in DEB_DISTROS:
            run_container_cmd(container, "apt-get update")
            run_container_cmd(container, "apt-get -y install python3 gcc")
        else:
            run_container_cmd(container, "yum install -y python3 gcc-c++ python3-devel")

        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        try:
            # install package
            if distro in DEB_DISTROS:
                run_container_cmd(container, f"dpkg -i /test/{pkg_base}")
            elif distro in RPM_DISTROS:
                run_container_cmd(container, f"rpm -i /test/{pkg_base}")

            run_container_cmd(container, f"test -d {BUNDLE_DIR}")

            # verify service is running after install
            time.sleep(5)
            assert service_is_running(container, SERVICE_NAME, SERVICE_OWNER, SERVICE_PROC)

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
