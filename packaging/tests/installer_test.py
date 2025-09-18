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
    SERVICE_NAME,
    SERVICE_OWNER,
    TESTS_DIR,
)

from tests.instrumentation.instrumentation_test import (
    IMAGES_DIR as INSTR_IMAGES_DIR,
    DEB_DISTROS as INSTR_DEB_DISTROS,
    RPM_DISTROS as INSTR_RPM_DISTROS,
)


CUSTOM_COLLECTOR_CONFIG = TESTS_DIR / "custom-config.yaml"
INSTALLER_PATH = REPO_DIR / "packaging" / "installer" / "install.sh"

# Override default test parameters with the following env vars
STAGE = os.environ.get("STAGE", "release")
VERSION = os.environ.get("VERSION", "latest")
SPLUNK_ACCESS_TOKEN = os.environ.get("SPLUNK_ACCESS_TOKEN", "testing123")
SPLUNK_REALM = os.environ.get("SPLUNK_REALM", "fake-realm")
LOCAL_COLLECTOR_PACKAGE = os.environ.get("LOCAL_COLLECTOR_PACKAGE")
LOCAL_INSTRUMENTATION_PACKAGE = os.environ.get("LOCAL_INSTRUMENTATION_PACKAGE")
DEBUG = os.environ.get("DEBUG", "no")
TOTAL_MEMORY = "512"

SPLUNK_ENV_PATH = "/etc/otel/collector/splunk-otel-collector.conf"
OLD_SPLUNK_ENV_PATH = "/etc/otel/collector/splunk_env"
AGENT_CONFIG_PATH = "/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_PATH = "/etc/otel/collector/gateway_config.yaml"
OLD_CONFIG_PATH = "/etc/otel/collector/splunk_config_linux.yaml"
LIBSPLUNK_PATH = "/usr/lib/splunk-instrumentation/libsplunk.so"
JAVA_AGENT_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
JAVA_TOOL_OPTIONS = f"-javaagent:{JAVA_AGENT_PATH}"
PRELOAD_PATH = "/etc/ld.so.preload"
SYSTEMD_CONFIG_PATH = "/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
NODE_PACKAGE_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-js.tgz"
JAVA_ZEROCONFIG_PATH = "/etc/splunk/zeroconfig/java.conf"
NODE_ZEROCONFIG_PATH = "/etc/splunk/zeroconfig/node.conf"
DOTNET_ZEROCONFIG_PATH = "/etc/splunk/zeroconfig/dotnet.conf"
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

INSTALLER_TIMEOUT = "30m"


def container_file_exists(container, path):
    return container.exec_run(f"test -f {path}").exit_code == 0


def package_is_installed(container, distro, name):
    if distro in DEB_DISTROS:
        return container.exec_run(f"dpkg -s {name}").exit_code == 0
    else:
        return container.exec_run(f"rpm -q {name}").exit_code == 0


def get_installer_cmd():
    debug_flag = "-x" if DEBUG == "yes" else ""

    install_cmd = f"sh -l {debug_flag} /test/install.sh -- {SPLUNK_ACCESS_TOKEN} --realm {SPLUNK_REALM}"

    if VERSION != "latest":
        install_cmd = f"{install_cmd} --collector-version {VERSION.lstrip('v')}"

    if STAGE != "release":
        assert STAGE in ("test", "beta"), f"Unsupported stage '{STAGE}'!"
        install_cmd = f"{install_cmd} --{STAGE}"

    return install_cmd


def verify_config_file(container, path, key, value, exists=True):
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


def verify_env_file(container, mode="agent", config_path=None, memory=TOTAL_MEMORY, listen_addr=""):
    env_path = SPLUNK_ENV_PATH
    if container_file_exists(container, OLD_SPLUNK_ENV_PATH):
        env_path = OLD_SPLUNK_ENV_PATH

    if not config_path:
        config_path = AGENT_CONFIG_PATH if mode == "agent" else GATEWAY_CONFIG_PATH
        if container_file_exists(container, OLD_CONFIG_PATH):
            config_path = OLD_CONFIG_PATH
        elif mode == "gateway" and not container_file_exists(container, GATEWAY_CONFIG_PATH):
            config_path = AGENT_CONFIG_PATH

    ingest_url = f"https://ingest.{SPLUNK_REALM}.signalfx.com"
    api_url = f"https://api.{SPLUNK_REALM}.signalfx.com"

    verify_config_file(container, env_path, "SPLUNK_CONFIG", config_path)
    verify_config_file(container, env_path, "SPLUNK_ACCESS_TOKEN", SPLUNK_ACCESS_TOKEN)
    verify_config_file(container, env_path, "SPLUNK_REALM", SPLUNK_REALM)
    verify_config_file(container, env_path, "SPLUNK_API_URL", api_url)
    verify_config_file(container, env_path, "SPLUNK_INGEST_URL", ingest_url)
    verify_config_file(container, env_path, "SPLUNK_HEC_URL", f"{ingest_url}/v1/log")
    verify_config_file(container, env_path, "SPLUNK_HEC_TOKEN", SPLUNK_ACCESS_TOKEN)
    verify_config_file(container, env_path, "SPLUNK_MEMORY_TOTAL_MIB", memory)
    if listen_addr:
        verify_config_file(container, env_path, "SPLUNK_LISTEN_INTERFACE", listen_addr)
    else:
        verify_config_file(container, env_path, "SPLUNK_LISTEN_INTERFACE", False, False)


def verify_support_bundle(container):
    run_container_cmd(container, "/etc/otel/collector/splunk-support-bundle.sh -t /tmp/splunk-support-bundle")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/config/agent_config.yaml")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/logs/splunk-otel-collector.log")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/logs/splunk-otel-collector.txt")
    if container_file_exists(container, "/etc/otel/collector/fluentd/fluent.conf"):
        assert container_file_exists(container, "/tmp/splunk-support-bundle/logs/td-agent.log")
        assert container_file_exists(container, "/tmp/splunk-support-bundle/logs/td-agent.txt")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/metrics/collector-metrics.txt")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/metrics/df.txt")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/metrics/free.txt")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/metrics/top.txt")
    assert container_file_exists(container, "/tmp/splunk-support-bundle/zpages/tracez.html")
    assert container_file_exists(container, "/tmp/splunk-support-bundle.tar.gz")


def verify_uninstall(container, distro):
    debug_flag = "-x" if DEBUG == "yes" else ""

    run_container_cmd(container, f"sh -l {debug_flag} /test/install.sh --uninstall")

    for pkg in ("splunk-otel-collector", "td-agent", "splunk-otel-auto-instrumentation"):
        assert not package_is_installed(container, distro, pkg), f"{pkg} was not uninstalled"

    # verify libsplunk.so was removed from /etc/ld.so.preload after uninstall
    verify_config_file(container, PRELOAD_PATH, f".*{LIBSPLUNK_PATH}.*", None, exists=False)

    # verify the systemd config file was removed after uninstall
    assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)

    if container_file_exists(container, NODE_PACKAGE_PATH):
        # verify splunk-otel-js was uninstalled
        assert not node_package_installed(container)


def fluentd_supported(distro, arch):
    if "opensuse" in distro:
        return False
    elif distro == "amazonlinux-2023":
        return False
    elif distro == "debian-bookworm":
        return False
    elif distro == "ubuntu-noble":
        return False
    return True


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("mode", ["agent", "gateway"])
def test_installer_default(distro, arch, mode):
    install_cmd = get_installer_cmd()
    if mode != "agent":
        install_cmd = f"{install_cmd} --mode {mode}"

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        # run installer script
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
            time.sleep(5)

            for pkg in ("td-agent", "splunk-otel-auto-instrumentation"):
                assert not package_is_installed(container, distro, "td-agent"), f"{pkg} was installed"

            assert container.exec_run("systemctl status td-agent").exit_code != 0

            # verify env file created with configured parameters
            verify_env_file(container, mode=mode)

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

            # test support bundle script
            verify_support_bundle(container)

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")


@pytest.mark.installer
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
def test_installer_custom(distro, arch):
    collector_version = "0.126.0"
    service_owner = "test-user"
    custom_config = "/etc/my-custom-config.yaml"

    install_cmd = " ".join((
        get_installer_cmd(),
        "--with-fluentd",
        "--listen-interface 10.0.0.1",
        "--memory 256",
        f"--service-user {service_owner} --service-group {service_owner}",
        f"--collector-config {custom_config}",
        f"--collector-version {collector_version}",
    ))

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    with run_distro_container(distro, arch) as container:
        copy_file_into_container(container, CUSTOM_COLLECTOR_CONFIG, custom_config)
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")

        try:
            # run installer script
            run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
            time.sleep(5)

            # verify collector version
            _, output = run_container_cmd(container, "otelcol --version")
            assert output.decode("utf-8").strip() == f"otelcol version v{collector_version}"

            # verify env file created with configured parameters
            verify_env_file(container, config_path=custom_config, memory="256", listen_addr="10.0.0.1")

            # verify collector service status
            assert wait_for(lambda: service_is_running(container, service_owner=service_owner))

            # verify the default user/group was deleted
            assert container.exec_run(f"getent passwd {SERVICE_OWNER}").exit_code != 0
            assert container.exec_run(f"getent group {SERVICE_OWNER}").exit_code != 0

            # verify the installed directories are owned by test-user
            bundle_owner = container.exec_run("stat -c '%U:%G' /usr/lib/splunk-otel-collector").output.decode("utf-8")
            assert bundle_owner.strip() == f"{service_owner}:{service_owner}"
            config_owner = container.exec_run("stat -c '%U:%G' /etc/otel").output.decode("utf-8")
            assert config_owner.strip() == f"{service_owner}:{service_owner}"

            if fluentd_supported(distro, arch):
                assert package_is_installed(container, distro, "td-agent"), "td-agent was not installed"
                assert container.exec_run("systemctl status td-agent").exit_code == 0

            verify_uninstall(container, distro)

        finally:
            run_container_cmd(container, f"journalctl -u {SERVICE_NAME} --no-pager")
            if fluentd_supported(distro, arch):
                run_container_cmd(container, "journalctl -u td-agent --no-pager")
                if container_file_exists(container, "/var/log/td-agent/td-agent.log"):
                    run_container_cmd(container, "cat /var/log/td-agent/td-agent.log")


def get_instrumentation_dockerfile(distro):
    if distro in INSTR_DEB_DISTROS:
        return INSTR_IMAGES_DIR / "deb" / f"Dockerfile.{distro}"
    else:
        return INSTR_IMAGES_DIR / "rpm" / f"Dockerfile.{distro}"


def get_installed_package_version(container, distro):
    package = "splunk-otel-auto-instrumentation"

    if distro in INSTR_DEB_DISTROS:
        _, output = run_container_cmd(container, f"dpkg-query --showformat='${{Version}}' --show {package}")
    else:
        _, output = run_container_cmd(container, f"rpm -q --queryformat='%{{VERSION}}' {package}")

    return output.decode("utf-8").strip()


def get_zc_method(method, version):
    package = "splunk-otel-auto-instrumentation"
    version = version.replace("~", "-")
    zc_method = rf"{package}-{version}"
    if method == "systemd":
        zc_method = f"{zc_method}-systemd"

    return zc_method


def node_package_installed(container):
    cmd = f"sh -l -c 'cd {NODE_PREFIX} >/dev/null 2>&1 && npm ls --global=false @splunk/otel'"
    print(f"Running '{cmd}':")
    rc, output = container.exec_run(cmd)
    print(output.decode("utf-8"))
    return rc == 0


def verify_dotnet_config(container, path, exists=True):
    for key, val in DOTNET_VARS.items():
        val = val if exists else ".*"
        verify_config_file(container, path, key, val, exists=exists)


@pytest.mark.installer
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in INSTR_DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in INSTR_RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("method", ["preload", "systemd"])
def test_installer_with_instrumentation_default(distro, arch, method):
    # minimum supported node version required for profiling
    node_version = 18
    if arch == "arm64" and distro in ("centos-7"):
        # g++ for these distros is too old to build/compile splunk-otel-js with node v16:
        #   g++: error: unrecognized command line option '-std=gnu++14'
        # use the minimum supported node version without profiling instead
        node_version = 14

    buildargs = {"NODE_VERSION": f"v{node_version}"}

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    dockerfile = get_instrumentation_dockerfile(distro)
    with run_distro_container(distro, dockerfile=dockerfile, arch=arch, buildargs=buildargs) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")
        run_container_cmd(container, f"sh -c 'echo \"# This line should be preserved\" >> {PRELOAD_PATH}'")
        if LOCAL_INSTRUMENTATION_PACKAGE:
            copy_file_into_container(container, LOCAL_INSTRUMENTATION_PACKAGE, f"/test/instrumentation.pkg")

        # set global=true for npm to test that splunk-otel-js is still installed locally
        run_container_cmd(container, "sh -l -c 'npm config set global true'")

        install_cmd = " ".join((
            get_installer_cmd(),
            "--with-systemd-instrumentation" if method == "systemd" else "--with-instrumentation",
        ))
        if LOCAL_INSTRUMENTATION_PACKAGE:
            install_cmd = f"{install_cmd} --instrumentation-version /test/instrumentation.pkg"

        # run installer script
        run_container_cmd(container, install_cmd, env={"VERIFY_ACCESS_TOKEN": "false"}, timeout=INSTALLER_TIMEOUT)
        time.sleep(5)

        version = get_installed_package_version(container, distro)
        zc_method = get_zc_method(method, version)

        # verify env file created with configured parameters
        verify_env_file(container)

        verify_config_file(container, PRELOAD_PATH, "# This line should be preserved", None)

        # verify collector service status
        assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

        # verify splunk-otel-auto-instrumentation is installed
        assert package_is_installed(container, distro, "splunk-otel-auto-instrumentation"), \
            "splunk-otel-auto-instrumentation was not installed"

        assert node_package_installed(container)

        rc, _ = run_container_cmd(container, "sh -l -c 'npm ls --global=true @splunk/otel'", exit_code=None)
        assert rc != 0, "splunk-otel-js installed globally"

        if arch == "amd64":
            assert container_file_exists(container, DOTNET_AGENT_PATH)

        config_attributes = rf"splunk\.zc\.method={zc_method}"

        if method == "preload":
            # verify libsplunk.so was added to /etc/ld.so.preload
            verify_config_file(container, PRELOAD_PATH, LIBSPLUNK_PATH, None)

            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)

            # verify default options for java, node.js, and .NET
            verify_config_file(container, JAVA_ZEROCONFIG_PATH, "JAVA_TOOL_OPTIONS", JAVA_TOOL_OPTIONS)
            verify_config_file(container, NODE_ZEROCONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            configs_to_verify = [JAVA_ZEROCONFIG_PATH, NODE_ZEROCONFIG_PATH]
            if arch == "amd64":
                verify_dotnet_config(container, DOTNET_ZEROCONFIG_PATH)
                configs_to_verify.append(DOTNET_ZEROCONFIG_PATH)
            else:
                assert not container_file_exists(container, DOTNET_ZEROCONFIG_PATH)
            for config_path in configs_to_verify:
                verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", config_attributes)
                verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "false")
                verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "false")
                verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "false")
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_SERVICE_NAME", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_METRICS_EXPORTER", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_LOGS_EXPORTER", ".*", exists=False)
                verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_PROTOCOL", ".*", exists=False)
        else:
            # verify libsplunk.so was not added to /etc/ld.so.preload
            verify_config_file(container, PRELOAD_PATH, f".*{LIBSPLUNK_PATH}.*", None, exists=False)

            # verify default options
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", JAVA_TOOL_OPTIONS)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_RESOURCE_ATTRIBUTES", config_attributes)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_MEMORY_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_METRICS_ENABLED", "false")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_ENDPOINT", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_SERVICE_NAME", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_METRICS_EXPORTER", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_LOGS_EXPORTER", ".*", exists=False)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_PROTOCOL", ".*", exists=False)
            verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=True if arch == "amd64" else False)

        verify_uninstall(container, distro)

        verify_config_file(container, PRELOAD_PATH, "# This line should be preserved", None)


@pytest.mark.installer
@pytest.mark.instrumentation
@pytest.mark.parametrize(
    "distro",
    [pytest.param(distro, marks=pytest.mark.deb) for distro in INSTR_DEB_DISTROS]
    + [pytest.param(distro, marks=pytest.mark.rpm) for distro in INSTR_RPM_DISTROS],
    )
@pytest.mark.parametrize("arch", ["amd64", "arm64"])
@pytest.mark.parametrize("method", ["preload", "systemd"])
@pytest.mark.parametrize("sdk", ["java", "node", "dotnet"])
def test_installer_with_instrumentation_custom(distro, arch, method, sdk):
    # minimum supported node version required for profiling
    node_version = 18
    if arch == "arm64" and distro in ("centos-7"):
        # g++ for these distros is too old to build/compile splunk-otel-js with node v16:
        #   g++: error: unrecognized command line option '-std=gnu++14'
        # use the minimum supported node version without profiling instead
        node_version = 14

    buildargs = {"NODE_VERSION": f"v{node_version}"}

    print(f"Testing installation on {distro} from {STAGE} stage ...")
    dockerfile = get_instrumentation_dockerfile(distro)
    with run_distro_container(distro, dockerfile=dockerfile, arch=arch, buildargs=buildargs) as container:
        copy_file_into_container(container, INSTALLER_PATH, "/test/install.sh")
        run_container_cmd(container, f"sh -c 'echo \"# This line should be preserved\" >> {PRELOAD_PATH}'")
        if LOCAL_INSTRUMENTATION_PACKAGE:
            copy_file_into_container(container, LOCAL_INSTRUMENTATION_PACKAGE, f"/test/instrumentation.pkg")

        # set global=true for npm to test that splunk-otel-js is still installed locally
        run_container_cmd(container, "sh -l -c 'npm config set global true'")

        service_name = f"service_name_from_{method}"
        environment = f"deployment_environment_from_{method}"

        install_cmd = " ".join((
            get_installer_cmd(),
            "--with-systemd-instrumentation" if method == "systemd" else "--with-instrumentation",
            f"--with-instrumentation-sdk {sdk}",
            f"--deployment-environment {environment}",
            f"--service-name {service_name}",
            "--enable-profiler",
            "--enable-profiler-memory",
            "--enable-metrics",
            "--otlp-endpoint http://0.0.0.0:4318",
            "--otlp-endpoint-protocol http/protobuf",
            "--metrics-exporter none",
            "--logs-exporter none",
        ))
        if LOCAL_INSTRUMENTATION_PACKAGE:
            install_cmd = f"{install_cmd} --instrumentation-version /test/instrumentation.pkg"

        # run installer script
        _, output = run_container_cmd(
            container,
            install_cmd,
            env={"VERIFY_ACCESS_TOKEN": "false"},
            exit_code=1 if sdk == "dotnet" and arch != "amd64" else 0,
            timeout=INSTALLER_TIMEOUT,
        )

        if sdk == "dotnet" and arch != "amd64":
            verify_config_file(container, PRELOAD_PATH, f".*{LIBSPLUNK_PATH}.*", None, exists=False)
            verify_config_file(container, PRELOAD_PATH, "# This line should be preserved", None)
            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)
            assert ".NET auto instrumentation is not currently supported" in output.decode("utf-8")
            pytest.xfail("installer script successfully failed for .NET on arm64")

        time.sleep(5)

        version = get_installed_package_version(container, distro)
        zc_method = get_zc_method(method, version)

        # verify env file created with configured parameters
        verify_env_file(container)

        verify_config_file(container, PRELOAD_PATH, "# This line should be preserved", None)

        # verify collector service status
        assert wait_for(lambda: service_is_running(container, service_owner=SERVICE_OWNER))

        # verify splunk-otel-auto-instrumentation is installed
        assert package_is_installed(container, distro, "splunk-otel-auto-instrumentation"), \
            "splunk-otel-auto-instrumentation was not installed"

        if sdk == "node":
            assert node_package_installed(container)
        else:
            assert not node_package_installed(container)

        rc, _ = run_container_cmd(container, "sh -l -c 'npm ls --global=true @splunk/otel'", exit_code=None)
        assert rc != 0, "splunk-otel-js installed globally"

        if arch == "amd64":
            assert container_file_exists(container, DOTNET_AGENT_PATH)

        config_attributes = ",".join((
            rf"splunk\.zc\.method={zc_method}",
            rf"deployment\.environment={environment}",
        ))

        if method == "preload":
            # verify libsplunk.so was added to /etc/ld.so.preload
            verify_config_file(container, PRELOAD_PATH, LIBSPLUNK_PATH, None)

            assert not container_file_exists(container, SYSTEMD_CONFIG_PATH)

            # verify configured options
            if sdk == "java":
                config_path = JAVA_ZEROCONFIG_PATH
                verify_config_file(container, config_path, "JAVA_TOOL_OPTIONS", JAVA_TOOL_OPTIONS)
                assert not container_file_exists(container, NODE_ZEROCONFIG_PATH)
                assert not container_file_exists(container, DOTNET_ZEROCONFIG_PATH)
            elif sdk == "node":
                config_path = NODE_ZEROCONFIG_PATH
                verify_config_file(container, config_path, "NODE_OPTIONS", NODE_OPTIONS)
                assert not container_file_exists(container, JAVA_ZEROCONFIG_PATH)
                assert not container_file_exists(container, DOTNET_ZEROCONFIG_PATH)
            else:
                config_path = DOTNET_ZEROCONFIG_PATH
                verify_dotnet_config(container, config_path)
                assert not container_file_exists(container, JAVA_ZEROCONFIG_PATH)
                assert not container_file_exists(container, NODE_ZEROCONFIG_PATH)

            verify_config_file(container, config_path, "OTEL_RESOURCE_ATTRIBUTES", config_attributes)
            verify_config_file(container, config_path, "SPLUNK_PROFILER_ENABLED", "true")
            verify_config_file(container, config_path, "SPLUNK_PROFILER_MEMORY_ENABLED", "true")
            verify_config_file(container, config_path, "SPLUNK_METRICS_ENABLED", "true")
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_ENDPOINT", "http://0.0.0.0:4318")
            verify_config_file(container, config_path, "OTEL_SERVICE_NAME", service_name)
            verify_config_file(container, config_path, "OTEL_METRICS_EXPORTER", "none")
            verify_config_file(container, config_path, "OTEL_LOGS_EXPORTER", "none")
            verify_config_file(container, config_path, "OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
        else:
            # verify libsplunk.so was not added to /etc/ld.so.preload
            verify_config_file(container, PRELOAD_PATH, f".*{LIBSPLUNK_PATH}.*", None, exists=False)

            # verify configured options
            if sdk == "java":
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", JAVA_TOOL_OPTIONS)
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
            elif sdk == "node":
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", NODE_OPTIONS)
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", ".*", exists=False)
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH, exists=False)
            else:
                verify_dotnet_config(container, SYSTEMD_CONFIG_PATH)
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "JAVA_TOOL_OPTIONS", ".*", exists=False)
                verify_config_file(container, SYSTEMD_CONFIG_PATH, "NODE_OPTIONS", ".*", exists=False)

            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_RESOURCE_ATTRIBUTES", config_attributes)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_PROFILER_MEMORY_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "SPLUNK_METRICS_ENABLED", "true")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_ENDPOINT", "http://0.0.0.0:4318")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_SERVICE_NAME", service_name)
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_METRICS_EXPORTER", "none")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_LOGS_EXPORTER", "none")
            verify_config_file(container, SYSTEMD_CONFIG_PATH, "OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

        verify_uninstall(container, distro)

        verify_config_file(container, PRELOAD_PATH, "# This line should be preserved", None)
