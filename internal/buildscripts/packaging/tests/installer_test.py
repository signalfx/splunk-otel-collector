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
    TESTS_DIR,
)

INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"

# Override default test parameters with the following env vars
STAGE = os.environ.get("STAGE", "release")
VERSIONS = os.environ.get("VERSIONS", "latest").split(",")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
)
@pytest.mark.parametrize("service_owner", ["default_service_owner", "new_service_owner"])
@pytest.mark.parametrize("version", VERSIONS)
def test_installer(distro, service_owner, version):
    install_cmd = "sh -x /test/install.sh -- testing123 --realm us0 --ballast 64"

    if version != "latest":
        install_cmd = f"{install_cmd} --collector-version {version.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    if service_owner == "default_service_owner":
        service_owner = "splunk-otel-collector"
    else:
        install_cmd = f"{install_cmd} --service-user {service_owner} --service-group {service_owner}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro) as container:
        # run installer script
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")
        run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"})
        time.sleep(5)

        try:
            # verify env file created with configured parameters
            run_container_cmd(container, "grep '^SPLUNK_ACCESS_TOKEN=testing123$' /etc/otel/collector/splunk_env")
            run_container_cmd(container, "grep '^SPLUNK_REALM=us0$' /etc/otel/collector/splunk_env")
            run_container_cmd(container, "grep '^SPLUNK_BALLAST_SIZE_MIB=64$' /etc/otel/collector/splunk_env")

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=service_owner))

            # the td-agent service should only be running when installing
            # collector packages that have our custom fluent config
            if container.exec_run("test -f /etc/otel/collector/fluentd/fluent.conf").exit_code == 0:
                assert container.exec_run("systemctl status td-agent").exit_code == 0
            else:
                assert container.exec_run("systemctl status td-agent").exit_code != 0

        finally:
            run_container_cmd(container, "journalctl -u td-agent --no-pager")
            run_container_cmd(container, "journalctl -u splunk-otel-collector --no-pager")
