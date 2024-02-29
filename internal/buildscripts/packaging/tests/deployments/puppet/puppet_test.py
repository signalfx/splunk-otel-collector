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
import re
import shutil
import string
import sys
import tempfile

from pathlib import Path

import psutil
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

if sys.platform == "win32":
    from tests.helpers.win_utils import has_choco, run_win_command, get_registry_value

IMAGES_DIR = Path(__file__).parent.resolve() / "images"
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "rpm" / "Dockerfile.*"))]
CONFIG_DIR = "/etc/otel/collector"
SPLUNK_ENV_PATH = f"{CONFIG_DIR}/splunk-otel-collector.conf"
SPLUNK_ACCESS_TOKEN = "testing123"
SPLUNK_REALM = "test"
SPLUNK_INGEST_URL = f"https://ingest.{SPLUNK_REALM}.signalfx.com"
SPLUNK_API_URL = f"https://api.{SPLUNK_REALM}.signalfx.com"
PUPPET_RELEASE = os.environ.get("PUPPET_RELEASE", "6,7").split(",")
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
INSTRUMENTATION_CONFIG_PATH = "/usr/lib/splunk-instrumentation/instrumentation.conf"
SYSTEMD_CONFIG_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
JAVA_CONFIG_PATH = "/etc/splunk/zeroconfig/java.conf"


def run_puppet_apply(container, config):
    with tempfile.NamedTemporaryFile(mode="w+") as fd:
        print(config)
        fd.write(config)
        fd.flush()
        copy_file_into_container(container, fd.name, "/root/test.pp")

    run_container_cmd(container, "puppet apply /root/test.pp")


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


def verify_env_file(container, api_url=SPLUNK_API_URL, ingest_url=SPLUNK_INGEST_URL, hec_token=SPLUNK_ACCESS_TOKEN):
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_ACCESS_TOKEN", SPLUNK_ACCESS_TOKEN)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_API_URL", api_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_TOKEN", hec_token)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_HEC_URL", f"{ingest_url}/v1/log")
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_INGEST_URL", ingest_url)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_REALM", SPLUNK_REALM)
    verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_TRACE_URL", f"{ingest_url}:443")


def skip_if_necessary(distro, puppet_release):
    if distro == "ubuntu-focal":
        pytest.skip("requires https://github.com/puppetlabs/puppetlabs-release/issues/271 to be resolved")


DEFAULT_CONFIG = f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
}}
"""


@pytest.mark.puppet
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
def test_puppet_default(distro, puppet_release):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            run_puppet_apply(container, DEFAULT_CONFIG)
            verify_env_file(container)
            verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_LISTEN_INTERFACE", ".*", exists=False)
            assert wait_for(lambda: service_is_running(container))
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


CUSTOM_VARS_CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    splunk_api_url => '$api_url',
    splunk_ingest_url => '$ingest_url',
    splunk_hec_token => 'fake-hec-token',
    splunk_listen_interface => '0.0.0.0',
    collector_version => '$version',
    with_fluentd => true,
    collector_additional_env_vars => {{ 'MY_CUSTOM_VAR1' => 'value1', 'MY_CUSTOM_VAR2' => 'value2' }},
}}
"""
)


@pytest.mark.puppet
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
def test_puppet_with_custom_vars(distro, puppet_release):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            api_url = "https://fake-splunk-api.com"
            ingest_url = "https://fake-splunk-ingest.com"
            config = CUSTOM_VARS_CONFIG.substitute(api_url=api_url, ingest_url=ingest_url, version="0.86.0")
            run_puppet_apply(container, config)
            verify_package_version(container, "splunk-otel-collector", "0.86.0")
            verify_env_file(container, api_url, ingest_url, "fake-hec-token")
            verify_config_file(container, SPLUNK_ENV_PATH, "SPLUNK_LISTEN_INTERFACE", "0.0.0.0")
            verify_config_file(container, SPLUNK_ENV_PATH, "MY_CUSTOM_VAR1", "value1")
            verify_config_file(container, SPLUNK_ENV_PATH, "MY_CUSTOM_VAR2", "value2")
            assert wait_for(lambda: service_is_running(container))
            if "opensuse" not in distro and distro != "amazonlinux-2023":
                assert container.exec_run("systemctl status td-agent").exit_code == 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            if "opensuse" not in distro and distro != "amazonlinux-2023":
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")


DEFAULT_INSTRUMENTATION_CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    with_auto_instrumentation => true,
    auto_instrumentation_version => '$version',
    auto_instrumentation_systemd => $with_systemd,
}}
"""
)


@pytest.mark.puppet
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
@pytest.mark.parametrize("version", ["0.86.0", "latest"])
@pytest.mark.parametrize("with_systemd", ["true", "false"])
def test_puppet_with_default_instrumentation(distro, puppet_release, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        config = DEFAULT_INSTRUMENTATION_CONFIG.substitute(version=version, with_systemd=with_systemd)
        run_puppet_apply(container, config)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        assert container.exec_run("systemctl status td-agent").exit_code != 0
        resource_attributes = r"splunk.zc.method=splunk-otel-auto-instrumentation-.*"
        if with_systemd == "true":
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        if version == "latest" or with_systemd == "true":
            config_path = SYSTEMD_CONFIG_PATH if with_systemd == "true" else JAVA_CONFIG_PATH
            verify_config_file(container, config_path, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
            verify_config_file(container, config_path, "OTEL_SERVICE_NAME", ".*", exists=False)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "false")
            verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "false")
            verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "false")
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://127.0.0.1:4317")
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


CUSTOM_INSTRUMENTATION_CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    with_auto_instrumentation => true,
    auto_instrumentation_version => '$version',
    auto_instrumentation_systemd => $with_systemd,
    auto_instrumentation_ld_so_preload => '# my extra library',
    auto_instrumentation_resource_attributes => 'deployment.environment=test',
    auto_instrumentation_generate_service_name => false,
    auto_instrumentation_disable_telemetry => true,
    auto_instrumentation_service_name => 'test',
    auto_instrumentation_enable_profiler => true,
    auto_instrumentation_enable_profiler_memory => true,
    auto_instrumentation_enable_metrics => true,
    auto_instrumentation_otlp_endpoint => 'http://0.0.0.0:4317',
}}
"""
)


@pytest.mark.puppet
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
@pytest.mark.parametrize("version", ["0.86.0", "latest"])
@pytest.mark.parametrize("with_systemd", ["true", "false"])
def test_puppet_with_custom_instrumentation(distro, puppet_release, version, with_systemd):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        config = CUSTOM_INSTRUMENTATION_CONFIG.substitute(version=version, with_systemd=with_systemd)
        run_puppet_apply(container, config)
        verify_env_file(container)
        assert wait_for(lambda: service_is_running(container))
        assert container.exec_run("systemctl status td-agent").exit_code != 0
        resource_attributes = r"splunk.zc.method=splunk-otel-auto-instrumentation-.*"
        if with_systemd == "true":
            resource_attributes = rf"{resource_attributes}-systemd"
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH, exists=False)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
        else:
            verify_config_file(container, "/etc/ld.so.preload", LIBSPLUNK_PATH)
            verify_config_file(container, "/etc/ld.so.preload", r"# my extra library")
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
        resource_attributes = rf"{resource_attributes},deployment.environment=test"
        if version == "latest" or with_systemd == "true":
            config_path = SYSTEMD_CONFIG_PATH if with_systemd == "true" else JAVA_CONFIG_PATH
            verify_config_file(container, config_path, "JAVA_TOOL_OPTIONS", rf"-javaagent:{JAVA_AGENT_PATH}")
            verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", resource_attributes)
            verify_config_file(container, config_path, "OTEL_SERVICE_NAME", "test")
            verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "true")
            verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "true")
            verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "true")
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", r"http://0.0.0.0:4317")
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


WIN_PUPPET_RELEASE = os.environ.get("PUPPET_RELEASE", "latest")
WIN_PUPPET_BIN_DIR = r"C:\Program Files\Puppet Labs\Puppet\bin"
WIN_PUPPET_MODULE_SRC_DIR = os.path.join(REPO_DIR, "deployments", "puppet")
WIN_PUPPET_MODULE_DEST_DIR = r"C:\ProgramData\PuppetLabs\code\environments\production\modules\splunk_otel_collector"
WIN_INSTALL_DIR = r"C:\Program Files\Splunk\OpenTelemetry Collector"
WIN_CONFIG_PATH = r"C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml"


def run_win_puppet_setup(puppet_release):
    assert has_choco(), "choco not installed!"
    if puppet_release == "latest":
        run_win_command(f"choco upgrade -y -f puppet-agent")
    else:
        run_win_command(f"choco upgrade -y -f puppet-agent --version {puppet_release}")
    if WIN_PUPPET_BIN_DIR not in os.environ.get("PATH"):
        os.environ["PATH"] = WIN_PUPPET_BIN_DIR + ";" + os.environ.get("PATH")
    if os.path.isdir(WIN_PUPPET_MODULE_DEST_DIR):
        shutil.rmtree(WIN_PUPPET_MODULE_DEST_DIR)
    shutil.copytree(WIN_PUPPET_MODULE_SRC_DIR, WIN_PUPPET_MODULE_DEST_DIR)
    run_win_command("puppet module install puppet-archive")
    run_win_command("puppet module install puppetlabs-powershell")
    run_win_command("puppet module install puppetlabs-registry")


def run_win_puppet_agent(config):
    with tempfile.TemporaryDirectory() as tmpdir:
        manifest_path = os.path.join(tmpdir, "agent.pp")
        print(config)
        with open(manifest_path, "w+", encoding="utf-8") as fd:
            fd.write(config)
        cmd = f"puppet apply {manifest_path}"
        run_win_command(cmd, returncodes=[0, 2])



@pytest.mark.windows
@pytest.mark.skipif(sys.platform != "win32", reason="only runs on windows")
def test_win_puppet_default():
    run_win_puppet_setup(WIN_PUPPET_RELEASE)

    config = f"""
    class {{ splunk_otel_collector:
        splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
        splunk_realm => '{SPLUNK_REALM}',
        collector_version => '0.86.0',
    }}
    """
    run_win_puppet_agent(config)

    assert get_registry_value("SPLUNK_REALM") == SPLUNK_REALM
    assert get_registry_value("SPLUNK_ACCESS_TOKEN") == SPLUNK_ACCESS_TOKEN
    assert get_registry_value("SPLUNK_API_URL") == SPLUNK_API_URL
    assert get_registry_value("SPLUNK_INGEST_URL") == SPLUNK_INGEST_URL
    assert get_registry_value("SPLUNK_HEC_URL") == f"{SPLUNK_INGEST_URL}/v1/log"
    assert get_registry_value("SPLUNK_TRACE_URL") == f"{SPLUNK_INGEST_URL}:443"
    assert get_registry_value("SPLUNK_HEC_TOKEN") == SPLUNK_ACCESS_TOKEN
    try:
        listen_interface = get_registry_value("SPLUNK_LISTEN_INTERFACE")
    except FileNotFoundError:
        listen_interface = None
    assert listen_interface is None

    assert psutil.win_service_get("splunk-otel-collector").status() == psutil.STATUS_RUNNING
    for service in psutil.win_service_iter():
        assert service.name() != "fluentdwinsvc"


@pytest.mark.windows
@pytest.mark.skipif(sys.platform != "win32", reason="only runs on windows")
def test_win_puppet_custom_vars():
    run_win_puppet_setup(WIN_PUPPET_RELEASE)

    api_url = "https://fake-splunk-api.com"
    ingest_url = "https://fake-splunk-ingest.com"
    config = CUSTOM_VARS_CONFIG.substitute(api_url=api_url, ingest_url=ingest_url, version="0.48.0")

    run_win_puppet_agent(config)

    assert get_registry_value("SPLUNK_REALM") == SPLUNK_REALM
    assert get_registry_value("SPLUNK_ACCESS_TOKEN") == SPLUNK_ACCESS_TOKEN
    assert get_registry_value("SPLUNK_API_URL") == api_url
    assert get_registry_value("SPLUNK_INGEST_URL") == ingest_url
    assert get_registry_value("SPLUNK_HEC_URL") == f"{ingest_url}/v1/log"
    assert get_registry_value("SPLUNK_LISTEN_INTERFACE") == "0.0.0.0"
    assert get_registry_value("SPLUNK_TRACE_URL") == f"{ingest_url}:443"
    assert get_registry_value("SPLUNK_HEC_TOKEN") == "fake-hec-token"
    assert get_registry_value("MY_CUSTOM_VAR1") == "value1"
    assert get_registry_value("MY_CUSTOM_VAR2") == "value2"

    assert psutil.win_service_get("splunk-otel-collector").status() == psutil.STATUS_RUNNING
    assert psutil.win_service_get("fluentdwinsvc").status() == psutil.STATUS_RUNNING
