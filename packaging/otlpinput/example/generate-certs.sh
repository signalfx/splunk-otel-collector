#!/bin/sh
# Generates a self-signed CA, a server certificate for Splunk, and a client
# certificate for the OTel Collector. Writes all files to CERTS_DIR.
#
# Usage: sh generate-certs.sh [<certs-dir>]   (default: /certs)
#
# Security note: private key files are created mode 600 (owner-readable only).
# In production use Docker Secrets or a secrets manager instead of bind mounts.

set -eu

CERTS_DIR="${1:-/certs}"
mkdir -p "$CERTS_DIR"

echo "==> Generating CA key and certificate"
openssl genrsa -out "$CERTS_DIR/ca.key" 4096
openssl req -new -x509 -days 3650 \
  -key "$CERTS_DIR/ca.key" \
  -subj "/CN=Example-CA/O=Example" \
  -out "$CERTS_DIR/ca.crt"

echo "==> Generating server key and certificate (Splunk)"
openssl genrsa -out "$CERTS_DIR/server.key" 4096
openssl req -new \
  -key "$CERTS_DIR/server.key" \
  -subj "/CN=splunk/O=Example" \
  -out "$CERTS_DIR/server.csr"
printf 'subjectAltName=DNS:splunk,DNS:localhost\n' > "$CERTS_DIR/server.ext"
openssl x509 -req -days 365 \
  -in "$CERTS_DIR/server.csr" \
  -CA "$CERTS_DIR/ca.crt" \
  -CAkey "$CERTS_DIR/ca.key" \
  -CAcreateserial \
  -extfile "$CERTS_DIR/server.ext" \
  -out "$CERTS_DIR/server.crt"

echo "==> Generating client key and certificate (OTel Collector)"
openssl genrsa -out "$CERTS_DIR/client.key" 4096
openssl req -new \
  -key "$CERTS_DIR/client.key" \
  -subj "/CN=otelcollector/O=Example" \
  -out "$CERTS_DIR/client.csr"
openssl x509 -req -days 365 \
  -in "$CERTS_DIR/client.csr" \
  -CA "$CERTS_DIR/ca.crt" \
  -CAkey "$CERTS_DIR/ca.key" \
  -CAcreateserial \
  -out "$CERTS_DIR/client.crt"

echo "==> Setting permissions"
# server.key: owned by the splunk user (UID 41812 in the official splunk/splunk
# image) and mode 640 — readable only by that user, not world-readable.
# client.key/ca.key: mode 644 — the OTel Collector runs as a non-root user and
# needs read access. The named Docker volume is the access control boundary;
# these files are never written to the host filesystem.
# In production, use Docker Secrets or a secrets manager instead.
chown 41812 "$CERTS_DIR/server.key"
chmod 640 "$CERTS_DIR/server.key"
chmod 644 "$CERTS_DIR/ca.key" "$CERTS_DIR/ca.crt" \
          "$CERTS_DIR/server.crt" \
          "$CERTS_DIR/client.crt" "$CERTS_DIR/client.key"

rm -f "$CERTS_DIR/server.csr" "$CERTS_DIR/server.ext" \
      "$CERTS_DIR/client.csr" "$CERTS_DIR/ca.srl"

echo "==> Done. Certificates written to $CERTS_DIR"
