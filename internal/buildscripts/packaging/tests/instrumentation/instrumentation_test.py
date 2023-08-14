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
import time

from pathlib import Path

import pytest
from iterators import TimeoutIterator

from tests.helpers.util import (
    copy_file_into_container,
    retry,
    run_container_cmd,
    run_distro_container,
    wait_for,
    wait_for_container_cmd,
    REPO_DIR,
    TESTS_DIR,
)


IMAGES_DIR = Path(__file__).parent.resolve() / "images"
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(IMAGES_DIR / "rpm" / "Dockerfile.*"))]

OTELCOL_BIN_DIR = REPO_DIR / "bin"
INSTALLER_PATH = REPO_DIR / "internal" / "buildscripts" / "packaging" / "installer" / "install.sh"
COLLECTOR_CONFIG_PATH = TESTS_DIR / "instrumentation" / "config.yaml"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"

PKG_NAME = "splunk-otel-auto-instrumentation"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
DEFAULT_INSTRUMENTATION_CONF = "/usr/lib/splunk-instrumentation/instrumentation.conf"
CUSTOM_INSTRUMENTATION_CONF = TESTS_DIR / "instrumentation" / "libsplunk-test.conf"
DEFAULT_SYSTEMD_CONF_PATH = "/usr/lib/splunk-instrumentation/examples/systemd/00-splunk-otel-javaagent.conf"
CUSTOM_SYSTEMD_CONF_PATH = TESTS_DIR / "instrumentation" / "systemd-test.conf"
SYSTEMD_CONF_DIR = "/usr/lib/systemd/system.conf.d"
PRELOAD_PATH = "/etc/ld.so.preload"


def get_dockerfile(distro):
    if distro in DEB_DISTROS:
        return IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        return IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"


def get_package(distro, name, arch):
    pkg_dir = REPO_DIR / "instrumentation" / "dist"
    pkg_paths = []
    if distro in DEB_DISTROS:
        pkg_paths = glob.glob(str(pkg_dir / f"{name}*{arch}.deb"))
    elif distro in RPM_DISTROS:
        if arch == "amd64":
            arch = "x86_64"
        elif arch == "arm64":
            arch = "aarch64"
        pkg_paths = glob.glob(str(pkg_dir / f"{name}*{arch}.rpm"))

    if pkg_paths:
        return sorted(pkg_paths)[-1]
    else:
        return None


def container_file_exists(container, path):
    return container.exec_run(f"test -f {path}").exit_code == 0


def install_package(container, distro, path):
    if distro in DEB_DISTROS:
        run_container_cmd(container, f"dpkg -i {path}")
    else:
        run_container_cmd(container, f"rpm -ivh {path}")


def verify_preload(container, line, exists=True):
    code, output = container.exec_run(f"cat {PRELOAD_PATH}")
    assert code == 0, f"failed to get contents from {PRELOAD_PATH}"
    config = output.decode("utf-8")

    match = re.search(f"^{line}$", config, re.MULTILINE)

    if exists:
        assert match, f"'{line}' not found in {PRELOAD_PATH}"
    else:
        assert not match, f"'{line}' found in {PRELOAD_PATH}"


def verify_tomcat_instrumentation(container, otelcol_path, test_case, source, attributes=[]):
    if source == "systemd":
        if otelcol_path is None:
            container.restart()
            wait_for_container_cmd(container, "systemctl show-environment", timeout=30)
            wait_for_container_cmd(container, "systemctl status splunk-otel-collector", timeout=30)
            # get the output stream for the collector from journald
            stream = container.exec_run("journalctl -u splunk-otel-collector -f", stream=True).output
        else:
            run_container_cmd(container, f"mkdir -p {SYSTEMD_CONF_DIR}")
            run_container_cmd(container, f"cp {DEFAULT_SYSTEMD_CONF_PATH} {SYSTEMD_CONF_DIR}/")
            if test_case == "custom":
                copy_file_into_container(container, CUSTOM_SYSTEMD_CONF_PATH,
                                         f"{SYSTEMD_CONF_DIR}/99-systemd-test.conf")
            container.restart()
            wait_for_container_cmd(container, "systemctl show-environment", timeout=30)
            # start the collector and get the output stream
            stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True).output
        print("Starting the tomcat systemd service ...")
        run_container_cmd(container, "systemctl start tomcat")
    else:
        if otelcol_path is None:
            wait_for_container_cmd(container, "systemctl status splunk-otel-collector", timeout=30)
            # get the output stream for the collector from journald
            stream = container.exec_run("journalctl -u splunk-otel-collector -f", stream=True).output
        else:
            run_container_cmd(container,
                              "sh -c 'echo /usr/lib/splunk-instrumentation/libsplunk.so > /etc/ld.so.preload'")
            if test_case == "custom":
                # overwrite the default instrumentation.conf with the custom one for testing
                copy_file_into_container(container, CUSTOM_INSTRUMENTATION_CONF, DEFAULT_INSTRUMENTATION_CONF)
            # start the collector and get the output stream
            stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True).output
        print("Starting tomcat from a shell ...")
        tomcat_env = {
            "JAVA_HOME": "/opt/java/openjdk",
            "CATALINA_PID": "/usr/local/tomcat/temp/tomcat.pid",
            "CATALINA_HOME": "/usr/local/tomcat",
            "CATALINA_BASE": "/usr/local/tomcat",
            "CATALINA_OPTS": "-Xms512M -Xmx1024M -server -XX:+UseParallelGC",
            "JAVA_OPTS": "-Djava.awt.headless=true",
        }
        run_container_cmd(container, "bash -c /usr/local/tomcat/bin/startup.sh", env=tomcat_env)

    print("Waiting for http://127.0.0.1:8080/sample ...")
    wait_for_container_cmd(container, "curl -sSL http://127.0.0.1:8080/sample", timeout=300)

    # check the collector output stream for attributes
    start_time = time.time()
    for output in TimeoutIterator(stream, timeout=10, sentinel=None):
        if output:
            output = output.decode("utf-8").rstrip()
            print(output)
            for attr in attributes:
                if attr["found"]:
                    continue
                if re.search(f"{attr['key']}: {attr['value']}", output, re.MULTILINE):
                    attr["found"] = True
        if False not in [ attr["found"] for attr in attributes ] or ((time.time() - start_time) > 300):
            break

    for attr in attributes:
        assert attr["found"], f"timed out waiting for '{attr['key']}: {attr['value']}'"


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("test_case", ["default", "custom"])
@pytest.mark.parametrize("source", ["systemd", "libsplunk"])
def test_tomcat_instrumentation(distro, arch, test_case, source):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"

    pkg_path = get_package(distro, PKG_NAME, arch)
    assert pkg_path, f"{PKG_NAME} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        install_package(container, distro, f"/test/{pkg_base}")
        for path in [JAVA_AGENT_PATH, LIBSPLUNK_PATH, DEFAULT_INSTRUMENTATION_CONF, DEFAULT_SYSTEMD_CONF_PATH]:
            assert container_file_exists(container, path), f"{path} not found"

        if test_case == "default":
            if source == "libsplunk":
                # service name auto-generated by libsplunk.so
                service_name = r"Str\(org-apache-catalina-startup-bootstrap\)"
            else:
                # service name auto-generated by java agent
                service_name = r"Str\(Hello, World Application\)"
            environment = None
            profiling = None
        else:
            service_name = rf"Str\(service_name_from_{source}\)"
            environment = rf"Str\(deployment_environment_from_{source}\)"
            profiling = r"Str\(otel\.profiling\)"

        attributes = [
            {"key": r"telemetry\.sdk\.language", "value": r"Str\(java\)", "found": False},
            {"key": r"service\.name", "value": service_name, "found": False},
            {"key": r"deployment\.environment", "value": environment, "found": False if environment else True},
            {"key": r"com\.splunk\.sourcetype", "value": profiling, "found": False if profiling else True},
        ]

        verify_tomcat_instrumentation(container, f"/test/{otelcol_bin}", test_case, source, attributes)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_package_uninstall(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    pkg_path = get_package(distro, PKG_NAME, arch)
    assert pkg_path, f"{PKG_NAME} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        run_container_cmd(container, f"sh -c 'echo \"# This line should be preserved\" >> {PRELOAD_PATH}'")

        # install the package
        install_package(container, distro, f"/test/{pkg_base}")

        verify_preload(container, "# This line should be preserved")

        for path in [JAVA_AGENT_PATH, LIBSPLUNK_PATH, DEFAULT_INSTRUMENTATION_CONF, DEFAULT_SYSTEMD_CONF_PATH]:
            assert container_file_exists(container, path), f"{path} not found"

        # verify libsplunk.so was not automatically added to /etc/ld.so.preload
        verify_preload(container, LIBSPLUNK_PATH, exists=False)

        # explicitly add libsplunk.so to /etc/ld.so.preload
        run_container_cmd(container, f"sh -c 'echo {LIBSPLUNK_PATH} >> {PRELOAD_PATH}'")

        # uninstall the package
        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -P {PKG_NAME}")
        else:
            run_container_cmd(container, f"rpm -e {PKG_NAME}")

        # verify the package was uninstalled
        if distro in DEB_DISTROS:
            assert container.exec_run(f"dpkg -s {PKG_NAME}").exit_code != 0
        else:
            assert container.exec_run(f"rpm -q {PKG_NAME}").exit_code != 0

        # verify files were uninstalled
        for path in [JAVA_AGENT_PATH, LIBSPLUNK_PATH, DEFAULT_INSTRUMENTATION_CONF, DEFAULT_SYSTEMD_CONF_PATH]:
            assert not container_file_exists(container, path)

        # verify libsplunk.so was removed from /etc/ld.so.preload
        verify_preload(container, LIBSPLUNK_PATH, exists=False)

        verify_preload(container, "# This line should be preserved")
