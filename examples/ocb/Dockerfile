FROM registry.access.redhat.com/ubi8/ubi

# copy the OpenTelemetry Collector binary to a runnable location

RUN mkdir -p /opt/bin /etc/otel /opt/share/otel

COPY otelcol /opt/bin
# expose build configuration directly inside container for debugging
COPY build.yaml /opt/share/otel

RUN chmod a+x /opt/bin/otelcol

# by default, expose OTLP/gRPC, OTLP/HTTP
EXPOSE 4317 4318

ENTRYPOINT ["/opt/bin/otelcol"]
CMD ["--config", "/etc/otel/config.yaml"]
