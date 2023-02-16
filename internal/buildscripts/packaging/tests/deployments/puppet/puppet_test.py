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

CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    collector_version => '$collector_version',
    with_fluentd => $with_fluentd,
}}
"""
)


def run_puppet_apply(container, config):
    with tempfile.NamedTemporaryFile(mode="w+") as fd:
        print(config)
        fd.write(config)
        fd.flush()
        copy_file_into_container(container, fd.name, "/root/test.pp")

    run_container_cmd(container, "puppet apply /root/test.pp")


def verify_env_file(container):
    run_container_cmd(container, f"grep '^SPLUNK_ACCESS_TOKEN={SPLUNK_ACCESS_TOKEN}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_API_URL={SPLUNK_API_URL}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_TOKEN={SPLUNK_ACCESS_TOKEN}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_URL={SPLUNK_INGEST_URL}/v1/log$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_INGEST_URL={SPLUNK_INGEST_URL}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_REALM={SPLUNK_REALM}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_TRACE_URL={SPLUNK_INGEST_URL}/v2/trace$' {SPLUNK_ENV_PATH}")


def skip_if_necessary(distro, puppet_release):
    if distro == "ubuntu-focal":
        pytest.skip("requires https://github.com/puppetlabs/puppetlabs-release/issues/271 to be resolved")


@pytest.mark.puppet
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
def test_puppet_with_fluentd(distro, puppet_release):
    skip_if_necessary(distro, puppet_release)
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            for collector_version in ["0.34.0", "latest"]:
                config = CONFIG.substitute(collector_version=collector_version, with_fluentd="true")
                run_puppet_apply(container, config)
                verify_env_file(container)
                assert wait_for(lambda: service_is_running(container))
                if "opensuse" not in distro:
                    assert container.exec_run("systemctl status td-agent").exit_code == 0
                if collector_version == "latest":
                    verify_package_version(container, "splunk-otel-collector", collector_version, "0.34.0")
                else:
                    verify_package_version(container, "splunk-otel-collector", collector_version)
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            if "opensuse" not in distro:
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")


@pytest.mark.puppet
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
def test_puppet_without_fluentd(distro, puppet_release):
    skip_if_necessary(distro, puppet_release)
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            config = CONFIG.substitute(collector_version="latest", with_fluentd="false")
            run_puppet_apply(container, config)
            verify_env_file(container)
            assert wait_for(lambda: service_is_running(container))
            assert container.exec_run("systemctl status td-agent").exit_code != 0
            if distro in DEB_DISTROS:
                assert container.exec_run("dpkg -s splunk-otel-auto-instrumentation").exit_code != 0
            else:
                assert container.exec_run("rpm -q splunk-otel-auto-instrumentation").exit_code != 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


INSTRUMENTATION_CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    collector_version => '$version',
    with_auto_instrumentation => true,
    auto_instrumentation_version => '$version',
    auto_instrumentation_resource_attributes => 'deployment.environment=test',
    auto_instrumentation_service_name => 'test',
}}
"""
)


def verify_instrumentation_config(container):
    config_path = "/usr/lib/splunk-instrumentation/instrumentation.conf"
    libsplunk_path = "/usr/lib/splunk-instrumentation/libsplunk.so"
    java_agent_path = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"

    try:
        run_container_cmd(container, f"grep '^{libsplunk_path}' /etc/ld.so.preload")
        run_container_cmd(container, f"grep '^java_agent_jar={java_agent_path}' {config_path}")
        run_container_cmd(container, f"grep '^resource_attributes=deployment.environment=test' {config_path}")
        run_container_cmd(container, f"grep '^service_name=test' {config_path}")
    finally:
        run_container_cmd(container, "cat /etc/ld.so.preload")
        run_container_cmd(container, f"cat {config_path}")


@pytest.mark.puppet
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("puppet_release", PUPPET_RELEASE)
def test_puppet_with_instrumentation(distro, puppet_release):
    skip_if_necessary(distro, puppet_release)
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            for version in ["0.48.0", "latest"]:
                config = INSTRUMENTATION_CONFIG.substitute(version=version)
                run_puppet_apply(container, config)
                verify_env_file(container)
                assert wait_for(lambda: service_is_running(container))
                if "opensuse" not in distro:
                    assert container.exec_run("systemctl status td-agent").exit_code == 0
                if version == "latest":
                    verify_package_version(container, "splunk-otel-auto-instrumentation", version, "0.48.0")
                else:
                    verify_package_version(container, "splunk-otel-auto-instrumentation", version)
                verify_instrumentation_config(container)
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            if "opensuse" not in distro:
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")


CUSTOM_VARS_CONFIG = string.Template(
    f"""
class {{ splunk_otel_collector:
    splunk_access_token => '{SPLUNK_ACCESS_TOKEN}',
    splunk_realm => '{SPLUNK_REALM}',
    splunk_api_url => '$api_url',
    splunk_ingest_url => '$ingest_url',
    splunk_hec_token => 'fake-hec-token',
    collector_version => '$version',
    with_fluentd => false,
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
    skip_if_necessary(distro, puppet_release)
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    api_url = "https://fake-splunk-api.com"
    ingest_url = "https://fake-splunk-ingest.com"
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            config = CUSTOM_VARS_CONFIG.substitute(api_url=api_url, ingest_url=ingest_url, version="latest")
            run_puppet_apply(container, config)
            run_container_cmd(container, f"grep '^SPLUNK_ACCESS_TOKEN={SPLUNK_ACCESS_TOKEN}$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_API_URL={api_url}$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_HEC_TOKEN=fake-hec-token$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_HEC_URL={ingest_url}/v1/log$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_INGEST_URL={ingest_url}$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_REALM={SPLUNK_REALM}$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^SPLUNK_TRACE_URL={ingest_url}/v2/trace$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^MY_CUSTOM_VAR1=value1$' {SPLUNK_ENV_PATH}")
            run_container_cmd(container, f"grep '^MY_CUSTOM_VAR2=value2$' {SPLUNK_ENV_PATH}")

            assert wait_for(lambda: service_is_running(container))
            assert container.exec_run("systemctl status td-agent").exit_code != 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


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

    run_win_puppet_agent(CONFIG.substitute(collector_version="0.48.0", with_fluentd="true"))

    assert get_registry_value("SPLUNK_REALM") == SPLUNK_REALM
    assert get_registry_value("SPLUNK_ACCESS_TOKEN") == SPLUNK_ACCESS_TOKEN
    assert get_registry_value("SPLUNK_API_URL") == SPLUNK_API_URL
    assert get_registry_value("SPLUNK_INGEST_URL") == SPLUNK_INGEST_URL
    assert get_registry_value("SPLUNK_HEC_URL") == f"{SPLUNK_INGEST_URL}/v1/log"
    assert get_registry_value("SPLUNK_TRACE_URL") == f"{SPLUNK_INGEST_URL}/v2/trace"
    assert get_registry_value("SPLUNK_HEC_TOKEN") == SPLUNK_ACCESS_TOKEN

    assert psutil.win_service_get("splunk-otel-collector").status() == psutil.STATUS_RUNNING
    assert psutil.win_service_get("fluentdwinsvc").status() == psutil.STATUS_RUNNING


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
    assert get_registry_value("SPLUNK_TRACE_URL") == f"{ingest_url}/v2/trace"
    assert get_registry_value("SPLUNK_HEC_TOKEN") == "fake-hec-token"
    assert get_registry_value("MY_CUSTOM_VAR1") == "value1"
    assert get_registry_value("MY_CUSTOM_VAR2") == "value2"

    assert psutil.win_service_get("splunk-otel-collector").status() == psutil.STATUS_RUNNING
    for service in psutil.win_service_iter():
        assert service.name() != "fluentdwinsvc"
