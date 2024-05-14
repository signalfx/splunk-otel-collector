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

import re
import string

PRELOAD_PATH = "/etc/ld.so.preload"
LIB_DIR = "/usr/lib/splunk-instrumentation"
LIBSPLUNK_PATH = f"{LIB_DIR}/libsplunk.so"
INSTRUMENTATION_CONFIG_PATH = f"{LIB_DIR}/instrumentation.conf"
SYSTEMD_CONF_DIR = "/usr/lib/systemd/system.conf.d"
SYSTEMD_CONFIG_PATH = f"{SYSTEMD_CONF_DIR}/00-splunk-otel-auto-instrumentation.conf"

JAVA_AGENT_PATH = f"{LIB_DIR}/splunk-otel-javaagent.jar"
JAVA_CONFIG_PATH = "/etc/splunk/zeroconfig/java.conf"

NODE_AGENT_PATH = f"{LIB_DIR}/splunk-otel-js.tgz"
NODE_PREFIX = f"{LIB_DIR}/splunk-otel-js"
NODE_OPTIONS = f"-r {NODE_PREFIX}/node_modules/@splunk/otel/instrument"
NODE_CONFIG_PATH = "/etc/splunk/zeroconfig/node.conf"

DOTNET_HOME = f"{LIB_DIR}/splunk-otel-dotnet"
DOTNET_AGENT_PATH = string.Template(f"{DOTNET_HOME}/linux-$arch/OpenTelemetry.AutoInstrumentation.Native.so")
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
DOTNET_CONFIG_PATH = "/etc/splunk/zeroconfig/dotnet.conf"


def container_file_exists(container, path):
    return container.exec_run(f"test -f {path}").exit_code == 0


def get_dotnet_agent_path(arch):
    return DOTNET_AGENT_PATH.substitute(arch="x64" if arch == "amd64" else arch)


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


def verify_preload(container, line, exists=True):
    code, output = container.exec_run(f"cat {PRELOAD_PATH}")
    assert code == 0, f"failed to get contents from {PRELOAD_PATH}"
    config = output.decode("utf-8")

    match = re.search(f"^{line}$", config, re.MULTILINE)

    if exists:
        assert match, f"'{line}' not found in {PRELOAD_PATH}"
    else:
        assert not match, f"'{line}' found in {PRELOAD_PATH}"
