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
    TESTS_DIR,
)


def get_package(distro, name, path):
    if distro in DEB_DISTROS:
        pkg_paths = glob.glob(str(path / f"{name}*amd64.deb"))
    else:
        pkg_paths = glob.glob(str(path / f"{name}*x86_64.rpm"))

    if pkg_paths:
        return sorted(pkg_paths)[-1]
    else:
        return None


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
def test_collector_package_install(distro):
    pkg_name = "splunk-otel-collector"
    pkg_dir = REPO_DIR / "dist"
    service_name = "splunk-otel-collector"
    service_owner = "splunk-otel-collector"
    service_proc = "otelcol"
    env_path = "/etc/otel/collector/splunk-otel-collector.conf"
    agent_config_path = "/etc/otel/collector/agent_config.yaml"
    gateway_config_path = "/etc/otel/collector/gateway_config.yaml"
    bundle_dir = "/usr/lib/splunk-otel-collector/agent-bundle"

    pkg_path = get_package(distro, pkg_name, pkg_dir)
    assert pkg_path, f"{pkg_name} package not found in {pkg_dir}"

    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro) as container:
        # install setcap dependency
        if distro in RPM_DISTROS:
            if container.exec_run("command -v yum").exit_code == 0:
                run_container_cmd(container, "yum install -y libcap")
            else:
                run_container_cmd(container, "dnf install -y libcap")
        else:
            run_container_cmd(container, "apt-get update")
            run_container_cmd(container, "apt-get install -y libcap2-bin")

        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        try:
            # install package
            if distro in DEB_DISTROS:
                run_container_cmd(container, f"dpkg -i /test/{pkg_base}")
            else:
                run_container_cmd(container, f"rpm -i /test/{pkg_base}")

            run_container_cmd(container, f"test -d {bundle_dir}")
            run_container_cmd(container, f"test -d {bundle_dir}/run/collectd")

            run_container_cmd(container, f"test -f {agent_config_path}")
            run_container_cmd(container, f"test -f {gateway_config_path}")

            # verify service is not running after install without config file
            time.sleep(5)
            assert not service_is_running(container, service_name, service_owner, service_proc)

            # verify service starts with config file
            run_container_cmd(container, f"cp -f {env_path}.example {env_path}")
            run_container_cmd(container, f"systemctl start {service_name}")
            time.sleep(5)
            assert wait_for(lambda: service_is_running(container, service_name, service_owner, service_proc))

            # verify service restart
            run_container_cmd(container, f"systemctl restart {service_name}")
            time.sleep(5)
            assert wait_for(lambda: service_is_running(container, service_name, service_owner, service_proc))

            # verify service stop
            run_container_cmd(container, f"systemctl stop {service_name}")
            time.sleep(5)
            assert not service_is_running(container, service_name, service_owner, service_proc)
        finally:
            run_container_cmd(container, f"journalctl -u {service_name} --no-pager")

        # verify uninstall
        run_container_cmd(container, f"systemctl start {service_name}")

        time.sleep(5)

        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -P {pkg_name}")
        else:
            run_container_cmd(container, f"rpm -e {pkg_name}")

        time.sleep(5)
        assert not service_is_running(container, service_name, service_owner, service_proc)

        # verify config file is not removed
        run_container_cmd(container, f"test -f {env_path}")
