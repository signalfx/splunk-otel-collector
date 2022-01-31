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
import tarfile
import time
from contextlib import contextmanager
from io import BytesIO
from pathlib import Path

import docker

TESTS_DIR = Path(__file__).parent.parent.resolve()
REPO_DIR = TESTS_DIR.parent.parent.parent.parent.resolve()
DEB_DISTROS = [df.split(".")[-1] for df in glob.glob(str(TESTS_DIR / "images" / "deb" / "Dockerfile.*"))]
RPM_DISTROS = [df.split(".")[-1] for df in glob.glob(str(TESTS_DIR / "images" / "rpm" / "Dockerfile.*"))]
SERVICE_NAME = "splunk-otel-collector"
SERVICE_OWNER = "splunk-otel-collector"
OTELCOL_BIN = "/usr/bin/otelcol"
DEFAULT_TIMEOUT = 10


def retry(function, exception, max_attempts=5, interval=5):
    for attempt in range(max_attempts):
        try:
            return function()
        except exception as e:
            assert attempt < (max_attempts - 1), "%s failed after %d attempts!\n%s" % (function, max_attempts, str(e))
        time.sleep(interval)


@contextmanager
def run_distro_container(distro, dockerfile=None, path=TESTS_DIR, buildargs=None, timeout=DEFAULT_TIMEOUT):
    client = docker.from_env()

    if not dockerfile:
        if distro in DEB_DISTROS:
            dockerfile = TESTS_DIR / "images" / "deb" / f"Dockerfile.{distro}"
        else:
            dockerfile = TESTS_DIR / "images" / "rpm" / f"Dockerfile.{distro}"

    assert os.path.isfile(str(dockerfile)), f"{dockerfile} not found!"

    image, _ = retry(
        lambda: client.images.build(path=str(path), dockerfile=str(dockerfile), pull=True, rm=True, forcerm=True, buildargs=buildargs),
        docker.errors.BuildError,
    )

    if (distro.find("windows") == -1):
        container = client.containers.create(
            image.id, detach=True, privileged=True, volumes={"/sys/fs/cgroup": {"bind": "/sys/fs/cgroup", "mode": "ro"}}
        )
    else:
        container = client.containers.create(
            image.id, detach=True
        )

    try:
        container.start()

        start_time = time.time()
        while True:
            container.reload()
            if (distro.find("windows") == -1) and container.attrs["NetworkSettings"]["IPAddress"]:
                break
            elif container.attrs["NetworkSettings"]["Networks"]["nat"]["IPAddress"]:
                break
            assert (time.time() - start_time) < timeout, "timed out waiting for container to start"

        yield container
    finally:
        container.remove(force=True, v=True)


def run_container_cmd(container, cmd, env=None, exit_code=0):
    print(f"Running '{cmd}' ...")
    code, output = container.exec_run(cmd, environment=env)
    print(output.decode("utf-8"))
    if exit_code is not None:
        assert code == exit_code
    return code, output

def copy_file_into_win_container(container, path, target_path):
    os.chdir(os.path.dirname(path))
    srcname = os.path.basename(path)
    with tarfile.open("temp.tar", 'w') as tar:
        try:
            tar.add(srcname)
        finally:
            tar.close()
    with open('temp.tar', 'rb') as fd:
        ok = container.put_archive(path=target_path, data=fd)
        if not ok:
            raise Exception('Put file failed')

def copy_file_into_container(container, path, target_path, size=None):
    with open(path, "rb") as fd:
        tario = BytesIO()
        tar = tarfile.TarFile(fileobj=tario, mode="w")

        info = tarfile.TarInfo(name=target_path)
        if size is None:
            size = os.fstat(fd.fileno()).st_size
        info.size = size

        tar.addfile(info, fd)

        tar.close()

        container.put_archive("/", tario.getvalue())

        time.sleep(2)


def wait_for(test, timeout=DEFAULT_TIMEOUT, interval=1):
    start_time = time.time()

    while (time.time() - start_time) < timeout:
        try:
            if test():
                return True
        except AssertionError:
            pass
        time.sleep(interval)

    return False


def ensure_always(test, timeout=DEFAULT_TIMEOUT, interval=1):
    start_time = time.time()

    while (time.time() - start_time) < timeout:
        try:
            if test():
                time.sleep(interval)
            else:
                return False
        except AssertionError:
            return False

    return True


def service_is_running(container, service_name=SERVICE_NAME, service_owner=SERVICE_OWNER, process=OTELCOL_BIN):
    cmd = f"sh -ec 'systemctl status {service_name} && pgrep -a -u {service_owner} -f {process}'"
    code, _ = run_container_cmd(container, cmd, exit_code=None)
    return code == 0

def service_is_running_win(container, service_name=SERVICE_NAME):
    cmd = f"powershell \"(Get-Service {service_name}).Status | FINDSTR 'Running'\""
    code, _ = run_container_cmd(container, cmd, exit_code=None)
    return code == 0
