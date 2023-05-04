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
VERSION = os.environ.get("VERSION", "latest")
SPLUNK_ACCESS_TOKEN = os.environ.get("SPLUNK_ACCESS_TOKEN", "testing123")
SPLUNK_REALM = os.environ.get("SPLUNK_REALM", "fake-realm")
TOTAL_MEMORY = "512"

SPLUNK_ENV_PATH = "/etc/otel/collector/splunk-otel-collector.conf"
OLD_SPLUNK_ENV_PATH = "/etc/otel/collector/splunk_env"
AGENT_CONFIG_PATH = "/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_PATH = "/etc/otel/collector/gateway_config.yaml"
OLD_CONFIG_PATH = "/etc/otel/collector/splunk_config_linux.yaml"
INSTR_CONF_PATH = "/usr/lib/splunk-instrumentation/instrumentation.conf"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"

INSTALLER_TIMEOUT = "30m"


def get_installer_cmd():
    install_cmd = f"sh -x /test/install.sh -- {SPLUNK_ACCESS_TOKEN} --realm {SPLUNK_REALM}"

    if VERSION != "latest":
        install_cmd = f"{install_cmd} --collector-version {VERSION.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    return install_cmd


def verify_env_file(container, mode="agent", config_path=None, memory=TOTAL_MEMORY, ballast=None):
    env_path = SPLUNK_ENV_PATH
    if container.exec_run(f"test -f {OLD_SPLUNK_ENV_PATH}").exit_code == 0:
        env_path = OLD_SPLUNK_ENV_PATH

    if not config_path:
        config_path = AGENT_CONFIG_PATH if mode == "agent" else GATEWAY_CONFIG_PATH
        if container.exec_run(f"test -f {OLD_CONFIG_PATH}").exit_code == 0:
            config_path = OLD_CONFIG_PATH
        elif mode == "gateway" and container.exec_run(f"test -f {GATEWAY_CONFIG_PATH}").exit_code != 0:
            config_path = AGENT_CONFIG_PATH

    ingest_url = f"https://ingest.{SPLUNK_REALM}.signalfx.com"
    api_url = f"https://api.{SPLUNK_REALM}.signalfx.com"

    run_container_cmd(container, f"grep '^SPLUNK_CONFIG={config_path}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_ACCESS_TOKEN={SPLUNK_ACCESS_TOKEN}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_REALM={SPLUNK_REALM}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_API_URL={api_url}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_INGEST_URL={ingest_url}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_TRACE_URL={ingest_url}/v2/trace$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_URL={ingest_url}/v1/log$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_TOKEN={SPLUNK_ACCESS_TOKEN}$' {env_path}")
    run_container_cmd(container, f"grep '^SPLUNK_MEMORY_TOTAL_MIB={memory}$' {env_path}")

    if ballast:
        run_container_cmd(container, f"grep '^SPLUNK_BALLAST_SIZE_MIB={ballast}$' {env_path}")


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


def fluentd_supported(distro, arch):
    if "opensuse" in distro:
        return False
    elif distro == "amazonlinux-2023":
        return False
    elif distro in ("debian-stretch", "ubuntu-xenial") and arch == "arm64":
        return False

    return True


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("mode", ["agent", "gateway"])
def test_installer_default(distro, arch, mode):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    install_cmd = get_installer_cmd()
    if mode != "agent":
        install_cmd = f"{install_cmd} --mode {mode}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        # run installer script
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
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

            if fluentd_supported(distro, arch):
                assert container.exec_run("systemctl status td-agent").exit_code == 0

            # test support bundle script
            verify_support_bundle(container)

            verify_uninstall(container, distro)

        finally:
            if fluentd_supported(distro, arch):
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
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_installer_custom(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    collector_version = "0.74.0"
    service_owner = "test-user"
    custom_config = "/etc/my-config.yaml"
    config_url = f"https://raw.githubusercontent.com/signalfx/splunk-otel-collector/v{collector_version}/cmd/otelcol/config/collector/gateway_config.yaml"

    install_cmd = " ".join((
        get_installer_cmd(),
        "--without-fluentd",
        "--memory 256 --ballast 64",
        f"--service-user {service_owner} --service-group {service_owner}",
        f"--collector-config {custom_config}",
        f"--collector-version {collector_version}",
    ))

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        run_container_cmd(container, f"wget -nv -O /etc/my-config.yaml {config_url}")
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
            time.sleep(5)

            # verify collector version
            _, output = run_container_cmd(container, "otelcol --version")
            assert output.decode("utf-8").strip() == f"otelcol version v{collector_version}"

            # verify env file created with configured parameters
            verify_env_file(container, config_path=custom_config, memory="256", ballast="64")

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=service_owner))

            # verify the default user/group was deleted
            assert container.exec_run(f"getent passwd {SERVICE_OWNER}").exit_code != 0
            assert container.exec_run(f"getent group {SERVICE_OWNER}").exit_code != 0

            # verify the installed directories are owned by test-user
            bundle_owner = container.exec_run("stat -c '%U:%G' /usr/lib/splunk-otel-collector").output.decode("utf-8")
            assert bundle_owner.strip() == f"{service_owner}:{service_owner}"
            config_owner = container.exec_run("stat -c '%U:%G' /etc/otel").output.decode("utf-8")
            assert config_owner.strip() == f"{service_owner}:{service_owner}"

            # verify td-agent was not installed
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
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_installer_with_instrumentation_default(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    install_cmd = get_installer_cmd()
    install_cmd = f"{install_cmd} --without-fluentd"
    install_cmd = f"{install_cmd} --with-instrumentation"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
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

            # verify default options
            run_container_cmd(container, f"grep '^disable_telemetry=false$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^generate_service_name=true$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep -v '^service_name=.*$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_profiler=false$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_profiler_memory=false$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_metrics=false$' {INSTR_CONF_PATH}")

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
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_installer_with_instrumentation_custom(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    install_cmd = " ".join((
        get_installer_cmd(),
        "--without-fluentd",
        "--with-instrumentation",
        "--deployment-environment test",
        "--disable-telemetry",
        "--service-name test",
        "--no-generate-service-name",
        "--enable-profiler",
        "--enable-profiler-memory",
        "--enable-metrics",
    ))

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
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

            # verify custom options
            run_container_cmd(container, f"grep '^disable_telemetry=true$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^generate_service_name=false$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^service_name=test$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_profiler=true$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_profiler_memory=true$' {INSTR_CONF_PATH}")
            run_container_cmd(container, f"grep '^enable_metrics=true$' {INSTR_CONF_PATH}")

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
