FROM quay.io/signalfx/splunk-otel-collector:latest
COPY --from=quay.io/prometheus/node-exporter:v1.6.1 --chown=999 /bin/node_exporter /usr/bin/node_exporter
CMD ["otelcol"]
