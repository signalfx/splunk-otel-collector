OTEL_COLLECTOR_VERSION ?= 0.61.0
export OTEL_COLLECTOR_VERSION
SPLUNK_COLLECTOR_VERSION ?= $(OTEL_COLLECTOR_VERSION)
export SPLUNK_COLLECTOR_VERSION
PLUGINS_VERSION ?= 0.61.2
export PLUGINS_VERSION

SOURCE_DIR ?= $(shell pwd)
ARTIFACTS_DIR ?= $(SOURCE_DIR)/artifacts
TEST_RUNTIME_CONFIG ?= runtime_config.yaml
TMPDIR ?= /tmp
GOOS ?= darwin
GOARCH ?= amd64
RELEASE ?= dev

GITHUB_RELEASE_PREFIX ?= https://github.com
BUILD_CONFIG = builder_config.yaml
OCB_EXECUTABLE = ocb_$(OTEL_COLLECTOR_VERSION)_$(GOOS)_$(GOARCH)

all: fetch_ocb build test

fetch_ocb: $(ARTIFACTS_DIR)/$(OCB_EXECUTABLE)

$(ARTIFACTS_DIR)/otelcol build: | $(ARTIFACTS_DIR)
	envsubst < ${BUILD_CONFIG}.in >${BUILD_CONFIG}
	$(ARTIFACTS_DIR)/$(OCB_EXECUTABLE) --config=$(BUILD_CONFIG) --output-path=$(ARTIFACTS_DIR)

$(ARTIFACTS_DIR)/$(OCB_EXECUTABLE):
	( cd $(TMPDIR) && curl -OL "${GITHUB_RELEASE_PREFIX}/open-telemetry/opentelemetry-collector/releases/download/v${OTEL_COLLECTOR_VERSION}/${OCB_EXECUTABLE}" )
	( cd $(TMPDIR) && sha256sum --ignore-missing -c $(SOURCE_DIR)/checksums.txt )
	mv $(TMPDIR)/$(OCB_EXECUTABLE) $(ARTIFACTS_DIR)/$(OCB_EXECUTABLE) && chmod a+x $(ARTIFACTS_DIR)/$(OCB_EXECUTABLE)

$(ARTIFACTS_DIR):
	mkdir $@

test: $(ARTIFACTS_DIR)/otelcol
	$(ARTIFACTS_DIR)/otelcol --config=$(TEST_RUNTIME_CONFIG) & \
	PID=$$! ; \
	sleep 5 && kill $${PID}

.PHONY: fetch_ocb
