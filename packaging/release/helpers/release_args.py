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
    DEFAULT_TIMEOUT,
    EXTENSIONS,
    STAGES,
)

from .util import Asset


def check_signing_args(args):
    assert args.staging_user, f"Staging username not set"
    assert args.staging_token, f"Staging token not set"

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
        "--paths",
        type=str,
        nargs="+",
        default=None,
        metavar="PATH",
        required=False,
        help="""
            Release a batch of local files. Only supported for deb/rpm packages.
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
            By default, the latest.txt file on dl.observability.splunkcloud.com is automatically updated with the version of the MSI.
            If uploading an older version of the MSI, specify this option to skip this step.
        """,
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=DEFAULT_TIMEOUT,
        metavar="TIMEOUT",
        required=False,
        help=f"Artifactory metadata wait timeout in seconds. Defaults to {DEFAULT_TIMEOUT}.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        default=False,
        required=False,
        help="Never prompt when overwriting existing files.",
    )
    add_artifactory_args(parser)

    args = parser.parse_args()

    assert args.path or args.paths or args.installers, "Either --path, --paths, or --installers must be specified"
    assert not (args.path and args.paths), "Only one of --path or --paths may be specified"

    asset = None
    args.assets = []
    if args.path:
        asset = get_asset(args.path)
        args.assets = [asset]
    elif args.paths:
        args.assets = [get_asset(path) for path in args.paths]
        components = {asset.component for asset in args.assets}
        assert len(components) == 1, f"All --paths assets must have the same component: {components}"
        component = next(iter(components))
        assert component in ("deb", "rpm"), "--paths is only supported for deb/rpm packages"
        if len(args.assets) == 1:
            asset = args.assets[0]

    if any(asset.component in ("deb", "rpm") for asset in args.assets):
        check_artifactory_args(args)

    return args, asset
