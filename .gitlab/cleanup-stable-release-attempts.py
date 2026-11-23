#!/usr/bin/env python3
"""Remove staging tags and unpinned artifacts from archived Collector releases."""

import argparse
import json
import os
import re
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass, field
from http.client import HTTPException
from io import BytesIO
from urllib.error import HTTPError
from urllib.parse import urlencode, urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener
from zipfile import BadZipFile, ZipFile


GITHUB_REPO = "https://github.com/signalfx/splunk-otel-collector.git"
GITHUB_API = "https://api.github.com/repos/signalfx/splunk-otel-collector"
STABLE_TAG = re.compile(r"v[0-9]+\.[0-9]+\.[0-9]+")
STAGING_TAG = re.compile(r"([0-9]+\.[0-9]+\.[0-9]+)-staging-([1-9][0-9]*)(.*)")
PUBLISHED_REF = re.compile(
    r"(?:Building(?: and pushing)?|Publishing image|pushing sha256:[a-f0-9]{64} to) "
    r"([a-z0-9_./-]+):([a-z0-9.-]+)(?=$|\s|\x1b)"
)
LEGACY_RELEASE_LOG = re.compile(r"Manual release triggered for version (v[0-9]+\.[0-9]+\.[0-9]+)(?=$|\s|\x1b)")
SNAPSHOT = "Collector-Release-Snapshot: "
ATTEMPT_LOG = "Collector-Release-Attempt: "
ATTEMPT_FILE = "dist/release-inputs/attempt.json"
ARCHIVE_JOBS = {
    "prepare-stable-github-release": "release",
    "retain-stable-docker-records": "docker-records",
    "build-stable-github-release-tools": "release-tools",
}
QUAY_PUBLISHERS = {
    "build-push-linux-image", "build-push-linux-fips-image", "build-push-windows-image",
    "build-push-windows-fips-image", "push-multiarch-manifest",
}
HUB_PUBLISHER = "mirror-docker-images"
STOPPED_STATUSES = {"success", "failed", "canceled", "skipped", "manual"}
WINDOWS_TAGS = ("", "-2019", "-2022")
REPOSITORY_TAGS = {
    "": WINDOWS_TAGS + ("-amd64", "-arm64", "-ppc64le"),
    "-fips": WINDOWS_TAGS + ("-amd64", "-arm64"),
    "-windows": WINDOWS_TAGS,
    "-fips-windows": WINDOWS_TAGS,
}
PUBLISH_REPOSITORIES = (
    ("quay.io", "quay", "QUAY_COLLECTOR_REPO"),
    ("docker.io", "hub", "DOCKER_HUB_COLLECTOR_REPO"),
)
DOCKER_HUB_API = "https://hub.docker.com/v2"
IO_ERRORS = (KeyError, TypeError, ValueError, OSError, HTTPException)
TRACE_BLOCK_SIZE = 64 * 1024


@dataclass
class ArchivedAttempt:
    pipeline: str
    archives: list[str]
    receipts: list[str]
    receipt: tuple[str, dict[str, str]] | None
    receipt_status: str
    jobs: list[dict] = field(default_factory=list)


class ArtifactRedirects(HTTPRedirectHandler):
    def __init__(self, *, same_origin=False):
        self.same_origin = same_origin

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        if req.get_method() != "GET":
            return None
        cross_origin = urlsplit(newurl)[:2] != urlsplit(req.full_url)[:2]
        if self.same_origin and cross_origin:
            return None
        request = super().redirect_request(req, fp, code, msg, headers, newurl)
        if request and cross_origin:
            request.remove_header("Authorization")
        return request


class NotFound(ValueError):
    pass


def api(url, token=None, *, method="GET", payload=None, delete_statuses=(204,), missing_ok=False,
        binary=False, reader=None, same_origin=False):
    headers = {"Accept": "*/*" if binary or reader else "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    data = None
    if payload is not None:
        headers["Content-Type"] = "application/json"
        data = json.dumps(payload).encode()
    request = Request(url, method=method, headers=headers, data=data)
    try:
        with build_opener(ArtifactRedirects(same_origin=same_origin)).open(request, timeout=30) as response:
            if method == "DELETE":
                if response.status not in delete_statuses:
                    raise ValueError(f"{method} {url} returned unexpected HTTP {response.status}")
                return response.status
            if reader:
                result = reader(response)
                if getattr(response, "length", None):
                    raise HTTPException(f"Incomplete trace from {url}")
                return result
            if binary:
                # Docker records are small ZIP archives. Refuse oversized downloads.
                data = response.read(2 * 1024 * 1024 + 1)
                if len(data) > 2 * 1024 * 1024:
                    raise ValueError(f"Cleanup evidence from {url} exceeds 2 MiB")
                return data
            return json.load(response)
    except HTTPError as error:
        if error.code == 404:
            if missing_ok:
                return None
            raise NotFound(f"{method} {url} returned HTTP 404") from error
        raise ValueError(f"{method} {url} returned HTTP {error.code}") from error


def pages(url, token, **params):
    for page in range(1, 1001):
        query = urlencode(dict(params, page=page, per_page=100))
        batch = api(f"{url}?{query}", token)
        if not isinstance(batch, list):
            raise ValueError(f"Expected a paginated list from {url}")
        yield from batch
        if len(batch) < 100:
            return
    raise ValueError(f"Pagination limit reached for {url}")


def positive_id(value):
    if not re.fullmatch(r"[1-9][0-9]*", str(value)):
        raise ValueError("Invalid pipeline, project, or job ID")
    return str(value)


def pinned_release_inputs(project, github_token):
    output = subprocess.check_output(
        ["git", "ls-remote", "--tags", GITHUB_REPO, "refs/tags/v*"], text=True, timeout=60
    )
    refs = dict(line.split()[::-1] for line in output.splitlines())
    annotated = [
        ref[:-3]
        for ref in refs
        if ref.startswith("refs/tags/") and ref.endswith("^{}") and STABLE_TAG.fullmatch(ref[len("refs/tags/"):-3])
    ]
    if not annotated:
        raise ValueError("Cannot establish release pins without annotated stable tags")

    pinned_pipelines, pinned_jobs = set(), set()
    for ref in annotated:
        tag = api(f"{GITHUB_API}/git/tags/{refs[ref]}", github_token)
        if tag["sha"] != refs[ref] or tag["tag"] != ref.removeprefix("refs/tags/"):
            raise ValueError(f"Unexpected tag metadata for {ref}")
        entries = [line.removeprefix(SNAPSHOT) for line in tag["message"].splitlines() if line.startswith(SNAPSHOT)]
        if not entries:
            continue
        if len(entries) != 1:
            raise ValueError(f"Ambiguous release snapshot on {ref}")
        # Preserve pins even if GitHub cannot verify an older signing key today.
        snapshot = json.loads(entries[0])
        if positive_id(snapshot.get("project", project)) != project:
            continue
        pinned_pipelines.add(positive_id(snapshot["pipeline"]))
        pinned_jobs.update(positive_id(snapshot[part]["job"]) for part in ("assets", "docker", "tools"))
    return pinned_pipelines, pinned_jobs


def repository_root(registry, root):
    if not isinstance(root, str):
        raise ValueError(f"Missing {registry} release repository")
    if registry == "quay.io" and not root.startswith("quay.io/"):
        raise ValueError("Release attempt metadata must qualify the Quay repository")
    repo = root.removeprefix(f"{registry}/")
    if not re.fullmatch(r"[a-z0-9][a-z0-9_.-]*/[a-z0-9][a-z0-9_.-]*", repo):
        raise ValueError(f"Invalid {registry} repository in release attempt metadata")
    return f"{registry}/{repo}"


def tag_suffixes(registry, repository_suffix):
    tags = REPOSITORY_TAGS[repository_suffix]
    if registry == "quay.io" and repository_suffix in ("", "-fips"):
        return tags + ("-linux",)
    return tags


def attempt_metadata(data, project, pipeline):
    if data["project"] != project or data["pipeline"] != pipeline:
        raise ValueError("Release metadata belongs to a different project or pipeline")
    if not STABLE_TAG.fullmatch(data["version"]):
        raise ValueError("Invalid stable release version in attempt metadata")
    roots = {registry: repository_root(registry, data[field]) for registry, field, _ in PUBLISH_REPOSITORIES}
    return data["version"], roots


def has_archive(job):
    return any(artifact["file_type"] == "archive" for artifact in job.get("artifacts") or ())


def read_receipts(gitlab, token, project, pipeline, receipts):
    metadata = None
    readable = []
    conflicting = False
    status = "valid" if receipts else "missing"
    for job in receipts:
        try:
            receipt = api(f"{gitlab}/jobs/{job}/artifacts/{ATTEMPT_FILE}", token)
            parsed = attempt_metadata(receipt, project, pipeline)
            conflicting = conflicting or (metadata is not None and metadata != parsed)
            metadata = parsed
            readable.append(job)
        except IO_ERRORS:
            status = "unreadable"
            print(f"Warning: unreadable cleanup receipt on pipeline {pipeline}, job {job}.", file=sys.stderr)
    if conflicting:
        print(f"Warning: conflicting cleanup receipts for pipeline {pipeline}.", file=sys.stderr)
        return None, [], "unreadable"
    return metadata, readable, status


def archived_attempts(gitlab, token, project):
    for pipeline in pages(f"{gitlab}/pipelines", token, ref="main", source="parent_pipeline"):
        if pipeline["status"] not in STOPPED_STATUSES:
            continue
        pipeline_id = positive_id(pipeline["id"])
        jobs = list(pages(f"{gitlab}/pipelines/{pipeline_id}/jobs", token, include_retried="true"))
        if any(job.get("archived") is not True or job["status"] not in STOPPED_STATUSES for job in jobs):
            continue
        if ARCHIVE_JOBS.keys().isdisjoint(job["name"] for job in jobs):
            continue

        receipts = [
            positive_id(job["id"])
            for job in jobs
            if job["name"] == "manual_release_trigger" and has_archive(job)
        ]
        archives = [positive_id(job["id"]) for job in jobs if job["name"] in ARCHIVE_JOBS and has_archive(job)]
        receipt, receipts, status = read_receipts(gitlab, token, project, pipeline_id, receipts)
        yield ArchivedAttempt(pipeline_id, archives, receipts, receipt, status, jobs)


def docker_hub_access_token():
    token = api(f"{DOCKER_HUB_API}/auth/token", method="POST", payload={
        "identifier": os.environ["COLLECTOR_CLEANUP_DOCKER_HUB_USERNAME"],
        "secret": os.environ["COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN"],
    })["access_token"]
    if not token:
        raise ValueError("Missing Docker Hub cleanup token")
    return token


def repository_tags(repository, token=None):
    registry, repo = repository.split("/", 1)
    if registry == "quay.io":
        repo_url = f"https://quay.io/api/v1/repository/{repo}"
        url = f"{repo_url}/tag/"
        metadata_url = f"{repo_url}?includeTags=false"
        items, next_page = "tags", "has_additional"
        params = {"onlyActiveTags": "true", "limit": 100}
    else:
        namespace, name = repo.split("/", 1)
        metadata_url = f"{DOCKER_HUB_API}/namespaces/{namespace}/repositories/{name}"
        url = f"{metadata_url}/tags"
        items, next_page = "results", "next"
        params = {"page_size": 100}
    for page in range(1, 1001):
        query = urlencode(dict(params, page=page))
        try:
            result = api(f"{url}?{query}", token, same_origin=True)
        except NotFound:
            if page != 1 or not token:
                raise
            # A missing tag page alone cannot establish that the repository is gone.
            try:
                api(metadata_url, token, same_origin=True)
            except NotFound:
                return
            raise ValueError(f"Could not confirm absence of {repository} after its tag inventory returned 404")
        yield from (item["name"] for item in result[items])
        if not result[next_page]:
            return
    raise ValueError(f"Pagination limit reached for {repository}")


def publishing_job(job):
    if job["status"] in {"manual", "skipped"} or not job.get("started_at", True):
        return None
    name, _, matrix = job["name"].partition(":")
    if name in ("push-multiarch-manifest", HUB_PUBLISHER):
        registry = "docker.io" if name == HUB_PUBLISHER else "quay.io"
        variant = re.fullmatch(r"\[(multiarch|windows_multiarch),\s*(-fips)?\s*\]", matrix.strip())
        if not variant:
            # Keep the full scope when GitLab does not supply the matrix values.
            return registry, tuple(REPOSITORY_TAGS)
        fips = variant[2] or ""
        windows = "-windows" if variant[1] == "windows_multiarch" else ""
        return registry, (f"{fips}{windows}",)
    if name in QUAY_PUBLISHERS:
        fips = "-fips" if name.endswith("-fips-image") else ""
        if name.startswith("build-push-windows-"):
            return "quay.io", (fips, f"{fips}-windows")
        return "quay.io", (fips,)
    return None


def read_evidence(url, token, **options):
    try:
        return api(url, token, missing_ok=True, **options)
    except IO_ERRORS:
        print(f"Warning: could not read cleanup evidence from {url}.", file=sys.stderr)
        return None


def trace_lines(response):
    pending = b""
    oversized = False
    while block := response.read(TRACE_BLOCK_SIZE):
        *complete, tail = re.split(br"[\r\n]", block)
        if complete:
            if not oversized and len(pending) + len(complete[0]) <= TRACE_BLOCK_SIZE:
                yield pending + complete[0]
            yield from complete[1:]
            pending, oversized = tail, False
        elif not oversized:
            if len(pending) + len(tail) > TRACE_BLOCK_SIZE:
                pending, oversized = b"", True
            else:
                pending += tail
    if pending:
        yield pending


def release_trace_records(response):
    records = set()
    for raw in trace_lines(response):
        line = raw.decode(errors="replace")
        _, marker, data = line.partition(ATTEMPT_LOG)
        if marker and data.startswith("{"):
            records.add(marker + data)
        records.update(match[0] for match in LEGACY_RELEASE_LOG.finditer(line))
        records.update(match[0] for match in PUBLISHED_REF.finditer(line))
    return records


def historical_metadata(attempt, publishers, repositories, gitlab, token, project):
    versions = defaultdict(set)
    roots = defaultdict(lambda: defaultdict(set))
    if attempt.receipt:
        version, receipt_roots = attempt.receipt
        versions[version].update(attempt.receipts)
        for registry, root in receipt_roots.items():
            roots[registry][root].update(attempt.receipts)
    for job in attempt.jobs:
        archive_kind = ARCHIVE_JOBS.get(job["name"])
        filename = (job.get("artifacts_file") or {}).get("filename", "")
        job_id = positive_id(job["id"])
        if archive_kind and has_archive(job):
            name = re.fullmatch(rf"collector-(v[0-9]+\.[0-9]+\.[0-9]+)-{attempt.pipeline}-{archive_kind}\.zip", filename)
            if name:
                versions[name[1]].add(job_id)

        if job["name"] == "manual_release_trigger" and job.get("started_at", True):
            trace = read_evidence(f"{gitlab}/jobs/{job_id}/trace", token, reader=release_trace_records) or ()
            for line in trace:
                _, marker, data = line.partition(ATTEMPT_LOG)
                if not marker or not data.startswith("{"):
                    legacy = LEGACY_RELEASE_LOG.search(line)
                    if legacy:
                        versions[legacy[1]].add(None)
                    continue
                try:
                    metadata, _ = json.JSONDecoder().raw_decode(data)
                except (KeyError, TypeError, ValueError):
                    print(f"Warning: unreadable preflight metadata on pipeline {attempt.pipeline}.", file=sys.stderr)
                    continue
                if metadata.get("project") != project or metadata.get("pipeline") != attempt.pipeline:
                    raise ValueError("Preflight metadata belongs to a different project or pipeline")
                try:
                    version, trace_roots = attempt_metadata(metadata, project, attempt.pipeline)
                except IO_ERRORS:
                    print(f"Warning: unreadable preflight metadata on pipeline {attempt.pipeline}.", file=sys.stderr)
                    continue
                versions[version].add(None)
                for registry, root in trace_roots.items():
                    roots[registry][root].add(None)

    # A publisher trace can identify the repository, but its tag needs an independent version check.
    independent_version = bool(versions)
    trace_registries = {registry for registry in repositories if not any(None in ids for ids in roots[registry].values())}
    for job_id, (registry, suffixes) in publishers.items():
        if registry not in trace_registries:
            continue
        # Windows build logs show the primary repository before copying to its legacy sibling.
        if len(suffixes) > 2:
            continue
        suffix = suffixes[0]
        trace = read_evidence(f"{gitlab}/jobs/{job_id}/trace", token, reader=release_trace_records) or ()
        for record in trace:
            log = PUBLISHED_REF.fullmatch(record)
            if log:
                tag = STAGING_TAG.fullmatch(log[2])
                if tag and tag[2] == attempt.pipeline and tag[3] in tag_suffixes(registry, suffix) and log[1].endswith(suffix):
                    root = repository_root(registry, log[1].removesuffix(suffix))
                    versions[f"v{tag[1]}"].add(None)
                    roots[registry][root].add(None)

    # Only Docker job ZIPs contain repository records. Other release archives can be large.
    needed = {registry for registry in repositories if not any(None in ids for ids in roots[registry].values())}
    for job in attempt.jobs:
        name = job["name"].partition(":")[0]
        if name in QUAY_PUBLISHERS | {HUB_PUBLISHER, "retain-stable-docker-records"} and has_archive(job):
            job_id = positive_id(job["id"])
            archive = read_evidence(f"{gitlab}/jobs/{job_id}/artifacts", token, binary=True)
            if archive:
                try:
                    with ZipFile(BytesIO(archive)) as records:
                        for path in records.namelist():
                            match = re.fullmatch(r"dist/docker/((?:quay\.io/|docker\.io/)?[^/]+/[^/]+)/[^/]+", path)
                            if match:
                                root = match[1]
                                registry = "quay.io" if root.startswith("quay.io/") else "docker.io"
                                if registry in repositories:
                                    roots[registry][repository_root(registry, root)].add(job_id)
                except BadZipFile:
                    print(f"Warning: unreadable Docker records on pipeline {attempt.pipeline}, job {job['id']}.", file=sys.stderr)

    if attempt.receipt_status != "valid" or not any(None in ids for ids in versions.values()) or needed:
        variables = read_evidence(f"{gitlab}/pipelines/{attempt.pipeline}/variables", token) or []
        missing = {registry for registry in needed if not roots[registry]}
        for variable in variables:
            if variable["key"] == "RELEASE_VERSION":
                versions[variable["value"]].add(None)
                independent_version = True
            for registry, _, key in PUBLISH_REPOSITORIES:
                if variable["key"] == key and registry in needed:
                    root = repository_root(registry, variable["value"])
                    if registry in missing or root in roots[registry]:
                        roots[registry][root].add(None)

    if not independent_version or len(versions) != 1 or not STABLE_TAG.fullmatch(next(iter(versions))):
        raise ValueError("Could not verify a single historical release version")
    if any(len(roots[registry]) != 1 for registry in repositories):
        raise ValueError("Could not verify the original publishing repositories")
    version, version_sources = next(iter(versions.items()))
    verified_roots = {registry: next(iter(roots[registry])) for registry in repositories}
    sources = [roots[registry][root] for registry, root in verified_roots.items()] + [version_sources]
    return version, verified_roots, sources


def attempt_tags(attempt, gitlab, token, project, inventory, registry_tokens):
    publishers = {}
    repositories = defaultdict(set)
    for job in attempt.jobs:
        publish = publishing_job(job)
        if not publish:
            continue
        publishers[positive_id(job["id"])] = publish
        registry, suffixes = publish
        repositories[registry].update(suffixes)
    if any(job["name"] == "retain-stable-docker-records" and has_archive(job) for job in attempt.jobs):
        for registry, _, _ in PUBLISH_REPOSITORIES:
            repositories.setdefault(registry, set(REPOSITORY_TAGS))
    if not repositories:
        return set(), False, []
    version, roots, sources = historical_metadata(attempt, publishers, repositories, gitlab, token, project)
    refs = set()
    unresolved = False
    for registry, root in sorted(roots.items()):
        for suffix in sorted(repositories[registry]):
            repository = root + suffix
            suffixes = tag_suffixes(registry, suffix)
            try:
                if repository not in inventory:
                    if registry not in registry_tokens:
                        registry_tokens[registry] = inventory_token(registry)
                    inventory[repository] = set(repository_tags(repository, registry_tokens[registry]))
            except IO_ERRORS as error:
                unresolved = True
                print(f"Warning: could not verify staging tags in {repository} for pipeline {attempt.pipeline}. {error}",
                      file=sys.stderr)
                continue
            for tag in inventory[repository]:
                match = STAGING_TAG.fullmatch(tag)
                if match and match[2] == attempt.pipeline and match[3] in suffixes:
                    if f"v{match[1]}" == version:
                        refs.add(f"{repository}:{tag}")
                    else:
                        unresolved = True
                        print(f"Warning: {repository}:{tag} disagrees with pipeline {attempt.pipeline}'s version {version}.",
                              file=sys.stderr)
    return refs, unresolved, sources


def inventory_token(registry):
    if registry == "quay.io":
        return os.environ.get("COLLECTOR_CLEANUP_QUAY_TOKEN")
    if os.environ.get("COLLECTOR_CLEANUP_DOCKER_HUB_TOKEN"):
        return docker_hub_access_token()
    return None


def retain_evidence(attempt, sources, deletions):
    jobs = {positive_id(job["id"]): job for job in attempt.jobs if has_archive(job)}
    permanent = {job for job, data in jobs.items() if "artifacts_expire_at" in data and data["artifacts_expire_at"] is None}
    small = set(attempt.receipts) | {job for job, data in jobs.items() if data["name"] == "retain-stable-docker-records"}
    retained, keep = set(), set()
    for evidence in sources:
        if None in evidence or evidence & (retained | (permanent - deletions)):
            continue
        candidates = evidence & jobs.keys()
        if not candidates:
            raise ValueError("No surviving archive contains the cleanup evidence")
        preferred = candidates & permanent or candidates
        job = min(preferred & small or preferred, key=int)
        retained.add(job)
        if job not in permanent:
            keep.add(job)
    return retained, keep


def cleanup_plan(attempts, pinned_pipelines, pinned_jobs, gitlab, token, project):
    inventory, registry_tokens = {}, {}
    incomplete = False
    plan = []
    for attempt in attempts:
        pinned = attempt.pipeline in pinned_pipelines or not pinned_jobs.isdisjoint(attempt.archives + attempt.receipts)
        tags = set()
        retained, keep = set(), set()
        unresolved = False
        archives = set(attempt.archives) - pinned_jobs
        # Keep receipts with failed reads so they can be recovered on a later run.
        receipts = set(attempt.receipts) - pinned_jobs if attempt.receipt_status == "valid" else set()
        if not pinned:
            try:
                tags, unresolved, sources = attempt_tags(attempt, gitlab, token, project, inventory, registry_tokens)
                retained, keep = retain_evidence(attempt, sources, archives | receipts)
            except IO_ERRORS as error:
                tags = set()
                unresolved = True
                print(f"Warning: pipeline {attempt.pipeline} needs manual cleanup review. {error}", file=sys.stderr)
        incomplete |= unresolved
        entry = {
            "pipeline": attempt.pipeline,
            "receipt_status": attempt.receipt_status,
            "staging_tags": sorted(tags),
            "archives": [] if unresolved else sorted(archives - retained, key=int),
            "receipts": [] if unresolved else sorted(receipts - retained, key=int),
        }
        if retained:
            entry["retained_evidence"] = sorted(retained, key=int)
        if keep:
            entry["keep_evidence"] = sorted(keep, key=int)
        plan.append(entry)
    return plan, incomplete


def delete_staging_tag(ref, quay_token, hub_token):
    registry, repo_tag = ref.split("/", 1)
    repo, tag = repo_tag.split(":", 1)
    if registry == "quay.io":
        tag_url = f"https://quay.io/api/v1/repository/{repo}/tag/"
        deleted = api(f"{tag_url}{tag}", quay_token, method="DELETE", missing_ok=True)
        if deleted is None:
            query = urlencode({"specificTag": tag, "onlyActiveTags": "true"})
            tags = api(f"{tag_url}?{query}", quay_token, missing_ok=True)
            if tags is not None and tags["tags"]:
                raise ValueError(f"Quay did not delete {ref}")
    elif registry == "docker.io":
        # Use the tag endpoint from Docker's hub-tool. A staging tag can share a
        # manifest digest with a release tag.
        deleted = api(f"{DOCKER_HUB_API}/repositories/{repo}/tags/{tag}/", hub_token,
                      method="DELETE", delete_statuses=(200, 204), missing_ok=True)
        if deleted is None:
            namespace, name = repo.split("/", 1)
            tag_url = f"{DOCKER_HUB_API}/namespaces/{namespace}/repositories/{name}/tags/{tag}"
            if api(tag_url, hub_token, missing_ok=True) is not None:
                raise ValueError(f"Docker Hub did not delete {ref}")
    else:
        raise ValueError(f"Unsupported cleanup registry {registry}")


def apply_cleanup(plan, gitlab, token):
    registry_tokens = {}
    failures = []
    for entry in plan:
        try:
            for job in entry.get("keep_evidence", []):
                kept = api(f"{gitlab}/jobs/{job}/artifacts/keep", token, method="POST")
                if positive_id(kept["id"]) != job or kept["artifacts_expire_at"] is not None or not has_archive(kept):
                    raise ValueError(f"GitLab did not retain the cleanup evidence for job {job}")
            registries = {ref.split("/", 1)[0] for ref in entry["staging_tags"]}
            for registry in sorted(registries):
                if registry not in registry_tokens:
                    registry_token = inventory_token(registry)
                    if not registry_token:
                        raise ValueError(f"Missing {registry} cleanup credentials")
                    registry_tokens[registry] = registry_token
            for ref in entry["staging_tags"]:
                delete_staging_tag(ref, registry_tokens.get("quay.io"), registry_tokens.get("docker.io"))
            # Keep the receipt and original Docker records until tag cleanup completes.
            for job in entry["archives"] + entry["receipts"]:
                job_url = f"{gitlab}/jobs/{job}"
                api(f"{job_url}/artifacts", token, method="DELETE")
                if has_archive(api(job_url, token)):
                    raise ValueError(f"GitLab did not delete the archive for job {job}")
        except IO_ERRORS as error:
            failures.append(f"pipeline {entry['pipeline']}: {error}")
    if failures:
        raise ValueError("Cleanup incomplete for " + "\n".join(failures))


def main():
    parser = argparse.ArgumentParser(
        description=__doc__,
        epilog=(
            "Use plan to preview without changing remote data. Both modes read GitLab with "
            "COLLECTOR_CLEANUP_GITLAB_TOKEN and GitHub with GITHUB_TOKEN. Apply requires a GitLab "
            "Maintainer API token. If tags are eligible, configure COLLECTOR_CLEANUP_QUAY_TOKEN "
            "with Quay OAuth write access and COLLECTOR_CLEANUP_DOCKER_HUB_USERNAME/TOKEN "
            "with Docker Hub tag deletion permissions. If a receipt is missing or unreadable, "
            "historical GitLab evidence must identify the release version and publishing repositories. "
            "Keep traces and Docker records until cleanup has completed."
        ),
    )
    parser.add_argument("mode", choices=("plan", "apply"))
    args = parser.parse_args()
    try:
        project = positive_id(os.environ["CI_PROJECT_ID"])
        gitlab = f"{os.environ['CI_API_V4_URL']}/projects/{project}"
        token = os.environ["COLLECTOR_CLEANUP_GITLAB_TOKEN"]
        attempts = list(archived_attempts(gitlab, token, project))
        if not attempts:
            print("No archived stable release attempts.")
            return
        pinned_pipelines, pinned_jobs = pinned_release_inputs(project, os.environ["GITHUB_TOKEN"])
        plan, incomplete = cleanup_plan(attempts, pinned_pipelines, pinned_jobs, gitlab, token, project)
        for entry in plan:
            print(json.dumps(entry), flush=True)
        if args.mode == "apply":
            apply_cleanup(plan, gitlab, token)
        if incomplete:
            parser.exit(1, "Staging tag discovery was incomplete. See warnings above.\n")
    except (*IO_ERRORS, subprocess.SubprocessError) as error:
        parser.exit(1, f"Cleanup stopped: {error}\n")


if __name__ == "__main__":
    main()
