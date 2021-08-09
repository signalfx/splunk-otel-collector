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

import sys

from helpers.release_args import get_args_and_asset
from helpers.util import (
    release_deb_to_artifactory,
    release_installers_to_s3,
    release_msi_to_s3,
    release_rpm_to_artifactory,
    sign_exe,
)


def main():
    args, asset = get_args_and_asset()
    signing_args = {}
    for key, val in vars(args).items():
        if key in ("chaperone_token", "staging_token", "staging_user"):
            signing_args[key] = val

    if asset:
        if args.no_push:
            print("Signing the following asset:")
        elif args.no_sign_msi and asset.component == "msi":
            print(f"Releasing the following asset to the '{args.stage}' stage:")
        else:
            print(f"Signing/Releasing the following asset to the '{args.stage}' stage:")
        print(asset.path)

        if not args.force:
            resp = input("Continue? [y/N]: ")
            if resp.lower() not in ("y", "yes"):
                sys.exit(1)

        if asset.component == "deb":
            # Release deb to artifactory and sign metadata
            release_deb_to_artifactory(asset, args, **signing_args)
        elif asset.component == "rpm":
            # Sign rpm, release to artifactory, and sign metadata
            release_rpm_to_artifactory(asset, args, **signing_args)
        elif asset.component == "exe":
            # Sign Windows executable
            sign_exe(asset, args, **signing_args)
        elif asset.component == "msi":
            # Sign MSI and release to S3
            release_msi_to_s3(asset, args, **signing_args)
        elif asset.component == "osx":
            # Sign OSX executable
            sign_exe(asset, args, **signing_args)

    # Release installer scripts to S3
    if args.installers and not args.no_push:
        release_installers_to_s3(force=args.force)


if __name__ == "__main__":
    main()
