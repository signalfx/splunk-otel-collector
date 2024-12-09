include ./Makefile.Common
include ./packaging/technical-addon/Makefile

### VARIABLES

# BUILD_TYPE should be one of (dev, release).
BUILD_TYPE?=release
VERSION?=latest

GIT_SHA=$(shell git rev-parse --short HEAD)
GO_ACC=go-acc
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

FIND_MOD_ARGS=-type f -name "go.mod"
TO_MOD_DIR=dirname {} \; | sort | egrep  '^./'

ALL_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR)) $(PWD)

GOTEST=go test -p $(NUM_CORES)

# Currently integration tests are flakey when run in parallel due to internal metric and config server conflicts
GOTEST_SERIAL=go test -p 1

BUILD_INFO_IMPORT_PATH=github.com/signalfx/splunk-otel-collector/internal/version
BUILD_INFO_IMPORT_PATH_TESTS=github.com/signalfx/splunk-otel-collector/tests/internal/version
BUILD_INFO_IMPORT_PATH_CORE=go.opentelemetry.io/collector/internal/version
VERSION=$(shell git describe --match "v[0-9]*" HEAD)
BUILD_X1=-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)
BUILD_X2=-X $(BUILD_INFO_IMPORT_PATH_CORE).Version=$(VERSION)
BUILD_INFO=-ldflags "${BUILD_X1} ${BUILD_X2}"
BUILD_INFO_TESTS=-ldflags "-X $(BUILD_INFO_IMPORT_PATH_TESTS).Version=$(VERSION)"
CGO_ENABLED?=0

JMX_METRIC_GATHERER_RELEASE=$(shell cat packaging/jmx-metric-gatherer-release.txt)
SKIP_COMPILE=false
ARCH?=amd64
BUNDLE_SUPPORTED_ARCHS := amd64 arm64
SKIP_BUNDLE=false

# For integration testing against local changes you can run
# SPLUNK_OTEL_COLLECTOR_IMAGE='otelcol:latest' make -e docker-otelcol integration-test
# for local docker build testing or
# SPLUNK_OTEL_COLLECTOR_IMAGE='' make -e otelcol integration-test
# for local binary testing (agent-bundle configuration required)
export SPLUNK_OTEL_COLLECTOR_IMAGE?=otelcol:latest

# Docker repository used.
DOCKER_REPO?=docker.io

GOTESPLIT_TOTAL?=1
GOTESPLIT_INDEX?=0

### TARGETS

.DEFAULT_GOAL := all

all-modules:
	@echo $(ALL_MODS) | tr ' ' '\n' | sort

.PHONY: all
all: checklicense impi lint misspell test otelcol

.PHONY: for-all
for-all:
	@set -e; for dir in $(ALL_MODS); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

# Define a delegation target for each module
.PHONY: $(ALL_MODS)
$(ALL_MODS):
	@echo "Running target '$(TARGET)' in module '$@'"
	$(MAKE) --no-print-directory -C $@ $(TARGET)

# Triggers each module's delegation target
.PHONY: for-all-target
for-all-target: $(ALL_MODS)

.PHONY: integration-vet
integration-vet:
	@set -e; cd tests && go vet -tags integration,testutilsintegration,zeroconfig,testutils ./... && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) -tags testutils,testutilsintegration -v -timeout 5m -count 1 ./...

.PHONY: integration-test
integration-test:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=integration -v -timeout 5m -count 1 ./...

.PHONY: integration-test-mongodb-discovery
integration-test-mongodb-discovery:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=discovery_integration_mongodb -v -timeout 5m -count 1 ./...

.PHONY: integration-test-kafkametrics-discovery
integration-test-kafkametrics-discovery:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=discovery_integration_kafkametrics -v -timeout 5m -count 1 ./...

.PHONY: integration-test-jmx/cassandra-discovery
integration-test-jmx/cassandra-discovery:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=discovery_integration_jmx -v -timeout 5m -count 1 ./...

.PHONY: integration-test-apache-discovery
integration-test-apache-discovery:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=discovery_integration_apachewebserver -v -timeout 5m -count 1 ./...

.PHONY: integration-test-nginx-discovery
integration-test-nginx-discovery:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=discovery_integration_nginx -v -timeout 5m -count 1 ./...

.PHONY: smartagent-integration-test
smartagent-integration-test:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=smartagent_integration -v -timeout 5m -count 1 ./...

.PHONY: test-with-cover
test-with-cover:
	@echo Verifying that all packages have test files to count in coverage
	@echo pre-compiling tests
	@time go test -p $(NUM_CORES) ./...
	$(GO_ACC) ./...
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: gendependabot
gendependabot:
	.github/workflows/scripts/gendependabot.sh

.PHONY: tidy-all
tidy-all:
	$(MAKE) for-all-target TARGET="tidy"
	$(MAKE) tidy

.PHONY: install-tools
install-tools:
	cd ./internal/tools && go install github.com/client9/misspell/cmd/misspell
	cd ./internal/tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd ./internal/tools && go install github.com/google/addlicense
	cd ./internal/tools && go install github.com/jstemmer/go-junit-report
	cd ./internal/tools && go install go.opentelemetry.io/collector/cmd/mdatagen
	cd ./internal/tools && go install github.com/ory/go-acc
	cd ./internal/tools && go install github.com/pavius/impi/cmd/impi
	cd ./internal/tools && go install github.com/tcnksm/ghr
	cd ./internal/tools && go install golang.org/x/tools/cmd/goimports
	cd ./internal/tools && go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
	cd ./internal/tools && go install golang.org/x/vuln/cmd/govulncheck@master


.PHONY: generate-metrics
generate-metrics:
	go generate -tags mdatagen ./...
	$(MAKE) fmt

.PHONY: otelcol
otelcol:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -o ./bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/otelcol
ifeq ($(OS), Windows_NT)
	$(LINK_CMD) .\bin\otelcol$(EXTENSION) .\bin\otelcol_$(GOOS)_$(GOARCH)$(EXTENSION)
else
	$(LINK_CMD) otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/otelcol$(EXTENSION)
endif


.PHONY: migratecheckpoint
migratecheckpoint:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -o ./bin/migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/migratecheckpoint
ifeq ($(OS), Windows_NT)
	$(LINK_CMD) .\bin\migratecheckpoint$(EXTENSION) .\bin\migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION)
else
	$(LINK_CMD) migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/migratecheckpoint$(EXTENSION)
endif

.PHONY: bundle.d
bundle.d:
	go install github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/bundle/cmd/discoverybundler
	go generate -tags bootstrap.bundle.d ./...
	go generate -tags bundle.d ./...

.PHONY: add-tag
add-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -s -m "Version ${TAG}"

.PHONY: delete-tag
delete-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting tag ${TAG}"
	@git tag -d ${TAG}

.PHONY: docker-otelcol
docker-otelcol:
	ARCH=$(ARCH) FIPS=$(FIPS) SKIP_COMPILE=$(SKIP_COMPILE) SKIP_BUNDLE=$(SKIP_BUNDLE) DOCKER_REPO=$(DOCKER_REPO) JMX_METRIC_GATHERER_RELEASE=$(JMX_METRIC_GATHERER_RELEASE) ./packaging/docker-otelcol.sh

.PHONY: binaries-all-sys
binaries-all-sys: binaries-darwin_amd64 binaries-darwin_arm64 binaries-linux_amd64 binaries-linux_arm64 binaries-windows_amd64 binaries-linux_ppc64le

.PHONY: binaries-darwin_amd64
binaries-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) otelcol
	GOOS=darwin  GOARCH=amd64 $(MAKE) migratecheckpoint

.PHONY: binaries-darwin_arm64
binaries-darwin_arm64:
	GOOS=darwin  GOARCH=arm64 $(MAKE) otelcol
	GOOS=darwin  GOARCH=arm64 $(MAKE) migratecheckpoint

.PHONY: binaries-linux_amd64
binaries-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) otelcol
	GOOS=linux   GOARCH=amd64 $(MAKE) migratecheckpoint

.PHONY: binaries-linux_arm64
binaries-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) otelcol
	GOOS=linux   GOARCH=arm64 $(MAKE) migratecheckpoint

.PHONY: binaries-windows_amd64
binaries-windows_amd64:
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) otelcol
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) migratecheckpoint

.PHONY: binaries-linux_ppc64le
binaries-linux_ppc64le:
	GOOS=linux GOARCH=ppc64le $(MAKE) otelcol
	GOOS=linux GOARCH=ppc64le $(MAKE) migratecheckpoint

.PHONY: deb-rpm-tar-package
%-package:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-linux_$(ARCH)
endif
ifneq ($(filter $(ARCH), $(BUNDLE_SUPPORTED_ARCHS)),)
ifneq ($(SKIP_BUNDLE), true)
	$(MAKE) -C packaging/bundle agent-bundle-linux ARCH=$(ARCH) DOCKER_REPO=$(DOCKER_REPO)
endif
endif
	docker build -t otelcol-fpm packaging/fpm
	docker run --rm -v $(CURDIR):/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) -e JMX_METRIC_GATHERER_RELEASE=$(JMX_METRIC_GATHERER_RELEASE) otelcol-fpm

.PHONY: msi
msi:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-windows_amd64
endif

	!(find ./internal/buildscripts/packaging/msi -name "*.wxs" | xargs grep -q "RemoveFolderEx") || (echo "Custom action 'RemoveFolderEx' can't be used without corresponding WiX upgrade due to CVE-2024-29188." && exit 1)
	test -f ./dist/agent-bundle_windows_amd64.zip || (echo "./dist/agent-bundle_windows_amd64.zip not found! Either download a pre-built bundle to ./dist/, or run './packaging/bundle/scripts/windows/make.ps1 bundle' on a windows host and copy it to ./dist/." && exit 1)
	./packaging/msi/build.sh "$(VERSION)" "$(DOCKER_REPO)" "$(JMX_METRIC_GATHERER_RELEASE)"

.PHONY: update-examples
update-examples:
	cd examples && $(MAKE) update-examples

.PHONY: install-test-tools
install-test-tools:
	cd ./tests/tools && go install github.com/Songmu/gotesplit/cmd/gotesplit

.PHONY: integration-test-split
integration-test-split: install-test-tools
	@set -e; cd tests && gotesplit --total=$(GOTESPLIT_TOTAL) --index=$(GOTESPLIT_INDEX) ./... -- -p 1 $(BUILD_INFO_TESTS) --tags=integration -v -timeout 5m -count 1

.PHONY: otelcol-fips
otelcol-fips:
ifeq ($(GOOS), linux)
    ifeq ($(filter $(GOARCH), amd64 arm64),)
		$(error GOOS=$(GOOS) GOARCH=$(GOARCH) not supported)
    endif
	$(eval BUILD_INFO = -ldflags "${BUILD_X1} ${BUILD_X2} -linkmode=external -extldflags=-static")
else ifeq ($(GOOS), windows)
    ifeq ($(filter $(GOARCH), amd64),)
		$(error GOOS=$(GOOS) GOARCH=$(GOARCH) not supported)
    endif
	$(eval EXTENSION = .exe)
else
	$(error GOOS=$(GOOS) GOARCH=$(GOARCH) not supported)
endif
	docker buildx build --pull \
		--tag otelcol-fips-builder-$(GOOS)-$(GOARCH) \
		--platform linux/$(GOARCH) \
		--build-arg DOCKER_REPO=$(DOCKER_REPO) \
		--build-arg BUILD_INFO='$(BUILD_INFO)' \
		--file cmd/otelcol/fips/build/Dockerfile.$(GOOS) ./
	@docker rm -f otelcol-fips-builder-$(GOOS)-$(GOARCH) >/dev/null 2>&1 || true
	@mkdir -p ./bin
	docker create --platform linux/$(GOARCH) --name otelcol-fips-builder-$(GOOS)-$(GOARCH) otelcol-fips-builder-$(GOOS)-$(GOARCH) true >/dev/null
	docker cp otelcol-fips-builder-$(GOOS)-$(GOARCH):/src/bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/otelcol-fips_$(GOOS)_$(GOARCH)$(EXTENSION)
	@docker rm -f otelcol-fips-builder-$(GOOS)-$(GOARCH) >/dev/null


.PHONY: package-technical-addon
package-technical-addon: bundle.d otelcol generate-technical-addon copy-local-build-to-ta package-ta smoketest-ta
