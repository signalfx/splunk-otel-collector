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
    if "jessie" in distro and puppet_release != "6":
        pytest.skip(f"Puppet release version {puppet_release} not supported on debian jessie")


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
