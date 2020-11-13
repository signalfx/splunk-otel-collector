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
def run_distro_container(distro, timeout=DEFAULT_TIMEOUT):
    client = docker.from_env()

    assert distro in DEB_DISTROS + RPM_DISTROS, f"'{distro}' distro not supported!"

    if distro in DEB_DISTROS:
        dockerfile = TESTS_DIR / "images" / "deb" / f"Dockerfile.{distro}"
    else:
        dockerfile = TESTS_DIR / "images" / "rpm" / f"Dockerfile.{distro}"

    image, _ = retry(
        lambda: client.images.build(path=str(TESTS_DIR), dockerfile=str(dockerfile), pull=True, rm=True, forcerm=True),
        docker.errors.BuildError,
    )

    container = client.containers.create(
        image.id, detach=True, privileged=True, volumes={"/sys/fs/cgroup": {"bind": "/sys/fs/cgroup", "mode": "ro"}}
    )

    try:
        container.start()

        start_time = time.time()
        while True:
            container.reload()
            if container.attrs["NetworkSettings"]["IPAddress"]:
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
