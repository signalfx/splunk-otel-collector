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

LIBSPLUNK_PKG_NAME = "splunk-otel-auto-instrumentation"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
DEFAULT_INSTRUMENTATION_CONF = "/usr/lib/splunk-instrumentation/instrumentation.conf"
CUSTOM_INSTRUMENTATION_CONF = TESTS_DIR / "instrumentation" / "instrumentation.conf"

SYSTEMD_PKG_NAME = "splunk-otel-systemd-auto-instrumentation"
DEFAULT_SYSTEMD_CONF_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf"
CUSTOM_SYSTEMD_CONF_PATH = TESTS_DIR / "instrumentation" / "99-test.conf"


def get_dockerfile(distro):
    if distro in DEB_DISTROS:
        return IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        return IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"


def get_package(distro, name, arch):
    pkg_dir = REPO_DIR / "instrumentation" / name / "dist"
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


def verify_installed_files(container, package):
    if package == LIBSPLUNK_PKG_NAME:
        for path in [JAVA_AGENT_PATH, LIBSPLUNK_PATH, DEFAULT_INSTRUMENTATION_CONF]:
            assert container_file_exists(container, path), f"{path} not found"
        # verify /etc/ld.so.preload was updated for libsplunk.so
        run_container_cmd(container, f"grep '{LIBSPLUNK_PATH}' /etc/ld.so.preload")
    else:
        for path in [JAVA_AGENT_PATH, DEFAULT_SYSTEMD_CONF_PATH]:
            assert container_file_exists(container, path), f"{path} not found"
        # verify /etc/ld.so.preload was not updated for libsplunk.so
        if container_file_exists(container, "/etc/ld.so.preload"):
            assert container.exec_run(f"grep '{LIBSPLUNK_PATH}' /etc/ld.so.preload") != 0, \
                f"{LIBSPLUNK_PATH} found in /etc/ld.so.preload after {package} was installed"


def verify_tomcat_instrumentation(container, package, otelcol_path, test_case, with_systemd):
    if test_case == "custom" and package == LIBSPLUNK_PKG_NAME:
        # overwrite the default instrumentation.conf with the custom one for testing
        copy_file_into_container(container, CUSTOM_INSTRUMENTATION_CONF, DEFAULT_INSTRUMENTATION_CONF)
    elif package == SYSTEMD_PKG_NAME:
        if test_case == "custom":
            # add a custom systemd conf file for testing
            copy_file_into_container(container, CUSTOM_SYSTEMD_CONF_PATH, "/usr/lib/systemd/system.conf.d/99-test.conf")
        container.restart()
        wait_for_container_cmd(container, "systemctl show-environment", timeout=30)

    # start the collector and get the output stream
    stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True).output

    if with_systemd:
        print("Starting the tomcat systemd service ...")
        run_container_cmd(container, "systemctl start tomcat")
    else:
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

    sdk_lang = "telemetry.sdk.language: Str(java)"
    if test_case == "default":
        if package == LIBSPLUNK_PKG_NAME:
            # service name auto-generated by libsplunk.so
            service_name = "service.name: Str(org-apache-catalina-startup-bootstrap)"
        else:
            # service name auto-generated by the java agent
            service_name = "service.name: Str(Hello, World Application)"
        deployment_environment = None
        profiling = None
    else:
        source = "libsplunk" if package == LIBSPLUNK_PKG_NAME else "systemd"
        service_name = f"service: Str(service_name_from_{source})"
        deployment_environment = f"deployment_environment: Str(deployment_environment_from_{source})"
        profiling = "com.splunk.sourcetype: Str(otel.profiling)"

    sdk_lang_found = False
    service_name_found = False
    deployment_environment_found = False if deployment_environment else True
    profiling_found = False if profiling else True

    # check the collector output stream for attributes
    start_time = time.time()
    for output in TimeoutIterator(stream, timeout=10, sentinel=None):
        if output:
            output = output.decode("utf-8").rstrip()
            print(output)
            if sdk_lang in output:
                sdk_lang_found = True
            if service_name in output:
                service_name_found = True
            if deployment_environment and deployment_environment in output:
                deployment_environment_found = True
            if profiling and profiling in output:
                profiling_found = True
            if sdk_lang_found and service_name_found and deployment_environment_found and profiling_found:
                break
        if (time.time() - start_time) > 300:
            break

    assert sdk_lang_found, f"timed out waiting for '{sdk_lang}'"
    assert service_name_found, f"timed out waiting for '{service_name}'"
    assert deployment_environment_found, f"timed out waiting for '{deployment_environment}'"
    assert profiling_found, f"timed out waiting for '{profiling}'"


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("package", [LIBSPLUNK_PKG_NAME, SYSTEMD_PKG_NAME])
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("test_case", ["default", "custom"])
@pytest.mark.parametrize("with_systemd", [True, False])
def test_tomcat_instrumentation(distro, package, arch, test_case, with_systemd):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    if package == SYSTEMD_PKG_NAME and not with_systemd:
        pytest.skip(f"{SYSTEMD_PKG_NAME} only supported with systemd")

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"

    pkg_path = get_package(distro, package, arch)
    assert pkg_path, f"{package} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        install_package(container, distro, f"/test/{pkg_base}")
        verify_installed_files(container, package)

        verify_tomcat_instrumentation(container, package, f"/test/{otelcol_bin}", test_case, with_systemd)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("first_package", [LIBSPLUNK_PKG_NAME, SYSTEMD_PKG_NAME])
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_package_conflict(distro, first_package, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    first_package_path = get_package(distro, first_package, arch)
    assert first_package_path, f"{first_package} package not found"
    first_package_base = os.path.basename(first_package_path)

    second_package = SYSTEMD_PKG_NAME if first_package == LIBSPLUNK_PKG_NAME else LIBSPLUNK_PKG_NAME
    second_package_path = get_package(distro, second_package, arch)
    assert second_package_path, f"{second_package} package not found"
    second_package_base = os.path.basename(second_package_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, first_package_path, f"/test/{first_package_base}")
        copy_file_into_container(container, second_package_path, f"/test/{second_package_base}")

        # install and verify the first package
        install_package(container, distro, f"/test/{first_package_base}")
        verify_installed_files(container, first_package)

        # installation of the second package should fail
        try:
            install_package(container, distro, f"/test/{second_package_base}")
        except AssertionError:
            pass
        else:
            pytest.fail(f"Installation of {second_package} should have failed")

        # verify files from first package were unchanged
        verify_installed_files(container, first_package)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("package", [LIBSPLUNK_PKG_NAME, SYSTEMD_PKG_NAME])
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_package_uninstall(distro, package, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    pkg_path = get_package(distro, package, arch)
    assert pkg_path, f"{package} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")

        # install the package
        install_package(container, distro, f"/test/{pkg_base}")

        verify_installed_files(container, package)

        # uninstall the package
        if distro in DEB_DISTROS:
            run_container_cmd(container, f"dpkg -P {package}")
        else:
            run_container_cmd(container, f"rpm -e {package}")

        # verify the package was uninstalled
        if distro in DEB_DISTROS:
            assert container.exec_run(f"dpkg -s {package}").exit_code != 0
        else:
            assert container.exec_run(f"rpm -q {package}").exit_code != 0

        # verify files were uninstalled
        if package == LIBSPLUNK_PKG_NAME:
            for path in [JAVA_AGENT_PATH, DEFAULT_SYSTEMD_CONF_PATH]:
                assert not container_file_exists(container, path)
        else:
            for path in [JAVA_AGENT_PATH, LIBSPLUNK_PATH, DEFAULT_INSTRUMENTATION_CONF]:
                assert not container_file_exists(container, path)
