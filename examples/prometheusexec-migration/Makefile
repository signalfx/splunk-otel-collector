run-sample-exec:
	docker build -t prometheus-exec-example . && \
	docker run -v ./otel-collector-config-with-exec.yaml:/etc/otel-collector-config.yaml prometheus-exec-example  --config /etc/otel-collector-config.yaml --set=service.telemetry.logs.level=debug
run-sample-noexec:
	docker-compose up --build

.PHONY: run-sample-exec run-sample-noexec
