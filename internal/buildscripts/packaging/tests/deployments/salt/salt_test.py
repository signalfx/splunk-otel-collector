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
import re
import string
import tempfile

from pathlib import Path

import pytest

from tests.helpers.util import (
    copy_file_into_container,
    run_container_cmd,
    run_distro_container,
    service_is_running,
    verify_package_version,
    wait_for,
    REPO_DIR,
    SERVICE_NAME,
    SERVICE_OWNER,
)


IMAGES_DIR = Path(__file__).parent.resolve() / "images"
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "rpm" / "Dockerfile.*"))]
CONFIG_DIR = "/etc/otel/collector"
SPLUNK_CONFIG = f"{CONFIG_DIR}/agent_config.yaml"
SPLUNK_ENV_PATH = f"{CONFIG_DIR}/splunk-otel-collector.conf"
SPLUNK_ACCESS_TOKEN = "testing123"
SPLUNK_REALM = "test"
SPLUNK_INGEST_URL = f"https://ingest.{SPLUNK_REALM}.signalfx.com"
SPLUNK_API_URL = f"https://api.{SPLUNK_REALM}.signalfx.com"
SPLUNK_SERVICE_USER = "splunk-otel-collector"
SPLUNK_SERVICE_GROUP = "splunk-otel-collector"
SPLUNK_MEMORY_TOTAL_MIB = 512
SPLUNK_BUNDLE_DIR = "/usr/lib/splunk-otel-collector/agent-bundle"
SPLUNK_COLLECTD_DIR = f"{SPLUNK_BUNDLE_DIR}/run/collectd"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
INSTRUMENTATION_CONFIG_PATH = "/usr/lib/splunk-instrumentation/instrumentation.conf"
SYSTEMD_CONFIG_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
JAVA_CONFIG_PATH = "/etc/splunk/zeroconfig/java.conf"

PILLAR_PATH = "/srv/pillar/splunk-otel-collector.sls"
SALT_CMD = "salt-call --local state.apply"


def run_salt_apply(container, config):
    with tempfile.NamedTemporaryFile(mode="w+") as fd:
        print(config)
        fd.write(config)
        fd.flush()
        copy_file_into_container(container, fd.name, PILLAR_PATH)

    run_container_cmd(container, SALT_CMD)


def container_file_exists(container, path):
    return container.exec_run(f"test -f {path}").exit_code == 0


def verify_config_file(container, path, key, value=None, exists=True, systemd=False):
    if exists:
        assert container_file_exists(container, path), f"{path} does not exist"
    elif not container_file_exists(container, path):
        return True

    code, output = container.exec_run(f"cat {path}")
    config = output.decode("utf-8")
    assert code == 0, f"failed to get file content from {path}:\n{config}"

    line = key if value is None else f"{key}={value}"
    if systemd:
        line = f"DefaultEnvironment=\"{line}\""

    match = re.search(f"^{line}$", config, re.MULTILINE)

    if exists:
        assert match, f"'{line}' not found in {path}:\n{config}"
    else:
        assert not match, f"'{line}' found in {path}:\n{config}"


def verify_env_file(container, api_url=SPLUNK_API_URL, ingest_url=SPLUNK_INGEST_URL, hec_token=SPLUNK_ACCESS_TOKEN, listen_interface=None):
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_CONFIG", SPLUNK_CONFIG)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_ACCESS_TOKEN", SPLUNK_ACCESS_TOKEN)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_REALM", SPLUNK_REALM)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_API_URL", api_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_INGEST_URL", ingest_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_TRACE_URL", f"{ingest_url}:443")
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_URL", f"{ingest_url}/v1/log")
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_TOKEN", hec_token)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_MEMORY_TOTAL_MIB", SPLUNK_MEMORY_TOTAL_MIB)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_BUNDLE_DIR", SPLUNK_BUNDLE_DIR)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_COLLECTD_DIR", SPLUNK_COLLECTD_DIR)
    if listen_interface:
        verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_LISTEN_INTERFACE", listen_interface)
    else:
        verify_config_file(container, SPLUNK_ENV_PATH, ".*SPLUNK_LISTEN_INTERFACE.*", exists=False)


DEFAULT_CONFIG = f"""
splunk-otel-collector:
  splunk_access_token: '{SPLUNK_ACCESS_TOKEN}'
  splunk_realm: '{SPLUNK_REALM}'
  """


@pytest.mark.salt
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
def test_salt_default(distro):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR) as container:
        try:
            run_salt_apply(container, DEFAULT_CONFIG)
            verify_env_file(container)
            assert wait_for(lambda: service_is_running(container))
            assert container.exec_run("systemctl status td-agent").exit_code != 0
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code != 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code != 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


CUSTOM_CONFIG = f"""
splunk-otel-collector:
  splunk_access_token: '{SPLUNK_ACCESS_TOKEN}'
  splunk_realm: '{SPLUNK_REALM}'
  splunk_ingest_url: 'https://fake-ingest.com'
  splunk_api_url: 'https://fake-api.com'
  splunk_hec_token: 'fake-hec-token'
  collector_version: '0.86.0'
  splunk_service_user: 'test-user'
  splunk_service_group: 'test-user'
  splunk_listen_interface: '0.0.0.0'
  collector_additional_env_vars:
    MY_CUSTOM_VAR1: value1
    MY_CUSTOM_VAR2: value2
  install_fluentd: True
  """


@pytest.mark.salt
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
def test_salt_custom(distro):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR) as container:
        try:
            run_salt_apply(container, CUSTOM_CONFIG)
            verify_env_file(
                container,
                api_url="https://fake-api.com",
                ingest_url="https://fake-ingest.com",
                hec_token="fake-hec-token",
                listen_interface="0.0.0.0"
            )
            verify_config_file(container, SPLUNK_ENV_PATH, "MY_CUSTOM_VAR1", "value1")
            verify_config_file(container, SPLUNK_ENV_PATH, "MY_CUSTOM_VAR2", "value2")
            assert wait_for(lambda: service_is_running(container, service_owner="test-user"))
            _, owner = run_container_cmd(container, f"stat -c '%U:%G' {SPLUNK_ENV_PATH}")
            assert owner.decode("utf-8").strip() == "test-user:test-user"
            if "opensuse" not in distro and distro != "amazonlinux-2023":
                assert container.exec_run("systemctl status td-agent").exit_code == 0
                _, owner = run_container_cmd(container, f"stat -c '%U:%G' {CONFIG_DIR}/fluentd/fluent.conf")
                assert owner.decode("utf-8").strip() == "td-agent:td-agent"
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code != 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code != 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


DEFAULT_INSTRUMENTATION_CONFIG = string.Template(f"""
splunk-otel-collector:
  splunk_access_token: '{SPLUNK_ACCESS_TOKEN}'
  splunk_realm: '{SPLUNK_REALM}'
  install_auto_instrumentation: True
  auto_instrumentation_version: '$version'
  auto_instrumentation_systemd: $systemd
  """
)


@pytest.mark.salt
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("version", ["0.86.0", "latest"])
@pytest.mark.parametrize("with_systemd", [True, False])
def test_salt_default_instrumentation(distro, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR) as container:
        config = DEFAULT_INSTRUMENTATION_CONFIG.substitute(version=version, systemd=str(with_systemd))
        run_salt_apply(container, config)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        assert container.exec_run("systemctl status td-agent").exit_code != 0
        resource_attributes = rf"splunk.zc.method=splunk-otel-auto-instrumentation-{version}"
        if with_systemd:
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        if version == "latest" or with_systemd:
            config_path = SYSTEMD_CONFIG_PATH if with_systemd else JAVA_CONFIG_PATH
            verify_config_file(container, config_path, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}", systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes, systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_SERVICE_NAME", ".*", exists=False, systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "false", systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "false", systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "false", systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://127.0.0.1:4317", systemd=with_systemd)
        else:
            config_path = INSTRUMENTATION_CONFIG_PATH
            verify_package_version(container, "splunk-otel-auto-instrumentation", version)
            verify_config_file(container, config_path, "java_agent_jar", JAVA_AGENT_PATH)
            verify_config_file(container, config_path, "resource_attributes", resource_attributes)
            verify_config_file(container, config_path, "service_name", ".*", exists=False)
            verify_config_file(container, config_path, "generate_service_name", "true")
            verify_config_file(container, config_path, "disable_telemetry", "false")
            verify_config_file(container, config_path, "enable_profiler", "false")
            verify_config_file(container, config_path, "enable_profiler_memory", "false")
            verify_config_file(container, config_path, "enable_metrics", "false")


CUSTOM_INSTRUMENTATION_CONFIG = string.Template(f"""
splunk-otel-collector:
  splunk_access_token: '{SPLUNK_ACCESS_TOKEN}'
  splunk_realm: '{SPLUNK_REALM}'
  collector_version: '$version'
  install_auto_instrumentation: True
  auto_instrumentation_version: '$version'
  auto_instrumentation_systemd: $systemd
  auto_instrumentation_ld_so_preload: '# my extra library'
  auto_instrumentation_resource_attributes: 'deployment.environment=test'
  auto_instrumentation_service_name: 'test'
  auto_instrumentation_generate_service_name: False
  auto_instrumentation_disable_telemetry: True
  auto_instrumentation_enable_profiler: True
  auto_instrumentation_enable_profiler_memory: True
  auto_instrumentation_enable_metrics: True
  auto_instrumentation_otlp_endpoint: 'http://0.0.0.0:4317'
  """
)


@pytest.mark.salt
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("version", ["0.86.0", "latest"])
@pytest.mark.parametrize("with_systemd", [True, False])
def test_salt_custom_instrumentation(distro, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR) as container:
        config = CUSTOM_INSTRUMENTATION_CONFIG.substitute(version=version, systemd=str(with_systemd))
        run_salt_apply(container, config)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        assert container.exec_run("systemctl status td-agent").exit_code != 0
        resource_attributes = rf"splunk.zc.method=splunk-otel-auto-instrumentation-{version}"
        if with_systemd:
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        resource_attributes = f"{resource_attributes},deployment.environment=test"
        if version == "latest" or with_systemd:
            config_path = SYSTEMD_CONFIG_PATH if with_systemd else JAVA_CONFIG_PATH
            verify_config_file(container, config_path, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}", systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes, systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_SERVICE_NAME", "test", systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "true", systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "true", systemd=with_systemd)
            verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "true", systemd=with_systemd)
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://0.0.0.0:4317", systemd=with_systemd)
        else:
            config_path = INSTRUMENTATION_CONFIG_PATH
            verify_package_version(container, "splunk-otel-auto-instrumentation", version)
            verify_config_file(container, config_path, "java_agent_jar", JAVA_AGENT_PATH)
            verify_config_file(container, config_path, "resource_attributes", resource_attributes)
            verify_config_file(container, config_path, "service_name", "test")
            verify_config_file(container, config_path, "generate_service_name", "false")
            verify_config_file(container, config_path, "disable_telemetry", "true")
            verify_config_file(container, config_path, "enable_profiler", "true")
            verify_config_file(container, config_path, "enable_profiler_memory", "true")
            verify_config_file(container, config_path, "enable_metrics", "true")
