import hashlib
import os
import re
import tempfile
import time

import requests
from github import Github

ARTIFACTORY_URL = "https://splunk.jfrog.io/artifactory"
ARTIFACTORY_API_URL = f"{ARTIFACTORY_URL}/api"
CHAPERONE_API_URL = "https://chaperone.re.splunkdev.com/api-service"
SIGNED_ARTIFACTS_REPO_URL = "https://repo.splunk.com/artifactory/signed-artifacts"
STAGING_URL = "https://repo.splunk.com/artifactory"
STAGING_REPO = os.environ.get("STAGING_REPO", "otel-collector-local")
STAGING_REPO_URL = f"{STAGING_URL}/{STAGING_REPO}"
SIGN_TYPES = ("GPG", "RPM", "WIN")
DEFAULT_ARTIFACTORY_USERNAME = "otel-collector"
DEFAULT_STAGING_USERNAME = "srv-otel-collector"
DEFAULT_TIMEOUT = 300


def upload_file_to_artifactory(src, dest, user, token):
    print(f"uploading {src} to {dest} ...")

    with open(src, "rb") as fd:
        data = fd.read()
        headers = {"X-Checksum-MD5": hashlib.md5(data).hexdigest()}
        resp = requests.put(dest, auth=(user, token), headers=headers, data=data)

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


def wait_for_signed_artifact(
    item_key, artifact_name, chaperone_token, staging_user, staging_token, timeout=DEFAULT_TIMEOUT
):
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


def sign_file(
    src,
    dest,
    sign_type,
    chaperone_token,
    staging_user,
    staging_token,
    src_user=None,
    src_token=None,
    timeout=DEFAULT_TIMEOUT,
):
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

        signed_artifact_url = wait_for_signed_artifact(
            item_key, artifact_name, chaperone_token, staging_user, staging_token, timeout
        )

        download_file(signed_artifact_url, dest, staging_user, staging_token)
    finally:
        if artifactory_file_exists(staged_artifact_url, staging_user, staging_token):
            delete_artifactory_file(staged_artifact_url, staging_user, staging_token)


def sign_artifactory_metadata(
    src, artifactory_user, artifactory_token, chaperone_token, staging_user, staging_token, timeout=DEFAULT_TIMEOUT
):
    with tempfile.TemporaryDirectory() as tmpdir:
        base = os.path.basename(src)
        signature_ext = ".gpg" if base == "Release" else ".asc"
        signature_path = os.path.join(tmpdir, base) + signature_ext
        signature_url = src + signature_ext

        sign_file(
            src,
            signature_path,
            "GPG",
            chaperone_token,
            staging_user,
            staging_token,
            artifactory_user,
            artifactory_token,
            timeout,
        )

        upload_file_to_artifactory(signature_path, signature_url, artifactory_user, artifactory_token)


def add_signing_args(parser):
    parser.add_argument(
        "--chaperone-token",
        type=str,
        default=os.environ.get("CHAPERONE_TOKEN"),
        metavar="CHAPERONE_TOKEN",
        required=False,
        help="Chaperone token. Required if the CHAPERONE_TOKEN env var is not set.",
    )
    parser.add_argument(
        "--staging-user",
        type=str,
        default=os.environ.get("STAGING_USERNAME", DEFAULT_STAGING_USERNAME),
        metavar="STAGING_USERNAME",
        required=False,
        help=f"""
            {STAGING_URL} username. Defaults to the STAGING_USERNAME env var if set,
            otherwise '{DEFAULT_STAGING_USERNAME}'.
        """,
    )
    parser.add_argument(
        "--staging-token",
        type=str,
        default=os.environ.get("STAGING_TOKEN"),
        metavar="STAGING_TOKEN",
        required=False,
        help=f"{STAGING_URL} token. Required if the STAGING_TOKEN env var is not set.",
    )


def check_signing_args(args):
    assert args.chaperone_token, f"Chaperone token not set"
    assert args.staging_user, f"{STAGING_URL} username not set"
    assert args.staging_token, f"{STAGING_URL} token not set"


def add_artifactory_args(parser):
    parser.add_argument(
        "--artifactory-user",
        type=str,
        default=os.environ.get("ARTIFACTORY_USERNAME", DEFAULT_ARTIFACTORY_USERNAME),
        metavar="ARTIFACTORY_USERNAME",
        required=False,
        help=f"""
            {ARTIFACTORY_URL} username. Defaults to the ARTIFACTORY_USERNAME env var if set,
            otherwise '{DEFAULT_ARTIFACTORY_USERNAME}'.
        """,
    )
    parser.add_argument(
        "--artifactory-token",
        type=str,
        default=os.environ.get("ARTIFACTORY_TOKEN"),
        metavar="ARTIFACTORY_TOKEN",
        required=False,
        help=f"{ARTIFACTORY_URL} token. Required if the ARTIFACTORY_TOKEN env var is not set.",
    )


def check_artifactory_args(args):
    assert args.artifactory_user, f"{ARTIFACTORY_URL} username not set"
    assert args.artifactory_token, f"{ARTIFACTORY_URL} token not set"


def get_github_release_packages(repo_name, release_tag=None, github_token=None):
    gh = Github(github_token)
    repo = gh.get_repo(repo_name)
    if not release_tag or release_tag == "latest":
        release = repo.get_latest_release()
    else:
        release = repo.get_release(release_tag)

    packages = {"deb": [], "rpm": []}
    for asset in release.get_assets():
        package_type = asset.name.strip().split(".")[-1]
        if package_type in ("deb", "rpm"):
            packages[package_type].append(asset.browser_download_url)

    return release, packages
