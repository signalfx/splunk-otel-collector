# Copyright 2020 Splunk, Inc.
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

import hashlib
import gzip
import os
import re
import sys
import tempfile
import time
import urllib.parse
import xml.etree.ElementTree as ET

import boto3
import requests

from .constants import (
    ARTIFACTORY_DEB_REPO_URL,
    ARTIFACTORY_RPM_REPO_URL,
    CLOUDFRONT_DISTRIBUTION_ID,
    DEFAULT_TIMEOUT,
    EXTENSIONS,
    INSTALLER_SCRIPTS,
    PACKAGE_NAME,
    S3_BUCKET,
    S3_MSI_BASE_DIR,
)


class Asset(object):
    def __init__(self, url=None, path=None):
        assert url or path, "Either url= or path= is required!"
        self.url = url
        self.path = path
        self.name = self._get_name()
        self.component = self._get_component()

    def _get_name(self):
        if self.url:
            return os.path.basename(self.url)
        elif self.path:
            return os.path.basename(self.path)
        return None

    def _get_component(self):
        if self.name:
            ext = os.path.splitext(self.name)[-1].strip(".")
            if ext and ext in EXTENSIONS:
                return ext.lower()
            elif "darwin" in self.name:
                return "osx"
        return None

    def download(self, dest, user=None, token=None, overwrite=False):
        if self.url:
            if not overwrite and os.path.isfile(dest):
                resp = input(f"{dest} already exists.\nOverwrite? [y/N]: ")
                if resp.lower() not in ("y", "yes"):
                    return None
                os.remove(dest)
            download_file(self.url, dest, user, token)
            self.path = dest
            return self.path
        return None

def get_checksum(path, hash_obj):
    with open(path, "rb") as f:
        # Read and update hash string value in blocks of 4K
        for byte_block in iter(lambda: f.read(4096), b""):
            hash_obj.update(byte_block)
        return str(hash_obj.hexdigest()).lower()


def upload_file_to_artifactory(src, dest, user, token):
    print(f"uploading {src} to {dest} ...")

    with open(src, "rb") as fd:
        headers = {"X-Checksum-MD5": get_checksum(src, hashlib.md5())}
        resp = requests.put(dest, auth=(user, token), headers=headers, data=fd)

        assert resp.status_code == 201, f"upload failed:\n{resp.reason}\n{resp.text}"

        return resp

def artifactory_file_exists(url, user, token):
    return requests.head(url, auth=(user, token)).status_code == 200


def download_file(url, dest, user=None, token=None):
    print(f"downloading {url} to {dest} ...")

    resp = requests.get(url, auth=(user, token))

    assert resp.status_code == 200, f"download failed:\n{resp.reason}\n{resp.text}"

    os.makedirs(os.path.dirname(dest), exist_ok=True)

    with open(dest, "wb") as fd:
        fd.write(resp.content)


def get_url_bytes(url, user=None, token=None):
    auth = (user, token) if user and token else None
    resp = requests.get(url, auth=auth, timeout=60)
    assert resp.status_code == 200, f"download failed:\n{resp.reason}\n{resp.text}"
    return resp.content


def delete_artifactory_file(url, user, token):
    print(f"deleting {url} ...")

    resp = requests.delete(url, auth=(user, token))

    assert resp.status_code == 204, f"delete failed:\n{resp.reason}\n{resp.text}"


def verify_bytes_checksum(data, checksum_type, expected_checksum, expected_size=None):
    if expected_size is not None:
        assert len(data) == int(expected_size), (
            f"size mismatch: expected {expected_size}, got {len(data)}"
        )
    hash_name = checksum_type.replace("-", "")
    if hash_name == "sha":
        hash_name = "sha1"
    digest = hashlib.new(hash_name)
    digest.update(data)
    actual_checksum = digest.hexdigest().lower()
    assert actual_checksum == expected_checksum.lower(), (
        f"{checksum_type} mismatch: expected {expected_checksum}, got {actual_checksum}"
    )


def get_deb_package_info(asset):
    match = re.match(r"(?P<name>.+)_(?P<version>[^_]+)_(?P<arch>amd64|arm64)\.deb$", asset.name)
    assert match, f"Failed to get deb package info from {asset.path}!"
    return match.groupdict()


def get_rpm_package_info(asset):
    match = re.match(
        r"(?P<name>.+)-(?P<version>[0-9][^-]*)-(?P<release>[^-]+)\.(?P<arch>x86_64|aarch64)\.rpm$",
        asset.name,
    )
    assert match, f"Failed to get rpm package info from {asset.path}!"
    return match.groupdict()


def parse_deb_release_sha256(release_text):
    entries = {}
    in_sha256 = False
    for line in release_text.splitlines():
        if line == "SHA256:":
            in_sha256 = True
            continue
        if re.match(r"^[A-Za-z0-9-]+:", line):
            in_sha256 = False
        if in_sha256 and line.startswith(" "):
            checksum, size, path = line.split()
            entries[path] = (checksum, int(size))
    return entries


def parse_deb_packages(packages_text):
    packages = []
    current = {}
    for line in packages_text.splitlines():
        if not line:
            if current:
                packages.append(current)
                current = {}
            continue
        if line[0].isspace() or ":" not in line:
            continue
        key, value = line.split(":", 1)
        current[key] = value.strip()
    if current:
        packages.append(current)
    return packages


def get_deb_packages_from_metadata(stage, arch, user, token):
    # Verify the Packages index through the signed Release metadata before
    # using it. This mirrors what apt relies on after Artifactory recalculates
    # repo metadata, and avoids treating a transient metadata update as done.
    release_url = f"{ARTIFACTORY_DEB_REPO_URL}/dists/{stage}/Release"
    release_text = get_url_bytes(release_url, user, token).decode("utf-8")
    checksums = parse_deb_release_sha256(release_text)

    for packages_path in (f"main/binary-{arch}/Packages", f"main/binary-{arch}/Packages.gz"):
        if packages_path not in checksums:
            continue
        packages_url = f"{ARTIFACTORY_DEB_REPO_URL}/dists/{stage}/{packages_path}"
        packages_bytes = get_url_bytes(packages_url, user, token)
        checksum, size = checksums[packages_path]
        verify_bytes_checksum(packages_bytes, "sha256", checksum, size)
        if packages_path.endswith(".gz"):
            packages_bytes = gzip.decompress(packages_bytes)
        return parse_deb_packages(packages_bytes.decode("utf-8"))

    raise AssertionError(f"No Packages metadata found in Release for {stage}/{arch}")


def get_rpm_packages_from_metadata(stage, arch, user, token):
    # Resolve the current primary metadata from repomd.xml and verify its
    # checksum before checking packages. This follows the same repo metadata
    # chain dnf uses, so the release waits for the customer-visible index.
    repomd_url = f"{ARTIFACTORY_RPM_REPO_URL}/{stage}/{arch}/repodata/repomd.xml"
    repomd_bytes = get_url_bytes(repomd_url, user, token)
    repo_ns = "{http://linux.duke.edu/metadata/repo}"
    common_ns = "{http://linux.duke.edu/metadata/common}"

    repomd = ET.fromstring(repomd_bytes)
    primary = None
    for data in repomd.findall(f"{repo_ns}data"):
        if data.attrib.get("type") == "primary":
            primary = data
            break
    assert primary is not None, f"primary metadata not found in {repomd_url}"

    checksum = primary.find(f"{repo_ns}checksum")
    location = primary.find(f"{repo_ns}location")
    assert checksum is not None and location is not None, f"primary checksum/location missing in {repomd_url}"

    repo_root_url = repomd_url.rsplit("/repodata/repomd.xml", 1)[0] + "/"
    primary_url = urllib.parse.urljoin(repo_root_url, location.attrib["href"])
    primary_bytes = get_url_bytes(primary_url, user, token)
    verify_bytes_checksum(primary_bytes, checksum.attrib["type"], checksum.text)
    if primary_url.endswith(".gz"):
        primary_bytes = gzip.decompress(primary_bytes)

    primary_xml = ET.fromstring(primary_bytes)
    packages = []
    for package in primary_xml.findall(f"{common_ns}package"):
        version = package.find(f"{common_ns}version")
        packages.append(
            {
                "name": package.findtext(f"{common_ns}name"),
                "arch": package.findtext(f"{common_ns}arch"),
                "version": version.attrib.get("ver") if version is not None else None,
                "release": version.attrib.get("rel") if version is not None else None,
            }
        )
    return packages


def wait_for_packages_in_metadata(package_type, expected_packages, stage, user, token, timeout=DEFAULT_TIMEOUT):
    print(
        f"Waiting for {len(expected_packages)} {package_type} package(s) "
        f"to appear in {stage} repo metadata ..."
    )
    start_time = time.time()
    last_log = -60

    while True:
        elapsed = int(time.time() - start_time)
        assert elapsed < timeout, (
            f"Timed out after {elapsed}s waiting for {package_type} metadata "
            f"to reference: {expected_packages}"
        )

        try:
            missing = get_missing_packages_from_metadata(package_type, expected_packages, stage, user, token)
        except Exception as err:
            missing = [f"metadata not ready: {err}"]

        if not missing:
            print(f"All expected {package_type} package(s) are present in repo metadata.")
            return

        if elapsed - last_log >= 60:
            print(f"  Still waiting after {elapsed}s. Missing: {missing}")
            last_log = elapsed
        time.sleep(10)


def get_missing_packages_from_metadata(package_type, expected_packages, stage, user, token):
    missing = []
    for arch in sorted({package["arch"] for package in expected_packages}):
        expected_for_arch = [package for package in expected_packages if package["arch"] == arch]
        if package_type == "deb":
            live_packages = get_deb_packages_from_metadata(stage, arch, user, token)
            for expected in expected_for_arch:
                if not any(
                    package.get("Package") == expected["name"]
                    and package.get("Version") == expected["version"]
                    and package.get("Architecture") == expected["arch"]
                    for package in live_packages
                ):
                    missing.append(expected)
        elif package_type == "rpm":
            live_packages = get_rpm_packages_from_metadata(stage, arch, user, token)
            for expected in expected_for_arch:
                if not any(
                    package.get("name") == expected["name"]
                    and package.get("version") == expected["version"]
                    and package.get("release") == expected["release"]
                    and package.get("arch") == expected["arch"]
                    for package in live_packages
                ):
                    missing.append(expected)
        else:
            raise AssertionError(f"Unsupported package metadata type: {package_type}")
    return missing


def upload_package_to_artifactory(
    path,
    dest_url,
    user,
    token,
):
    clean_url = dest_url.split(";")[0]
    if artifactory_file_exists(clean_url, user, token):
        print(f"File already exists at {clean_url}; uploading replacement.")
    upload_file_to_artifactory(path, dest_url, user, token)


def release_deb_to_artifactory(asset, args, wait_for_metadata=True):

    user = args.artifactory_user
    token = args.artifactory_token

    package_info = get_deb_package_info(asset)
    arch = package_info["arch"]

    deb_url = f"{ARTIFACTORY_DEB_REPO_URL}/pool/{args.stage}/{arch}/{asset.name}"
    dest_opts = f"deb.distribution={args.stage};deb.component=main;deb.architecture={arch}"
    dest_url = f"{deb_url};{dest_opts}"

    if not args.force and artifactory_file_exists(deb_url, user, token):
        resp = input(f"{deb_url} already exists.\nOverwrite? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    upload_package_to_artifactory(
        asset.path,
        dest_url,
        user,
        token,
    )

    if wait_for_metadata:
        wait_for_packages_in_metadata("deb", [package_info], args.stage, user, token, args.timeout)

    return package_info


def release_debs_to_artifactory(assets, args):
    expected_packages = []
    for asset in assets:
        expected_packages.append(release_deb_to_artifactory(asset, args, wait_for_metadata=False))

    wait_for_packages_in_metadata(
        "deb",
        expected_packages,
        args.stage,
        args.artifactory_user,
        args.artifactory_token,
        args.timeout,
    )


def release_rpm_to_artifactory(asset, args, wait_for_metadata=True):
    user = args.artifactory_user
    token = args.artifactory_token

    package_info = get_rpm_package_info(asset)
    arch = package_info["arch"]

    dest_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/{asset.name}"

    if not args.force and artifactory_file_exists(dest_url, user, token):
        resp = input(f"{dest_url} already exists.\nOverwrite? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    upload_package_to_artifactory(
        asset.path,
        dest_url,
        user,
        token,
    )

    if wait_for_metadata:
        wait_for_packages_in_metadata("rpm", [package_info], args.stage, user, token, args.timeout)

    return package_info


def release_rpms_to_artifactory(assets, args):
    expected_packages = []
    arch = get_rpm_package_info(assets[0])["arch"]
    for asset in assets:
        package_info = get_rpm_package_info(asset)
        assert package_info["arch"] == arch, "RPM batch must use a single arch"
        expected_packages.append(release_rpm_to_artifactory(asset, args, wait_for_metadata=False))

    wait_for_packages_in_metadata(
        "rpm",
        expected_packages,
        args.stage,
        args.artifactory_user,
        args.artifactory_token,
        args.timeout,
    )


def s3_file_exists(s3_client, path):
    results = s3_client.list_objects(Bucket=S3_BUCKET, Prefix=f"{path}")
    for content in results.get("Contents", []):
        if content.get("Key") == path:
            return True
    return False


def invalidate_cloudfront(paths):
    session = boto3.Session()
    cloudfront = session.client("cloudfront")
    print(f"Invalidating cloudfront for {paths} ...")
    resp = cloudfront.create_invalidation(
        DistributionId=CLOUDFRONT_DISTRIBUTION_ID,
        InvalidationBatch={
            "Paths": {"Quantity": len(paths), "Items": [f"/{path}" for path in paths]},
            "CallerReference": f"splunk-otel-collector-{time.time()}",
        },
    )
    print(resp)
    assert resp.get("ResponseMetadata", {}).get("HTTPStatusCode") == 201, "Failed to submit invalidation request!"


def upload_file_to_s3(local_path, s3_path, force=False):
    session = boto3.Session()
    s3_client = session.client("s3")
    if not force and s3_file_exists(s3_client, s3_path):
        resp = input(f"{S3_BUCKET}/{s3_path} already exists.\nOverwrite [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)
    print(f"Uploading {local_path} to {S3_BUCKET}/{s3_path} ...")
    s3_client.upload_file(local_path, S3_BUCKET, s3_path)

def release_msi_to_s3(asset, args):
    msi_path = asset.path

    s3_path = f"{S3_MSI_BASE_DIR}/{args.stage}/{asset.name}"
    upload_file_to_s3(msi_path, s3_path, force=args.force)

    s3_latest_path = f"{S3_MSI_BASE_DIR}/{args.stage}/latest.txt"
    if not args.not_latest:
        with tempfile.TemporaryDirectory() as tmpdir:
            latest_txt = os.path.join(tmpdir, "latest.txt")
            match = re.match(f"^{PACKAGE_NAME}-(\d+\.\d+\.\d+(\.\d+)?)-(amd64|arm64)\.msi$", asset.name)
            assert match, f"Failed to get version from '{asset.name}'!"
            msi_version = match.group(1)
            with open(latest_txt, "w") as fd:
                fd.write(msi_version)
            print(f"Updating {S3_BUCKET}/{s3_latest_path} for version '{msi_version}' ...")
            upload_file_to_s3(latest_txt, s3_latest_path, force=True)

    invalidate_cloudfront([s3_path, s3_latest_path])

def release_installers_to_s3(force=False):
    if not force:
        resp = input("Releasing installer scripts to S3:\nContinue [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    for s3_path, local_path in INSTALLER_SCRIPTS.items():
        assert os.path.isfile(local_path), f"{local_path} not found!"
        upload_file_to_s3(str(local_path), s3_path, force=force)

    invalidate_cloudfront(INSTALLER_SCRIPTS.keys())
