# OTel Collector vs Universal Forwarder Tests

This acceptance-test harness is maintained inside the
`splunk-otel-collector` repository. Run its commands from this directory, or
use the `make otel-uf-*` wrapper targets from the collector repository root.

This directory provides an end-to-end test environment for comparing events
sent by:

- Universal Forwarder (UF), using Splunk-to-Splunk (S2S)
- Splunk OTel Collector, using HEC through `splunk_outputs`
- Splunk Enterprise, which receives both streams into separate indexes

The deterministic test TA is `Splunk_TA_otel_uf_test`. It monitors one file and
uses the sourcetype `otel_uf_test`, making it possible to send identical test
events through both pipelines.

## VM layout

```text
VM A: Universal Forwarder       -- S2S :9997 --> VM C
VM B: OTel Collector            -- HEC :8088 --> VM C
VM C: Splunk Enterprise         -- REST :8089 for pytest searches
```

The default indexes are:

```text
UF:  uf_nix
OTel: otel_nix
```

## Prerequisites

- Three reachable Linux VMs
- SSH access from the workstation to all VMs
- Go matching the collector repository's `go.mod`
- Python 3 and pip
- Splunk Enterprise credentials for VM C
- The collector checkout containing this directory

Install the Python dependencies:

```bash
pip install -r requirements.txt
```

## 1. Configure the environment

Create the local configuration file:

```bash
cd test-frameworks/otel-uf
cp config.env.example config.env
```

Update `config.env` with the VM addresses, SSH key, credentials, and ports.
The important settings are:

```bash
SSH_KEY=/path/to/key.pem
SSH_USER=splunker

UF_HOST=<VM-A-address>
OTEL_HOST=<VM-B-address>
SPLUNK_HOST=<VM-C-address>

SPLUNK_ADMIN_PASSWORD=<password>
SPLUNK_HEC_TOKEN=<HEC-token>

UF_INDEX=uf_nix
OTEL_INDEX=otel_nix
TA_SOURCE_DIR=./tas
```

For the deterministic TA, either set:

```bash
TA_LIST=Splunk_TA_otel_uf_test
```

or keep the existing TA list and deploy the deterministic TA separately with
`make add-ta`.

## 2. Add a TA

Place an extracted TA directory or a TA archive under `tas/`.

For the deterministic smoke test, the repository already contains:

```text
tas/Splunk_TA_otel_uf_test/
└── default/
    ├── app.conf
    ├── inputs.conf
    └── props.conf
```

Its single input is:

```text
/home/splunker/otel_uf_test.log
```

The TA contains `index = __INDEX__`. During deployment it becomes:

```text
VM A: uf_nix
VM B: otel_nix
```

## 3. Build and provision the VMs

Build the OTel Collector binary first:

```bash
make build
```

Run the complete setup:

```bash
make setup
```

`make setup` runs these steps in order:

```bash
make setup-splunk   # VM C: Splunk Enterprise, indexes, HEC, S2S
make setup-uf       # VM A: UF and configured TA(s)
make setup-otel     # VM B: OTel Collector and configured TA(s)
```

Run the individual targets instead of `make setup` when provisioning or
repairing one VM at a time.

If the deterministic TA was not included in `TA_LIST`, deploy it to both
agents after setup:

```bash
make add-ta TA=Splunk_TA_otel_uf_test
```

This copies the TA to UF and OTel, substitutes the correct index independently,
and restarts both services.

## 4. Add identical test events

The event must be appended to the input file on both VM A and VM B. The easiest
way is the repository target:

```bash
make append-test-event \
  EVENT='otel_uf_test_id=case-001 user=alice src=10.0.0.1 dest=10.0.0.2 action=allowed app=ssh'
```

This uses `sudo tee` remotely and appends one identical line to:

```text
/home/splunker/otel_uf_test.log
```

If adding the event manually, run the command on each agent VM. A local command
such as `sudo tee -a /home/splunker/otel_uf_test.log` writes to the local
workstation, not to a remote VM.

Example manual commands:

```bash
EVENT='<34>Sep 10 06:00:00 test-host otel_uf_test_id=case-001 user=alice src=10.0.0.1 dest=10.0.0.2 action=allowed app=ssh'

printf '%s\n' "$EVENT" | \
  ssh -i "$SSH_KEY" "$SSH_USER@$UF_HOST" \
  'sudo tee -a /home/splunker/otel_uf_test.log >/dev/null'

printf '%s\n' "$EVENT" | \
  ssh -i "$SSH_KEY" "$SSH_USER@$OTEL_HOST" \
  'sudo tee -a /home/splunker/otel_uf_test.log >/dev/null'
```

Wait a few seconds for both agents to forward the event to VM C.

## 5. Verify events in Splunk

On VM C, verify the sourcetypes and counts:

```spl
index=uf_nix earliest=-24h
| stats count by sourcetype
```

```spl
index=otel_nix earliest=-24h
| stats count by sourcetype
```

For the deterministic TA, both searches should contain:

```text
sourcetype=otel_uf_test
```

Check the exact counts:

```spl
index=uf_nix sourcetype=otel_uf_test earliest=-24h
| stats count
```

```spl
index=otel_nix sourcetype=otel_uf_test earliest=-24h
| stats count
```

## 6. Run pytest

Run only the strict deterministic count test against the Splunk REST API on VM
C:

```bash
pytest \
  'tests/tas/test_splunk_ta_otel_uf_test.py::test_exact_count_parity' \
  --splunk-host "$SPLUNK_HOST" \
  --port 8089 \
  --username "$SPLUNK_ADMIN_USERNAME" \
  --password "$SPLUNK_ADMIN_PASSWORD" \
  --uf-index "$UF_INDEX" \
  --otel-index "$OTEL_INDEX" \
  -vv
```

The test requires both indexes to contain at least one `otel_uf_test` event and
requires the counts to be exactly equal.

Run all framework tests with:

```bash
make compare
```

`make compare` passes `SPLUNK_HOST` as the search target and uses port `8089`,
the Splunk management/REST port. Port `8000` is the Splunk web UI and port
`8088` is HEC; neither should be used for the pytest SDK connection.

## Troubleshooting

### Events exist but pytest reports zero events

Check the actual sourcetype first:

```spl
index=uf_nix earliest=0
| stats count by sourcetype source
```

The test searches specifically for `sourcetype=otel_uf_test`. Also note that
the test uses `earliest=-24h`; older events will not be included.

### Pytest connects to the wrong Splunk instance

When running pytest directly, always provide `--splunk-host`. Without it, the
fixture defaults to `localhost`. Confirm the REST endpoint:

```bash
curl -sk -u "admin:$SPLUNK_ADMIN_PASSWORD" \
  "https://$SPLUNK_HOST:8089/services/server/info?output_mode=json"
```

### OTel is running but does not ingest the file

Check the deployed input and collector logs:

```bash
ssh -i "$SSH_KEY" "$SSH_USER@$OTEL_HOST" \
  "grep -A5 'otel_uf_test.log' $OTEL_SPLUNK_HOME/etc/apps/Splunk_TA_otel_uf_test/default/inputs.conf"

ssh -i "$SSH_KEY" "$SSH_USER@$OTEL_HOST" \
  'sudo journalctl -u otelcol --no-pager -n 50'
```

The OTel service user must be able to read the input file. The standard
`append-test-event` target uses `sudo tee` to create a readable file.
