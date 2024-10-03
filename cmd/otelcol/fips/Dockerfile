ARG DOCKER_REPO=docker.io

FROM ${DOCKER_REPO}/alpine:3.17.0 AS certs
RUN apk --update add ca-certificates

FROM ${DOCKER_REPO}/alpine:3.17.0 AS otelcol
ARG TARGETARCH
COPY --chmod=755 dist/otelcol-fips_linux_${TARGETARCH} /otelcol
RUN echo "splunk-otel-collector:x:999:999::/:" > /etc_passwd
# create base dirs since we cannot chown in scratch image except via COPY
RUN mkdir -p /otel/collector /splunk-otel-collector

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=otelcol /etc_passwd /etc/passwd
COPY --from=otelcol --chown=999 /otelcol /
COPY --from=otelcol --chown=999 /otel /etc/otel
COPY --from=otelcol --chown=999 /otel/collector /etc/otel/collector

COPY --chown=999 config/gateway_config.yaml /etc/otel/collector/gateway_config.yaml
COPY --chown=999 config/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
COPY --chown=999 config/agent_config.yaml /etc/otel/collector/agent_config.yaml
COPY --chown=999 config/fargate_config.yaml /etc/otel/collector/fargate_config.yaml
COPY --chown=999 config/ecs_ec2_config.yaml /etc/otel/collector/ecs_ec2_config.yaml

USER splunk-otel-collector
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 4317 4318 6060 8006 8888 9411 9443 9080
