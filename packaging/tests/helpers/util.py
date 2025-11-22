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
from typing import Optional, Tuple

import docker

TESTS_DIR = Path(__file__).parent.parent.resolve()
REPO_DIR = TESTS_DIR.parent.parent.resolve()
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
            if attempt < (max_attempts - 1):
                print(f"Attempt {attempt + 1}/{max_attempts} failed: {str(e)}")
                print(f"Retrying in {interval} seconds...")
            else:
                error_msg = f"{function} failed after {max_attempts} attempts!\n{str(e)}"
                error_msg += "\n\nTroubleshooting tips:"
                error_msg += "\n1. Check your internet connection"
                error_msg += "\n2. Verify Docker is running: docker ps"
                error_msg += "\n3. Test Docker registry access: docker pull ubuntu:22.04"
                error_msg += "\n4. Check Docker proxy settings: docker info | grep -i proxy"
                error_msg += "\n5. Try restarting Docker Desktop"
                assert False, error_msg
        time.sleep(interval)


def wait_for_container_cmd(container, cmd, timeout=DEFAULT_TIMEOUT):
    start_time = time.time()
    while True:
        code, output = container.exec_run(cmd)
        elapsed = time.time() - start_time
        if code == 0:
            print(f"'{cmd}' completed in {elapsed}s:\n{output.decode('utf-8')}")
            return code, output
        output_str = output.decode('utf-8', errors='ignore')
        if 'exec format error' in output_str.lower():
            error_msg = f"Architecture mismatch detected: 'exec format error'\n"
            error_msg += f"This usually means Docker Desktop on ARM64 Mac doesn't have emulation enabled.\n"
            error_msg += f"To fix:\n"
            error_msg += f"1. Open Docker Desktop\n"
            error_msg += f"2. Go to Settings > General\n"
            error_msg += f"3. Enable 'Use Rosetta for x86/amd64 emulation on Apple Silicon'\n"
            error_msg += f"4. Restart Docker Desktop\n"
            error_msg += f"5. Verify: docker run --platform linux/amd64 --rm ubuntu:22.04 uname -m\n"
            error_msg += f"   (should output 'x86_64', not 'exec format error')\n"
            assert False, error_msg
        assert elapsed < timeout, f"timed out waiting for '{cmd}':\n{output_str}"
        time.sleep(1)


@contextmanager
def run_distro_container(distro, arch="amd64", dockerfile=None, path=TESTS_DIR, buildargs=None, timeout=DEFAULT_TIMEOUT):
    client = docker.from_env()

    # Check if Docker can run amd64 containers on ARM64 Mac (requires Rosetta emulation)
    import platform as plat
    host_arch = plat.machine().lower()
    if arch == "amd64" and host_arch in ["arm64", "aarch64"]:
        print(f"Detected ARM64 host ({host_arch}) running amd64 container - checking emulation support...")
        try:
            test_container = client.containers.run(
                "ubuntu:22.04",
                command=["uname", "-m"],
                platform="linux/amd64",
                remove=True,
                stdout=True,
                stderr=True,
            )
            result = test_container.decode('utf-8').strip()
            if result != "x86_64":
                error_msg = f"Docker emulation check failed: expected 'x86_64', got '{result}'\n"
                error_msg += "\nDocker Desktop on ARM64 Mac doesn't have amd64 emulation enabled.\n"
                error_msg += "To fix:\n"
                error_msg += "1. Open Docker Desktop\n"
                error_msg += "2. Go to Settings > General\n"
                error_msg += "3. Enable 'Use Rosetta for x86/amd64 emulation on Apple Silicon'\n"
                error_msg += "4. Restart Docker Desktop\n"
                error_msg += "5. Verify: docker run --platform linux/amd64 --rm ubuntu:22.04 uname -m\n"
                raise RuntimeError(error_msg)
            print(f"✓ Docker emulation verified: {result}")
        except docker.errors.ContainerError as e:
            if "exec format error" in str(e).lower():
                error_msg = "Docker Desktop on ARM64 Mac doesn't have amd64 emulation enabled.\n"
                error_msg += "\nTo fix:\n"
                error_msg += "1. Open Docker Desktop\n"
                error_msg += "2. Go to Settings > General\n"
                error_msg += "3. Enable 'Use Rosetta for x86/amd64 emulation on Apple Silicon'\n"
                error_msg += "4. Restart Docker Desktop\n"
                error_msg += "5. Verify: docker run --platform linux/amd64 --rm ubuntu:22.04 uname -m\n"
                raise RuntimeError(error_msg)
            raise
        except Exception as e:
            print(f"Warning: Could not verify Docker emulation: {e}")
            print("Continuing anyway...")

    if not dockerfile:
        if distro in DEB_DISTROS:
            dockerfile = TESTS_DIR / "images" / "deb" / f"Dockerfile.{distro}"
        elif distro in RPM_DISTROS:
            dockerfile = TESTS_DIR / "images" / "rpm" / f"Dockerfile.{distro}"
        else:
            dockerfile = TESTS_DIR / "images" / "tar" / f"Dockerfile.{distro}"

    assert os.path.isfile(str(dockerfile)), f"{dockerfile} not found!"

    print(f"Building {dockerfile} ...")

    if buildargs is None:
        buildargs = {}
    buildargs["TARGETARCH"] = arch

    # Try to pull base image first to avoid network issues during build
    try:
        with open(dockerfile, 'r') as f:
            for line in f:
                line = line.strip()
                if line.startswith('FROM'):
                    # Extract base image (handles FROM image:tag and FROM image)
                    parts = line.split()
                    if len(parts) >= 2:
                        base_image = parts[1]
                        print(f"Pre-pulling base image: {base_image}")
                        try:
                            client.images.pull(base_image, platform=f"linux/{arch}")
                            print(f"Successfully pulled {base_image}")
                        except Exception as e:
                            print(f"Warning: Could not pre-pull {base_image}: {e}")
                            print("Continuing with build (will pull during build)...")
                    break
    except Exception as e:
        print(f"Warning: Could not determine base image: {e}")
        print("Continuing with build...")

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

        # increase default timeout for emulated containers (qemu/rosetta)
        # On ARM64 Mac running amd64 containers, emulation is slower
        import platform as plat
        host_arch = plat.machine().lower()
        is_emulated = (arch == "amd64" and host_arch in ["arm64", "aarch64"]) or (arch != "amd64" and timeout == DEFAULT_TIMEOUT)
        if is_emulated and timeout == DEFAULT_TIMEOUT:
            timeout = DEFAULT_TIMEOUT * 5  # Increase timeout for emulated containers
            print(f"Using extended timeout ({timeout}s) for emulated {arch} container on {host_arch} host")

        print("Waiting for container to be ready ...")

        start_time = time.time()
        while True:
            container.reload()
            ip_address = container.attrs["NetworkSettings"]["IPAddress"]
            if ip_address:
                print(f"Container started with IP: {ip_address}")
                break
            elapsed = time.time() - start_time
            if elapsed >= timeout:
                # Get container logs for debugging
                try:
                    logs = container.logs(tail=50).decode('utf-8', errors='ignore')
                    error_msg = f"timed out waiting for container to start (waited {elapsed:.1f}s)\n"
                    error_msg += f"Container status: {container.status}\n"
                    error_msg += f"Last 50 lines of container logs:\n{logs}"
                except Exception as e:
                    error_msg = f"timed out waiting for container to start (waited {elapsed:.1f}s)\n"
                    error_msg += f"Container status: {container.status}\n"
                    error_msg += f"Could not retrieve logs: {e}"
                assert False, error_msg
            time.sleep(1)

        # qemu is slow, so wait for systemd to be ready
        wait_for_container_cmd(container, "systemctl show-environment", timeout=timeout)

        yield container
    finally:
        container.remove(force=True, v=True)


def run_container_cmd(container, cmd, env=None, exit_code:Optional[int]=0, timeout=None, user='', workdir=None) -> Tuple[int, bytes]:
    if timeout:
        cmd = f"timeout {timeout} {cmd}"
    print(f"Running '{cmd}' ...")
    code, output = container.exec_run(cmd, environment=env, user=user, workdir=workdir)
    print(output.decode("utf-8"))
    if exit_code is not None:
        assert code == exit_code
    return code, output


def copy_file_into_container(container, path, target_path, size=None):
    """Copy a file from the host into the container using Docker's archive API."""
    target_path = os.path.normpath(target_path)
    rel_path = target_path.lstrip("/")
    assert rel_path, f"Invalid target_path: {target_path}"
    target_dir = os.path.dirname(target_path)

    if target_dir not in ("", "/"):
        code, output = container.exec_run(f"mkdir -p {target_dir}")
        assert code == 0, f"Failed to create directory {target_dir}:\n{output.decode('utf-8', errors='ignore')}"

    with open(path, "rb") as fd:
        tario = BytesIO()
        with tarfile.open(fileobj=tario, mode="w") as tar:
            info = tarfile.TarInfo(name=rel_path)
            if size is None:
                size = os.fstat(fd.fileno()).st_size
            info.size = size
            tar.addfile(info, fd)

        tario.seek(0)
        assert container.put_archive("/", tario.getvalue()), \
            f"Failed to copy {path} to {target_path}"

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
    cmd = f"""sh -ec 'systemctl status {service_name}'"""
    sysctl_code, _ = run_container_cmd(container, cmd, exit_code=None)
    cmd = f"""sh -ec 'pgrep -a -u {service_owner} -f {process}'"""
    pgrep_code, _ = run_container_cmd(container, cmd, exit_code=None)
    return (sysctl_code | pgrep_code) == 0


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
