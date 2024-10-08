# Used for building
SOURCE_DIR?=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR?=./build
OTEL_COLLECTOR_VERSION?=0.110.0
SPLUNK_OTELCOL_DOWNLOAD_BASE?=https://github.com/signalfx/splunk-otel-collector/releases/download
PLATFORM?=linux
ARCH?=amd64

# Used for testing & validation
ORCA_CLOUD?=kubernetes
UF_VERSION?=8.2.7
SPLUNK_PLATFORM?=x64_centos_7
SPLUNK_ARCH?=x86_64

# Used for releases
SPLUNK_OTEL_VERSION?=
TA_VERSION?=


.PHONY: generate-technical-addon
generate-technical-addon: env-guard-all
	BUILD_DIR="$(BUILD_DIR)" \
	SOURCE_DIR="$(SOURCE_DIR)" \
	$(SOURCE_DIR)/packaging-scripts/make-buildspace.sh

.PHONY: env-guard-all
env-guard-all:
	SOURCE_DIR="$(SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	$(SOURCE_DIR)/packaging-scripts/env/all.sh

.PHONY: env-guard-ta
env-guard-ta: env-guard-all
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(SOURCE_DIR)/packaging-scripts/env/ta.sh


.PHONY: env-guard-verify
env-guard-verify: env-guard-ta
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	$(SOURCE_DIR)/packaging-scripts/env/verify.sh

.PHONY: download
download: env-guard-ta
	BUILD_DIR="$(BUILD_DIR)" \
	SOURCE_DIR="$(SOURCE_DIR)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	SPLUNK_OTELCOL_DOWNLOAD_BASE="$(SPLUNK_OTELCOL_DOWNLOAD_BASE)" \
	PLATFORM="$(PLATFORM)" \
	ARCH="$(ARCH)" \
	$(SOURCE_DIR)/packaging-scripts/download-release.sh


.PHONY: package
package: env-guard-ta
	BUILD_DIR="$(BUILD_DIR)" \
	SOURCE_DIR="$(SOURCE_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	$(SOURCE_DIR)/packaging-scripts/package-ta.sh

.PHONY: distribute-ta
distribute-ta: generate-technical-addon download package

.PHONY: verify
verify: env-guard-verify
	SOURCE_DIR="$(SOURCE_DIR)" \
	BUILD_DIR="$(BUILD_DIR)" \
	ARCH="$(ARCH)" \
	PLATFORM="$(PLATFORM)" \
	OTEL_COLLECTOR_VERSION="$(OTEL_COLLECTOR_VERSION)" \
	ORCA_CLOUD="$(ORCA_CLOUD)" \
	UF_VERSION="$(UF_VERSION)" \
	SPLUNK_PLATFORM="$(SPLUNK_PLATFORM)" \
	OLLY_ACCESS_TOKEN="$(OLLY_ACCESS_TOKEN)" \
	$(SOURCE_DIR)/packaging-scripts/deploy-and-verify.sh

.PHONY: update-deps
update-deps:
	SOURCE_DIR="$(SOURCE_DIR)" \
	SPLUNK_OTEL_VERSION="$(SPLUNK_OTEL_VERSION)" \
	$(SOURCE_DIR)/packaging-scripts/update-otel-version.sh
	
	SOURCE_DIR="$(SOURCE_DIR)" \
	TA_VERSION="$(TA_VERSION)" \
	$(SOURCE_DIR)/packaging-scripts/update-ta-version.sh

.PHONY: release
release:
	SOURCE_DIR="$(SOURCE_DIR)" \
	$(SOURCE_DIR)/packaging-scripts/release-ta-version.sh

.PHONY: update-and-release
update-and-release: update-deps release

.PHONY: clean
clean:
	rm -rf build
