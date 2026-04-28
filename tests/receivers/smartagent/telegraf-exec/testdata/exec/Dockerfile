ARG SPLUNK_OTEL_COLLECTOR_IMAGE
ARG IMAGE_PLATFORM

FROM --platform=${IMAGE_PLATFORM} golang:1.26.2 AS golang
COPY telegraf-exec.go /opt/telegraf-exec.go
RUN cd /opt && go mod init telegraf-exec && go mod tidy && go build -o /opt/telegraf-exec /opt/telegraf-exec.go

FROM ${SPLUNK_OTEL_COLLECTOR_IMAGE}
COPY config.yaml /etc/config.yaml
COPY --from=golang /opt/telegraf-exec /opt/telegraf-exec
