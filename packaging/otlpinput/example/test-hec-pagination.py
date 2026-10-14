#!/usr/bin/env python3
# Copyright Splunk Inc.
# SPDX-License-Identifier: Apache-2.0

"""Seed and exercise more than two default HEC token pages (OTL-4676).

Called by test-ip-sans.sh with TEST_HEC_PAGINATION=1. All tokens are configured
before Splunk starts because the TA reads tokens only once, at startup.
"""

import json
import subprocess
import sys
import uuid
from pathlib import Path


TOKEN_COUNT = 61
COMPOSE = [
    "docker", "compose",
    "-f", "docker-compose.yml",
    "-f", "docker-compose.test.yml",
    "--profile", "splunk",
]
TOKEN_NAMES = {f"http://pagination-{i:03d}" for i in range(TOKEN_COUNT)}


def seed(inputs_conf):
    with Path(inputs_conf).open("a") as inputs:
        inputs.write("\n[http]\ndisabled = 0\n")
        for i in range(TOKEN_COUNT):
            inputs.write(
                f"\n[http://pagination-{i:03d}]\n"
                f"token = {uuid.UUID(int=i + 1)}\n"
                "index = main\nindexes = main\ndisabled = 0\n"
            )


def read_tokens(query):
    response = subprocess.check_output(
        COMPOSE + [
            "exec", "-T", "splunk", "sh", "-c",
            'curl -fsSk --max-time 30 -u "admin:$SPLUNK_PASSWORD" '
            '"https://localhost:8089/servicesNS/-/-/data/inputs/http?'
            + query + '"',
        ],
        text=True,
    )
    return json.loads(response)["entry"]


def status_for(token, ca_cert):
    return subprocess.check_output(
        [
            "curl", "-sS", "--max-time", "10", "--cacert", ca_cert,
            "-o", "/dev/null", "-w", "%{http_code}",
            "-H", f"Authorization: Splunk {token}",
            "-H", "Content-Type: application/json", "-d", "{}",
            "https://localhost:4318/v1/logs",
        ],
        text=True,
    )


def check(ca_cert):
    # Use the same default listing as internal/auth.New, plus an unlimited
    # listing to prove that valid tokens really were omitted from that page.
    first_page = read_tokens("output_mode=json")
    all_tokens = read_tokens("output_mode=json&count=0")
    tokens = {
        e["name"]: e["content"]["token"]
        for e in all_tokens if e["name"] in TOKEN_NAMES
    }
    first_names = {e["name"] for e in first_page} & TOKEN_NAMES
    omitted_names = TOKEN_NAMES - first_names
    if set(tokens) != TOKEN_NAMES or not first_names or len(omitted_names) < 31:
        raise RuntimeError(
            f"Pagination fixture invalid: {len(tokens)} seeded tokens visible, "
            f"{len(first_names)} on the default page, {len(omitted_names)} omitted"
        )

    # A negative control ensures the receiver actually enforces authentication.
    if status_for(str(uuid.UUID(int=0)), ca_cert) != "401":
        raise RuntimeError("OTLP receiver did not reject an unknown HEC token")

    failures = []
    for name, token in sorted(tokens.items()):
        status = status_for(token, ca_cert)
        if status != "200":
            page = "default page" if name in first_names else "later page"
            failures.append(f"{name} ({page}): expected HTTP 200, got {status}")
    if failures:
        raise RuntimeError("HEC token pagination regression:\n" + "\n".join(failures))
    print(
        f"OK: all {TOKEN_COUNT} HEC tokens authenticated, including "
        f"{len(omitted_names)} omitted from the default page"
    )


if __name__ == "__main__":
    if len(sys.argv) != 3 or sys.argv[1] not in ("seed", "check"):
        sys.exit(f"Usage: {sys.argv[0]} seed INPUTS_CONF | check CA_CERT")
    if sys.argv[1] == "seed":
        seed(sys.argv[2])
    else:
        check(sys.argv[2])
