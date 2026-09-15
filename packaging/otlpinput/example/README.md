# Splunk HEC Example

This example showcases how the collector can collect data from files and send it to Splunk Enterprise over **TLS**, following the [OpenTelemetry security configuration best practices](https://opentelemetry.io/docs/security/config-best-practices/#store-your-configuration-securely).

The example runs as a Docker Compose deployment with three services:

| Service | Role |
| --- | --- |
| `cert-init` | One-shot container that generates TLS certificates before any other service starts |
| `otelcollector` | Reads log files and forwards them to Splunk over HTTPS |
| `splunk` | Splunk Enterprise with the Splunk Connect for OTLP TA, configured for TLS |

## Certificate strategy

`cert-init` runs `generate-certs.sh`, which creates:

- A self-signed CA (`ca.crt` / `ca.key`)
- A server certificate for Splunk (`server.crt` / `server.key`)
- A client certificate for the OTel Collector (`client.crt` / `client.key`)

All files are written to a **named Docker volume** (`certs`). Private keys are created with mode `600` (owner-readable only) and never written to the host filesystem, limiting unintended exposure. In production, replace this with Docker Secrets or a secrets manager.

> **Note:** The current TA release supports one-way TLS (server authentication only). Currently, the TA does not verify the collector's client certificate.

## Running the example

```bash
docker-compose --profile splunk up
```

Splunk will become available on port 18000. Log in at [http://localhost:18000](http://localhost:18000) with `admin` / `VH#ou812@5150`.

The `SPLUNK_APPS_URL` environment variable in `docker-compose.yml` installs the latest Splunk Connect for OTLP release automatically at startup. To use a locally built package instead, replace it with a bind-mount and a `file:///` URL (see the root `Makefile`'s `splunk` target for an example).

The `splunk-inputs-local.conf` file is mounted into the Splunk container as the app's `local/inputs.conf`, overriding the defaults to enable SSL and point to the certificate files.

Once Splunk is running, visit the [Search app](http://localhost:18000/en-US/app/search) to see the logs collected by Splunk.

## TLS configuration

### OTel Collector (`otel-collector-config.yml`)

The `otlphttp` exporter connects to `https://splunk:4318` and verifies the server certificate against the shared CA:

```yaml
tls:
  ca_file: /certs/ca.crt  # verifies the Splunk server certificate
```

### Splunk / TA (`splunk-inputs-local.conf`)

The local inputs override enables TLS on the TA listener:

```ini
[splunk-connect-for-otlp]
enableSSL  = 1
serverCert = /certs/server.crt
serverKey  = /certs/server.key
```
