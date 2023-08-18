NODE_EXPORTER_VERSION=1.6.1
ARCH=amd64
OS=linux
OTEL_VERSION=0.82.0

run-sample-exec:
	wget --no-clobber https://github.com/signalfx/splunk-otel-collector/releases/download/v$(OTEL_VERSION)/otelcol_$(OS)_$(ARCH)
	wget --no-clobber https://github.com/prometheus/node_exporter/releases/download/v$(NODE_EXPORTER_VERSION)/node_exporter-$(NODE_EXPORTER_VERSION).$(OS)-$(ARCH).tar.gz
	tar -xzvf node_exporter-$(NODE_EXPORTER_VERSION).$(OS)-$(ARCH).tar.gz
	mv node_exporter-$(NODE_EXPORTER_VERSION).$(OS)-$(ARCH)/node_exporter ./
	chmod +x ./otelcol_$(OS)_$(ARCH)
	./otelcol_$(OS)_$(ARCH) --config otel-collector-config-with-exec.yaml
run-sample-noexec:
	docker-compose up --build

.PHONY: run-sample-exec run-sample-noexec
