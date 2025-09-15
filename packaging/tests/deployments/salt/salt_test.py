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
import yaml

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
DEB_DOCKERFILE = IMAGES_DIR / "Dockerfile.deb"
RPM_DOCKERFILE = IMAGES_DIR / "Dockerfile.rpm"
DISTRO_YAML = IMAGES_DIR / "distro_docker_opts.yaml"
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
NODE_CONFIG_PATH = "/etc/splunk/zeroconfig/node.conf"
DOTNET_CONFIG_PATH = "/etc/splunk/zeroconfig/dotnet.conf"
NODE_PREFIX = "/usr/lib/splunk-instrumentation/splunk-otel-js"
NODE_OPTIONS = f"-r {NODE_PREFIX}/node_modules/@splunk/otel/instrument"
DOTNET_HOME = "/usr/lib/splunk-instrumentation/splunk-otel-dotnet"
DOTNET_AGENT_PATH = f"{DOTNET_HOME}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
DOTNET_VARS = {
    "CORECLR_ENABLE_PROFILING": "1",
    "CORECLR_PROFILER": "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
    "CORECLR_PROFILER_PATH": DOTNET_AGENT_PATH,
    "DOTNET_ADDITIONAL_DEPS": f"{DOTNET_HOME}/AdditionalDeps",
    "DOTNET_SHARED_STORE": f"{DOTNET_HOME}/store",
    "DOTNET_STARTUP_HOOKS": f"{DOTNET_HOME}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll",
    "OTEL_DOTNET_AUTO_HOME": DOTNET_HOME,
    "OTEL_DOTNET_AUTO_PLUGINS":
        "Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation",
}

PILLAR_PATH = "/srv/pillar/splunk-otel-collector.sls"
SALT_CMD = "salt-call --local state.apply"


def load_distro_opts(yaml_file):
    with open(yaml_file, 'r') as file:
        return yaml.safe_load(file)

# Load distro options from YAML file
DISTRO_OPTS = load_distro_opts(DISTRO_YAML)

# Extract DEB and RPM distributions
DEB_DISTROS = [distro for distro in DISTRO_OPTS.get('deb', {})]
RPM_DISTROS = [distro for distro in DISTRO_OPTS.get('rpm', {})]


def run_salt_apply(container, config):
    with tempfile.NamedTemporaryFile(mode="w+") as fd:
        print(config)
        fd.write(config)
        fd.flush()
        copy_file_into_container(container, fd.name, PILLAR_PATH)

    run_container_cmd(container, SALT_CMD)


def container_file_exists(container, path):
    return container.exec_run(f"test -f {path}").exit_code == 0


def verify_config_file(container, path, key, value=None, exists=True):
    if exists:
        assert container_file_exists(container, path), f"{path} does not exist"
    elif not container_file_exists(container, path):
        return True

    code, output = container.exec_run(f"cat {path}")
    config = output.decode("utf-8")
    assert code == 0, f"failed to get file content from {path}:\n{config}"

    line = key if value is None else f"{key}={value}"
    if path == SYSTEMD_CONFIG_PATH:
        line = f"DefaultEnvironment=\"{line}\""

    match = re.search(f"^{line}$", config, re.MULTILINE)

    if exists:
        assert match, f"'{line}' not found in {path}:\n{config}"
    else:
        assert not match, f"'{line}' found in {path}:\n{config}"


def verify_env_file(container, api_url=SPLUNK_API_URL, ingest_url=SPLUNK_INGEST_URL, hec_token=SPLUNK_ACCESS_TOKEN, listen_interface=None, command_line_args=None):
    if command_line_args:
        verify_config_file(container, SPLUNK_ENV_PATH, "OTELCOL_OPTIONS", command_line_args)
    else:
        verify_config_file(container, SPLUNK_ENV_PATH, "OTELCOL_OPTIONS=", None, exists=True)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_CONFIG", SPLUNK_CONFIG)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_ACCESS_TOKEN", SPLUNK_ACCESS_TOKEN)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_REALM", SPLUNK_REALM)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_API_URL", api_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_INGEST_URL", ingest_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_URL", f"{ingest_url}/v1/log")
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_TOKEN", hec_token)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_MEMORY_TOTAL_MIB", SPLUNK_MEMORY_TOTAL_MIB)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_BUNDLE_DIR", SPLUNK_BUNDLE_DIR)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_COLLECTD_DIR", SPLUNK_COLLECTD_DIR)
    if listen_interface:
        verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_LISTEN_INTERFACE", listen_interface)
    else:
        verify_config_file(container, SPLUNK_ENV_PATH, ".*SPLUNK_LISTEN_INTERFACE.*", exists=False)


def verify_dotnet_config(container, path, exists=True):
    for key, val in DOTNET_VARS.items():
        val = val if exists else ".*"
        verify_config_file(container, path, key, val, exists=exists)


def node_package_installed(container):
    cmd = "npm ls --global=false @splunk/otel"
    print(f"Running '{cmd}' in {NODE_PREFIX}:")
    rc, output = container.exec_run(cmd, workdir=NODE_PREFIX)
    print(output.decode("utf-8"))
    return rc == 0

def get_build_args(distro):
    if distro in DEB_DISTROS:
        build_args = DISTRO_OPTS.get('deb', {}).get(distro, [])
    else:
        build_args = DISTRO_OPTS.get('rpm', {}).get(distro, [])
    return {arg.split('=')[0]: arg.split('=')[1] for arg in build_args}

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
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=build_args) as container:
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
  collector_version: '0.126.0'
  splunk_service_user: 'test-user'
  splunk_service_group: 'test-user'
  splunk_listen_interface: '0.0.0.0'
  splunk_otel_collector_command_line_args: '--discovery --set=processors.batch.timeout=10s'
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
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=build_args) as container:
        try:
            run_salt_apply(container, CUSTOM_CONFIG)
            verify_env_file(
                container,
                api_url="https://fake-api.com",
                ingest_url="https://fake-ingest.com",
                hec_token="fake-hec-token",
                listen_interface="0.0.0.0",
                command_line_args="--discovery --set=processors.batch.timeout=10s"
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
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=build_args) as container:
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
        if version == "latest":
            assert node_package_installed(container)
        if with_systemd:
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_SERVICE_NAME", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_MEMORY_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_METRICS_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_ENDPOINT", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_PROTOCOL", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_METRICS_EXPORTER", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_LOGS_EXPORTER", ".*", exists=False)
            if version == "latest":
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH)
            else:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
        elif version == "latest":
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
            verify_config_file(container, JAVA_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, NODE_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            verify_dotnet_config(container, DOTNET_CONFIG_PATH)
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH]:
                verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
                verify_config_file(container, config_path, "OTEL_SERVICE_NAME", ".*", exists=False)
                verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "false")
                verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "false")
                verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "false")
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_PROTOCOL", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_METRICS_EXPORTER", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_LOGS_EXPORTER", ".*", exists=False)
        else:
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH, SYSTEMD_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
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
  auto_instrumentation_otlp_endpoint_protocol: 'grpc'
  auto_instrumentation_metrics_exporter: 'none'
  auto_instrumentation_logs_exporter: 'none'
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
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=build_args) as container:
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
        if version == "latest":
            assert node_package_installed(container)
        resource_attributes = f"{resource_attributes},deployment.environment=test"
        if with_systemd:
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_SERVICE_NAME", "test")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_MEMORY_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_METRICS_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://0.0.0.0:4317")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_METRICS_EXPORTER", "none")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_LOGS_EXPORTER", "none")
            if version == "latest":
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH)
            else:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
        elif version == "latest":
            for config_path in [SYSTEMD_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
            verify_config_file(container, JAVA_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, NODE_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            verify_dotnet_config(container, DOTNET_CONFIG_PATH)
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH]:
                verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
                verify_config_file(container, config_path, "OTEL_SERVICE_NAME", "test")
                verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "true")
                verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "true")
                verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "true")
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://0.0.0.0:4317")
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
                verify_config_file(container, config_path, "OTEL_METRICS_EXPORTER", "none")
                verify_config_file(container, config_path, "OTEL_LOGS_EXPORTER", "none")
        else:
            for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH, SYSTEMD_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
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
