# Used for building
THIS_MAKEFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
include $(abspath $(THIS_MAKEFILE_DIR)/../../Makefile.Common)

# Used for packaging
OTEL_COLLECTOR_VERSION?=0.140.0
SPLUNK_OTELCOL_DOWNLOAD_BASE?=https://github.com/signalfx/splunk-otel-collector/releases/download
PLATFORM?=linux
ARCH?=amd64

# include autoinstrumentation
AUTOINSTRUMENTATION_DIR:=$(SRC_ROOT)/instrumentation

# Used for testing & validation
ORCA_CLOUD?=kubernetes
UF_VERSION?=8.2.7
SPLUNK_PLATFORM?=x64_centos_7
SPLUNK_ARCH?=x86_64

# Used for releases
SPLUNK_OTEL_VERSION?=
TA_VERSION?=
TA_PLATFORMS=linux windows
MODULAR_INPUT_SCHEMAS=Splunk_TA_otel_linux_autoinstrumentation

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
	go test -v `go list ./... | grep --invert-match pkg`

# Targets for building the golang style TAs
.PHONY: build-tas
build-tas: gen-modinput-config ta-build-deps-all build-ta-runners package-tas

.PHONY: gen-modinput-config
gen-modinput-config:
	@echo "Generating runner configs..."
	@for schema in $(MODULAR_INPUT_SCHEMAS); do \
		schema_lower=$$(echo $$schema | tr '[:upper:]' '[:lower:]'); \
		echo "Generating config and TA scaffold for $$schema..."; \
		rm -rf $(BUILD_DIR)/$$schema/; \
		mkdir -p $(BUILD_DIR)/$$schema/; \
		$(MODINPUT_CONFIG_GENERATOR) -source-dir=$(ADDONS_SOURCE_DIR) -schema-name=$$schema --build-dir=$(BUILD_DIR) || exit 1; \
		ls -lAh $(BUILD_DIR); \
	done

.PHONY: build-ta-runners
build-ta-runners:
	@echo "Building runner binaries..."
	@for schema in $(MODULAR_INPUT_SCHEMAS); do \
		echo "Building $$schema..."; \
		schema_lower=$$(echo $$schema | tr '[:upper:]' '[:lower:]'); \
		for platform in $(TA_PLATFORMS); do \
		  	if [ "$(SPLUNK_ARCH)" = "x86_64" ]; then \
		  		export arch="amd64"; \
		  	fi; \
			echo "Building $${platform}_$$arch for $$schema..."; \
			mkdir -p $(BUILD_DIR)/$$schema/$${platform}_$(SPLUNK_ARCH)/bin; \
			GOOS=$$platform GOARCH=$$arch go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$$schema/$${platform}_$(SPLUNK_ARCH)/bin/$$schema $(ADDONS_SOURCE_DIR)/pkg/$$schema_lower/runner || exit 1; \
		done \
	done

.PHONY: generate-technical-addon-linux-autoinstrumentation
generate-technical-addon-linux-autoinstrumentation:
	cd $(AUTOINSTRUMENTATION_DIR) && $(MAKE) dist
	mkdir -p $(BUILD_DIR)/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin;
	mkdir -p $(BUILD_DIR)
	PLATFORM="all" \
	BUILD_DIR="$(BUILD_DIR)" \
	SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	AUTOINSTRUMENTATION_DIR="$(AUTOINSTRUMENTATION_DIR)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/download-autoinstrumentation.sh


.PHONY: ta-build-deps-all
ta-build-deps-all: generate-technical-addon-linux-autoinstrumentation

.PHONY: package-tas
package-tas:
	mkdir -p $(BUILD_DIR)/out/distribution/
	@echo "Building runner binaries..."
	@for schema in $(MODULAR_INPUT_SCHEMAS); do \
		echo "packaging $$schema"; \
		APP_CONFIG="$(BUILD_DIR)/$$schema/default/app.conf" ;\
		if ! [ -f "$$APP_CONFIG" ]; then \
			echo "file $$APP_CONFIG not found" ;\
			exit 1; \
		fi ; \
		VERSION=$$(sed -n '/^\[launcher\]/,/^\[/{s/^version = \(.*\)/\1/p}' "$$APP_CONFIG" | tr -d ' \t\n\r') ; \
		COPYFILE_DISABLE=1 tar -C "$(BUILD_DIR)" -hcz -f "$(BUILD_DIR)/out/distribution/$${schema}-$$VERSION.tgz" "$${schema}" ;\
	done

# Targets for building/releasing the "original" Splunk_TA_otel, which is essentially just the collector + agent bundle + discovery mode.
# Some of these can be reused with the new golang-style TAs

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


.PHONY: env-guard-verify
env-guard-verify: env-guard-ta
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/env/verify.sh

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

.PHONY: orca-test-ta
orca-test-ta: env-guard-verify
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	SPLUNK_PLATFORM="$(SPLUNK_PLATFORM)" \
	OLLY_ACCESS_TOKEN="$(OLLY_ACCESS_TOKEN)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/happypath-test.sh

.PHONY: orca-test-ta
orca-collectd-test-ta: env-guard-verify
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	SPLUNK_PLATFORM="$(SPLUNK_PLATFORM)" \
	OLLY_ACCESS_TOKEN="$(OLLY_ACCESS_TOKEN)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/agent-bundle/orca-collectd-test.sh

.PHONY: orca-gateway-test-ta
orca-gateway-test-ta:  env-guard-verify
	ADDONS_SOURCE_DIR="$(ADDONS_SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	SPLUNK_PLATFORM="$(SPLUNK_PLATFORM)" \
	PLATFORM="$(PLATFORM)" \
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	OLLY_ACCESS_TOKEN="$(OLLY_ACCESS_TOKEN)" \
	$(ADDONS_SOURCE_DIR)/packaging-scripts/cicd-tests/gateway/gateway.sh

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
