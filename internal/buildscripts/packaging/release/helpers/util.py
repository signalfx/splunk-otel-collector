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
import os
import random
import re
import shutil
import string
import sys
import tempfile
import time

import boto3
import docker
import requests
from github import Github

from .constants import (
    ARTIFACTORY_API_URL,
    ARTIFACTORY_DEB_REPO,
    ARTIFACTORY_DEB_REPO_URL,
    ARTIFACTORY_RPM_REPO,
    ARTIFACTORY_RPM_REPO_URL,
    CHAPERONE_API_URL,
    CLOUDFRONT_DISTRIBUTION_ID,
    DEFAULT_TIMEOUT,
    EXTENSIONS,
    INSTALLER_SCRIPTS,
    PACKAGE_NAME,
    REPO_DIR,
    S3_BUCKET,
    S3_MSI_BASE_DIR,
    SIGN_TYPES,
    SIGNED_ARTIFACTS_REPO_URL,
    SMART_AGENT_RELEASE_PATH,
    STAGING_REPO_URL,
    STAGING_URL,
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


def wait_for_artifactory_metadata(url, orig_md5, user, token, timeout=DEFAULT_TIMEOUT):
    print(f"waiting for {url} to be updated ...")

    start_time = time.time()
    while True:
        assert (time.time() - start_time) < timeout, f"timed out waiting for {url} to be updated"

        new_md5 = get_md5_from_artifactory(url, user, token)

        if new_md5 and str(orig_md5).lower() != str(new_md5).lower():
            break

        time.sleep(5)

def upload_package_to_artifactory(
    path,
    dest_url,
    user,
    token,
    metadata_api_url,
    metadata_url,
    sign_metadata=True,
    timeout=DEFAULT_TIMEOUT,
):
    orig_md5 = None
    if sign_metadata:
        orig_md5 = get_md5_from_artifactory(metadata_api_url, user, token)

    upload_file_to_artifactory(path, dest_url, user, token)

    if sign_metadata:
        wait_for_artifactory_metadata(metadata_api_url, orig_md5, user, token, timeout=timeout)
        # don't sign the metadata; just download it so that it can be signed externally
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
    staging_url = f"{STAGING_URL}/otel-collector-deb/pool/{args.stage}/{arch}/{asset.name};{dest_opts}"

    if not args.force and artifactory_file_exists(deb_url, user, token):
        resp = input(f"{deb_url} already exists.\nOverwrite? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    upload_package_to_artifactory(
        asset.path,
        dest_url,
        user,
        token,
        metadata_api_url,
        metadata_url,
        sign_metadata=True,
        timeout=args.timeout,
    )

    upload_file_to_artifactory(asset.path, staging_url, args.staging_user, args.staging_token)


def release_rpm_to_artifactory(asset, args):
    user = args.artifactory_user
    token = args.artifactory_token

    match = re.match(r".*\.(x86_64|aarch64)\.rpm$", asset.name)
    assert match, f"Failed to get arch from {asset.path}!"
    arch = match.groups()[0]

    metadata_api_url = f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_RPM_REPO}/{args.stage}/{arch}/repodata/repomd.xml"
    metadata_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/repodata/repomd.xml"
    dest_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/{asset.name}"
    staging_url = f"{STAGING_URL}/otel-collector-rpm/{args.stage}/{arch}/{asset.name}"

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
        sign_metadata=True,
        timeout=args.timeout,
    )
    upload_file_to_artifactory(asset.path, staging_url, args.staging_user, args.staging_token)


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
            match = re.match(f"^{PACKAGE_NAME}-(\d+\.\d+\.\d+(\.\d+)?)-amd64.msi$", asset.name)
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
