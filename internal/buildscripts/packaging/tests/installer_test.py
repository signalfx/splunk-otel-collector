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
    SERVICE_NAME,
    SERVICE_OWNER,
    TESTS_DIR,
)

INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"

# Override default test parameters with the following env vars
STAGE = os.environ.get("STAGE", "release")
VERSIONS = os.environ.get("VERSIONS", "latest").split(",")

SPLUNK_ENV_PATH = "/etc/otel/collector/splunk-otel-collector.conf"
OLD_SPLUNK_ENV_PATH = "/etc/otel/collector/splunk_env"
AGENT_CONFIG_PATH = "/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_PATH = "/etc/otel/collector/gateway_config.yaml"
OLD_CONFIG_PATH = "/etc/otel/collector/splunk_config_linux.yaml"
TOTAL_MEMORY = "256"
BALLAST = "64"
REALM = "test"
INSTR_CONF_PATH = "/usr/lib/splunk-instrumentation/instrumentation.conf"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"


def verify_env_file(container, mode="agent", ballast=None):
    env_path = SPLUNK_ENV_PATH
    if container.exec_run(f"test -f {OLD_SPLUNK_ENV_PATH}").exit_code == 0:
        env_path = OLD_SPLUNK_ENV_PATH

    config_path = AGENT_CONFIG_PATH if mode == "agent" else GATEWAY_CONFIG_PATH
    if container.exec_run(f"test -f {OLD_CONFIG_PATH}").exit_code == 0:
        config_path = OLD_CONFIG_PATH
    elif mode == "gateway" and container.exec_run(f"test -f {GATEWAY_CONFIG_PATH}").exit_code != 0:
        config_path = AGENT_CONFIG_PATH

    run_container_cmd(container, f"grep '^SPLUNK_CONFIG={config_path}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_ACCESS_TOKEN=testing123$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_REALM={REALM}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_API_URL=https://api.{REALM}.signalfx.com$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_INGEST_URL=https://ingest.{REALM}.signalfx.com$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_TRACE_URL=https://ingest.{REALM}.signalfx.com/v2/trace$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_URL=https://ingest.{REALM}.signalfx.com/v1/log$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_TOKEN=testing123$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_MEMORY_TOTAL_MIB={TOTAL_MEMORY}$' {env_path}")

    if ballast:
        run_container_cmd(container, f"grep '^SPLUNK_BALLAST_SIZE_MIB={BALLAST}$' {env_path}")


def verify_support_bundle(container):
    run_container_cmd(container, "/etc/otel/collector/splunk-support-bundle.sh -t /tmp/splunk-support-bundle")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/config/agent_config.yaml")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/logs/splunk-otel-collector.log")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/logs/splunk-otel-collector.txt")
    if container.exec_run("test -f /etc/otel/collector/fluentd/fluent.conf").exit_code == 0:
        run_container_cmd(container, "test -f /tmp/splunk-support-bundle/logs/td-agent.log")
        run_container_cmd(container, "test -f /tmp/splunk-support-bundle/logs/td-agent.txt")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/metrics/collector-metrics.txt")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/metrics/df.txt")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/metrics/free.txt")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/metrics/top.txt")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle/zpages/tracez.html")
    run_container_cmd(container, "test -f /tmp/splunk-support-bundle.tar.gz")


def verify_uninstall(container, distro):
    run_container_cmd(container, "sh -x /test/install.sh --uninstall")

    for pkg in ("splunk-otel-collector", "td-agent", "splunk-otel-auto-instrumentation"):
        if distro in DEB_DISTROS:
            assert container.exec_run(f"dpkg -s {pkg}").exit_code != 0
        else:
            assert container.exec_run(f"rpm -q {pkg}").exit_code != 0


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", VERSIONS)
@pytest.mark.parametrize("mode", ["agent", "gateway"])
def test_installer_mode(distro, version, mode):
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY} --mode {mode}"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        # run installer script
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify splunk-otel-auto-instrumentation is not installed
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code != 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code != 0

            # verify env file created with configured parameters
            verify_env_file(container, mode=mode)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            if "opensuse" not in distro and distro != "ubuntu-jammy":
                assert container.exec_run("systemctl status td-agent").exit_code == 0

            # test support bundle script
            verify_support_bundle(container)

            verify_uninstall(container, distro)

        finally:
            if "opensuse" not in distro and distro != "ubuntu-jammy":
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", VERSIONS)
def test_installer_ballast(distro, version):
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY} --ballast {BALLAST}"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        # run installer script
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify env file created with configured parameters
            verify_env_file(container, ballast=BALLAST)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            if "opensuse" not in distro and distro != "ubuntu-jammy":
                assert container.exec_run("systemctl status td-agent").exit_code == 0

            verify_uninstall(container, distro)

        finally:
            if "opensuse" not in distro and distro != "ubuntu-jammy":
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("version", VERSIONS)
def test_installer_service_owner(distro, version):
    service_owner = "test-user"
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY}"
    install_cmd = f"{install_cmd} --service-user {service_owner} --service-group {service_owner}"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify env file created with configured parameters
            verify_env_file(container)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=service_owner))

            if "opensuse" not in distro and distro != "ubuntu-jammy":
                assert container.exec_run("systemctl status td-agent").exit_code == 0

            verify_uninstall(container, distro)

        finally:
            if "opensuse" not in distro and distro != "ubuntu-jammy":
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", VERSIONS)
def test_installer_without_fluentd(distro, version):
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY} --without-fluentd"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify env file created with configured parameters
            verify_env_file(container)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s td-agent").exit_code != 0
            else:
                assert container.exec_run("rpm -q td-agent").exit_code != 0

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", VERSIONS)
def test_installer_with_instrumentation(distro, version):
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY} --with-instrumentation"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify env file created with configured parameters
            verify_env_file(container)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            # verify splunk-otel-auto-instrumentation is installed
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code == 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code == 0

            # verify /etc/ld.so.preload is configured
            run_container_cmd(container, f"grep '^{LIBSPLUNK_PATH}$' /etc/ld.so.preload")

            # verify deployment.environment attribute is not set
            run_container_cmd(container, f"grep -v '^resource_attributes=deployment.environment=.*$' {INSTR_CONF_PATH}")

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", VERSIONS)
def test_installer_with_deployment_environment(distro, version):
    install_cmd = f"sh -x /test/install.sh -- testing123 --realm {REALM} --memory {TOTAL_MEMORY} --with-instrumentation"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    install_cmd = f"{install_cmd} --deployment-environment test"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
            time.sleep(5)

            # verify env file created with configured parameters
            verify_env_file(container)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            # verify splunk-otel-auto-instrumentation is installed
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code == 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code == 0

            # verify /etc/ld.so.preload is configured
            run_container_cmd(container, f"grep '^{LIBSPLUNK_PATH}$' /etc/ld.so.preload")

            # verify deployment.environment is set
            run_container_cmd(container, f"grep '^resource_attributes=deployment.environment=test$' {INSTR_CONF_PATH}")

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
