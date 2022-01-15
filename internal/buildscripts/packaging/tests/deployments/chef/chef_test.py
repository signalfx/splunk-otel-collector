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

import json
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

# allow CHEF_VERSIONS env var with comma-separated chef versions for test parameterization
CHEF_VERSIONS = os.environ.get("CHEF_VERSIONS", "16.0,latest").split(",")
# CHEF_VERSIONS = os.environ.get("CHEF_VERSIONS", "latest").split(",")

CHEF_CMD = "chef-client -z -o 'recipe[splunk-otel-collector::default]' -j /root/test_attrs.json"

def run_chef_apply(container, configs, chef_version, CHEF_CMD):
    with tempfile.NamedTemporaryFile(mode="w+") as fd:
        print(json.dumps(configs))
        fd.write(json.dumps(configs))
        fd.flush()
        if chef_version == "latest" or int(chef_version.split(".")[0]) >= 15:
            CHEF_CMD += " --chef-license accept-silent"
        copy_file_into_container(container, fd.name, "/root/test_attrs.json")

    run_container_cmd(container, CHEF_CMD)


def verify_env_file(container):
    run_container_cmd(container, f"grep '^SPLUNK_CONFIG={SPLUNK_CONFIG}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_ACCESS_TOKEN={SPLUNK_ACCESS_TOKEN}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_REALM={SPLUNK_REALM}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_API_URL={SPLUNK_API_URL}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_INGEST_URL={SPLUNK_INGEST_URL}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_TRACE_URL={SPLUNK_INGEST_URL}/v2/trace$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_URL={SPLUNK_INGEST_URL}/v1/log$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_HEC_TOKEN={SPLUNK_ACCESS_TOKEN}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_MEMORY_TOTAL_MIB={SPLUNK_MEMORY_TOTAL_MIB}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_BUNDLE_DIR={SPLUNK_BUNDLE_DIR}$' {SPLUNK_ENV_PATH}")
    run_container_cmd(container, f"grep '^SPLUNK_COLLECTD_DIR={SPLUNK_COLLECTD_DIR}$' {SPLUNK_ENV_PATH}")

@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("chef_version", CHEF_VERSIONS)
def test_chef_without_fluentd(distro, chef_version):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    buildargs = {"CHEF_INSTALLER_ARGS": ""}
    if chef_version != "latest":
        buildargs["CHEF_INSTALLER_ARGS"] = f"-v {chef_version}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            configs = {}
            configs["splunk-otel-collector"] = {}
            configs["splunk-otel-collector"]["splunk_access_token"] = SPLUNK_ACCESS_TOKEN
            configs["splunk-otel-collector"]["splunk_realm"] = SPLUNK_REALM
            configs["splunk-otel-collector"]["splunk_ingest_url"] = SPLUNK_INGEST_URL
            configs["splunk-otel-collector"]["splunk_api_url"] = SPLUNK_API_URL
            configs["splunk-otel-collector"]["splunk_service_user"] = SPLUNK_SERVICE_USER
            configs["splunk-otel-collector"]["splunk_service_group"] = SPLUNK_SERVICE_GROUP
            configs["splunk-otel-collector"]["with_fluentd"] = False
            configs["splunk-otel-collector"]["collector_version"] = 'latest'
            run_chef_apply(container, configs, chef_version, CHEF_CMD)
            verify_env_file(container)
            assert wait_for(lambda: service_is_running(container))
            assert container.exec_run("systemctl status td-agent").exit_code != 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")

@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("chef_version", CHEF_VERSIONS)
def test_chef_with_fluentd(distro, chef_version):
    if distro in DEB_DISTROS:
        dockerfile = IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"

    if "opensuse" in distro:
        pytest.skip(f"FluentD is not supported on opensuse")

    buildargs = {"CHEF_INSTALLER_ARGS": ""}
    if chef_version != "latest":
        buildargs["CHEF_INSTALLER_ARGS"] = f"-v {chef_version}"

    with run_distro_container(distro, dockerfile=dockerfile, path=REPO_DIR, buildargs=buildargs) as container:
        try:
            for collector_version in ["0.34.0", "latest"]:
                configs = {}
                configs["splunk-otel-collector"] = {}
                configs["splunk-otel-collector"]["splunk_access_token"] = SPLUNK_ACCESS_TOKEN
                configs["splunk-otel-collector"]["splunk_realm"] = SPLUNK_REALM
                configs["splunk-otel-collector"]["splunk_ingest_url"] = SPLUNK_INGEST_URL
                configs["splunk-otel-collector"]["splunk_api_url"] = SPLUNK_API_URL
                configs["splunk-otel-collector"]["splunk_service_user"] = SPLUNK_SERVICE_USER
                configs["splunk-otel-collector"]["splunk_service_group"] = SPLUNK_SERVICE_GROUP
                configs["splunk-otel-collector"]["with_fluentd"] = True
                configs["splunk-otel-collector"]["collector_version"] = collector_version
                run_chef_apply(container, configs, chef_version, CHEF_CMD)
                verify_env_file(container)
                assert wait_for(lambda: service_is_running(container))
                if "opensuse" not in distro:
                    assert container.exec_run("systemctl status td-agent").exit_code == 0
        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            if "opensuse" not in distro:
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container.exec_run("test -f /var/log/td-agent/td-agent.log").exit_code == 0:
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")
