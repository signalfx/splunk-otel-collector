ARG SPLUNK_OTEL_COLLECTOR_IMAGE
ARG IMAGE_PLATFORM

FROM --platform=${IMAGE_PLATFORM} golang:1.25.2 as golang
COPY telegraf-exec.go /opt/telegraf-exec.go
RUN go build -o /opt/telegraf-exec /opt/telegraf-exec.go

FROM ${SPLUNK_OTEL_COLLECTOR_IMAGE}
COPY config.yaml /etc/config.yaml
COPY --from=golang /opt/telegraf-exec /opt/telegraf-exec
