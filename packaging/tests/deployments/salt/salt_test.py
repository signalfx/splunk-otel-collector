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
import re
import shlex
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
PKG_DIR = REPO_DIR / "dist"
LOCAL_ARTIFACTS_DIR = "/opt/splunk-otel-local-artifacts"
COLLECTOR_PKG_NAME = "splunk-otel-collector"
AUTO_INSTRUMENTATION_PKG_NAME = "splunk-otel-auto-instrumentation"
CONFIG_DIR = "/etc/otel/collector"
SPLUNK_CONFIG = f"{CONFIG_DIR}/agent_config.yaml"
SPLUNK_ENV_PATH = f"{CONFIG_DIR}/splunk-otel-collector.conf"
SPLUNK_ACCESS_TOKEN = "testing123"
SPLUNK_REALM = "test"
SPLUNK_INGEST_URL = f"https://ingest.{SPLUNK_REALM}.observability.splunkcloud.com"
SPLUNK_API_URL = f"https://api.{SPLUNK_REALM}.observability.splunkcloud.com"
SPLUNK_SERVICE_USER = "splunk-otel-collector"
SPLUNK_SERVICE_GROUP = "splunk-otel-collector"
SPLUNK_MEMORY_TOTAL_MIB = 512
LOCAL_ARTIFACT_TESTING_ENABLED = os.environ.get("LOCAL_ARTIFACT_TESTING_ENABLED", "false").lower() == "true"
COLLECTOR_VERSION = os.environ.get("VERSION", "latest")
AUTO_INSTRUMENTATION_VERSION = os.environ.get("AUTO_INSTRUMENTATION_VERSION", "latest")
INSTRUMENTATION_VERSIONS = (
    [AUTO_INSTRUMENTATION_VERSION]
    if LOCAL_ARTIFACT_TESTING_ENABLED
    else ["0.86.0", "0.158.0", "latest"]
)
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
LIBOTELINJECT_PATH = "/usr/lib/splunk-instrumentation/libotelinject.so"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
INSTRUMENTATION_CONFIG_PATH = "/usr/lib/splunk-instrumentation/instrumentation.conf"
SYSTEMD_CONFIG_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
INJECTOR_CONFIG_PATH = "/etc/opentelemetry/injector/injector.conf"
INJECTOR_DEFAULT_ENV_PATH = "/etc/opentelemetry/injector/default_env.conf"
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
    if listen_interface:
        verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_LISTEN_INTERFACE", listen_interface)
    else:
        verify_config_file(container, SPLUNK_ENV_PATH, ".*SPLUNK_LISTEN_INTERFACE.*", exists=False)


def verify_dotnet_config(container, path, exists=True):
    for key, val in DOTNET_VARS.items():
        val = val if exists else ".*"
        verify_config_file(container, path, key, val, exists=exists)


def verify_otel_injector_config(
    container,
    with_systemd,
    resource_attributes,
    service_name=None,
    profiler_enabled=False,
    profiler_memory_enabled=False,
    metrics_enabled=False,
    otlp_endpoint=None,
    otlp_protocol=None,
    metrics_exporter=None,
    logs_exporter=None,
):
    for config_path in [JAVA_CONFIG_PATH, NODE_CONFIG_PATH, DOTNET_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
        assert not container_file_exists(container, config_path)

    verify_config_file(container, INJECTOR_CONFIG_PATH, "jvm_auto_instrumentation_agent_path", JAVA_AGENT_PATH)
    verify_config_file(
        container,
        INJECTOR_CONFIG_PATH,
        "nodejs_auto_instrumentation_agent_path",
        f"{NODE_PREFIX}/node_modules/@splunk/otel/instrument.js",
    )
    verify_config_file(container, INJECTOR_CONFIG_PATH, "dotnet_auto_instrumentation_agent_path_prefix", DOTNET_HOME)
    verify_config_file(container, INJECTOR_CONFIG_PATH, "auto_instrumentation_disabled=.*", exists=False)

    if with_systemd:
        verify_config_file(container, "/etc/ld.so.preload", LIBOTELINJECT_PATH, exists=False)
        verify_config_file(container, SYSTEMD_CONFIG_PATH, "LD_PRELOAD", LIBOTELINJECT_PATH)
    else:
        verify_config_file(container, "/etc/ld.so.preload", LIBOTELINJECT_PATH)
        assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)

    verify_config_file(
        container,
        INJECTOR_DEFAULT_ENV_PATH,
        "OTEL_DOTNET_AUTO_PLUGINS",
        DOTNET_VARS["OTEL_DOTNET_AUTO_PLUGINS"],
    )
    verify_config_file(container, INJECTOR_DEFAULT_ENV_PATH, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
    verify_config_file(container, INJECTOR_DEFAULT_ENV_PATH, "OTEL_SERVICE_NAME", service_name or ".*", exists=bool(service_name))
    verify_config_file(container, INJECTOR_DEFAULT_ENV_PATH, "SPLUNK_PROFILER_ENABLED", str(profiler_enabled).lower())
    verify_config_file(
        container,
        INJECTOR_DEFAULT_ENV_PATH,
        "SPLUNK_PROFILER_MEMORY_ENABLED",
        str(profiler_memory_enabled).lower(),
    )
    verify_config_file(container, INJECTOR_DEFAULT_ENV_PATH, "SPLUNK_METRICS_ENABLED", str(metrics_enabled).lower())
    for key, value in {
        "OTEL_EXPORTER_OTLP_ENDPOINT": otlp_endpoint,
        "OTEL_EXPORTER_OTLP_PROTOCOL": otlp_protocol,
        "OTEL_METRICS_EXPORTER": metrics_exporter,
        "OTEL_LOGS_EXPORTER": logs_exporter,
    }.items():
        verify_config_file(container, INJECTOR_DEFAULT_ENV_PATH, key, value or ".*", exists=bool(value))


def node_package_installed(container):
    cmd = "npm ls --global=false @splunk/otel"
    print(f"Running '{cmd}' in {NODE_PREFIX}:")
    rc, output = container.exec_run(cmd, workdir=NODE_PREFIX)
    print(output.decode("utf-8"))
    return rc == 0


def package_version_at_least(version, minimum):
    if version == "latest":
        return True

    version_match = re.match(r"^(\d+)\.(\d+)\.(\d+)", version)
    minimum_match = re.match(r"^(\d+)\.(\d+)\.(\d+)", minimum)
    assert version_match is not None, f"unexpected package version: {version}"
    assert minimum_match is not None, f"unexpected minimum package version: {minimum}"
    return tuple(map(int, version_match.groups())) >= tuple(map(int, minimum_match.groups()))


def package_uses_otel_injector(version):
    if LOCAL_ARTIFACT_TESTING_ENABLED or version == "latest":
        return True

    match = re.match(r"^(\d+)\.(\d+)\.(\d+)(.*)$", version)
    assert match is not None, f"unexpected package version: {version}"
    release = tuple(map(int, match.groups()[:3]))
    return release > (0, 158, 0) or (release == (0, 158, 0) and bool(match.group(4)))


def find_local_package(package_name, distro):
    if distro in DEB_DISTROS:
        pattern = f"{package_name}_*amd64.deb"
    else:
        pattern = f"{package_name}-*x86_64.rpm"

    matches = sorted(PKG_DIR.glob(pattern))
    assert len(matches) == 1, f"expected exactly one {pattern} package in {PKG_DIR}, found {matches}"
    return matches[0]


def local_artifact_volumes():
    if not LOCAL_ARTIFACT_TESTING_ENABLED:
        return None
    return {str(PKG_DIR): {"bind": LOCAL_ARTIFACTS_DIR, "mode": "ro"}}


def get_local_package_source(package_name, distro):
    package_path = find_local_package(package_name, distro)
    return f"{LOCAL_ARTIFACTS_DIR}/{package_path.name}"


def verify_local_package_source(container, package_source):
    run_container_cmd(container, f"ls -l {shlex.quote(LOCAL_ARTIFACTS_DIR)}", exit_code=None)
    run_container_cmd(container, f"test -f {shlex.quote(package_source)}")


def with_local_artifacts(container, distro, config, install_auto_instrumentation=False):
    if not LOCAL_ARTIFACT_TESTING_ENABLED:
        return config

    collector_package_source = get_local_package_source(COLLECTOR_PKG_NAME, distro)
    verify_local_package_source(container, collector_package_source)
    local_config = f"""
  local_artifact_testing_enabled: True
  collector_package_source: '{collector_package_source}'
"""
    if install_auto_instrumentation:
        auto_instrumentation_package_source = get_local_package_source(AUTO_INSTRUMENTATION_PKG_NAME, distro)
        verify_local_package_source(container, auto_instrumentation_package_source)
        local_config += f"  auto_instrumentation_package_source: '{auto_instrumentation_package_source}'\n"

    return f"{config}\n{local_config}"


def verify_collector_version(container):
    if LOCAL_ARTIFACT_TESTING_ENABLED:
        verify_package_version(container, COLLECTOR_PKG_NAME, COLLECTOR_VERSION)


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
    with run_distro_container(
        distro,
        dockerfile=dockerfile,
        path=REPO_DIR,
        buildargs=build_args,
        extra_volumes=local_artifact_volumes(),
    ) as container:
        try:
            config = with_local_artifacts(container, distro, DEFAULT_CONFIG)
            run_salt_apply(container, config)
            verify_collector_version(container)
            verify_env_file(container)
            assert wait_for(lambda: service_is_running(container))
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
    SPLUNK_OPAMP_SUPERVISOR_ENABLED: 'true'
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

    with run_distro_container(
        distro,
        dockerfile=dockerfile,
        path=REPO_DIR,
        buildargs=build_args,
        extra_volumes=local_artifact_volumes(),
    ) as container:
        try:
            config = with_local_artifacts(container, distro, CUSTOM_CONFIG)
            run_salt_apply(container, config)
            verify_collector_version(container)
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
            verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_OPAMP_SUPERVISOR_ENABLED", "true")
            assert wait_for(lambda: service_is_running(container, service_owner="test-user"))
            assert wait_for(lambda: service_is_running(
                container,
                service_owner="test-user",
                process="opampsupervisor",
            ))
            _, owner = run_container_cmd(container, f"stat -c '%U:%G' {SPLUNK_ENV_PATH}")
            assert owner.decode("utf-8").strip() == "test-user:test-user"
            for path in [CONFIG_DIR, "/var/lib/otelcol"]:
                _, owner = run_container_cmd(container, f"stat -c '%U:%G:%a' {path}")
                assert owner.decode("utf-8").strip() == "test-user:test-user:755"
            for path in [
                SPLUNK_CONFIG,
                f"{CONFIG_DIR}/supervisor",
                f"{CONFIG_DIR}/supervisor/supervisor_config.yaml",
                "/var/lib/otelcol/supervisor",
            ]:
                _, owner = run_container_cmd(container, f"stat -c '%U:%G' {path}")
                assert owner.decode("utf-8").strip() == "test-user:test-user"
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
@pytest.mark.parametrize("version", INSTRUMENTATION_VERSIONS)
@pytest.mark.parametrize("with_systemd", [True, False])
def test_salt_default_instrumentation(distro, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(
        distro,
        dockerfile=dockerfile,
        path=REPO_DIR,
        buildargs=build_args,
        extra_volumes=local_artifact_volumes(),
    ) as container:
        config = DEFAULT_INSTRUMENTATION_CONFIG.substitute(version=version, systemd=str(with_systemd))
        config = with_local_artifacts(container, distro, config, install_auto_instrumentation=True)
        run_salt_apply(container, config)
        verify_collector_version(container)
        verify_package_version(container, AUTO_INSTRUMENTATION_PKG_NAME, version)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        with_new_instrumentation = package_version_at_least(version, "0.87.0")
        with_dotnet_instrumentation = package_version_at_least(version, "0.99.0")
        resource_attributes = rf"splunk.zc.method=splunk-otel-auto-instrumentation-{version}"
        if package_uses_otel_injector(version):
            if with_systemd:
                resource_attributes = rf"{resource_attributes}-systemd"
            assert node_package_installed(container)
            verify_otel_injector_config(container, with_systemd, resource_attributes)
            return

        if with_systemd:
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        if with_new_instrumentation:
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
            if with_new_instrumentation:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=with_dotnet_instrumentation)
            else:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
        elif with_new_instrumentation:
            for config_path in [SYSTEMD_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
            verify_config_file(container, JAVA_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, NODE_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            if with_dotnet_instrumentation:
                verify_dotnet_config(container, DOTNET_CONFIG_PATH)
            else:
                assert not container_file_exists(container, DOTNET_CONFIG_PATH)
            config_paths = [JAVA_CONFIG_PATH, NODE_CONFIG_PATH]
            if with_dotnet_instrumentation:
                config_paths.append(DOTNET_CONFIG_PATH)
            for config_path in config_paths:
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
  collector_version: '$collector_version'
  install_auto_instrumentation: True
  auto_instrumentation_version: '$version'
  auto_instrumentation_systemd: $systemd
  auto_instrumentation_ld_so_preload: '# my extra library'
  auto_instrumentation_resource_attributes: 'deployment.environment.name=test'
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
@pytest.mark.parametrize("version", INSTRUMENTATION_VERSIONS)
@pytest.mark.parametrize("with_systemd", [True, False])
def test_salt_custom_instrumentation(distro, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(
        distro,
        dockerfile=dockerfile,
        path=REPO_DIR,
        buildargs=build_args,
        extra_volumes=local_artifact_volumes(),
    ) as container:
        collector_version = COLLECTOR_VERSION if LOCAL_ARTIFACT_TESTING_ENABLED else version
        config = CUSTOM_INSTRUMENTATION_CONFIG.substitute(
            collector_version=collector_version,
            version=version,
            systemd=str(with_systemd),
        )
        config = with_local_artifacts(container, distro, config, install_auto_instrumentation=True)
        run_salt_apply(container, config)
        verify_collector_version(container)
        verify_package_version(container, AUTO_INSTRUMENTATION_PKG_NAME, version)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        with_new_instrumentation = package_version_at_least(version, "0.87.0")
        with_dotnet_instrumentation = package_version_at_least(version, "0.99.0")
        resource_attributes = rf"splunk.zc.method=splunk-otel-auto-instrumentation-{version}"
        if package_uses_otel_injector(version):
            if with_systemd:
                resource_attributes = rf"{resource_attributes}-systemd"
            resource_attributes = f"{resource_attributes},deployment.environment.name=test"
            assert node_package_installed(container)
            verify_otel_injector_config(
                container,
                with_systemd,
                resource_attributes,
                service_name="test",
                profiler_enabled=True,
                profiler_memory_enabled=True,
                metrics_enabled=True,
                otlp_endpoint=r"http://0.0.0.0:4317",
                otlp_protocol="grpc",
                metrics_exporter="none",
                logs_exporter="none",
            )
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
            return

        if with_systemd:
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        if with_new_instrumentation:
            assert node_package_installed(container)
        resource_attributes = f"{resource_attributes},deployment.environment.name=test"
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
            if with_new_instrumentation:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=with_dotnet_instrumentation)
            else:
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
        elif with_new_instrumentation:
            for config_path in [SYSTEMD_CONFIG_PATH, INSTRUMENTATION_CONFIG_PATH]:
                assert not container_file_exists(container, config_path)
            verify_config_file(container, JAVA_CONFIG_PATH, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, NODE_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            if with_dotnet_instrumentation:
                verify_dotnet_config(container, DOTNET_CONFIG_PATH)
            else:
                assert not container_file_exists(container, DOTNET_CONFIG_PATH)
            config_paths = [JAVA_CONFIG_PATH, NODE_CONFIG_PATH]
            if with_dotnet_instrumentation:
                config_paths.append(DOTNET_CONFIG_PATH)
            for config_path in config_paths:
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
            verify_config_file(container, config_path, "java_agent_jar", JAVA_AGENT_PATH)
            verify_config_file(container, config_path, "resource_attributes", resource_attributes)
            verify_config_file(container, config_path, "service_name", "test")
            verify_config_file(container, config_path, "generate_service_name", "false")
            verify_config_file(container, config_path, "disable_telemetry", "true")
            verify_config_file(container, config_path, "enable_profiler", "true")
            verify_config_file(container, config_path, "enable_profiler_memory", "true")
            verify_config_file(container, config_path, "enable_metrics", "true")


@pytest.mark.salt
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("with_systemd", [True, False])
def test_salt_instrumentation_upgrade_from_libsplunk(distro, with_systemd):
    # Simulates a fleet-wide upgrade: a host previously provisioned by salt with the legacy,
    # libsplunk.so-based auto-instrumentation is re-converged with a version past the otel
    # injector threshold, and should end up fully migrated to libotelinject.so.
    if LOCAL_ARTIFACT_TESTING_ENABLED:
        pytest.skip("local artifact testing only builds the current, otel-injector-based package")

    legacy_version = "0.158.0"
    assert not package_uses_otel_injector(legacy_version), \
        f"test setup error: {legacy_version} is expected to predate the otel injector"
    new_version = AUTO_INSTRUMENTATION_VERSION
    assert package_uses_otel_injector(new_version), (
        f"AUTO_INSTRUMENTATION_VERSION={new_version} does not use the otel injector; set it to "
        "'latest' or a version greater than 0.158.0 to exercise the upgrade path"
    )

    if distro in DEB_DISTROS:
        dockerfile = DEB_DOCKERFILE
        build_args = get_build_args(distro)
    else:
        dockerfile = RPM_DOCKERFILE
        build_args = get_build_args(distro)

    with run_distro_container(
        distro,
        dockerfile=dockerfile,
        path=REPO_DIR,
        buildargs=build_args,
    ) as container:
        legacy_config = DEFAULT_INSTRUMENTATION_CONFIG.substitute(version=legacy_version, systemd=str(with_systemd))
        run_salt_apply(container, legacy_config)
        verify_package_version(container, AUTO_INSTRUMENTATION_PKG_NAME, legacy_version)
        for config_path in (JAVA_CONFIG_PATH, NODE_CONFIG_PATH):
            assert container_file_exists(container, config_path), f"{config_path} missing after legacy install"
        if with_systemd:
            assert container_file_exists(container, SYSTEMD_CONFIG_PATH)
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)

        # Re-apply the same pillar with the version bumped past the otel injector threshold, as a
        # mass-deployment tool would when rolling out an upgrade fleet-wide.
        new_config = DEFAULT_INSTRUMENTATION_CONFIG.substitute(version=new_version, systemd=str(with_systemd))
        run_salt_apply(container, new_config)
        verify_collector_version(container)
        verify_package_version(container, AUTO_INSTRUMENTATION_PKG_NAME, new_version)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        assert node_package_installed(container)
        assert not container_file_exists(container, LIBSPLUNK_PATH), "libsplunk.so was not removed by the upgrade"

        resource_attributes = rf"splunk.zc.method=splunk-otel-auto-instrumentation-{new_version}"
        if with_systemd:
            resource_attributes = rf"{resource_attributes}-systemd"
        verify_otel_injector_config(container, with_systemd, resource_attributes)
        if not with_systemd:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
