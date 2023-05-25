ARG SPLUNK_OTEL_COLLECTOR_IMAGE
ARG IMAGE_PLATFORM
FROM --platform=${IMAGE_PLATFORM} golang:1.19 as golang

RUN mkdir -p /some/path

FROM ${SPLUNK_OTEL_COLLECTOR_IMAGE}

ENV PATH=/some/path/go/bin:$PATH
ENV GOROOT=/some/path/go
ENV GOCACHE=/tmp/.cache/go-build

COPY telegraf-exec.go /opt/telegraf-exec.go
COPY config.yaml /etc/config.yaml
COPY --from=golang --chown=999 /tmp /tmp
COPY --from=golang --chown=999 /some/path /some/path
COPY --from=golang --chown=999 /usr/local/go /some/path/go