ARG SPLUNK_OTEL_COLLECTOR_IMAGE

FROM golang:1.26.2 AS golang
ARG IMAGE_PLATFORM
COPY telegraf-exec.go /opt/telegraf-exec.go
RUN cd /opt && go mod init telegraf-exec && go mod tidy && \
    GOOS=linux GOARCH=$(echo ${IMAGE_PLATFORM} | cut -d/ -f2) CGO_ENABLED=0 \
    go build -o /opt/telegraf-exec /opt/telegraf-exec.go

FROM ${SPLUNK_OTEL_COLLECTOR_IMAGE}
COPY config.yaml /etc/config.yaml
COPY --from=golang /opt/telegraf-exec /opt/telegraf-exec
