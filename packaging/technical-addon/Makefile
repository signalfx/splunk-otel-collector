# Used for building
THIS_MAKEFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
include $(abspath $(THIS_MAKEFILE_DIR)/../../Makefile.Common)

# Used for packaging
OTEL_COLLECTOR_VERSION?=0.149.0
SPLUNK_OTELCOL_DOWNLOAD_BASE?=https://github.com/signalfx/splunk-otel-collector/releases/download
PLATFORM?=linux
ARCH?=amd64

SPLUNK_ARCH?=x86_64

# Used for releases
SPLUNK_OTEL_VERSION?=
TA_VERSION?=

# Tools directory
TOOLS_DIR := $(BUILD_DIR)/.tools
MODINPUT_CONFIG_GENERATOR := $(TOOLS_DIR)/modinput_config_generator

# Overrides
# Packaging addons and starting splunk takes some time...
GOTEST_OPT= -v -timeout 16m --tags=$(GO_BUILD_TAGS)

## Make targets

# Targets for: TA Build tools (currently just the stuff genreated from modular-inputs.yaml)
.PHONY: ta-build-tools
ta-build-tools:
	@echo "Building tools..."
	@mkdir -p $(TOOLS_DIR)
	cd $(ADDONS_SOURCE_DIR) && go build -o $(MODINPUT_CONFIG_GENERATOR) ./cmd/modinput_config_generator

.PHONY: test-ta-build-tools
test-ta-build-tools: ta-build-tools
	$(MODINPUT_CONFIG_GENERATOR) -source-dir=$(ADDONS_SOURCE_DIR)/cmd/modinput_config_generator/internal/testdata -schema-name=Sample_Addon --build-dir=$(BUILD_DIR) || exit 1; \
	mkdir -p $(BUILD_DIR)/Sample_Addon/$(PLATFORM)_$(SPLUNK_ARCH)/bin; \
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/Sample_Addon/$(PLATFORM)_$(SPLUNK_ARCH)/bin/Sample_Addon $(ADDONS_SOURCE_DIR)/cmd/modinput_config_generator/internal/testdata/pkg/sample_addon/runner || exit 1; \
	echo "built addon, testing..."
	go test -v ./...

# Targets for building/releasing Splunk_TA_otel, which is essentially just the collector + agent bundle + discovery mode.

.PHONY: generate-technical-addon
generate-technical-addon: env-guard-all
	BUILD_DIR="$(BUILD_DIR)" \
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/make-buildspace.sh

.PHONY: copy-local-build-to-ta
copy-local-build-to-ta: env-guard-ta
	# Should only have affect if otelcol or download make targets were run first
	# Running make download-otelcol will overwrite anything locally created
	mkdir -p $(BUILD_DIR)/out/bin/
	cp -Rv $(SRC_ROOT)/bin/* $(BUILD_DIR)/out/bin/
	BUILD_DIR="$(BUILD_DIR)" \
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	SPLUNK_OTELCOL_DOWNLOAD_BASE="$(SPLUNK_OTELCOL_DOWNLOAD_BASE)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/download-agent-bundle.sh

.PHONY: env-guard-all
env-guard-all:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/env/all.sh

.PHONY: env-guard-ta
env-guard-ta: env-guard-all
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/env/ta.sh


.PHONY: download-release
download-release: env-guard-ta
	BUILD_DIR="$(BUILD_DIR)" \
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	SPLUNK_OTELCOL_DOWNLOAD_BASE="$(SPLUNK_OTELCOL_DOWNLOAD_BASE)" \
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/download-otelcol.sh
	
	BUILD_DIR="$(BUILD_DIR)" \
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	SPLUNK_OTELCOL_DOWNLOAD_BASE="$(SPLUNK_OTELCOL_DOWNLOAD_BASE)" \
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/download-agent-bundle.sh


.PHONY: package-ta
package-ta: env-guard-ta
	BUILD_DIR="$(BUILD_DIR)" \
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/package-ta.sh

.PHONY: distribute-ta
distribute-ta: generate-technical-addon download-release package-ta

.PHONY: smoketest-ta
smoketest-ta:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/smoketests/smoketest.sh

.PHONY: discovery-test-ta
discovery-test-ta:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/discovery/discovery-docker-test.sh

.PHONY: envvar-test-ta
envvar-test-ta:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/envvars/envvartest.sh

.PHONY: update-ta-deps
update-ta-deps:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/update-otel-version.sh
	
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	TA_VERSION="$(TA_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/update-ta-version.sh

.PHONY: release-ta
release-ta:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/release-ta-version.sh

.PHONY: update-and-release-ta
update-and-release: update-ta-deps release-ta

.PHONY: validate-release
validate-release:
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/validate-ta-release.sh

.PHONY: clean-ta
clean-ta:
	rm -rf $(BUILD_DIR)
