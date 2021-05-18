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
import string
import tempfile

from pathlib import Path

import pytest

from tests.helpers.util import (
    copy_file_into_container,
    run_container_cmd,
    run_distro_container,
    service_is_running,
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


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("puppet_release", ["6", "7"])
def test_puppet(distro, puppet_release):
    if "jessie" in distro and puppet_release != "6":
        pytest.skip(f"Puppet release version {puppet_release} not supported on debian jessie")

    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"PUPPET_RELEASE": puppet_release}
    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            for collector_version in ["0.24.0", "latest"]:
                config = CONFIG.substitute(collector_version=collector_version, with_fluentd="true")
                run_puppet_apply(container, config)
                verify_env_file(container)
                assert wait_for(lambda: service_is_running(container))
                assert container.exec_run("systemctl status td-agent").exit_code == 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            run_container_cmd(container, "journalctl -u td-agent --no-pager")
            if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("puppet_release", ["6", "7"])
def test_puppet_without_fluentd(distro, puppet_release):
    if "jessie" in distro and puppet_release != "6":
        pytest.skip(f"Puppet release version {puppet_release} not supported on debian jessie")

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
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
