FROM alpine:3.14.2 as certs
RUN apk --update add ca-certificates

FROM alpine:3.14.2 AS otelcol
COPY otelcol /
COPY translatesfx /
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see #1317)
RUN chmod 755 /otelcol
RUN chmod 755 /translatesfx

FROM alpine:3.14.2 AS smartagent
ARG SMART_AGENT_RELEASE
RUN mkdir -p /usr/lib/splunk-otel-collector && \
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
ENV PATH=${PATH}:${SPLUNK_BUNDLE_DIR}/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol /
COPY --from=otelcol /translatesfx /
# Maintaining interpreter location as required.
COPY --from=smartagent /usr/lib/splunk-otel-collector/agent-bundle /usr/lib/splunk-otel-collector/agent-bundle
COPY config/collector/gateway_config.yaml    /etc/otel/collector/gateway_config.yaml
COPY config/collector/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
COPY config/collector/agent_config.yaml      /etc/otel/collector/agent_config.yaml
COPY config/collector/fargate_config.yaml    /etc/otel/collector/fargate_config.yaml
COPY config/collector/ecs_ec2_config.yaml    /etc/otel/collector/ecs_ec2_config.yaml
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 4317 6060 8888 9411 9443 9080
