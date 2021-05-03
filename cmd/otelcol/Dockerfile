ARG SMART_AGENT_RELEASE=v5.10.2
ARG ALPINE_VERSION=3.12

FROM alpine:${ALPINE_VERSION} as certs
RUN apk --update add ca-certificates

FROM alpine:${ALPINE_VERSION} AS otelcol
COPY otelcol /
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see #1317)
RUN chmod 755 /otelcol

FROM alpine:${ALPINE_VERSION} AS smartagent
ARG SMART_AGENT_RELEASE
RUN mkdir -p "$(dirname "/usr/lib/splunk-otel-collector/agent-bundle")" && \
    SMART_AGENT=signalfx-agent-${SMART_AGENT_RELEASE#v}.tar.gz && \
    URL=https://github.com/signalfx/signalfx-agent/releases/download/${SMART_AGENT_RELEASE}/$SMART_AGENT && \
    wget "$URL" && tar -xzf $SMART_AGENT && mv signalfx-agent /usr/lib/splunk-otel-collector/agent-bundle && \
    # Absolute path of interpreter in smart agent dir is set in dependent binaries
    # requiring the interpreter location not to change.
    /usr/lib/splunk-otel-collector/agent-bundle/bin/patch-interpreter /usr/lib/splunk-otel-collector/agent-bundle && \
    rm -f /usr/lib/splunk-otel-collector/agent-bundle/bin/signalfx-agent \
          /usr/lib/splunk-otel-collector/agent-bundle/bin/agent-status \
          $SMART_AGENT

FROM scratch
ENV SPLUNK_BUNDLE_DIR=/usr/lib/splunk-otel-collector/agent-bundle
ENV SPLUNK_COLLECTD_DIR=${SPLUNK_BUNDLE_DIR}/run/collectd
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol /
# Maintaining interpreter location as required.
COPY --from=smartagent /usr/lib/splunk-otel-collector/agent-bundle /usr/lib/splunk-otel-collector/agent-bundle
COPY config/collector/gateway_config.yaml    /etc/otel/collector/gateway_config.yaml
COPY config/collector/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
COPY config/collector/agent_config.yaml      /etc/otel/collector/agent_config.yaml
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 55680 4317 6060 7276 8888 9411 9443 9080
