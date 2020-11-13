#!/usr/bin/env python3

import argparse
import glob
import os
import re
import sys
import tempfile

from helpers.util import (
    ARTIFACTORY_URL,
    ARTIFACTORY_API_URL,
    add_artifactory_args,
    add_signing_args,
    artifactory_file_exists,
    download_file,
    check_artifactory_args,
    check_signing_args,
    get_github_release_packages,
    get_md5_from_artifactory,
    sign_artifactory_metadata,
    sign_file,
    upload_file_to_artifactory,
    wait_for_artifactory_metadata,
)

COLLECTOR_REPO = os.environ.get("COLLECTOR_REPO", "signalfx/splunk-otel-collector")
PACKAGE_NAME = os.environ.get("PACKAGE_NAME", "splunk-otel-collector")
ARTIFACTORY_DEB_REPO = "otel-collector-deb"
ARTIFACTORY_DEB_REPO_URL = f"{ARTIFACTORY_URL}/{ARTIFACTORY_DEB_REPO}"
ARTIFACTORY_RPM_REPO = "otel-collector-rpm"
ARTIFACTORY_RPM_REPO_URL = f"{ARTIFACTORY_URL}/{ARTIFACTORY_RPM_REPO}"
DEFAULT_TIMEOUT = 600
STAGES = ("test", "beta", "release")


def getargs():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description="Sign and release deb/rpm packages from Github to Artifactory.",
    )
    parser.add_argument(
        "--path",
        type=str,
        default=[],
        metavar="PATH",
        action="append",
        help="""
            Instead of releasing packages from Github, release packages on the local filesystem.
            PATH can be a deb/rpm file or a directory containing deb/rpm files. "
            This option may be specified multiple times for multiple paths.
        """,
    )
    parser.add_argument(
        "--stage",
        type=str,
        default="release",
        metavar="STAGE",
        choices=STAGES,
        help=f"Stage for artifactory packages. Should be one of {STAGES}. Defaults to release.",
    )
    parser.add_argument(
        "--version",
        type=str,
        default=None,
        metavar="TAG",
        required=False,
        help="Github release tag (e.g. 'v1.2.3'). Defaults to the latest release tag if not specified.",
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
        "--yes",
        "-y",
        action="store_true",
        default=False,
        required=False,
        help="Never prompt and assume yes when overwriting existing files in artifactory.",
    )

    add_artifactory_args(parser)
    add_signing_args(parser)

    args = parser.parse_args()

    check_artifactory_args(args)

    if args.stage != "test":
        check_signing_args(args)

    return args


def sign_metadata(metadata_url, args):
    sign_artifactory_metadata(
        metadata_url,
        args.artifactory_user,
        args.artifactory_token,
        args.chaperone_token,
        args.staging_user,
        args.staging_token,
        args.timeout,
    )


def add_debs_to_repo(paths, args):
    metadata_api_url = f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_DEB_REPO}/dists/{args.stage}/Release"
    metadata_url = f"{ARTIFACTORY_DEB_REPO_URL}/dists/{args.stage}/Release"

    for path in paths:
        base = os.path.basename(path)
        match = re.match(r".*_(amd64|arm64)\.deb$", base)
        assert match, f"Failed to get arch from {path}!"
        arch = match.groups()[0]
        deb_url = f"{ARTIFACTORY_DEB_REPO_URL}/pool/{args.stage}/{arch}/{base}"
        dest_opts = f"deb.distribution={args.stage};deb.component=main;deb.architecture={arch}"
        dest_url = f"{deb_url};{dest_opts}"

        if not args.yes and artifactory_file_exists(deb_url, args.artifactory_user, args.artifactory_token):
            overwrite = input(f"package {deb_url} already exists. Overwrite? [y/N] ")
            if overwrite.lower() not in ("y", "yes"):
                sys.exit(1)

        orig_metadata_md5 = get_md5_from_artifactory(metadata_api_url, args.artifactory_user, args.artifactory_token)

        with tempfile.TemporaryDirectory() as tmpdir:
            if re.match(r"^http(s)?://.*", path):
                download_file(path, os.path.join(tmpdir, base))
                path = os.path.join(tmpdir, base)

            upload_file_to_artifactory(path, dest_url, args.artifactory_user, args.artifactory_token)

            wait_for_artifactory_metadata(
                metadata_api_url, orig_metadata_md5, args.artifactory_user, args.artifactory_token, args.timeout
            )

    if paths and args.stage != "test":
        sign_metadata(metadata_url, args)


def add_rpms_to_repo(paths, args):
    for path in paths:
        base = os.path.basename(path)
        match = re.match(r".*\.(x86_64|aarch64)\.rpm$", base)
        assert match, f"Failed to get arch from {path}!"
        arch = match.groups()[0]
        dest_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/{base}"
        metadata_api_url = (
            f"{ARTIFACTORY_API_URL}/storage/{ARTIFACTORY_RPM_REPO}/{args.stage}/{arch}/repodata/repomd.xml"
        )
        metadata_url = f"{ARTIFACTORY_RPM_REPO_URL}/{args.stage}/{arch}/repodata/repomd.xml"

        if not args.yes and artifactory_file_exists(dest_url, args.artifactory_user, args.artifactory_token):
            overwrite = input(f"package {dest_url} already exists. Overwrite? [y/N] ")
            if overwrite.lower() not in ("y", "yes"):
                sys.exit(1)

        orig_metadata_md5 = get_md5_from_artifactory(metadata_api_url, args.artifactory_user, args.artifactory_token)

        with tempfile.TemporaryDirectory() as tmpdir:
            unsigned_rpm_path = path
            if re.match(r"^http(s)?://.*", path):
                unsigned_rpm_path = os.path.join(tmpdir, "unsigned", base)
                download_file(path, unsigned_rpm_path)
            signed_rpm_path = os.path.join(tmpdir, "signed", base)
            if args.stage != "test":
                print(f"Signing {base} (may take 10+ minutes):")
                sign_file(
                    unsigned_rpm_path,
                    signed_rpm_path,
                    "RPM",
                    args.chaperone_token,
                    args.staging_user,
                    args.staging_token,
                    timeout=args.timeout,
                )
            else:
                signed_rpm_path = unsigned_rpm_path
            upload_file_to_artifactory(signed_rpm_path, dest_url, args.artifactory_user, args.artifactory_token)

        wait_for_artifactory_metadata(
            metadata_api_url, orig_metadata_md5, args.artifactory_user, args.artifactory_token, args.timeout
        )

        if args.stage != "test":
            sign_metadata(metadata_url, args)


def get_packages(args):
    packages = {"deb": [], "rpm": []}

    if args.path:
        for path in args.path:
            if os.path.isdir(path):
                packages["deb"] = glob.glob(f"{path}/**/*.deb", recursive=True)
                packages["rpm"] = glob.glob(f"{path}/**/*.rpm", recursive=True)
            else:
                assert os.path.isfile(path), f"File {path} not found!"
                if os.path.splitext(path)[-1] == ".deb":
                    packages["deb"].append(path)
                elif os.path.splitext(path)[-1] == ".rpm":
                    packages["rpm"].append(path)
                else:
                    print(f"Unsupported file '{path}'")
                    sys.exit(1)

        assert packages["deb"] or packages["rpm"], f"No files found to release from {args.path}!"

        print(f"Releasing {packages} to {args.stage} stage.")
        if not args.yes:
            resp = input("Continue? [y/n]: ")
            if resp.lower() not in ("y", "yes"):
                sys.exit(0)
    else:
        gh_release, packages = get_github_release_packages(COLLECTOR_REPO, args.version)
        args.version = gh_release.tag_name if args.version is None else args.version

        assert packages["rpm"], f"Failed to get rpms from {COLLECTOR_REPO} {args.version} github release!"

        assert packages["deb"], f"Failed to get debs from {COLLECTOR_REPO} {args.version} github release!"

        print(f"Releasing deb/rpm from {COLLECTOR_REPO}:{args.version} to {args.stage} stage.")
        if not args.yes:
            resp = input("Continue? [y/n]: ")
            if resp.lower() not in ("y", "yes"):
                sys.exit(0)

    return packages


def main():
    args = getargs()

    packages = get_packages(args)

    add_debs_to_repo(packages["deb"], args)

    add_rpms_to_repo(packages["rpm"], args)


if __name__ == "__main__":
    main()
