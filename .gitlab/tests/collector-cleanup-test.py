#!/usr/bin/env python3

import contextlib
import importlib.util
import io
import json
import os
import unittest
from pathlib import Path
from unittest.mock import Mock, patch
from zipfile import ZipFile


spec = importlib.util.spec_from_file_location(
    "cleanup", Path(__file__).resolve().parents[1] / "cleanup-stable-release-attempts.py"
)
cleanup = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cleanup)

ATTEMPT = {
    "project": "42",
    "pipeline": "100",
    "version": "v1.2.3",
    "quay": "quay.io/splunk/collector",
    "hub": "splunk/collector",
}
METADATA = cleanup.attempt_metadata(ATTEMPT, "42", "100")


def job(job_id, name, *, archived=True, status="success", archive=True):
    return {
        "id": job_id,
        "name": name,
        "archived": archived,
        "status": status,
        "artifacts": [{"file_type": "archive"}] if archive else [],
        "artifacts_expire_at": None,
    }


def archived_release(pipeline, *, receipt=True, archive=True, publisher="build-push-linux-image"):
    return [
        job(pipeline * 10, "manual_release_trigger", archive=receipt),
        job(pipeline * 10 + 1, "prepare-stable-github-release", archive=archive),
        job(pipeline * 10 + 2, publisher, archive=False, status="failed"),
    ]


def plan_cleanup(attempts, pinned_pipelines=None, pinned_jobs=None):
    return cleanup.cleanup_plan(attempts, pinned_pipelines or set(), pinned_jobs or set(), "gitlab", "fixture", "42")


def trace_records(text):
    return cleanup.release_trace_records(io.BytesIO(text.encode()))


class CleanupTest(unittest.TestCase):
    def test_archived_attempts_include_retried_archives(self):
        jobs = [
            job(11, "manual_release_trigger"),
            job(12, "prepare-stable-github-release"),
            job(13, "prepare-stable-github-release"),
            job(14, "retain-stable-docker-records"),
            job(15, "build-stable-github-release-tools"),
            job(16, "unrelated"),
            job(17, "prepare-stable-github-release", status="failed", archive=False),
        ]
        responses = [
            [{"id": 100, "status": "failed"}, {"id": 101, "status": "success"},
             {"id": 102, "status": "running"}, {"id": 103, "status": "success"},
             {"id": 104, "status": "success"}],
            jobs,
            jobs + [job(18, "unrelated", archived=False)],
            jobs + [job(19, "unrelated", status="running")],
            [job(21, "manual_release_trigger")],
        ]
        with patch.object(cleanup, "pages", side_effect=responses) as pages, patch.object(cleanup, "api", return_value=ATTEMPT) as api:
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
        self.assertEqual(attempts, [cleanup.ArchivedAttempt(
            "100", ["12", "13", "14", "15"], ["11"], METADATA, "valid", jobs
        )])
        pages.assert_any_call("gitlab/pipelines/100/jobs", "fixture", include_retried="true")
        self.assertNotIn("gitlab/pipelines/102/jobs", [call.args[0] for call in pages.call_args_list])
        api.assert_called_once_with(f"gitlab/jobs/11/artifacts/{cleanup.ATTEMPT_FILE}", "fixture")

    def test_pins_even_when_signature_verification_is_unavailable(self):
        refs = (
            "tag-a\trefs/tags/v1.2.3\ncommit-a\trefs/tags/v1.2.3^{}\n"
            "legacy\trefs/tags/v1.2.2\ncommit-b\trefs/tags/v1.2.2^{}\n"
            "lightweight\trefs/tags/v1.2.1\n"
        )
        snapshot = dict(project="42", pipeline="100", assets={"job": "12"}, docker={"job": "14"}, tools={"job": "15"})
        tags = [
            {"sha": "tag-a", "tag": "v1.2.3", "message": cleanup.SNAPSHOT + json.dumps(snapshot), "verification": {"verified": False}},
            {"sha": "legacy", "tag": "v1.2.2", "message": "Legacy release"},
        ]
        with patch.object(cleanup.subprocess, "check_output", return_value=refs), patch.object(cleanup, "api", side_effect=tags):
            pipelines, archives = cleanup.pinned_release_inputs("42", "fixture")
        self.assertEqual(pipelines, {"100"})
        self.assertEqual(archives, {"12", "14", "15"})

    def test_bad_snapshot_prevents_planning(self):
        refs = "tag\trefs/tags/v1.2.3\ncommit\trefs/tags/v1.2.3^{}\n"
        tag = {"sha": "tag", "tag": "v1.2.3", "message": cleanup.SNAPSHOT + "invalid json"}
        with patch.object(cleanup.subprocess, "check_output", return_value=refs), patch.object(cleanup, "api", return_value=tag):
            with self.assertRaises(ValueError):
                cleanup.pinned_release_inputs("42", "fixture")

    def test_plan_preserves_pins_and_only_generates_staging_tags(self):
        jobs = archived_release(100)
        attempts = [cleanup.ArchivedAttempt("100", ["12", "13", "14", "15"], ["11"], METADATA, "valid", jobs)]
        plan, incomplete = plan_cleanup(attempts, {"100"}, {"12", "14", "15"})
        self.assertFalse(incomplete)
        self.assertEqual(plan, [{"pipeline": "100", "receipt_status": "valid", "staging_tags": [], "archives": ["13"], "receipts": ["11"]}])
        self.assertEqual(plan_cleanup(attempts, pinned_jobs={"12", "14", "15"}), (plan, False))

        trace = trace_records(cleanup.ATTEMPT_LOG + json.dumps(ATTEMPT))
        catalog = ["1.2.3-staging-100-linux", "1.2.3-staging-101-linux", "1.2.3-staging-100-unknown", "latest", "1.2.3"]
        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "api", return_value=trace), \
                patch.object(cleanup, "repository_tags", return_value=catalog) as registry:
            plan, incomplete = plan_cleanup(attempts)
        self.assertFalse(incomplete)
        self.assertEqual(plan[0]["staging_tags"], ["quay.io/splunk/collector:1.2.3-staging-100-linux"])
        self.assertEqual(plan[0]["receipts"], ["11"])
        registry.assert_called_once_with("quay.io/splunk/collector", None)

    def test_plan_recovers_missing_and_unreadable_receipts_without_touching_pins(self):
        responses = [
            [{"id": i, "status": "failed"} for i in range(100, 106)],
            archived_release(100, receipt=False),
            archived_release(101, publisher="mirror-docker-images: [windows_multiarch, -fips]"),
            archived_release(102, publisher="mirror-docker-images: [windows_multiarch, -fips]"),
            archived_release(103, receipt=False),
            archived_release(104, receipt=False, archive=False),
            archived_release(105),
        ]
        quay = ATTEMPT["quay"]
        hub = "docker.io/" + ATTEMPT["hub"]
        inventory = {
            quay: ("1.2.3-staging-100-linux", "1.2.3-staging-103-linux", "1.2.3-staging-104-amd64", "1.2.3-staging-105-linux",
                   "1.2.3-staging-999-amd64", "1.2.3", "latest", "1.2.3-staging-100-unrecognized"),
            hub + "-fips-windows": ("1.2.3-staging-101-2022", "1.2.3-staging-102", "1.2.3-staging-100-amd64"),
        }
        def read(url, token, **kwargs):
            if url.endswith("/variables"):
                return [{"key": "RELEASE_VERSION", "value": "v1.2.3"}]
            job_id = url.split("/")[2]
            pipeline = job_id[:-1]
            if url.endswith("/trace"):
                self.assertNotEqual(pipeline, "103")
                return trace_records("\x1b[0m" + cleanup.ATTEMPT_LOG + json.dumps(dict(ATTEMPT, pipeline=pipeline)))
            if pipeline == "101":
                raise ValueError("HTTP 404")
            if pipeline == "102":
                raise json.JSONDecodeError("invalid JSON", "", 0)
            self.assertEqual(pipeline, "105")
            return dict(ATTEMPT, pipeline="105")

        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "pages", side_effect=responses), \
                patch.object(cleanup, "api", side_effect=read), contextlib.redirect_stderr(io.StringIO()) as warnings, \
                patch.object(cleanup, "repository_tags", side_effect=lambda repo, *_args, **_kwargs: inventory.get(repo, ())) as tags:
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
            plan, incomplete = plan_cleanup(attempts, {"103"}, {"1031"})

        self.assertFalse(incomplete)
        self.assertEqual([entry["receipt_status"] for entry in plan], ["missing", "unreadable", "unreadable", "missing", "missing", "valid"])
        self.assertEqual([entry["archives"] for entry in plan], [["1001"], ["1011"], ["1021"], [], [], ["1051"]])
        self.assertEqual([entry["receipts"] for entry in plan], [[], [], [], [], [], ["1050"]])
        self.assertEqual([entry["staging_tags"] for entry in plan[:5]], [
            [f"{quay}:1.2.3-staging-100-linux"],
            [f"{hub}-fips-windows:1.2.3-staging-101-2022"],
            [f"{hub}-fips-windows:1.2.3-staging-102"],
            [],
            [f"{quay}:1.2.3-staging-104-amd64"],
        ])
        self.assertEqual(plan[5]["staging_tags"], [f"{quay}:1.2.3-staging-105-linux"])
        self.assertEqual(tags.call_count, 2)
        self.assertIn("pipeline 101, job 1010", warnings.getvalue())
        self.assertIn("pipeline 102, job 1020", warnings.getvalue())

    def test_registry_lists_follow_provider_pagination(self):
        fixtures = (
            ("quay.io/splunk/collector", "https://quay.io/api/v1/repository/splunk/collector/tag/", "tags", "has_additional"),
            ("docker.io/splunk/collector", f"{cleanup.DOCKER_HUB_API}/namespaces/splunk/repositories/collector/tags", "results", "next"),
        )
        for repo, endpoint, items, more in fixtures:
            with self.subTest(repo=repo), patch.object(cleanup, "api", side_effect=[
                {items: [{"name": "page-one"}], more: True},
                {items: [{"name": "page-two"}], more: False},
            ]) as api:
                self.assertEqual(list(cleanup.repository_tags(repo)), ["page-one", "page-two"])
            self.assertTrue(api.call_args.args[0].startswith(endpoint + "?"))
            self.assertTrue(api.call_args.args[0].endswith("&page=2"))

    def test_registry_absence_requires_authenticated_confirmation(self):
        repositories = {
            "quay.io/old/collector": "https://quay.io/api/v1/repository/old/collector?includeTags=false",
            "docker.io/old/collector": f"{cleanup.DOCKER_HUB_API}/namespaces/old/repositories/collector",
        }
        for repo, metadata_url in repositories.items():
            with self.subTest(repo=repo), patch.object(cleanup, "api", side_effect=[cleanup.NotFound(), cleanup.NotFound()]) as api:
                self.assertEqual(list(cleanup.repository_tags(repo, "fixture")), [])
                api.assert_called_with(metadata_url, "fixture", same_origin=True)

        repo = "docker.io/old/collector"
        with patch.object(cleanup, "api", side_effect=cleanup.NotFound("HTTP 404")) as api:
            with self.assertRaises(cleanup.NotFound):
                list(cleanup.repository_tags(repo))
            self.assertEqual(api.call_count, 1)
        for metadata, message in (({}, "Could not confirm absence"), (None, "Could not confirm absence"), (ValueError("HTTP 403"), "HTTP 403")):
            with self.subTest(metadata=metadata), patch.object(cleanup, "api", side_effect=[cleanup.NotFound(), metadata]):
                with self.assertRaisesRegex(ValueError, message):
                    list(cleanup.repository_tags(repo, "fixture"))

        listed_page = {"results": [{"name": "1.2.3-staging-100"}], "next": "page-two"}
        with patch.object(cleanup, "api", side_effect=[listed_page, cleanup.NotFound("page two HTTP 404")]) as api:
            with self.assertRaises(cleanup.NotFound):
                list(cleanup.repository_tags(repo, "fixture"))
            self.assertEqual(api.call_count, 2)

    def test_removed_repository_allows_archive_cleanup_and_rerun_without_touching_pins(self):
        publisher = "mirror-docker-images: [multiarch, ]"
        jobs = archived_release(100, receipt=False, publisher=publisher)
        signed_jobs = archived_release(101, receipt=False, publisher=publisher)
        pipelines = [{"id": 100, "status": "failed"}, {"id": 101, "status": "success"}]
        repo_url = f"{cleanup.DOCKER_HUB_API}/namespaces/splunk/repositories/collector"
        tags_url = f"{repo_url}/tags?page_size=100&page=1"
        real_api = cleanup.api
        opener = Mock()
        opener.open.side_effect = [cleanup.HTTPError(url, 404, "Not Found", None, None)
                                   for url in (tags_url, repo_url, tags_url, repo_url)]
        removed = []

        def request(url, token=None, *, method="GET", **options):
            if url in (tags_url, repo_url):
                self.assertEqual(method, "GET")
                return real_api(url, token, **options)
            if url == f"{cleanup.DOCKER_HUB_API}/auth/token":
                self.assertEqual(method, "POST")
                return {"access_token": "fixture"}
            if url == "gitlab/jobs/1000/trace":
                return trace_records(cleanup.ATTEMPT_LOG + json.dumps(ATTEMPT))
            if url == "gitlab/pipelines/100/variables":
                return [{"key": "RELEASE_VERSION", "value": "v1.2.3"}]
            if url == "gitlab/jobs/1001/artifacts" and method == "DELETE":
                jobs[1]["artifacts"].clear()
                removed.append(url)
                return 204
            if url == "gitlab/jobs/1001" and method == "GET":
                return jobs[1]
            self.fail(f"Unexpected {method} {url}")

        env = {"COLLECTOR_CLEANUP_DOCKER_HUB_USERNAME": "fixture", "COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN": "fixture"}
        with patch.dict(os.environ, env, clear=True), patch.object(cleanup, "build_opener", return_value=opener), \
                patch.object(cleanup, "pages", side_effect=[pipelines, jobs, signed_jobs] * 2), \
                patch.object(cleanup, "api", side_effect=request):
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
            plan, incomplete = plan_cleanup(attempts, {"101"}, {"1011"})
            self.assertFalse(incomplete)
            self.assertEqual(plan[0]["archives"], ["1001"])
            self.assertEqual(plan[0]["staging_tags"], [])
            self.assertEqual(plan[1]["archives"], [])
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
            retry, incomplete = plan_cleanup(attempts, {"101"}, {"1011"})
            self.assertFalse(incomplete)
            cleanup.apply_cleanup(retry, "gitlab", "fixture")
        self.assertEqual(removed, ["gitlab/jobs/1001/artifacts"])
        self.assertTrue(cleanup.has_archive(signed_jobs[1]))
        self.assertEqual([call.args[0].full_url for call in opener.open.call_args_list], [tags_url, repo_url] * 2)
        self.assertTrue(all(call.args[0].get_header("Authorization") == "Bearer fixture" for call in opener.open.call_args_list))

    def test_missing_evidence_or_inventory_failure_preserves_archives_and_other_attempts(self):
        attempts = [cleanup.ArchivedAttempt("100", ["1001"], [], None, "missing", archived_release(100, receipt=False)),
                    cleanup.ArchivedAttempt("101", ["23"], ["21"], METADATA, "valid")]
        variables = [{"key": "RELEASE_VERSION", "value": "v1.2.3"}, {"key": "QUAY_COLLECTOR_REPO", "value": ATTEMPT["quay"]}]
        for evidence in (None, variables):
            with self.subTest(evidence=evidence), patch.dict(os.environ, {}, clear=True), contextlib.redirect_stderr(io.StringIO()), \
                    patch.object(cleanup, "api", side_effect=[None, None, evidence]), \
                    patch.object(cleanup, "repository_tags", side_effect=OSError("registry unavailable")) as inventory:
                plan, incomplete = plan_cleanup(attempts)
            self.assertTrue(incomplete)
            self.assertEqual(plan[0]["archives"], [])
            self.assertEqual(plan[0]["staging_tags"], [])
            self.assertEqual(plan[1]["receipts"], ["21"])
            self.assertEqual(plan[1]["staging_tags"], [])
            self.assertEqual(inventory.called, evidence is not None)

    def test_archive_names_and_docker_records_verify_original_repositories_and_version(self):
        records = io.BytesIO()
        with ZipFile(records, "w") as archive:
            archive.writestr("dist/docker/quay.io/old/collector/multiarch_digest.txt", "fixture")
            archive.writestr("dist/docker/old/collector/multiarch_digest.txt", "fixture")
        docker = job(10, "retain-stable-docker-records")
        docker["artifacts_file"] = {"filename": "collector-v1.2.3-100-docker-records.zip"}
        preflight = job(9, "manual_release_trigger", archive=False)
        attempt = cleanup.ArchivedAttempt("100", ["10", "11"], [], None, "missing", [preflight, docker])
        environment = {"QUAY_COLLECTOR_REPO": "quay.io/current/collector", "DOCKER_HUB_COLLECTOR_REPO": "current/collector"}
        catalog = {"quay.io/old/collector": ["9.9.9-staging-100-linux", "1.2.3-staging-100-linux"],
                   "docker.io/old/collector-windows": ["1.2.3-staging-100-2019"]}
        corrupted_trace = trace_records(cleanup.ATTEMPT_LOG + "{\n")
        for mismatch in (True, False):
            with self.subTest(mismatch=mismatch), patch.dict(os.environ, environment, clear=True), \
                    patch.object(cleanup, "api", side_effect=[corrupted_trace, records.getvalue(), None]) as api, \
                    patch.object(cleanup, "repository_tags", side_effect=lambda repo, *args, **kwargs: catalog.get(repo, ())) as tags, \
                    contextlib.redirect_stderr(io.StringIO()) as warnings:
                plan, incomplete = plan_cleanup([attempt])
            self.assertEqual(api.call_count, 3)
            api.assert_any_call("gitlab/jobs/10/artifacts", "fixture", missing_ok=True, binary=True)
            self.assertTrue(all("/old/collector" in call.args[0] for call in tags.call_args_list))
            self.assertEqual(plan[0]["staging_tags"], ["docker.io/old/collector-windows:1.2.3-staging-100-2019",
                                                       "quay.io/old/collector:1.2.3-staging-100-linux"])
            self.assertEqual(incomplete, mismatch)
            self.assertEqual(plan[0]["archives"], [] if mismatch else ["11"])
            self.assertEqual(plan[0]["retained_evidence"], ["10"])
            self.assertEqual("9.9.9-staging-100-linux" in warnings.getvalue(), mismatch)
            catalog["quay.io/old/collector"] = ["1.2.3-staging-100-linux"]

    def test_partial_inventory_preserves_evidence_and_keeps_other_verified_tags(self):
        jobs = archived_release(100, receipt=False) + [job(1003, "build-push-linux-fips-image", archive=False)]
        attempt = cleanup.ArchivedAttempt("100", ["1001"], [], None, "missing", jobs)
        trace = trace_records(cleanup.ATTEMPT_LOG + json.dumps(ATTEMPT))
        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "api", side_effect=[trace, []]), \
                patch.object(cleanup, "repository_tags", side_effect=[ValueError("temporary inventory failure"),
                             ["1.2.3-staging-100-linux"]]), contextlib.redirect_stderr(io.StringIO()):
            plan, incomplete = plan_cleanup([attempt])
        self.assertTrue(incomplete)
        self.assertEqual(plan[0]["archives"], [])
        self.assertEqual(plan[0]["staging_tags"], ["quay.io/splunk/collector-fips:1.2.3-staging-100-linux"])

    def test_canceled_unstarted_publisher_does_not_require_registry_evidence(self):
        jobs = archived_release(100, receipt=False)
        jobs[2].update(status="canceled", started_at=None)
        attempt = cleanup.ArchivedAttempt("100", ["1001"], [], None, "missing", jobs)
        with patch.object(cleanup, "api") as api, patch.object(cleanup, "repository_tags") as registry:
            plan, incomplete = plan_cleanup([attempt])
        self.assertFalse(incomplete)
        self.assertEqual(plan[0]["archives"], ["1001"])
        self.assertEqual(plan[0]["staging_tags"], [])
        api.assert_not_called()
        registry.assert_not_called()

    def test_trace_only_attempt_cleans_its_repository_and_reruns_cleanly(self):
        jobs = archived_release(100, receipt=False)
        windows = job(1003, "build-push-windows-fips-image: [2022]", status="canceled", archive=False)
        windows["started_at"] = None
        jobs.append(windows)
        repo = "quay.io/old/collector"
        tag = "1.2.3-staging-100-linux"
        refs = {tag, "latest", "1.2.3", "1.2.3-staging-101-linux"}
        archive = jobs[1]
        published = jobs[2]

        def read(url, token, *, method="GET", **kwargs):
            if url == "gitlab/pipelines/100/variables":
                return [{"key": "RELEASE_VERSION", "value": "v1.2.3"}]
            if url == "gitlab/jobs/1000/trace":
                return trace_records("Manual release triggered for version v1.2.3\n")
            if url == f"gitlab/jobs/{published['id']}/trace":
                return trace_records(f"\x1b[0mBuilding and pushing {repo}:{tag}\r\n")
            if url == f"https://quay.io/api/v1/repository/old/collector/tag/{tag}" and method == "DELETE":
                refs.remove(tag)
                return 204
            if url == f"gitlab/jobs/{archive['id']}/artifacts" and method == "DELETE":
                archive["artifacts"].clear()
                return 204
            if url == f"gitlab/jobs/{archive['id']}":
                return archive
            self.fail(f"Unexpected {method} {url}")

        def inventory(repository, *_args, **_kwargs):
            self.assertEqual(repository, repo)
            return iter(refs)

        signed_jobs = archived_release(101, receipt=False)
        pipelines = [{"id": 100, "status": "failed"}, {"id": 101, "status": "success"}]
        responses = [pipelines, jobs, signed_jobs] * 2
        with patch.dict(os.environ, {"COLLECTOR_CLEANUP_QUAY_TOKEN": "fixture"}, clear=True), \
                patch.object(cleanup, "pages", side_effect=responses), patch.object(cleanup, "api", side_effect=read), \
                patch.object(cleanup, "repository_tags", side_effect=inventory):
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
            plan, incomplete = plan_cleanup(attempts, {"101"}, {"1011"})
            self.assertFalse(incomplete)
            self.assertEqual(plan[0]["staging_tags"], [f"{repo}:{tag}"])
            self.assertEqual(plan[0]["archives"], ["1001"])
            self.assertEqual(plan[1]["archives"], [])
            self.assertEqual(plan[1]["staging_tags"], [])
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
            attempts = list(cleanup.archived_attempts("gitlab", "fixture", "42"))
            retry, incomplete = plan_cleanup(attempts, {"101"}, {"1011"})
            self.assertFalse(incomplete)
            self.assertEqual(retry[0]["staging_tags"], [])
            self.assertEqual(retry[0]["archives"], [])
            cleanup.apply_cleanup(retry, "gitlab", "fixture")
        self.assertEqual(refs, {"latest", "1.2.3", "1.2.3-staging-101-linux"})
        self.assertTrue(cleanup.has_archive(signed_jobs[1]))

    def test_docker_records_as_last_evidence_survive_successful_cleanup_and_rerun(self):
        jobs = archived_release(100, receipt=False)
        docker = job(1003, "retain-stable-docker-records")
        docker["artifacts_file"] = {"filename": "collector-v1.2.3-100-docker-records.zip"}
        jobs.extend([docker, job(1004, "mirror-docker-images: [multiarch, ]", archive=False)])
        retained = io.BytesIO()
        with ZipFile(retained, "w") as records:
            records.writestr("dist/docker/quay.io/old/collector/multiarch_digest.txt", "fixture")
            records.writestr("dist/docker/docker.io/old/collector/multiarch_digest.txt", "fixture")
        refs = ["1.2.3-staging-100-linux"]

        def read(url, token, *, method="GET", **kwargs):
            if method == "GET":
                if url.endswith("/trace") or url.endswith("/variables"):
                    return None
                if url == "gitlab/jobs/1003/artifacts":
                    return retained.getvalue()
                if url == "gitlab/jobs/1001":
                    return jobs[1]
            if method == "DELETE":
                if url == "https://quay.io/api/v1/repository/old/collector/tag/1.2.3-staging-100-linux":
                    refs.clear()
                    return 204
                if url == "gitlab/jobs/1001/artifacts":
                    jobs[1]["artifacts"].clear()
                    return 204
            self.fail(f"Unexpected {method} {url}")

        responses = [[{"id": 100, "status": "failed"}], jobs] * 2
        with patch.dict(os.environ, {"COLLECTOR_CLEANUP_QUAY_TOKEN": "fixture"}, clear=True), \
                patch.object(cleanup, "pages", side_effect=responses), patch.object(cleanup, "api", side_effect=read), \
                patch.object(cleanup, "repository_tags", side_effect=lambda repo, *args, **kwargs: iter(refs if repo.startswith("quay.io/") else [])):
            first, incomplete = plan_cleanup(list(cleanup.archived_attempts("gitlab", "fixture", "42")))
            self.assertFalse(incomplete)
            self.assertEqual(first[0]["retained_evidence"], ["1003"])
            self.assertEqual(first[0]["archives"], ["1001"])
            cleanup.apply_cleanup(first, "gitlab", "fixture")
            retry, incomplete = plan_cleanup(list(cleanup.archived_attempts("gitlab", "fixture", "42")))
            self.assertFalse(incomplete)
            self.assertEqual(retry[0]["retained_evidence"], ["1003"])
            self.assertEqual(retry[0]["archives"], [])
            self.assertEqual(retry[0]["staging_tags"], [])
            cleanup.apply_cleanup(retry, "gitlab", "fixture")

    def test_expiring_publisher_evidence_is_kept_before_cleanup_and_survives_rerun(self):
        jobs = archived_release(100, receipt=False, publisher="mirror-docker-images: [multiarch, ]")
        release_archive, publisher = jobs[1:]
        publisher.update(status="success", artifacts=[{"file_type": "archive"}], artifacts_expire_at="2026-11-21T00:00:00Z")
        publisher_archive = io.BytesIO()
        with ZipFile(publisher_archive, "w") as records:
            records.writestr("dist/docker/docker.io/old/collector/multiarch_digest.txt", "fixture")
        repo = "docker.io/old/collector"
        tag = "1.2.3-staging-100"
        refs = [tag]
        changes = []

        def request(url, token=None, *, method="GET", **options):
            if url == "gitlab/jobs/1000/trace":
                return trace_records("Manual release triggered for version v1.2.3\n")
            if url == "gitlab/jobs/1002/trace" or url.endswith("/variables"):
                return None
            if url == "gitlab/jobs/1002/artifacts" and method == "GET":
                return publisher_archive.getvalue()
            if url == f"{cleanup.DOCKER_HUB_API}/auth/token":
                return {"access_token": "fixture"}
            changes.append((method, url))
            if url == "gitlab/jobs/1002/artifacts/keep" and method == "POST":
                publisher["artifacts_expire_at"] = None
                return publisher
            if url == f"{cleanup.DOCKER_HUB_API}/repositories/old/collector/tags/{tag}/" and method == "DELETE":
                refs.clear()
                return 204
            if url == "gitlab/jobs/1001/artifacts" and method == "DELETE":
                release_archive["artifacts"].clear()
                return 204
            if url == "gitlab/jobs/1001" and method == "GET":
                return release_archive
            self.fail(f"Unexpected {method} {url}")

        responses = [[{"id": 100, "status": "failed"}], jobs] * 2
        env = {"COLLECTOR_CLEANUP_DOCKER_HUB_USERNAME": "fixture", "COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN": "fixture"}
        with patch.dict(os.environ, env, clear=True), patch.object(cleanup, "pages", side_effect=responses), \
                patch.object(cleanup, "api", side_effect=request), \
                patch.object(cleanup, "repository_tags", side_effect=lambda *args, **kwargs: iter(refs)):
            first, incomplete = plan_cleanup(list(cleanup.archived_attempts("gitlab", "fixture", "42")))
            self.assertFalse(incomplete)
            self.assertEqual(first[0]["keep_evidence"], ["1002"])
            self.assertEqual(first[0]["staging_tags"], [f"{repo}:{tag}"])
            self.assertEqual(changes, [])
            cleanup.apply_cleanup(first, "gitlab", "fixture")
            self.assertEqual(changes, [
                ("POST", "gitlab/jobs/1002/artifacts/keep"),
                ("DELETE", f"{cleanup.DOCKER_HUB_API}/repositories/old/collector/tags/{tag}/"),
                ("DELETE", "gitlab/jobs/1001/artifacts"),
                ("GET", "gitlab/jobs/1001"),
            ])
            self.assertIsNone(publisher["artifacts_expire_at"])
            retry, incomplete = plan_cleanup(list(cleanup.archived_attempts("gitlab", "fixture", "42")))
            self.assertFalse(incomplete)
            self.assertNotIn("keep_evidence", retry[0])
            self.assertEqual(retry[0]["staging_tags"], [])
            self.assertEqual(retry[0]["archives"], [])
            changes.clear()
            cleanup.apply_cleanup(retry, "gitlab", "fixture")
            self.assertEqual(changes, [])

    def test_legacy_buildx_trace_over_two_mib_supplies_receipt_free_evidence(self):
        repo = "docker.io/old/collector-fips"
        tag = "1.2.3-staging-100-amd64"
        digest = "a" * 64
        buildx_record = f"pushing sha256:{digest} to {repo}:{tag}"
        # The useful reference appears after multiple read blocks and an oversized line.
        trace = ("copying layer\n" * 170000 + "x" * 70000 + "\r#1 524.2 " + buildx_record).encode()
        self.assertGreater(len(trace), 2 * 1024 * 1024)
        jobs = archived_release(100, receipt=False, publisher="mirror-docker-images: [multiarch, -fips]")
        attempt = cleanup.ArchivedAttempt("100", ["1001"], [], None, "missing", jobs)
        opener = Mock()
        opener.open.side_effect = [
            io.BytesIO(b"Manual release triggered for version v1.2.3\r\n"),
            io.BytesIO(trace),
            io.BytesIO(b'[{"key":"RELEASE_VERSION","value":"v1.2.3"}]'),
        ]
        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "build_opener", return_value=opener), \
                patch.object(cleanup, "repository_tags", return_value=[tag]) as inventory:
            plan, incomplete = cleanup.cleanup_plan([attempt], set(), set(), "https://gitlab.example.invalid", "fixture", "42")
        self.assertFalse(incomplete)
        self.assertEqual(plan[0]["staging_tags"], [f"{repo}:{tag}"])
        self.assertEqual(plan[0]["archives"], ["1001"])
        inventory.assert_called_once_with(repo, None)
        self.assertEqual(opener.open.call_count, 3)

    def test_publisher_log_version_needs_matching_release_evidence(self):
        repo = "docker.io/old/collector"
        tag = "9.9.9-staging-100"
        output = trace_records(f"#1 524.2 pushing sha256:{'a' * 64} to {repo}:{tag}\n")
        jobs = archived_release(100, receipt=False, publisher="mirror-docker-images: [multiarch, ]")
        attempt = cleanup.ArchivedAttempt("100", ["1001"], [], None, "missing", jobs)
        for recorded, expected in (("v1.2.3", False), (None, False), ("v9.9.9", True)):
            variables = [{"key": "RELEASE_VERSION", "value": recorded}] if recorded else []
            with self.subTest(recorded=recorded), patch.dict(os.environ, {}, clear=True), \
                    patch.object(cleanup, "api", side_effect=[None, output, variables]) as api, \
                    patch.object(cleanup, "repository_tags", return_value=[tag]) as inventory, \
                    contextlib.redirect_stderr(io.StringIO()):
                plan, incomplete = plan_cleanup([attempt])
            api.assert_called_with("gitlab/pipelines/100/variables", "fixture", missing_ok=True)
            self.assertEqual(incomplete, not expected)
            self.assertEqual(inventory.called, expected)
            self.assertEqual(plan[0]["staging_tags"], [f"{repo}:{tag}"] if expected else [])
            self.assertEqual(plan[0]["archives"], ["1001"] if expected else [])

    def test_stream_preserves_complete_records_at_read_boundaries_and_keeps_zip_limit(self):
        version = "Manual release triggered for version v1.2.3"
        staging = "Building and pushing quay.io/old/collector:1.2.3-staging-100-linux"
        with patch.object(cleanup, "TRACE_BLOCK_SIZE", 128):
            trace = "x" * 109 + "\r" + version + "\r\n" + "x" * 127 + "\n" + staging + "\n"
            self.assertEqual(trace_records(trace), {version, staging})
        opener = Mock()
        opener.open.return_value = io.BytesIO(b"x" * (2 * 1024 * 1024 + 1))
        with patch.object(cleanup, "build_opener", return_value=opener):
            with self.assertRaisesRegex(ValueError, "exceeds 2 MiB"):
                cleanup.api("https://gitlab.example.invalid/artifacts", "fixture", binary=True)

    def test_proof_from_an_interrupted_trace_is_discarded(self):
        response = io.BytesIO(b"Building and pushing quay.io/old/collector:1.2.3-staging-100-linux\n")
        response.length = 100
        opener = Mock()
        opener.open.return_value = response
        with patch.object(cleanup, "build_opener", return_value=opener), contextlib.redirect_stderr(io.StringIO()):
            self.assertIsNone(cleanup.read_evidence("https://gitlab.example.invalid/trace", "fixture",
                                                    reader=cleanup.release_trace_records))

    def test_failed_evidence_keep_blocks_its_deletions_and_continues_other_attempts(self):
        plan = [
            {"pipeline": "100", "keep_evidence": ["1002"], "staging_tags": [], "archives": ["1001"], "receipts": []},
            {"pipeline": "101", "staging_tags": [], "archives": ["1011"], "receipts": []},
        ]
        expiring = job(1002, "mirror-docker-images")
        expiring["artifacts_expire_at"] = "2026-11-21T00:00:00Z"
        for failed_keep in (ValueError("HTTP 404"), expiring, job(1002, "mirror-docker-images", archive=False)):
            with self.subTest(failed_keep=failed_keep), \
                    patch.object(cleanup, "api", side_effect=[failed_keep, 204, {"artifacts": []}]) as api:
                with self.assertRaisesRegex(ValueError, "Cleanup incomplete for pipeline 100"):
                    cleanup.apply_cleanup(plan, "gitlab", "fixture")
            self.assertEqual([(call.args[0], call.kwargs.get("method", "GET")) for call in api.call_args_list], [
                ("gitlab/jobs/1002/artifacts/keep", "POST"),
                ("gitlab/jobs/1011/artifacts", "DELETE"),
                ("gitlab/jobs/1011", "GET"),
            ])

    def test_evidence_is_kept_even_if_registry_credentials_are_unavailable(self):
        plan = [{"pipeline": "100", "keep_evidence": ["1002"],
                 "staging_tags": ["quay.io/old/collector:1.2.3-staging-100-linux"], "archives": ["1001"], "receipts": []}]
        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "api", return_value=job(1002, "build-push-linux-image")) as api:
            with self.assertRaisesRegex(ValueError, "Missing quay.io cleanup credentials"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
        api.assert_called_once_with("gitlab/jobs/1002/artifacts/keep", "fixture", method="POST")

    def test_scope_matches_publish_job_matrices(self):
        cases = {
            "build-push-linux-image": ("quay.io", ("",)),
            "build-push-linux-fips-image": ("quay.io", ("-fips",)),
            "build-push-windows-image: [2019]": ("quay.io", ("", "-windows")),
            "build-push-windows-fips-image: [2022]": ("quay.io", ("-fips", "-fips-windows")),
            "push-multiarch-manifest: [multiarch, ]": ("quay.io", ("",)),
            "push-multiarch-manifest: [windows_multiarch, -fips]": ("quay.io", ("-fips-windows",)),
            "mirror-docker-images: [multiarch, -fips]": ("docker.io", ("-fips",)),
            "mirror-docker-images: [windows_multiarch, ]": ("docker.io", ("-windows",)),
        }
        for name, expected in cases.items():
            with self.subTest(name=name):
                self.assertEqual(cleanup.publishing_job(job(100, name, archive=False)), expected)

    def test_historical_pipeline_override_is_used_when_a_publisher_has_no_archive(self):
        jobs = archived_release(100, receipt=False, archive=False, publisher="mirror-docker-images: [multiarch, -fips]")
        attempt = cleanup.ArchivedAttempt("100", [], [], None, "missing", jobs)
        variables = [{"key": "RELEASE_VERSION", "value": "v1.2.3"}, {"key": "DOCKER_HUB_COLLECTOR_REPO", "value": "other/collector"}]
        environment = {"COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN": "fixture"}
        with patch.dict(os.environ, environment, clear=True), patch.object(cleanup, "api", side_effect=[None, None, variables]), \
                patch.object(cleanup, "repository_tags", return_value=["1.2.3-staging-100"]) as tags, \
                patch.object(cleanup, "docker_hub_access_token", return_value="hub-read-fixture"):
            plan, incomplete = plan_cleanup([attempt])
        self.assertFalse(incomplete)
        self.assertIn("docker.io/other/collector-fips:1.2.3-staging-100", plan[0]["staging_tags"])
        tags.assert_called_once_with("docker.io/other/collector-fips", "hub-read-fixture")

    def test_conflicting_receipts_fall_back_to_registry_evidence(self):
        alternate = dict(ATTEMPT, hub="other/collector")
        with patch.object(cleanup, "api", side_effect=[ATTEMPT, alternate]), contextlib.redirect_stderr(io.StringIO()):
            tags, receipts, status = cleanup.read_receipts("gitlab", "fixture", "42", "100", ["11", "12"])
        self.assertEqual((tags, receipts, status), (None, [], "unreadable"))

    def test_apply_only_requires_credentials_for_registries_with_tags(self):
        plan = [{"pipeline": "100", "staging_tags": ["quay.io/splunk/collector:1.2.3-staging-100-linux"], "archives": [], "receipts": []}]
        with patch.dict(os.environ, {"COLLECTOR_CLEANUP_QUAY_TOKEN": "quay-fixture"}, clear=True), \
                patch.object(cleanup, "api", return_value=204) as api:
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
        api.assert_called_once_with("https://quay.io/api/v1/repository/splunk/collector/tag/1.2.3-staging-100-linux",
                                    "quay-fixture", method="DELETE", missing_ok=True)

    def test_cleanup_only_deletes_tags_and_unpinned_archive_ids(self):
        tag = "1.2.3-staging-100"
        plan = [{"pipeline": "100", "staging_tags": [f"quay.io/splunk/collector:{tag}", f"docker.io/splunk/collector:{tag}"],
                 "archives": ["13"], "receipts": ["11"]}]
        credentials = {
            "COLLECTOR_CLEANUP_QUAY_TOKEN": "quay-fixture",
            "COLLECTOR_CLEANUP_DOCKER_HUB_USERNAME": "hub-fixture",
            "COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN": "hub-secret-fixture",
        }
        auth = {"access_token": "hub-access-fixture"}
        removed = {"artifacts": []}
        with patch.dict(os.environ, credentials), patch.object(cleanup, "api", side_effect=[auth, 204, 204, 204, removed, 204, removed]) as api:
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
            api.assert_any_call(f"{cleanup.DOCKER_HUB_API}/auth/token", method="POST",
                                payload={"identifier": "hub-fixture", "secret": "hub-secret-fixture"})
            deletes = [call for call in api.call_args_list if call.kwargs.get("method") == "DELETE"]
            self.assertEqual([call.args for call in deletes], [
                (f"https://quay.io/api/v1/repository/splunk/collector/tag/{tag}", "quay-fixture"),
                (f"{cleanup.DOCKER_HUB_API}/repositories/splunk/collector/tags/{tag}/", "hub-access-fixture"),
                ("gitlab/jobs/13/artifacts", "fixture"),
                ("gitlab/jobs/11/artifacts", "fixture"),
            ])
            api.assert_called_with("gitlab/jobs/11", "fixture")

            api.reset_mock(side_effect=True)
            api.side_effect = [auth, 204, ValueError("registry failure")]
            with self.assertRaisesRegex(ValueError, "registry failure"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
            self.assertEqual(api.call_count, 3)

            api.reset_mock(side_effect=True)
            api.side_effect = [auth, 204, None, {"name": tag}]
            with self.assertRaisesRegex(ValueError, "Docker Hub did not delete"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
            api.assert_called_with(f"{cleanup.DOCKER_HUB_API}/namespaces/splunk/repositories/collector/tags/{tag}",
                                   "hub-access-fixture", missing_ok=True)

            api.reset_mock(side_effect=True)
            api.side_effect = [auth, None, {"tags": [{"name": tag}]}]
            with self.assertRaisesRegex(ValueError, "Quay did not delete"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
            self.assertEqual(api.call_count, 3)

            api.reset_mock(side_effect=True)
            api.side_effect = [auth, None, {"tags": []}, None, None, 204, removed, 204, removed]
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
            api.assert_called_with("gitlab/jobs/11", "fixture")

    def test_apply_with_pinned_staging_tags_only_uses_gitlab(self):
        plan = [{"pipeline": "100", "staging_tags": [], "archives": ["13"], "receipts": ["11"]}]
        removed = {"artifacts": []}
        with patch.dict(os.environ, {}, clear=True), patch.object(cleanup, "api", side_effect=[204, removed, 204, removed]) as api:
            cleanup.apply_cleanup(plan, "gitlab", "fixture")
            self.assertEqual([call.args[0] for call in api.call_args_list], [
                "gitlab/jobs/13/artifacts", "gitlab/jobs/13", "gitlab/jobs/11/artifacts", "gitlab/jobs/11",
            ])
            api.reset_mock(side_effect=True)
            api.side_effect = ValueError("archive deletion failure")
            with self.assertRaisesRegex(ValueError, "archive deletion failure"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
            api.assert_called_once_with("gitlab/jobs/13/artifacts", "fixture", method="DELETE")

            api.reset_mock(side_effect=True)
            api.side_effect = [204, job(13, "prepare-stable-github-release")]
            with self.assertRaisesRegex(ValueError, "GitLab did not delete the archive"):
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
            api.assert_called_with("gitlab/jobs/13", "fixture")

    def test_apply_continues_after_tag_archive_and_credential_failures(self):
        plan = [
            {"pipeline": "100", "staging_tags": ["quay.io/old/collector:1.2.3-staging-100"], "archives": ["10"], "receipts": ["11"]},
            {"pipeline": "101", "staging_tags": [], "archives": ["20", "22"], "receipts": ["21"]},
            {"pipeline": "102", "staging_tags": ["docker.io/old/collector:1.2.3-staging-102"], "archives": ["30"], "receipts": []},
            {"pipeline": "103", "staging_tags": [], "archives": ["40"], "receipts": []},
        ]
        with patch.dict(os.environ, {"COLLECTOR_CLEANUP_QUAY_TOKEN": "quay-fixture"}, clear=True), \
                patch.object(cleanup, "delete_staging_tag", side_effect=ValueError("tag deletion failure")), \
                patch.object(cleanup, "api", side_effect=[ValueError("archive deletion failure"), 204, {"artifacts": []}]) as api:
            with self.assertRaises(ValueError) as error:
                cleanup.apply_cleanup(plan, "gitlab", "fixture")
        self.assertIn("pipeline 100: tag deletion failure", str(error.exception))
        self.assertIn("pipeline 101: archive deletion failure", str(error.exception))
        self.assertIn("pipeline 102: Missing docker.io cleanup credentials", str(error.exception))
        self.assertEqual([call.args[0] for call in api.call_args_list], [
            "gitlab/jobs/20/artifacts", "gitlab/jobs/40/artifacts", "gitlab/jobs/40",
        ])

    def test_api_does_not_forward_credentials_outside_artifact_origin(self):
        redirects = cleanup.ArtifactRedirects()
        req = cleanup.Request("https://gitlab.example.invalid", headers={"Authorization": "Bearer fixture"})
        headers = redirects.redirect_request(req, None, 302, "Found", {}, "https://storage.example.invalid/receipt").headers
        self.assertNotIn("Authorization", headers)
        post = cleanup.Request("https://hub.docker.com/v2/auth/token", method="POST")
        self.assertIsNone(redirects.redirect_request(post, None, 302, "Found", {}, "https://other.example.invalid/"))

    def test_registry_lookups_reject_redirects_that_would_drop_authentication(self):
        redirects = cleanup.ArtifactRedirects(same_origin=True)
        url = f"{cleanup.DOCKER_HUB_API}/namespaces/old/repositories/collector"
        req = cleanup.Request(url, headers={"Authorization": "Bearer fixture"})
        self.assertIsNone(redirects.redirect_request(req, None, 302, "Found", {}, "https://other.example.invalid/"))
        redirected = redirects.redirect_request(req, None, 302, "Found", {}, f"{url}/")
        self.assertEqual(redirected.get_header("Authorization"), "Bearer fixture")

    def test_preview_does_not_apply(self):
        environment = {"CI_PROJECT_ID": "42", "CI_API_V4_URL": "https://gitlab.example.invalid", "COLLECTOR_CLEANUP_GITLAB_TOKEN": "fixture", "GITHUB_TOKEN": "fixture"}
        with patch.dict(os.environ, environment), patch("sys.argv", ["cleanup", "plan"]), \
                patch.object(cleanup, "archived_attempts", return_value=[cleanup.ArchivedAttempt("100", ["12"], ["11"], METADATA, "valid")]), \
                patch.object(cleanup, "pinned_release_inputs", return_value=({"100"}, {"12"})), \
                patch.object(cleanup, "apply_cleanup") as apply, contextlib.redirect_stdout(io.StringIO()) as output:
            cleanup.main()
        self.assertEqual(json.loads(output.getvalue())["archives"], [])
        apply.assert_not_called()


if __name__ == "__main__":
    unittest.main()
