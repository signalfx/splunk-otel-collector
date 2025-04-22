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
INSTALLER_PATH = REPO_DIR / "packaging" / "installer" / "install.sh"
COLLECTOR_CONFIG_PATH = TESTS_DIR / "instrumentation" / "config.yaml"

PKG_NAME = "splunk-otel-auto-instrumentation"
LIB_DIR = "/usr/lib/splunk-instrumentation"
LIBSPLUNK_PATH = f"{LIB_DIR}/libsplunk.so"
PRELOAD_PATH = "/etc/ld.so.preload"
SYSTEMD_CONF_DIR = "/usr/lib/systemd/system.conf.d"
SAMPLE_SYSTEMD_CONF_PATH = f"{LIB_DIR}/examples/systemd/00-splunk-otel-auto-instrumentation.conf"

JAVA_AGENT_PATH = f"{LIB_DIR}/splunk-otel-javaagent.jar"
JAVA_CONFIG_PATH = "/etc/splunk/zeroconfig/java.conf"
CUSTOM_JAVA_CONFIG_PATH = TESTS_DIR / "instrumentation" / "libsplunk-java-test.conf"
CUSTOM_JAVA_SYSTEMD_CONF_PATH = TESTS_DIR / "instrumentation" / "systemd-java-test.conf"

NODE_AGENT_PATH = f"{LIB_DIR}/splunk-otel-js.tgz"
NODE_CONFIG_PATH = "/etc/splunk/zeroconfig/node.conf"
CUSTOM_NODE_CONFIG_PATH = TESTS_DIR / "instrumentation" / "libsplunk-node-test.conf"
CUSTOM_NODE_SYSTEMD_CONF_PATH = TESTS_DIR / "instrumentation" / "systemd-node-test.conf"

DOTNET_AGENT_PATH = f"{LIB_DIR}/splunk-otel-dotnet/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
DOTNET_CONFIG_PATH = "/etc/splunk/zeroconfig/dotnet.conf"
CUSTOM_DOTNET_CONFIG_PATH = TESTS_DIR / "instrumentation" / "libsplunk-dotnet-test.conf"
CUSTOM_DOTNET_SYSTEMD_CONF_PATH = TESTS_DIR / "instrumentation" / "systemd-dotnet-test.conf"

INSTALLED_FILES = [
    JAVA_AGENT_PATH,
    NODE_AGENT_PATH,
    DOTNET_AGENT_PATH,
    LIBSPLUNK_PATH,
    JAVA_CONFIG_PATH,
    NODE_CONFIG_PATH,
    DOTNET_CONFIG_PATH,
    SAMPLE_SYSTEMD_CONF_PATH,
]

TOMCAT_PIDFILE = "/usr/local/tomcat/temp/tomcat.pid"
TOMCAT_ENV = {
    "JAVA_HOME": "/opt/java/openjdk",
    "CATALINA_PID": TOMCAT_PIDFILE,
    "CATALINA_HOME": "/usr/local/tomcat",
    "CATALINA_BASE": "/usr/local/tomcat",
    "CATALINA_OPTS": "-Xms512M -Xmx1024M -server -XX:+UseParallelGC",
    "JAVA_OPTS": "-Djava.awt.headless=true",
}

EXPRESS_PIDFILE = "/opt/express/express.pid"
DOTNET_PIDFILE = "/opt/dotnet/dotnet.pid"


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


def install_package(container, distro, path, arch="amd64"):
    if distro in DEB_DISTROS:
        run_container_cmd(container, f"dpkg -i {path}")
    else:
        run_container_cmd(container, f"rpm -ivh {path}")

    for path in INSTALLED_FILES:
        if arch == "arm64" and path in [DOTNET_AGENT_PATH, DOTNET_CONFIG_PATH]:
            # the arm64 package shouldn't include splunk-otel-dotnet files
            assert not container_file_exists(container, path), f"{path} found"
        else:
            assert container_file_exists(container, path), f"{path} not found"


def verify_preload(container, line, exists=True):
    code, output = container.exec_run(f"cat {PRELOAD_PATH}")
    assert code == 0, f"failed to get contents from {PRELOAD_PATH}"
    config = output.decode("utf-8")

    match = re.search(f"^{line}$", config, re.MULTILINE)

    if exists:
        assert match, f"'{line}' not found in {PRELOAD_PATH}"
    else:
        assert not match, f"'{line}' found in {PRELOAD_PATH}"


def start_app(container, app, systemd, timeout=300):
    if systemd:
        print(f"Starting the {app} systemd service ...")
        run_container_cmd(container, f"systemctl start {app}")
    else:
        print(f"Starting {app} from a shell ...")
        if app == "tomcat":
            run_container_cmd(container, "bash -c /usr/local/tomcat/bin/startup.sh", env=TOMCAT_ENV, user='tomcat:tomcat')
        elif app == "express":
            run_container_cmd(
                container, f"bash -l -c 'node /opt/express/app.js & echo $! > {EXPRESS_PIDFILE}'", user='express:express',
            )
        elif app == "dotnet":
            run_container_cmd(
                container,
                f"bash -c '/opt/dotnet-sdk/dotnet /opt/dotnet/myWebApp.dll & echo $! > {DOTNET_PIDFILE}'",
                user='dotnet:dotnet',
                workdir="/opt/dotnet",
            )

    if app == "tomcat":
        print("Waiting for http://127.0.0.1:8080/sample ...")
        wait_for_container_cmd(container, "curl -sSL http://127.0.0.1:8080/sample", timeout=timeout)
    elif app == "express":
        print("Waiting for http://127.0.0.1:3000 ...")
        wait_for_container_cmd(container, "curl -sSL http://127.0.0.1:3000", timeout=timeout)
    elif app == "dotnet":
        print("Waiting for http://127.0.0.1:5000 ...")
        wait_for_container_cmd(container, "curl -sSL http://127.0.0.1:5000", timeout=timeout)


def stop_app(container, app):
    run_container_cmd(container, f"systemctl stop {app}")

    pidfile = None
    if app == "tomcat":
        pidfile = TOMCAT_PIDFILE
    elif app == "express":
        pidfile = EXPRESS_PIDFILE
    elif app == "dotnet":
        pidfile = DOTNET_PIDFILE

    if pidfile and container_file_exists(container, pidfile):
        if app == "tomcat":
            run_container_cmd(container, "bash -c /usr/local/tomcat/bin/shutdown.sh", env=TOMCAT_ENV)
        elif app in ("express", "dotnet"):
            run_container_cmd(container, f"bash -c 'kill -TERM `cat {pidfile}`'")
            run_container_cmd(container, f"rm -f {pidfile}")


def verify_attributes(stream, attributes, timeout=300):
    found = {}
    for key, value in attributes.items():
        found[key] = False if value else True

    start_time = time.time()
    for output in TimeoutIterator(stream, timeout=10, sentinel=None):
        if output:
            output = output.decode("utf-8").rstrip()
            print(output)
            for key, value in attributes.items():
                if found[key]:
                    continue
                if re.search(f"{key}: {value}", output, re.MULTILINE):
                    found[key] = True
        if False not in found.values() or ((time.time() - start_time) > timeout):
            break

    for key, value in attributes.items():
        assert found[key], f"timed out waiting for '{key}: {value}'"


def verify_app_instrumentation(container, app, method, attributes, otelcol_path=None, timeout=300):
    systemd = True if method == "systemd" else False

    try:
        stop_app(container, app)
    except AssertionError:
        pass

    container.restart()
    wait_for_container_cmd(container, "systemctl show-environment", timeout=30)

    if otelcol_path is None:
        # start the collector systemd service
        run_container_cmd(container, "systemctl start splunk-otel-collector")
        wait_for_container_cmd(container, "systemctl status splunk-otel-collector", timeout=30)
        # get the output stream for the collector from journald
        stream = container.exec_run("journalctl -u splunk-otel-collector -f", stream=True).output
    else:
        # start the collector from the shell and get the output stream
        stream = container.exec_run(f"{otelcol_path} --config=/test/config.yaml", stream=True).output

    start_app(container, app, systemd)

    # check the collector output stream for attributes
    verify_attributes(stream, attributes, timeout=timeout)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_tomcat_instrumentation(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"
    otelcol = f"/test/{otelcol_bin}"

    pkg_path = get_package(distro, PKG_NAME, arch)
    assert pkg_path, f"{PKG_NAME} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        install_package(container, distro, f"/test/{pkg_base}", arch=arch)

        for method in ["systemd", "libsplunk"]:
            # attributes from default config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(java\)",
                r"service\.name": r"Str\(Hello, World Application\)",  # auto-generated for the sample app
            }

            if method == "systemd":
                # install the sample drop-in file to enable the agent
                run_container_cmd(container, f"mkdir -p {SYSTEMD_CONF_DIR}")
                run_container_cmd(container, f"cp -f {SAMPLE_SYSTEMD_CONF_PATH} {SYSTEMD_CONF_DIR}/")
                if container_file_exists(container, "/etc/ld.so.preload"):
                    run_container_cmd(container, "rm -f /etc/ld.so.preload")
            else:
                # add libsplunk.so to /etc/ld.so.preload
                run_container_cmd(container, f"sh -c 'echo {LIBSPLUNK_PATH} > /etc/ld.so.preload'")

            # verify default config
            verify_app_instrumentation(container, "tomcat", method, attributes, otelcol_path=otelcol)

            # attributes from custom config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(java\)",
                r"service\.name": rf"Str\(service_name_from_{method}_java\)",
                r"deployment\.environment": rf"Str\(deployment_environment_from_{method}_java\)",
                r"com\.splunk\.sourcetype": r"Str\(otel\.profiling\)",
            }

            if method == "systemd":
                # install the custom drop-in file to configure the agent
                copy_file_into_container(container, CUSTOM_JAVA_SYSTEMD_CONF_PATH, f"{SYSTEMD_CONF_DIR}/test.conf")
            else:
                # overwrite the default libsplunk config with the custom one for testing
                copy_file_into_container(container, CUSTOM_JAVA_CONFIG_PATH, JAVA_CONFIG_PATH)

            # verify custom config
            verify_app_instrumentation(container, "tomcat", method, attributes, otelcol_path=otelcol)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_express_instrumentation(distro, arch):
    if distro == "opensuse-12" and arch == "arm64":
        pytest.skip("opensuse-12 arm64 no longer supported")

    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"
    otelcol = f"/test/{otelcol_bin}"

    pkg_path = get_package(distro, PKG_NAME, arch)
    assert pkg_path, f"{PKG_NAME} package not found"
    pkg_base = os.path.basename(pkg_path)

    # minimum supported node version
    node_version = 18

    buildargs = {"NODE_VERSION": f"v{node_version}"}

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch, buildargs=buildargs) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, otelcol)
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        install_package(container, distro, f"/test/{pkg_base}", arch=arch)

        # install splunk-otel-js to /usr/lib/splunk-instrumentation/splunk-otel-js
        run_container_cmd(container, f"mkdir -p {LIB_DIR}/splunk-otel-js")
        run_container_cmd(container, f"bash -l -c 'cd {LIB_DIR}/splunk-otel-js && npm install {NODE_AGENT_PATH}'")

        for method in ["systemd", "libsplunk"]:
            # attributes from default config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(nodejs\)",
                r"service\.name": r"Str\(unnamed-node-service\)",  # auto-generated for the sample app
            }

            if method == "systemd":
                # install the sample drop-in file to enable the agent
                run_container_cmd(container, f"mkdir -p {SYSTEMD_CONF_DIR}")
                run_container_cmd(container, f"cp -f {SAMPLE_SYSTEMD_CONF_PATH} {SYSTEMD_CONF_DIR}/")
                if container_file_exists(container, "/etc/ld.so.preload"):
                    run_container_cmd(container, "rm -f /etc/ld.so.preload")
            else:
                # add libsplunk.so to /etc/ld.so.preload
                run_container_cmd(container, f"sh -c 'echo {LIBSPLUNK_PATH} > /etc/ld.so.preload'")

            # verify default config
            verify_app_instrumentation(container, "express", method, attributes, otelcol_path=otelcol)

            # attributes from custom config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(nodejs\)",
                r"service\.name": rf"Str\(service_name_from_{method}_node\)",
                r"deployment\.environment": rf"Str\(deployment_environment_from_{method}_node\)",
                r"com\.splunk\.sourcetype": None if node_version < 16 else r"Str\(otel\.profiling\)",
            }

            if method == "systemd":
                # install the custom drop-in file to configure the agent
                copy_file_into_container(container, CUSTOM_NODE_SYSTEMD_CONF_PATH, f"{SYSTEMD_CONF_DIR}/test.conf")
            else:
                # overwrite the default libsplunk config with the custom one for testing
                copy_file_into_container(container, CUSTOM_NODE_CONFIG_PATH, NODE_CONFIG_PATH)

            # verify custom config
            verify_app_instrumentation(container, "express", method, attributes, otelcol_path=otelcol)


@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64"])
def test_dotnet_instrumentation(distro, arch):
    otelcol_bin = f"otelcol_linux_{arch}"
    otelcol_bin_path = OTELCOL_BIN_DIR / otelcol_bin
    assert os.path.isfile(otelcol_bin_path), f"{otelcol_bin_path} not found!"
    otelcol = f"/test/{otelcol_bin}"

    pkg_path = get_package(distro, PKG_NAME, arch)
    assert pkg_path, f"{PKG_NAME} package not found"
    pkg_base = os.path.basename(pkg_path)

    with run_distro_container(distro, dockerfile=get_dockerfile(distro), arch=arch) as container:
        copy_file_into_container(container, COLLECTOR_CONFIG_PATH, "/test/config.yaml")
        copy_file_into_container(container, pkg_path, f"/test/{pkg_base}")
        copy_file_into_container(container, otelcol_bin_path, f"/test/{otelcol_bin}")
        run_container_cmd(container, f"chmod a+x /test/{otelcol_bin}")

        install_package(container, distro, f"/test/{pkg_base}", arch=arch)

        for method in ["libsplunk", "systemd"]:
            # attributes from default config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(dotnet\)",
                r"service\.name": r"Str\(myWebApp\)",  # auto-generated for the sample app
            }

            if method == "systemd":
                # install the sample drop-in file to enable the agent
                run_container_cmd(container, f"mkdir -p {SYSTEMD_CONF_DIR}")
                run_container_cmd(container, f"cp -f {SAMPLE_SYSTEMD_CONF_PATH} {SYSTEMD_CONF_DIR}/")
                if container_file_exists(container, "/etc/ld.so.preload"):
                    run_container_cmd(container, "rm -f /etc/ld.so.preload")
            else:
                # add libsplunk.so to /etc/ld.so.preload
                run_container_cmd(container, f"sh -c 'echo {LIBSPLUNK_PATH} > /etc/ld.so.preload'")

            # verify default config
            verify_app_instrumentation(container, "dotnet", method, attributes, otelcol_path=otelcol)

            # attributes from custom config
            attributes = {
                r"telemetry\.sdk\.language": r"Str\(dotnet\)",
                r"service\.name": rf"Str\(service_name_from_{method}_dotnet\)",
                r"deployment\.environment": rf"Str\(deployment_environment_from_{method}_dotnet\)",
                r"com\.splunk\.sourcetype": r"Str\(otel\.profiling\)",
            }

            if method == "systemd":
                # install the custom drop-in file to configure the agent
                copy_file_into_container(container, CUSTOM_DOTNET_SYSTEMD_CONF_PATH, f"{SYSTEMD_CONF_DIR}/test.conf")
            else:
                # overwrite the default libsplunk config with the custom one for testing
                copy_file_into_container(container, CUSTOM_DOTNET_CONFIG_PATH, DOTNET_CONFIG_PATH)

            # verify custom config
            verify_app_instrumentation(container, "dotnet", method, attributes, otelcol_path=otelcol)


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
        install_package(container, distro, f"/test/{pkg_base}", arch=arch)

        verify_preload(container, "# This line should be preserved")

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
        for path in INSTALLED_FILES:
            assert not container_file_exists(container, path)

        # verify libsplunk.so was removed from /etc/ld.so.preload
        verify_preload(container, LIBSPLUNK_PATH, exists=False)

        verify_preload(container, "# This line should be preserved")
