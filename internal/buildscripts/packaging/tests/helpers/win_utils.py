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

import subprocess
import winreg

WIN_REGISTRY = winreg.HKEY_LOCAL_MACHINE
WIN_REGISTRY_KEY = r"SYSTEM\CurrentControlSet\Control\Session Manager\Environment"


def run_win_command(cmd, returncodes=None, shell=True, **kwargs):
    if returncodes is None:
        returncodes = [0]
    print('running "%s" ...' % cmd)
    # pylint: disable=subprocess-run-check
    proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, shell=shell, close_fds=False, **kwargs)
    output = proc.stdout.decode("utf-8")
    if returncodes:
        assert proc.returncode in returncodes, output
    print(output.encode('cp1252', errors='ignore'))
    return proc


def has_choco():
    return run_win_command("choco --version", []).returncode == 0


def get_registry_value(name, registry=WIN_REGISTRY, key=WIN_REGISTRY_KEY):
    access_key = winreg.OpenKeyEx(registry, key)
    return winreg.QueryValueEx(access_key, name)[0]
