FROM alpine:3.16.0 as certs
RUN apk --update add ca-certificates

FROM alpine:3.16.0 AS otelcol
COPY otelcol /
COPY translatesfx /
COPY migratecheckpoint /
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see https://github.com/open-telemetry/opentelemetry-collector/issues/1317)
RUN chmod 755 /otelcol
RUN chmod 755 /translatesfx
RUN chmod 755 /migratecheckpoint
RUN echo "splunk-otel-collector:x:999:999::/:" > /etc_passwd
# create base dirs since we cannot chown in scratch image except via COPY
RUN mkdir -p /otel/collector /splunk-otel-collector

FROM alpine:3.16.0 AS smartagent
ARG SMART_AGENT_RELEASE
ARG ARCH="amd64"

COPY --from=otelcol /etc_passwd /etc_passwd
RUN cat /etc_passwd >> /etc/passwd
COPY --from=otelcol --chown=999 /splunk-otel-collector /usr/lib/splunk-otel-collector
USER splunk-otel-collector
RUN if [ "$ARCH" = "amd64" ]; then \
        SMART_AGENT=signalfx-agent-${SMART_AGENT_RELEASE#v}.tar.gz && \
        URL=https://github.com/signalfx/signalfx-agent/releases/download/${SMART_AGENT_RELEASE}/$SMART_AGENT && \
        cd /usr/lib/splunk-otel-collector && wget "$URL" && tar -xzf $SMART_AGENT && mv signalfx-agent agent-bundle && \
        # Absolute path of interpreter in smart agent dir is set in dependent binaries
        # requiring the interpreter location not to change.
        /usr/lib/splunk-otel-collector/agent-bundle/bin/patch-interpreter /usr/lib/splunk-otel-collector/agent-bundle && \
        rm -f /usr/lib/splunk-otel-collector/agent-bundle/bin/signalfx-agent \
            /usr/lib/splunk-otel-collector/agent-bundle/bin/agent-status \
            $SMART_AGENT; \
    else \
        mkdir -p /usr/lib/splunk-otel-collector/agent-bundle; \
    fi
RUN find /usr/lib/splunk-otel-collector/agent-bundle -wholename "*test*.key" -delete -or -wholename "*test*.pem" -delete

FROM debian:11.3 as journalctl
RUN apt update
RUN apt install -y systemd
COPY collect-libs.sh /collect-libs.sh
RUN /collect-libs.sh /opt/journalctl /bin/journalctl

FROM scratch
ENV SPLUNK_BUNDLE_DIR=/usr/lib/splunk-otel-collector/agent-bundle
ENV SPLUNK_COLLECTD_DIR=${SPLUNK_BUNDLE_DIR}/run/collectd
ENV PATH=${PATH}:${SPLUNK_BUNDLE_DIR}/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /etc_passwd /etc/passwd
COPY --from=otelcol --chown=999 /otelcol /
COPY --from=otelcol --chown=999 /translatesfx /
COPY --from=otelcol --chown=999 /migratecheckpoint /
COPY --from=otelcol --chown=999 /otel /etc/otel
COPY --from=otelcol --chown=999 /otel/collector /etc/otel/collector
COPY --from=smartagent --chown=999 /usr/lib/splunk-otel-collector /usr/lib/splunk-otel-collector

COPY --from=journalctl --chown=999 /bin/journalctl /bin/journalctl
COPY --from=journalctl --chown=999 /opt/journalctl /

COPY --chown=999 config/collector/gateway_config.yaml /etc/otel/collector/gateway_config.yaml
COPY --chown=999 config/collector/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
COPY --chown=999 config/collector/agent_config.yaml /etc/otel/collector/agent_config.yaml
COPY --chown=999 config/collector/fargate_config.yaml /etc/otel/collector/fargate_config.yaml
COPY --chown=999 config/collector/ecs_ec2_config.yaml /etc/otel/collector/ecs_ec2_config.yaml

USER splunk-otel-collector
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 4317 4318 6060 8888 9411 9443 9080
