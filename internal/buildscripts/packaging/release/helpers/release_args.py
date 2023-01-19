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

from .constants import (
    ARTIFACTORY_URL,
    DEFAULT_ARTIFACTORY_USERNAME,
    DEFAULT_STAGING_USERNAME,
    DEFAULT_TIMEOUT,
    EXTENSIONS,
    S3_BUCKET,
    STAGES,
    STAGING_URL,
)

from .util import Asset

def add_staging_args(parser):
    signing_args = parser.add_argument_group("Signing Credentials")

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


def get_asset(path):
    assert os.path.isfile(path), f"{path} not found!"
    asset = Asset(path=path)
    assert asset.component, f"{path} is not a supported file!"
    return asset


def get_args_and_asset():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description="Sign and release splunk-otel-collector assets.",
    )
    parser.add_argument(
        "--stage",
        type=str,
        default="test",
        metavar="STAGE",
        choices=STAGES,
        required=False,
        help=f"""
            Stage for pushing the assets to Artifactory and S3.
            Should be one of {STAGES}.
            Defaults to 'test'.
        """,
    )
    parser.add_argument(
        "--path",
        type=str,
        default=None,
        metavar="PATH",
        required=False,
        help=f"""
            Sign/release a local file.
            Only files with {EXTENSIONS} extensions or is named 'otelcol_darwin_*' are supported.
            Required if the --installers option is not specified.
        """,
    )
    parser.add_argument(
        "--installers",
        action="store_true",
        default=False,
        required=False,
        help="""
            Release the installer scripts to S3.
            Required if the --path option is not specified.
        """,
    )
    parser.add_argument(
        "--not-latest",
        action="store_true",
        default=False,
        required=False,
        help="""
            Only applicable if the --path option is specified for a MSI file.
            By default, the latest.txt file on dl.signalfx.com is automatically updated with the version of the MSI.
            If uploading an older version of the MSI, specify this option to skip this step.
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
    add_staging_args(parser)

    args = parser.parse_args()

    assert args.path or args.installers, "Either --path or --installers must be specified"

    asset = None
    if args.path:
        asset = get_asset(args.path)

    if asset and asset.component in ("deb", "rpm"):
        check_artifactory_args(args)

    return args, asset
