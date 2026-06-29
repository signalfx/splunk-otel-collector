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

import gzip
import hashlib
import os
import re
import sys
import tempfile
import time
import xml.etree.ElementTree as ET

import boto3
import requests

from .constants import (
    ARTIFACTORY_API_URL,
    ARTIFACTORY_DEB_REPO,
    ARTIFACTORY_DEB_REPO_URL,
    ARTIFACTORY_RPM_REPO,
    ARTIFACTORY_RPM_REPO_URL,
    ARTIFACTORY_URL,
    CLOUDFRONT_DISTRIBUTION_ID,
    DEFAULT_TIMEOUT,
    EXTENSIONS,
    INSTALLER_SCRIPTS,
    METADATA_SETTLE_DELAY,
    PACKAGE_NAME,
    REPO_DIR,
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


def delete_artifactory_file(url, user, token):
    print(f"deleting {url} ...")

    resp = requests.delete(url, auth=(user, token))

    assert resp.status_code == 204, f"delete failed:\n{resp.reason}\n{resp.text}"


def get_md5_from_artifactory(url, user, token):
    if not artifactory_file_exists(url, user, token):
        return None

    resp = requests.get(url, auth=(user, token))

    assert resp.status_code == 200, f"md5 request failed:\n{resp.reason}\n{resp.text}"

    md5 = resp.json().get("checksums", {}).get("md5", "")

    assert md5, f"md5 not found in response:\n{resp.text}"

    return md5


def trigger_metadata_calculation(
    calculate_url, user, token, sync=False, timeout=DEFAULT_TIMEOUT, params=None
):
    params = dict(params or {})
    params.setdefault("async", "0" if sync else "1")
    print(f"Triggering metadata calculation: {calculate_url} params={params}")
    resp = requests.post(calculate_url, auth=(user, token), params=params, timeout=timeout)
    print(f"Calculation response: {resp.status_code} {resp.text.strip()}")
    resp.raise_for_status()


def wait_for_artifactory_metadata(
    url, orig_md5, user, token, timeout=DEFAULT_TIMEOUT, settle_delay=METADATA_SETTLE_DELAY
):
    print(f"Waiting for {url} to be updated (original MD5: {orig_md5}) ...")

    start_time = time.time()

    # wait for the metadata to change from the pre-upload snapshot.
    last_md5 = orig_md5
    while True:
        elapsed = int(time.time() - start_time)
        assert elapsed < timeout, (
            f"Timed out after {elapsed}s waiting for {url} to be updated "
            f"(MD5 still {orig_md5})"
        )

        new_md5 = get_md5_from_artifactory(url, user, token)

        if new_md5 and str(orig_md5).lower() != str(new_md5).lower():
            last_md5 = new_md5
            print(
                f"Metadata updated after {int(time.time() - start_time)}s "
                f"(new MD5: {new_md5}); reconfirming it has settled ..."
            )
            break

        if elapsed > 0 and elapsed % 60 == 0:
            print(f"  Still waiting after {elapsed}s (MD5 unchanged: {new_md5}) ...")

        time.sleep(5)

    # Treat the first metadata change as provisional until it stays stable.
    while True:
        elapsed = int(time.time() - start_time)
        assert elapsed + settle_delay <= timeout, (
            f"Timed out after {elapsed}s waiting for {url} to settle "
            f"(last seen MD5: {last_md5})"
        )

        print(f"Waiting {settle_delay}s to confirm the metadata has settled ...")
        time.sleep(settle_delay)
        new_md5 = get_md5_from_artifactory(url, user, token)
        if new_md5 and str(new_md5).lower() == str(last_md5).lower():
            print(
                f"Metadata settled after {int(time.time() - start_time)}s "
                f"(unchanged for {settle_delay}s, MD5: {new_md5})"
            )
            return
        print(
            f"Metadata changed again during settle window "
            f"(MD5 {last_md5} -> {new_md5}); reconfirming ..."
        )
        last_md5 = new_md5


def wait_for_package_in_metadata(package_present_check, package_name, timeout=DEFAULT_TIMEOUT):
    print(f"Waiting for {package_name} to appear in live repo metadata ...")
    start_time = time.time()

    while True:
        present = package_present_check()
        if present is True:
            print(f"{package_name} is present in live repo metadata.")
            return

        elapsed = int(time.time() - start_time)
        assert elapsed < timeout, (
            f"Timed out after {elapsed}s waiting for {package_name} to appear "
            f"in live repo metadata (last result: {present})"
        )

        if elapsed > 0 and elapsed % 60 == 0:
            print(
                f"  Still waiting after {elapsed}s for {package_name} "
                f"to appear in live repo metadata (last result: {present}) ..."
            )

        time.sleep(5)


def rpm_package_in_metadata(stage, arch, package_name, user, token):
    # Return True/False if package_name (the rpm filename) is listed in the live
    # repodata primary.xml, or None if it can't be determined.
    repomd_url = f"{ARTIFACTORY_RPM_REPO_URL}/{stage}/{arch}/repodata/repomd.xml"
    resp = requests.get(repomd_url, auth=(user, token))
    if resp.status_code != 200:
        print(f"Could not fetch {repomd_url}: {resp.status_code}")
        return None

    try:
        repomd = ET.fromstring(resp.content)
    except ET.ParseError as err:
        print(f"Could not parse {repomd_url}: {err}")
        return None

    primary_href = None
    for data in repomd.findall(".//{*}data"):
        if data.attrib.get("type") == "primary":
            location = data.find("{*}location")
            if location is not None:
                primary_href = location.attrib.get("href")
            break

    if not primary_href:
        print(f"Could not find primary metadata href in {repomd_url}")
        return None

    if primary_href.startswith("http://") or primary_href.startswith("https://"):
        primary_url = primary_href
    else:
        primary_url = f"{ARTIFACTORY_RPM_REPO_URL}/{stage}/{arch}/{primary_href}"

    resp = requests.get(primary_url, auth=(user, token))
    if resp.status_code != 200:
        print(f"Could not fetch {primary_url}: {resp.status_code}")
        return None

    primary_content = resp.content
    if primary_href.endswith(".gz"):
        try:
            primary_content = gzip.decompress(primary_content)
        except OSError as err:
            print(f"Could not decompress {primary_url}: {err}")
            return None

    try:
        primary = ET.fromstring(primary_content)
    except ET.ParseError as err:
        print(f"Could not parse {primary_url}: {err}")
        return None

    for location in primary.findall(".//{*}location"):
        href = location.attrib.get("href", "")
        if os.path.basename(href) == package_name:
            return True
    return False


def deb_package_in_metadata(stage, arch, package_name, user, token):
    # Return True/False if package_name (the .deb filename) is listed in the live
    # Packages index for this distribution/arch, or None if it can't be
    # determined.
    base = f"{ARTIFACTORY_DEB_REPO_URL}/dists/{stage}/main/binary-{arch}"
    text = None
    resp = requests.get(f"{base}/Packages.gz", auth=(user, token))
    if resp.status_code == 200:
        try:
            text = gzip.decompress(resp.content).decode("utf-8", "replace")
        except OSError:
            text = None
    if text is None:
        resp = requests.get(f"{base}/Packages", auth=(user, token))
        if resp.status_code == 200:
            text = resp.text
    if text is None:
        return None
    # Packages lists each entry as "Filename: pool/<stage>/<arch>/<file>.deb"
    return any(
        line.startswith("Filename:") and line.strip().endswith(package_name)
        for line in text.splitlines()
    )


def upload_package_to_artifactory(
    path,
    dest_url,
    user,
    token,
    metadata_api_url,
    metadata_url,
    sign_metadata=True,
    timeout=DEFAULT_TIMEOUT,
    calculate_url=None,
    calculate_params=None,
    package_present_check=None,
    sync_calculate=False,
):
    local_md5 = get_checksum(path, hashlib.md5())
    clean_url = dest_url.split(";")[0]
    storage_api_url = clean_url.replace(
        ARTIFACTORY_URL + "/", ARTIFACTORY_API_URL + "/storage/", 1
    )

    pre_upload_md5 = None
    if artifactory_file_exists(clean_url, user, token):
        pre_upload_md5 = get_md5_from_artifactory(storage_api_url, user, token)
        print(f"Pre-upload file MD5 (remote): {pre_upload_md5}")
    else:
        print(f"File does not yet exist at {clean_url}")

    print(f"Local file MD5:                {local_md5}")
    content_changed = pre_upload_md5 is None or pre_upload_md5.lower() != local_md5.lower()
    print(f"Content changed:               {content_changed}")

    orig_md5 = None
    if sign_metadata:
        orig_md5 = get_md5_from_artifactory(metadata_api_url, user, token)
        print(f"Pre-upload metadata MD5:       {orig_md5}")

    upload_file_to_artifactory(path, dest_url, user, token)

    if sign_metadata:
        do_wait = content_changed
        if not content_changed:
            # Existing bytes do not prove live metadata references the package.
            present = package_present_check() if package_present_check else None
            if present is False:
                print(
                    "Upload content is identical to the existing file, but the "
                    "package is NOT present in the live repo metadata; the index "
                    "is stale. Forcing metadata regeneration before signing."
                )
                do_wait = True
            elif present is True:
                print(
                    "Upload content is identical and the package is already present "
                    "in the live repo metadata."
                )
                do_wait = False
            else:
                print(
                    "Upload content is identical to existing file and metadata "
                    "presence could not be confirmed."
                )
                do_wait = True
        if do_wait:
            if calculate_url:
                trigger_metadata_calculation(
                    calculate_url,
                    user,
                    token,
                    sync=sync_calculate,
                    timeout=timeout,
                    params=calculate_params,
                )
                current_md5 = get_md5_from_artifactory(metadata_api_url, user, token)
                if current_md5 and str(orig_md5).lower() != str(current_md5).lower():
                    wait_for_artifactory_metadata(
                        metadata_api_url, orig_md5, user, token, timeout=timeout
                    )
                else:
                    print(
                        "Metadata MD5 did not change after calculation; "
                        "checking package presence before signing."
                    )
            else:
                wait_for_artifactory_metadata(
                    metadata_api_url, orig_md5, user, token, timeout=timeout
                )
        if package_present_check:
            wait_for_package_in_metadata(
                package_present_check, os.path.basename(path), timeout=timeout
            )
        dest = os.path.join(REPO_DIR, os.path.basename(metadata_url))
        download_file(metadata_url, dest, user, token)


def release_deb_to_artifactory(asset, args):

    user = args.artifactory_user
    token = args.artifactory_token

    match = re.match(r".*_(amd64|arm64)\.deb$", asset.name)
    assert match, f"Failed to get arch from {asset.path}!"
    arch = match.groups()[0]

    metadata_api_url = f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_DEB_REPO}/dists/{args.stage}/Release"
    metadata_url = f"{ARTIFACTORY_DEB_REPO_URL}/dists/{args.stage}/Release"
    deb_url = f"{ARTIFACTORY_DEB_REPO_URL}/pool/{args.stage}/{arch}/{asset.name}"
    dest_opts = f"deb.distribution={args.stage};deb.component=main;deb.architecture={arch}"
    dest_url = f"{deb_url};{dest_opts}"

    if not args.force and artifactory_file_exists(deb_url, user, token):
        resp = input(f"{deb_url} already exists.\nOverwrite? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    calculate_url = f"{ARTIFACTORY_API_URL}/deb/reindex/{ARTIFACTORY_DEB_REPO}"

    upload_package_to_artifactory(
        asset.path,
        dest_url,
        user,
        token,
        metadata_api_url,
        metadata_url,
        sign_metadata=True,
        timeout=args.timeout,
        calculate_url=calculate_url,
        package_present_check=lambda: deb_package_in_metadata(
            args.stage, arch, asset.name, user, token
        ),
        sync_calculate=getattr(args, "sync_calculate_metadata", False),
    )


def get_rpm_arch(asset):
    match = re.match(r".*\.(x86_64|aarch64)\.rpm$", asset.name)
    assert match, f"Failed to get arch from {asset.path}!"
    return match.groups()[0]


def release_rpm_to_artifactory(asset, args):
    release_rpms_to_artifactory([asset], args)


def release_rpms_to_artifactory(assets, args):
    assert assets, "No RPM assets specified"

    user = args.artifactory_user
    token = args.artifactory_token

    arch = get_rpm_arch(assets[0])
    for asset in assets:
        assert get_rpm_arch(asset) == arch, "RPM batch must use a single arch"

    metadata_api_url = f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_RPM_REPO}/{args.stage}/{arch}/repodata/repomd.xml"
    metadata_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/repodata/repomd.xml"

    # Auto Calculate RPM Metadata is disabled for this repo, so the release
    # pipeline recalculates once after all RPMs are uploaded.
    calculate_url = f"{ARTIFACTORY_API_URL}/yum/{ARTIFACTORY_RPM_REPO}"
    calculate_params = {"async": "0", "path": f"{args.stage}/{arch}"}

    orig_md5 = get_md5_from_artifactory(metadata_api_url, user, token)
    print(f"Pre-upload metadata MD5:       {orig_md5}")

    for asset in assets:
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
            metadata_api_url,
            metadata_url,
            sign_metadata=False,
            timeout=args.timeout,
        )

    trigger_metadata_calculation(
        calculate_url,
        user,
        token,
        sync=True,
        timeout=args.timeout,
        params=calculate_params,
    )

    current_md5 = get_md5_from_artifactory(metadata_api_url, user, token)
    if current_md5 and str(orig_md5).lower() != str(current_md5).lower():
        wait_for_artifactory_metadata(
            metadata_api_url, orig_md5, user, token, timeout=args.timeout
        )
    else:
        print(
            "Metadata MD5 did not change after calculation; "
            "checking package presence before signing."
        )

    for asset in assets:
        wait_for_package_in_metadata(
            lambda asset=asset: rpm_package_in_metadata(
                args.stage, arch, asset.name, user, token
            ),
            asset.name,
            timeout=args.timeout,
        )

    dest = os.path.join(REPO_DIR, os.path.basename(metadata_url))
    download_file(metadata_url, dest, user, token)


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
            match = re.match(rf"^{PACKAGE_NAME}-(\d+\.\d+\.\d+(\.\d+)?)-(amd64|arm64)\.msi$", asset.name)
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
