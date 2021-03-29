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

import argparse
import os
import sys

from .constants import (
    ARTIFACTORY_URL,
    COMPONENTS,
    DEFAULT_ARTIFACTORY_USERNAME,
    DEFAULT_STAGING_USERNAME,
    DEFAULT_TIMEOUT,
    STAGES,
    STAGING_URL,
)


def add_signing_args(parser):
    signing_args = parser.add_argument_group("Signing Credentials")
    signing_args.add_argument(
        "--chaperone-token",
        type=str,
        default=os.environ.get("CHAPERONE_TOKEN"),
        metavar="CHAPERONE_TOKEN",
        required=False,
        help="Chaperone token. Required if the CHAPERONE_TOKEN env var is not set.",
    )
    signing_args.add_argument(
        "--staging-user",
        type=str,
        default=os.environ.get("STAGING_USERNAME", DEFAULT_STAGING_USERNAME),
        metavar="STAGING_USERNAME",
        required=False,
        help=f"""
            Staging username for {STAGING_URL}.
            Defaults to the STAGING_USERNAME env var if set, otherwise '{DEFAULT_STAGING_USERNAME}'.
        """,
    )
    signing_args.add_argument(
        "--staging-token",
        type=str,
        default=os.environ.get("STAGING_TOKEN"),
        metavar="STAGING_TOKEN",
        required=False,
        help=f"""
            Staging token for {STAGING_URL}.
            Required if the STAGING_TOKEN env var is not set.
        """,
    )


def check_signing_args(args):
    assert args.chaperone_token, f"Chaperone token not set"
    assert args.staging_user, f"Staging username not set for {STAGING_URL}"
    assert args.staging_token, f"Staging token not set for {STAGING_URL}"


def add_artifactory_args(parser):
    artifactory_args = parser.add_argument_group("Artifactory Credentials")
    artifactory_args.add_argument(
        "--artifactory-user",
        type=str,
        default=os.environ.get("ARTIFACTORY_USERNAME", DEFAULT_ARTIFACTORY_USERNAME),
        metavar="ARTIFACTORY_USERNAME",
        required=False,
        help=f"""
            Artifactory username for {ARTIFACTORY_URL}.
            Defaults to the ARTIFACTORY_USERNAME env var if set, otherwise '{DEFAULT_ARTIFACTORY_USERNAME}'.
        """,
    )
    artifactory_args.add_argument(
        "--artifactory-token",
        type=str,
        default=os.environ.get("ARTIFACTORY_TOKEN"),
        metavar="ARTIFACTORY_TOKEN",
        required=False,
        help=f"""
            Artifactory token for {ARTIFACTORY_URL}.
            Required if the ARTIFACTORY_TOKEN env var is not set.
        """,
    )


def check_artifactory_args(args):
    assert args.artifactory_user, f"Artifactory username not set for {ARTIFACTORY_URL}"
    assert args.artifactory_token, f"Artifactory token not set for {ARTIFACTORY_URL}"


def add_github_args(parser):
    github_args = parser.add_argument_group("Github Credentials")
    github_args.add_argument(
        "--github-token",
        type=str,
        default=os.environ.get("GITHUB_TOKEN"),
        metavar="GITHUB_TOKEN",
        required=False,
        help=f"""
            Personal Github token.
            Required if the GITHUB_TOKEN env var is not set and STAGE is 'release'.
        """,
    )


def check_github_args(args):
    assert args.github_token, "Github token not set"


def get_args():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description="Sign and release assets from Github Releases.",
    )
    parser.add_argument(
        "--stage",
        type=str,
        default="test",
        metavar="STAGE",
        choices=STAGES,
        required=False,
        help=f"""
            Stage for pushing the packages to Artifactory and S3.
            Should be one of {STAGES}.
            If STAGE is 'test', the packages will *not* be signed and *only* pushed to Artifactory and S3.
            If STAGE is 'beta', the packages will be signed and *only* pushed to Artifactory and S3.
            If STAGE is 'github', the packages will be signed and *only* pushed to the Github release.
            If STAGE is 'release', the packages will be signed, pushed to Artifactory, S3, and Github, and the
            'Pre-release' label will be removed from the Github release.
            Defaults to 'test'.
        """,
    )
    parser.add_argument(
        "--no-push",
        action="store_true",
        default=False,
        required=False,
        help="Only download and sign the assets. Do not push the assets to Artifactory, S3, or Github.",
    )
    parser.add_argument(
        "--tag",
        type=str,
        default=None,
        metavar="TAG",
        required=False,
        help="Existing Github release tag (e.g. 'v1.2.3'). Defaults to the latest release tag.",
    )
    parser.add_argument(
        "--download-only",
        action="store_true",
        default=False,
        required=False,
        help="Download assets from the Github release and exit.",
    )
    parser.add_argument(
        "--assets-dir",
        type=str,
        default=None,
        metavar="DIR",
        required=False,
        help=f"""
            Directory to save the downloaded assets from the Github release.
            The directory will be created if it does not exist.
            This option is ignored if the '--path' option is also specified.
            Defaults to 'dist/release/<RELEASE_TAG>' in the repo root directory.
            Signed assets will be saved to the 'signed' sub-directory, e.g. 'dist/release/<RELEASE_TAG>/signed'.
        """,
    )
    parser.add_argument(
        "--component",
        type=str,
        default=[],
        metavar="COMPONENT",
        choices=COMPONENTS,
        action="append",
        required=False,
        help=f"""
            Only download, sign, and release the specified component from the Github release.
            Should be one of {COMPONENTS}.
            If COMPONENT is 'windows', the exe will be signed, and the msi will be rebuilt with the signed exe.
            This option may be specified multiple times for multiple components.
            The default is to perform a full sign/release for all supported components.
            This option is ignored if the '--path' option is also specified.
        """,
    )
    parser.add_argument(
        "--path",
        type=str,
        default=[],
        metavar="PATH",
        action="append",
        required=False,
        help="""
            Sign/release a local file instead of downloading the assets from the Github release.
            This option may be specified multiple times for multiple files.
            Only files with .deb, .rpm, or .exe extensions are supported.
            If the file is .exe, the msi will also be built and signed/released.
            NOTE: This option is only applicable if STAGE is 'test' or 'beta'.
        """,
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=DEFAULT_TIMEOUT,
        metavar="TIMEOUT",
        required=False,
        help=f"Signing request timeout in seconds. Defaults to {DEFAULT_TIMEOUT}.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        default=False,
        required=False,
        help="Never prompt when overwriting existing files.",
    )

    add_artifactory_args(parser)
    add_signing_args(parser)
    add_github_args(parser)

    args = parser.parse_args()

    if not args.component:
        args.component = COMPONENTS

    if args.stage != "test":
        check_signing_args(args)

    if not args.no_push:
        if args.stage in ("beta", "release"):
            check_artifactory_args(args)

        if args.stage in ("github", "release"):
            check_github_args(args)
            if args.path:
                print("The '--path' option is not supported for the 'github' or 'release' stage.")
                sys.exit(1)

    return args
