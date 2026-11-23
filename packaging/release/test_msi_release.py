# Copyright 2026 Splunk, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

import base64
import hashlib
import io
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import ANY, Mock, call, patch

import boto3
from botocore.exceptions import ClientError
from botocore.response import StreamingBody
from botocore.stub import Stubber

from helpers import util
from release import main as release_main


class StableMsiReleaseTest(unittest.TestCase):
    def setUp(self):
        self.s3 = boto3.client("s3", region_name="us-east-1", aws_access_key_id="fixture", aws_secret_access_key="fixture")
        self.stub = Stubber(self.s3)
        self.stub.activate()
        self.addCleanup(self.stub.deactivate)
        self.addCleanup(self.stub.assert_no_pending_responses)
        session = patch.object(util.boto3, "Session")
        session.start().return_value.client.return_value = self.s3
        self.addCleanup(session.stop)
        self.path = f"{util.S3_MSI_BASE_DIR}/release/latest.txt"
        self.request = {"Bucket": util.S3_BUCKET, "Key": self.path}
        workdir = tempfile.TemporaryDirectory()
        self.addCleanup(workdir.cleanup)
        self.workdir = Path(workdir.name)

    def expect_latest(self, version, etag='"previous"', content_length=None):
        body = version.encode("ascii")
        length = len(body) if content_length is None else content_length
        self.stub.add_response("get_object", {
            "Body": StreamingBody(io.BytesIO(body), length), "ContentLength": length, "ETag": etag,
        }, self.request)

    def expect_write(self, version, **condition):
        params = dict(self.request, Body=version.encode("ascii"), **condition)
        self.stub.add_response("put_object", {}, params)

    def msi(self, arch, version="1.2.3"):
        path = self.workdir / f"{util.PACKAGE_NAME}-{version}-{arch}.msi"
        path.write_bytes(f"signed {version} {arch} MSI".encode("ascii"))
        return util.Asset(path=str(path))

    def package_key(self, asset):
        return f"{util.S3_MSI_BASE_DIR}/release/{asset.name}"

    def expect_msi_upload(self, asset, error=None, status=None):
        checksum = base64.b64encode(hashlib.sha256(Path(asset.path).read_bytes()).digest()).decode("ascii")
        request = {"Bucket": util.S3_BUCKET, "Key": self.package_key(asset), "Body": ANY,
                   "IfNoneMatch": "*", "ChecksumSHA256": checksum}
        if error:
            self.stub.add_client_error("put_object", error, http_status_code=status, expected_params=request)
        else:
            self.stub.add_response("put_object", {}, request)

    def expect_existing_msi(self, asset, content=None):
        body = Path(asset.path).read_bytes() if content is None else content
        self.stub.add_response("get_object", {
            "Body": StreamingBody(io.BytesIO(body), len(body)), "ContentLength": len(body),
        }, {"Bucket": util.S3_BUCKET, "Key": self.package_key(asset)})

    def test_first_release_creates_pointer_conditionally(self):
        self.stub.add_client_error("get_object", "NoSuchKey", http_status_code=404, expected_params=self.request)
        self.expect_write("1.2.3", IfNoneMatch="*")
        util.promote_stable_msi_latest(self.path, "1.2.3")

    def test_stale_and_equivalent_versions_leave_pointer_alone(self):
        for current, candidate in (("1.2.10", "1.2.9"), ("1.2.10", "1.2.10.0"), ("1.2.10.1", "1.2.10")):
            with self.subTest(current=current, candidate=candidate):
                self.expect_latest(current + "\n")
                util.promote_stable_msi_latest(self.path, candidate)

    def test_higher_version_updates_only_the_pointer_it_read(self):
        self.expect_latest("1.2.9.1")
        self.expect_write("1.2.10", IfMatch='"previous"')
        util.promote_stable_msi_latest(self.path, "1.2.10")

    def test_newer_concurrent_release_wins_a_creation_race(self):
        self.stub.add_client_error("get_object", "NoSuchKey", http_status_code=404, expected_params=self.request)
        self.stub.add_client_error("put_object", "PreconditionFailed", http_status_code=412,
                                   expected_params=dict(self.request, Body=b"1.2.3", IfNoneMatch="*"))
        self.expect_latest("1.2.4", etag='"concurrent"')
        util.promote_stable_msi_latest(self.path, "1.2.3")

    def test_newer_candidate_retries_with_fresh_etag_after_conflict(self):
        self.expect_latest("1.2.2")
        self.stub.add_client_error("put_object", "ConditionalRequestConflict", http_status_code=409,
                                   expected_params=dict(self.request, Body=b"1.2.4", IfMatch='"previous"'))
        self.expect_latest("1.2.3", etag='"concurrent"')
        self.expect_write("1.2.4", IfMatch='"concurrent"')
        util.promote_stable_msi_latest(self.path, "1.2.4")

    def test_unverifiable_pointer_prevents_write(self):
        for version, etag in (("invalid", '"previous"'), ("1.2.2", "*")):
            with self.subTest(version=version, etag=etag):
                self.expect_latest(version, etag)
                with self.assertRaises(ValueError):
                    util.promote_stable_msi_latest(self.path, "1.2.3")
        self.stub.add_client_error("get_object", "AccessDenied", http_status_code=403, expected_params=self.request)
        with self.assertRaises(ClientError):
            util.promote_stable_msi_latest(self.path, "1.2.3")

    def test_partial_version_pointer_prevents_write(self):
        # A truncated 1.2.30 must not be read as 1.2.3.
        self.expect_latest("1.2.3", content_length=len("1.2.30"))
        with self.assertRaises(ValueError):
            util.promote_stable_msi_latest(self.path, "1.2.4")

    def test_conflict_retry_budget_is_bounded(self):
        for _ in range(util.MSI_WRITE_ATTEMPTS):
            self.expect_latest("1.2.3")
            self.stub.add_client_error("put_object", "PreconditionFailed", http_status_code=412,
                                       expected_params=dict(self.request, Body=b"1.2.4", IfMatch='"previous"'))
        with self.assertRaises(ClientError):
            util.promote_stable_msi_latest(self.path, "1.2.4")

    def test_cli_promotes_latest_after_both_msi_packages_are_available(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64")
        self.expect_msi_upload(amd64)
        self.expect_msi_upload(arm64)
        self.expect_latest("1.2.2")
        self.expect_write("1.2.3", IfMatch='"previous"')
        calls = Mock()
        with patch("sys.argv", ["release.py", "--stage", "release", "--force", "--paths", arm64.path, amd64.path]), \
                patch.object(util, "invalidate_cloudfront") as invalidate, \
                patch.object(util, "promote_stable_msi_latest", wraps=util.promote_stable_msi_latest) as promote:
            calls.attach_mock(invalidate, "invalidate")
            calls.attach_mock(promote, "promote")
            release_main()
        self.assertEqual(calls.mock_calls, [
            call.invalidate([self.package_key(amd64), self.package_key(arm64)], wait=True),
            call.promote(self.path, "1.2.3"),
            call.invalidate([self.path]),
        ])

    def test_failed_package_invalidation_leaves_latest_unchanged(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64")
        self.expect_msi_upload(amd64)
        self.expect_msi_upload(arm64)
        with patch.object(util, "invalidate_cloudfront", side_effect=TimeoutError), self.assertRaises(TimeoutError):
            util.release_stable_msi_to_s3([amd64, arm64], SimpleNamespace(stage="release", force=True, not_latest=False))

    def test_cloudfront_waits_for_the_requested_invalidation(self):
        cloudfront = Mock()
        cloudfront.create_invalidation.return_value = {
            "ResponseMetadata": {"HTTPStatusCode": 201},
            "Invalidation": {"Id": "fixture-invalidation"},
        }
        with patch.object(util, "boto3") as boto3_mock:
            boto3_mock.Session.return_value.client.return_value = cloudfront
            util.invalidate_cloudfront([self.path], wait=True)
        cloudfront.get_waiter.assert_called_once_with("invalidation_completed")
        cloudfront.get_waiter.return_value.wait.assert_called_once_with(
            DistributionId=util.CLOUDFRONT_DISTRIBUTION_ID, Id="fixture-invalidation",
        )

    def test_failed_second_architecture_does_not_advance_latest(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64")
        self.expect_msi_upload(amd64)
        self.expect_msi_upload(arm64, "AccessDenied", 403)
        with patch.object(util, "invalidate_cloudfront") as invalidate, self.assertRaises(ClientError):
            util.release_stable_msi_to_s3([amd64, arm64], SimpleNamespace(stage="release", force=True, not_latest=False))
        invalidate.assert_not_called()

    def test_retry_verifies_existing_bytes_and_only_creates_missing_architecture(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64")
        self.expect_msi_upload(amd64, "PreconditionFailed", 412)
        self.expect_existing_msi(amd64)
        self.expect_msi_upload(arm64, "ConditionalRequestConflict", 409)
        self.expect_msi_upload(arm64)
        self.expect_latest("1.2.2")
        self.expect_write("1.2.3", IfMatch='"previous"')
        with patch.object(util, "invalidate_cloudfront"):
            util.release_stable_msi_to_s3([amd64, arm64], SimpleNamespace(stage="release", force=True, not_latest=False))

    def test_existing_checksum_mismatch_blocks_promotion_even_with_force(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64")
        self.expect_msi_upload(amd64, "PreconditionFailed", 412)
        self.expect_existing_msi(amd64)
        self.expect_msi_upload(arm64, "PreconditionFailed", 412)
        self.expect_existing_msi(arm64, b"different signed bytes")
        with patch.object(util, "invalidate_cloudfront") as invalidate, self.assertRaisesRegex(ValueError, "checksum mismatch"):
            util.release_stable_msi_to_s3([amd64, arm64], SimpleNamespace(stage="release", force=True, not_latest=False))
        invalidate.assert_not_called()

    def test_promotion_requires_a_single_version_with_both_architectures(self):
        amd64, arm64 = self.msi("amd64"), self.msi("arm64", "1.2.4")
        args = SimpleNamespace(stage="release", force=True, not_latest=False)
        for assets in ([amd64], [amd64, arm64], [amd64, amd64]):
            with self.subTest(assets=assets), self.assertRaises(ValueError):
                util.release_stable_msi_to_s3(assets, args)

    def test_single_backfill_is_create_only_and_keeps_latest(self):
        asset = self.msi("arm64")
        self.expect_msi_upload(asset)
        with patch.object(util, "invalidate_cloudfront") as invalidate:
            util.release_msi_to_s3(asset, SimpleNamespace(stage="release", force=True, not_latest=True))
        invalidate.assert_called_once_with([self.package_key(asset)])

    def test_beta_pointer_keeps_existing_publication_behavior(self):
        asset = SimpleNamespace(name=f"{util.PACKAGE_NAME}-1.2.3.4-amd64.msi", path="signed.msi")
        uploads = {}

        def upload(path, key, force):
            self.assertTrue(force)
            uploads[key] = Path(path).read_text() if key.endswith("latest.txt") else path

        with patch.object(util, "upload_file_to_s3", side_effect=upload), patch.object(util, "invalidate_cloudfront"):
            util.release_msi_to_s3(asset, SimpleNamespace(stage="beta", force=True, not_latest=False))
        self.assertEqual(uploads[f"{util.S3_MSI_BASE_DIR}/beta/latest.txt"], "1.2.3.4")


if __name__ == "__main__":
    unittest.main()
