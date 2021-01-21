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
import re
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
    MSI_CONFIG,
    PACKAGE_NAME,
    REPO_DIR,
    S3_BUCKET,
    S3_MSI_BASE_DIR,
    SIGN_TYPES,
    SIGNED_ARTIFACTS_REPO_URL,
    STAGING_REPO_URL,
    WIX_IMAGE,
    WXS_PATH,
)


class Asset(object):
    def __init__(self, url=None, path=None):
        assert url or path, "Either url= or path= is required!"
        self.url = url
        self.path = path
        self.name = self._get_name()
        self.component = self._get_component()
        self.sign_type = self._get_sign_type()
        self.checksum = None
        self.signed = False
        self.signed_path = None
        self.signed_checksum = None
        if self.path:
            self.checksum = get_checksum(self.path, hashlib.sha256())

    def _get_name(self):
        if self.url:
            return os.path.basename(self.url)
        elif self.path:
            return os.path.basename(self.path)
        return None

    def _get_component(self):
        if self.name:
            ext = os.path.splitext(self.name)[-1].strip(".")
            if ext:
                return ext.lower()
        return None

    def _get_sign_type(self):
        if self.component:
            if self.component == "rpm":
                return "RPM"
            elif self.component in ("exe", "msi"):
                return "WIN"
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
            self.checksum = get_checksum(self.path, hashlib.sha256())
            return self.path
        return None

    def sign(self, dest=None, overwrite=False, timeout=DEFAULT_TIMEOUT, **signing_args):
        if self.path:
            if not dest:
                dest = os.path.join(os.path.dirname(self.path), "signed", self.name)
            if not overwrite and os.path.isfile(dest):
                resp = input(f"{dest} already exists.\nOverwrite? [y/N]: ")
                if resp.lower() not in ("y", "yes"):
                    return None
                os.remove(dest)
            sign_file(self.path, dest, self.sign_type, timeout=timeout, **signing_args)
            self.signed_path = dest
            self.signed_checksum = get_checksum(self.signed_path, hashlib.sha256())
            self.signed = True
            return self.signed_path
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


def submit_signing_request(src, sign_type, token):
    print(f"submitting '{sign_type}' signing request for {src} ...")

    headers = {"Accept": "application/json", "Authorization": f"Bearer {token}"}
    data = {"artifact_url": src, "sign_type": sign_type, "project_key": "otel-collector"}

    resp = requests.post(CHAPERONE_API_URL + "/sign/submit", headers=headers, data=data)

    assert resp.status_code == 200, f"signing request failed:\n{resp.reason}"
    assert "item_key" in resp.json().keys(), f"'item_key' not found in response:\n{resp.text}"

    print(resp.text)

    return resp.json().get("item_key")


def run_chaperone_check(item_key, token):
    url = f"{CHAPERONE_API_URL}/{item_key}/check"
    headers = {"Accept": "application/json", "Authorization": f"Bearer {token}"}

    print(f"running chaperone check {url}:")

    resp = requests.get(url, headers=headers)

    assert resp.status_code == 200, f"chaperone check failed for {item_key}:\n{resp.reason}\n{resp.text}"

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


def wait_for_signed_artifact(item_key, artifact_name, timeout=DEFAULT_TIMEOUT, **signing_args):
    chaperone_token = signing_args.get("chaperone_token")
    staging_token = signing_args.get("staging_token")
    staging_user = signing_args.get("staging_user")
    url = f"{SIGNED_ARTIFACTS_REPO_URL}/{item_key}/{artifact_name}"

    print(f"waiting for {url} ...")

    start_time = time.time()
    while True:
        assert (time.time() - start_time) < timeout, f"timed out waiting for {url}"

        resp = run_chaperone_check(item_key, chaperone_token)
        status = resp.json().get("status", "").lower()
        node = resp.json().get("node", "").lower()

        assert status and node and status != "exception" and node != "failed", f"signing request failed:\n{resp.text}"

        print(resp.text)

        if node == "signed" and artifactory_file_exists(url, staging_user, staging_token):
            break

        time.sleep(10)

    return url


def sign_file(src, dest, sign_type, src_user=None, src_token=None, timeout=DEFAULT_TIMEOUT, **signing_args):
    chaperone_token = signing_args.get("chaperone_token")
    staging_token = signing_args.get("staging_token")
    staging_user = signing_args.get("staging_user")

    if not re.match("^http(s)?://.*", src):
        assert os.path.isfile(src), f"{src} not found"

    assert sign_type.upper() in SIGN_TYPES, f"sign type '{sign_type}' not supported"

    base = os.path.basename(src)
    staged_artifact_url = f"{STAGING_REPO_URL}/{base}"

    try:
        with tempfile.TemporaryDirectory() as tmpdir:
            if not os.path.isfile(src):
                tmpsrc = os.path.join(tmpdir, base)
                download_file(src, tmpsrc, src_user, src_token)
                src = tmpsrc
            upload_file_to_artifactory(src, staged_artifact_url, staging_user, staging_token)

        item_key = submit_signing_request(staged_artifact_url, sign_type.upper(), chaperone_token)

        artifact_name = f"{base}.asc" if sign_type.lower() == "gpg" else base

        signed_artifact_url = wait_for_signed_artifact(item_key, artifact_name, timeout=timeout, **signing_args)

        download_file(signed_artifact_url, dest, staging_user, staging_token)
    finally:
        if artifactory_file_exists(staged_artifact_url, staging_user, staging_token):
            delete_artifactory_file(staged_artifact_url, staging_user, staging_token)


def sign_artifactory_metadata(src, artifactory_user, artifactory_token, timeout=DEFAULT_TIMEOUT, **signing_args):
    with tempfile.TemporaryDirectory() as tmpdir:
        base = os.path.basename(src)
        signature_ext = ".gpg" if base == "Release" else ".asc"
        signature_path = os.path.join(tmpdir, base) + signature_ext
        signature_url = src + signature_ext

        sign_file(
            src,
            signature_path,
            "GPG",
            src_user=artifactory_user,
            src_token=artifactory_token,
            timeout=timeout,
            **signing_args,
        )

        upload_file_to_artifactory(signature_path, signature_url, artifactory_user, artifactory_token)


def upload_package_to_artifactory(
    path,
    dest_url,
    user,
    token,
    metadata_api_url,
    metadata_url,
    sign_metadata=True,
    timeout=DEFAULT_TIMEOUT,
    **signing_args,
):
    orig_md5 = get_md5_from_artifactory(metadata_api_url, user, token)

    upload_file_to_artifactory(path, dest_url, user, token)

    wait_for_artifactory_metadata(metadata_api_url, orig_md5, user, token, timeout=timeout)

    if sign_metadata:
        sign_artifactory_metadata(metadata_url, user, token, timeout=timeout, **signing_args)


def release_deb_to_artifactory(asset, args, **signing_args):
    if args.no_push or args.stage == "github":
        # nothing to do for deb packages
        return

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

    sign_metadata = False if args.stage == "test" else True

    upload_package_to_artifactory(
        asset.path,
        dest_url,
        user,
        token,
        metadata_api_url,
        metadata_url,
        sign_metadata=sign_metadata,
        timeout=args.timeout,
        **signing_args,
    )


def release_rpm_to_artifactory(asset, args, **signing_args):
    user = args.artifactory_user
    token = args.artifactory_token

    match = re.match(r".*\.(x86_64|aarch64)\.rpm$", asset.name)
    assert match, f"Failed to get arch from {asset.path}!"
    arch = match.groups()[0]

    metadata_api_url = f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_RPM_REPO}/{args.stage}/{arch}/repodata/repomd.xml"
    metadata_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/repodata/repomd.xml"
    dest_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/{asset.name}"

    if args.stage != "github" and not args.no_push:
        if not args.force and artifactory_file_exists(dest_url, user, token):
            resp = input(f"{dest_url} already exists.\nOverwrite? [y/N]: ")
            if resp.lower() not in ("y", "yes"):
                sys.exit(1)

    rpm_path = asset.path
    sign_metadata = False

    if args.stage != "test":
        print(f"Signing {asset.name} (may take 10+ minutes):")
        if not asset.sign(overwrite=args.force, timeout=args.timeout, **signing_args):
            sys.exit(1)
        rpm_path = asset.signed_path
        sign_metadata = True

    if args.stage != "github" and not args.no_push:
        upload_package_to_artifactory(
            rpm_path,
            dest_url,
            user,
            token,
            metadata_api_url,
            metadata_url,
            sign_metadata=sign_metadata,
            timeout=args.timeout,
            **signing_args,
        )


def msi_exists_in_s3(s3_client, path):
    results = s3_client.list_objects(Bucket=S3_BUCKET, Prefix=f"{path}")
    for content in results.get("Contents", []):
        if content.get("Key") == path:
            return True
    return False


def invalidate_cloudfront(paths):
    cloudfront = boto3.client('cloudfront')
    print(f"Invalidating cloudfront for {paths} ...")
    resp = cloudfront.create_invalidation(
        DistributionId=CLOUDFRONT_DISTRIBUTION_ID,
        InvalidationBatch={
            'Paths': {
                'Quantity': len(paths),
                'Items': [f"/{path}" for path in paths]
            },
            'CallerReference': f"splunk-otel-collector-msi-{time.time()}",
        }
    )
    print(resp)
    assert resp.get("ResponseMetadata", {}).get("HTTPStatusCode") == 201, "Failed to submit invalidation request!"


def build_msi(exe_path, args):
    assert exe_path, f"{exe_path} not found!"

    msi_version = args.tag.strip("v")
    msi_name = f"{PACKAGE_NAME}-{msi_version}-amd64.msi"
    msi_path = os.path.join(args.assets_dir, msi_name)
    wixobj = f"{PACKAGE_NAME}.wixobj"

    print(f"Building {msi_path} with {exe_path} ...")
    if not args.force and os.path.isfile(msi_path):
        resp = input(f"{msi_path} already exists.\nOverwrite? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)
        os.remove(msi_path)

    client = docker.from_env()
    container_options = {
        "remove": True,
        "volumes": {
            REPO_DIR: {"bind": "/work", "mode": "rw"},
            os.path.abspath(exe_path): {"bind": "/otelcol/otelcol.exe", "mode": "ro"},
        },
        "working_dir": "/work",
    }

    candle_opts = f'-arch x64 -dOtelcol="/otelcol/otelcol.exe" -dVersion="{msi_version}" -dConfig="{MSI_CONFIG}"'
    cmd = f"candle {candle_opts} {WXS_PATH}"
    output = client.containers.run(WIX_IMAGE, command=cmd, **container_options)
    print(output.decode("utf-8"))
    assert os.path.isfile(wixobj), f"{wixobj} not found!"

    cmd = f"light {wixobj} -sval -spdb -out {msi_name}"
    output = client.containers.run(WIX_IMAGE, command=cmd, **container_options)
    print(output.decode("utf-8"))
    assert os.path.isfile(msi_name), f"{msi_name} not found!"

    os.makedirs(args.assets_dir, exist_ok=True)
    os.rename(msi_name, msi_path)
    if os.path.isfile(wixobj):
        os.remove(wixobj)

    return msi_path


def sign_exe(asset, args, **signing_args):
    if args.stage != "test":
        print(f"Signing {asset.name} (may take 10+ minutes):")
        if not asset.sign(timeout=args.timeout, overwrite=args.force, **signing_args):
            sys.exit(1)


def release_msi_to_s3(asset, args, **signing_args):
    if args.stage != "test":
        print(f"Signing {asset.name} (may take 10+ minutes):")
        if not asset.sign(timeout=args.timeout, overwrite=args.force, **signing_args):
            sys.exit(1)
        msi_path = asset.signed_path
    else:
        msi_path = asset.path

    if args.stage != "github" and not args.no_push:
        session = boto3.Session(profile_name="prod")
        s3_client = session.client('s3')
        s3_path = f"{S3_MSI_BASE_DIR}/{args.stage}/{asset.name}"
        print(f"Uploading {asset.name} to {S3_BUCKET}/{s3_path} ...")
        if not args.force and msi_exists_in_s3(s3_client, s3_path):
            resp = input(f"{S3_BUCKET}/{s3_path} already exists.\nOverwrite [y/N]: ")
            if resp.lower() not in ("y", "yes"):
                sys.exit(1)
        s3_client.upload_file(msi_path, S3_BUCKET, s3_path)
        with tempfile.TemporaryDirectory() as tmpdir:
            latest_txt = os.path.join(tmpdir, "latest.txt")
            match = re.match(f"^{PACKAGE_NAME}-(\d+\.\d+\.\d+(\.\d+)?)-amd64.msi$", asset.name)
            assert match, f"Failed to version from '{asset.name}'!"
            msi_version = match.group(1)
            with open(latest_txt, "w") as fd:
                fd.write(msi_version)
            s3_latest_path = f"{S3_MSI_BASE_DIR}/{args.stage}/latest.txt"
            print(f"Updating {S3_BUCKET}/{s3_latest_path} for version '{msi_version}' ...")
            s3_client.upload_file(latest_txt, S3_BUCKET, s3_latest_path)
            invalidate_cloudfront([s3_path, s3_latest_path])


def get_github_release(repo_name, tag=None, token=None):
    gh = Github(token)
    repo = gh.get_repo(repo_name)

    if tag:
        github_release = repo.get_release(tag)
    else:
        github_releases = repo.get_releases()
        assert github_releases, f"No releases found for '{repo_name}' repository!"
        github_release = github_releases[0]

    return github_release


def download_github_assets(github_release, args):
    assets = []
    checksums_asset = None

    for asset in github_release.get_assets():
        ext = os.path.splitext(asset.name)[-1].strip(".")
        if asset.name == "checksums.txt":
            checksums_asset = Asset(url=asset.browser_download_url)
        elif ext and ext in args.component or ("windows" in args.component and ext == "exe"):
            assets.append(Asset(url=asset.browser_download_url))

    assert checksums_asset, f"checksums.txt not found in {github_release.html_url}!"
    checksums_path = os.path.join(args.assets_dir, checksums_asset.name)
    if not checksums_asset.download(checksums_path, token=args.github_token, overwrite=args.force):
        sys.exit(1)

    for asset in assets:
        dest = os.path.join(args.assets_dir, asset.name)
        if not asset.download(dest, token=args.github_token, overwrite=args.force):
            sys.exit(1)

    return assets, checksums_asset
