NODE_EXPORTER_VERSION=1.6.1
ARCH=amd64
OS=linux
OTEL_VERSION=0.82.0

run-sample-exec:
	#cd ../../ && $$(make otelcol)
	docker build -t prometheus-exec-sample . && \
	docker run -v ./otel-collector-config-with-exec.yaml:/etc/otel-collector-config.yaml prometheus-exec-example  --config /etc/otel-collector-config.yaml --set=service.telemetry.logs.level=debug
run-sample-noexec:
	docker-compose up --build

.PHONY: run-sample-exec run-sample-noexec
