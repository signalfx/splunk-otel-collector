FROM alpine:3.12 as certs
RUN apk --update add ca-certificates

FROM alpine:3.12 AS otelcol
COPY otelcol /
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see #1317)
RUN chmod 755 /otelcol

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol /
COPY config/collector/gateway_config /etc/otel/collector/gateway_config
COPY config/collector/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 55680 4317 6060 7276 8888 9411 9443
