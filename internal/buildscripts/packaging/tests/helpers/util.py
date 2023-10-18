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
TAR_DISTROS = [df.split(".")[-1] for df in glob.glob(str(TESTS_DIR / "images" / "tar" / "Dockerfile.*"))]
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


def wait_for_container_cmd(container, cmd, timeout=DEFAULT_TIMEOUT):
    start_time = time.time()
    while True:
        code, output = container.exec_run(cmd)
        elapsed = time.time() - start_time
        if code == 0:
            print(f"'{cmd}' completed in {elapsed}s:\n{output.decode('utf-8')}")
            return code, output
        assert elapsed < timeout, f"timed out waiting for '{cmd}':\n{output.decode('utf-8')}"
        time.sleep(1)


@contextmanager
def run_distro_container(distro, arch="amd64", dockerfile=None, path=TESTS_DIR, buildargs=None, timeout=DEFAULT_TIMEOUT):
    client = docker.from_env()

    if not dockerfile:
        if distro in DEB_DISTROS:
            dockerfile = TESTS_DIR / "images" / "deb" / f"Dockerfile.{distro}"
        elif distro in RPM_DISTROS:
            dockerfile = TESTS_DIR / "images" / "rpm" / f"Dockerfile.{distro}"
        else:
            dockerfile = TESTS_DIR / "images" / "tar" / f"Dockerfile.{distro}"

    assert os.path.isfile(str(dockerfile)), f"{dockerfile} not found!"

    print(f"Building {dockerfile} ...")

    image, _ = retry(
        lambda: client.images.build(
            path=str(path),
            dockerfile=str(dockerfile),
            pull=True,
            rm=True,
            forcerm=True,
            platform=f"linux/{arch}",
            buildargs=buildargs,
        ),
        docker.errors.BuildError,
    )

    container = client.containers.create(
        image.id,
        detach=True,
        privileged=True,
        volumes={"/sys/fs/cgroup": {"bind": "/sys/fs/cgroup", "mode": "rw"}},
        platform=f"linux/{arch}",
        cgroupns="host",
    )

    try:
        container.start()

        # increase default timeout for qemu
        if arch != "amd64" and timeout == DEFAULT_TIMEOUT:
            timeout = DEFAULT_TIMEOUT * 3

        print("Waiting for container to be ready ...")

        start_time = time.time()
        while True:
            container.reload()
            if container.attrs["NetworkSettings"]["IPAddress"]:
                break
            assert (time.time() - start_time) < timeout, "timed out waiting for container to start"
            time.sleep(1)

        # qemu is slow, so wait for systemd to be ready
        wait_for_container_cmd(container, "systemctl show-environment", timeout=timeout)

        yield container
    finally:
        container.remove(force=True, v=True)


def run_container_cmd(container, cmd, env=None, exit_code=0, timeout=None):
    if timeout:
        cmd = f"timeout {timeout} {cmd}"
    print(f"Running '{cmd}' ...")
    code, output = container.exec_run(cmd, environment=env)
    print(output.decode("utf-8"))
    if exit_code is not None:
        assert code == exit_code
    return code, output


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


def verify_package_version(container, package, version, old_version=None):
    if container.exec_run("bash -c 'command -v dpkg-query'").exit_code == 0:
        _, output = run_container_cmd(container, f"dpkg-query --showformat='${{Version}}' --show {package}")
    else:
        _, output = run_container_cmd(container, f"rpm -q --queryformat '%{{VERSION}}' {package}")
    installed_version = output.decode("utf-8").strip()
    assert installed_version, f"failed to get {package} version"
    if version == "latest" and old_version:
        assert tuple(installed_version.split(".")) > tuple(old_version.split(".")), \
            f"installed version = {installed_version}, expected version > {old_version}"
    elif version != "latest":
        assert installed_version == version, \
            f"installed version = {installed_version}, expected version = {version}"
