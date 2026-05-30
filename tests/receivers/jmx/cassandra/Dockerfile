FROM otelcol:latest AS collector_image

FROM eclipse-temurin:11-jre

COPY --from=collector_image /etc/otel /etc/otel
COPY --from=collector_image /opt /opt
COPY --from=collector_image /otelcol /otelcol
