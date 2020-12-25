#!/usr/bin/env python3

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

import os
import re
import sys

from helpers.constants import ASSETS_BASE_DIR, COLLECTOR_REPO
from helpers.release_args import get_args
from helpers.util import (
    Asset,
    build_msi,
    download_github_assets,
    get_github_release,
    release_deb_to_artifactory,
    release_rpm_to_artifactory,
    release_msi_to_s3,
    sign_exe,
)


def get_local_assets(args):
    assets = []

    for local_path in args.path:
        assert os.path.isfile(local_path), f"{local_path} not found!"
        ext = os.path.splitext(local_path)[-1].strip(".")
        assert ext and ext in ("deb", "rpm", "exe", "msi"), f"{local_path} not supported"
        assets.append(Asset(path=local_path))

    return assets


def verify_checksums(checksums_path, assets, signed=False):
    print(f"Verifying checksums with {checksums_path} ...")
    checksums = {}
    with open(checksums_path, "r") as f:
        for line in f.readlines():
            match = re.match(r"^([a-f0-9]{64})\s+(.+)", line)
            assert match, f"Failed to parse {checksums_path}!"
            checksum = match.group(1)
            asset_name = match.group(2).strip()
            checksums[asset_name] = checksum

    for asset in assets:
        if asset.signed and signed:
            path = asset.signed_path
            checksum = asset.signed_checksum
        else:
            path = asset.path
            checksum = asset.checksum
        assert path, f"Path not set for asset {asset.name}!"
        assert os.path.isfile(path), f"{path} not found!"
        assert checksum, f"Checksum not calculated for {path}!"
        assert checksums.get(asset.name) == checksum, f"Checksum for {path} does not match {checksums_path}!"


def create_signed_checksums(checksums_asset, assets, assets_dir):
    new_checksums = {}
    for asset in assets:
        if asset.signed:
            new_checksums[asset.name] = asset.signed_checksum
        else:
            new_checksums[asset.name] = asset.checksum
        assert new_checksums[asset.name], f"Checksum not calculated for {asset.name}!"

    checksums = {}
    with open(checksums_asset.path, "r") as f:
        # update original checksums with the signed checksums
        for line in f.readlines():
            if line.strip():
                match = re.match(r"^([a-f0-9]{64})\s+(.+)", line)
                assert match, f"Failed to parse {checksums_asset.path}!"
                orig_checksum = match.group(1)
                asset_name = match.group(2).strip()
                if asset_name in new_checksums.keys():
                    checksums[asset_name] = new_checksums.get(asset_name)
                else:
                    checksums[asset_name] = orig_checksum
        # add checksums for any new assets
        for asset_name in new_checksums.keys():
            if not checksums.get(asset_name):
                checksums[asset_name] = new_checksums.get(asset_name)

    checksums_asset.signed_path = os.path.join(assets_dir, "signed", "checksums.txt")

    print(f"Creating {checksums_asset.signed_path} ...")
    os.makedirs(os.path.dirname(checksums_asset.signed_path), exist_ok=True)
    with open(checksums_asset.signed_path, "w") as f:
        for asset_name in sorted(checksums):
            f.write(f"{checksums.get(asset_name)}  {asset_name}\n")

    checksums_asset.signed = True


def main():
    args = get_args()
    signing_args = {}
    for key, val in vars(args).items():
        if key in ("chaperone_token", "staging_token", "staging_user"):
            signing_args[key] = val

    checksums_asset = None

    github_release = get_github_release(COLLECTOR_REPO, args.tag, args.github_token)
    if not args.tag:
        args.tag = github_release.tag_name

    if not args.no_push and not args.download_only and args.stage == "release" and not github_release.prerelease:
        resp = input(f"{github_release.html_url} is not labeled 'Pre-release'.\nContinue? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)

    if args.path:
        assets = get_local_assets(args)
    else:
        if not args.assets_dir:
            args.assets_dir = os.path.join(ASSETS_BASE_DIR, args.tag)
        resp = input(f"Downloading assets from {github_release.html_url} to '{args.assets_dir}'.\nContinue? [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)
        assets, checksums_asset = download_github_assets(github_release, args)
        verify_checksums(checksums_asset.path, assets)
        if args.download_only:
            sys.exit(0)

    if args.no_push:
        print("Signing the following asset(s):")
    elif args.stage == "release":
        print(f"Releasing the following asset(s) to the '{args.stage}' stage and {github_release.html_url}:")
    elif args.stage == "github":
        print(f"Releasing the following asset(s) to {github_release.html_url}:")
    else:
        print(f"Releasing the following asset(s) to the '{args.stage}' stage:")
    for asset in assets:
        print(asset.path)
    resp = input("Continue? [y/N]: ")
    if resp.lower() not in ("y", "yes"):
        sys.exit(1)

    for asset in assets:
        if asset.component == "deb":
            # Release deb to artifactory and sign metadata
            release_deb_to_artifactory(asset, args, **signing_args)
        elif asset.component == "rpm":
            # Sign/release rpm to artifactory and sign metadata
            release_rpm_to_artifactory(asset, args, **signing_args)
        elif asset.component == "exe":
            # Sign exe, build msi with signed exe, sign msi, and release msi to S3
            if args.stage == "test":
                exe_path = asset.path
            else:
                sign_exe(asset, args, **signing_args)
                exe_path = asset.signed_path
            msi_asset = Asset(path=build_msi(exe_path, args))
            release_msi_to_s3(msi_asset, args, **signing_args)
            assets.append(msi_asset)

    # Create a copy of checksums.txt updated for the signed assets
    if checksums_asset and args.stage != "test":
        create_signed_checksums(checksums_asset, assets, args.assets_dir)
        verify_checksums(checksums_asset.signed_path, assets, signed=True)

    # Replace assets in the Github release with the signed assets and the updated checksums.txt
    if assets and checksums_asset and args.stage in ("github", "release") and not args.no_push:
        resp = input(f"Releasing assets to {github_release.html_url}.\nContinue [y/N]: ")
        if resp.lower() not in ("y", "yes"):
            sys.exit(1)
        github_assets = github_release.get_assets()
        for asset in assets + [checksums_asset]:
            if asset.signed and asset.signed_path:
                print(f"Uploading {asset.signed_path} to {github_release.html_url} ...")
                for github_asset in github_assets:
                    # delete pre-existing assets before upload
                    if asset.name == github_asset.name:
                        github_asset.delete_asset()
                        break
                github_release.upload_asset(asset.signed_path, name=asset.name)
        if github_release.prerelease and args.stage == "release":
            print(f"Removing 'Pre-release' label from {github_release.html_url}")
            github_release.update_release(name=github_release.title, message=github_release.body, prerelease=False)


if __name__ == "__main__":
    main()
