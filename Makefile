include ./Makefile.Common

### VARIABLES

# BUILD_TYPE should be one of (dev, release).
BUILD_TYPE?=release
DEFAULT_VERSION=$(shell git describe --match "v[0-9]*" HEAD)
VERSION?=${DEFAULT_VERSION}

GIT_SHA=$(shell git rev-parse --short HEAD)
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

FIND_MOD_ARGS=-type f -name "go.mod"  -not -path "./packaging/technical-addon/*"
TO_MOD_DIR=dirname {} \; | sort | egrep  '^./'

ALL_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR)) $(PWD)

# Currently integration tests are flakey when run in parallel due to internal metric and config server conflicts
GOTEST_SERIAL=go test -p 1

BUILD_INFO_IMPORT_PATH=github.com/signalfx/splunk-otel-collector/internal/version
BUILD_INFO_IMPORT_PATH_TESTS=github.com/signalfx/splunk-otel-collector/tests/internal/version
BUILD_INFO_IMPORT_PATH_CORE=go.opentelemetry.io/collector/internal/version
BUILD_X1=-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)
BUILD_X2=-X $(BUILD_INFO_IMPORT_PATH_CORE).Version=$(VERSION)
BUILD_INFO=-ldflags "${BUILD_X1} ${BUILD_X2}"
BUILD_INFO_TESTS=-ldflags "-X $(BUILD_INFO_IMPORT_PATH_TESTS).Version=$(VERSION)"
CGO_ENABLED?=0

# This directory is used in tests hold code coverage results.
# It's mounted on docker containers which then write code coverage
# results to it, making coverage profiles available on the host after tests.
# 777 privileges are important to allow docker container write
# access to host dir.
MAKE_TEST_COVER_DIR=mkdir -m 777 -p $(TEST_COVER_DIR)

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
all: checklicense lint misspell test otelcol

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

.PHONY: integration-test-target
integration-test-target:
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=$(TARGET) -v -timeout 5m -count 1 ./...

.PHONY: integration-test-cover-target
integration-test-cover-target:
	@set -e; $(MAKE_TEST_COVER_DIR) && cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=$(TARGET) -v -timeout 10m -count 1 ./... $(COVER_TESTING_INTEGRATION_OPTS)
	$(GOCMD) tool covdata textfmt -i=$(TEST_COVER_DIR) -o ./$(TARGET)-coverage.txt

.PHONY: integration-test
integration-test:
	@make integration-test-target TARGET='integration'

.PHONY: integration-test-with-cover
integration-test-with-cover:
	@make integration-test-cover-target TARGET='integration'

.PHONY: integration-test-mongodb-discovery
integration-test-mongodb-discovery:
	@make integration-test-target TARGET='discovery_integration_mongodb'

.PHONY: integration-test-mongodb-discovery-with-cover
integration-test-mongodb-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_mongodb'

.PHONY: integration-test-mysql-discovery
integration-test-mysql-discovery:
	@make integration-test-target TARGET='discovery_integration_mysql'

.PHONY: integration-test-mysql-discovery-with-cover
integration-test-mysql-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_mysql'

.PHONY: integration-test-kafkametrics-discovery
integration-test-kafkametrics-discovery:
	@make integration-test-target TARGET='discovery_integration_kafkametrics'

.PHONY: integration-test-kafkametrics-discovery-with-cover
integration-test-kafkametrics-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_kafkametrics'

.PHONY: integration-test-jmx/cassandra-discovery
integration-test-jmx/cassandra-discovery:
	@make integration-test-target TARGET='discovery_integration_jmx'

.PHONY: integration-test-jmx/cassandra-discovery-with-cover
integration-test-jmx/cassandra-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_jmx'

.PHONY: integration-test-apache-discovery
integration-test-apache-discovery:
	@make integration-test-target TARGET='discovery_integration_apachewebserver'

.PHONY: integration-test-apache-discovery-with-cover
integration-test-apache-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_apachewebserver'

.PHONY: integration-test-envoy-discovery
integration-test-envoy-discovery:
	@make integration-test-target TARGET='discovery_integration_envoy'

.PHONY: integration-test-envoy-discovery-with-cover
integration-test-envoy-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_envoy'

.PHONY: integration-test-nginx-discovery
integration-test-nginx-discovery:
	@make integration-test-target TARGET='discovery_integration_nginx'

.PHONY: integration-test-nginx-discovery-with-cover
integration-test-nginx-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_nginx'

.PHONY: integration-test-redis-discovery
integration-test-redis-discovery:
	@make integration-test-target TARGET='discovery_integration_redis'

.PHONY: integration-test-weaviate-discovery-with-cover
integration-test-weaviate-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_weaviate'

.PHONY: integration-test-weaviate-discovery
integration-test-weaviate-discovery:
	@make integration-test-target TARGET='discovery_integration_weaviate'

.PHONY: integration-test-redis-discovery-with-cover
integration-test-redis-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_redis'

.PHONY: integration-test-oracledb-discovery
integration-test-oracledb-discovery:
	@make integration-test-target TARGET='discovery_integration_oracledb'

.PHONY: integration-test-oracledb-discovery-with-cover
integration-test-oracledb-discovery-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_oracledb'

.PHONY: smartagent-integration-test
smartagent-integration-test:
	@make integration-test-target TARGET='smartagent_integration'

.PHONY: smartagent-integration-test-with-cover
smartagent-integration-test-with-cover:
	@make integration-test-cover-target TARGET='smartagent_integration'

.PHONY: integration-test-envoy-discovery-k8s
integration-test-envoy-discovery-k8s:
	@make integration-test-target TARGET='discovery_integration_envoy_k8s'

.PHONY: integration-test-envoy-discovery-k8s-with-cover
integration-test-envoy-discovery-k8s-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_envoy_k8s'

.PHONY: integration-test-istio-discovery-k8s
integration-test-istio-discovery-k8s:
	@make integration-test-target TARGET='discovery_integration_istio_k8s'

.PHONY: integration-test-istio-discovery-k8s-with-cover
integration-test-istio-discovery-k8s-with-cover:
	@make integration-test-cover-target TARGET='discovery_integration_istio_k8s'

ifeq ($(COVER_TESTING),true)
# These targets are expensive to build, so only build if explicitly requested

.PHONY: gotest-with-codecov
gotest-with-codecov:
	@$(MAKE) for-all-target TARGET="test-with-codecov"
	$(GOCMD) tool covdata textfmt -i=./coverage -o ./coverage.txt

.PHONY: gotest-cover-without-race
gotest-cover-without-race:
	@$(MAKE) for-all-target TARGET="test-cover-without-race"
	$(GOCMD) tool covdata textfmt -i=./coverage  -o ./coverage.txt

endif

.PHONY: tidy-all
tidy-all:
	$(MAKE) for-all-target TARGET="tidy"
	$(MAKE) tidy

.PHONY: fmt-all
fmt-all:
	$(MAKE) for-all-target TARGET="fmt"
	$(MAKE) fmt

.PHONY: lint-all
lint-all:
	$(MAKE) for-all-target TARGET="lint"
	$(MAKE) lint

.PHONY: test-all
test-all:
	$(MAKE) for-all-target TARGET="test"
	$(MAKE) test

.PHONY: install-tools
install-tools:
	cd ./internal/tools && go install github.com/client9/misspell/cmd/misspell
	cd ./internal/tools && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	cd ./internal/tools && go install github.com/google/addlicense
	cd ./internal/tools && go install github.com/jstemmer/go-junit-report
	cd ./internal/tools && go install go.opentelemetry.io/collector/cmd/mdatagen
	cd ./internal/tools && go install github.com/tcnksm/ghr
	cd ./internal/tools && go install golang.org/x/tools/cmd/goimports
	cd ./internal/tools && go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
	cd ./internal/tools && go install golang.org/x/vuln/cmd/govulncheck@latest
	cd ./internal/tools && go install go.opentelemetry.io/build-tools/chloggen
	cd ./internal/tools && go install mvdan.cc/gofumpt

.PHONY: generate-metrics
generate-metrics:
	go generate -tags mdatagen ./...
	$(MAKE) fmt

.PHONY: otelcol
otelcol:
	go generate ./...
ifeq ($(COVER_TESTING), true)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) go build $(COVER_OPTS) -trimpath -o ./bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/otelcol
else
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -o ./bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/otelcol
endif
ifeq ($(OS), Windows_NT)
	$(LINK_CMD) .\bin\otelcol$(EXTENSION) .\bin\otelcol_$(GOOS)_$(GOARCH)$(EXTENSION)
else
	$(LINK_CMD) otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/otelcol$(EXTENSION)
endif


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

.PHONY: binaries-darwin_arm64
binaries-darwin_arm64:
	GOOS=darwin  GOARCH=arm64 $(MAKE) otelcol

.PHONY: binaries-linux_amd64
binaries-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) otelcol

.PHONY: binaries-linux_arm64
binaries-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) otelcol

.PHONY: binaries-windows_amd64
binaries-windows_amd64:
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) otelcol

.PHONY: binaries-linux_ppc64le
binaries-linux_ppc64le:
	GOOS=linux GOARCH=ppc64le $(MAKE) otelcol

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
	$(eval BUILD_INFO = -ldflags "${BUILD_X1} ${BUILD_X2}")
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

FILENAME?=$(shell git branch --show-current)
.PHONY: chlog-new
chlog-new:
	$(CHLOGGEN) new --filename $(FILENAME)

.PHONY: chlog-validate
chlog-validate:
	$(CHLOGGEN) validate

.PHONY: chlog-preview
chlog-preview:
	$(CHLOGGEN) update --dry

.PHONY: chlog-update
chlog-update:
	$(CHLOGGEN) update -v $(VERSION)

.PHONY: prepare-changelog
prepare-changelog:
	@if [ "$(VERSION)" = $(DEFAULT_VERSION) ]; then \
		echo "Error: VERSION is required. Usage: make prepare-changelog VERSION=v0.132.0"; \
		exit 1; \
	fi
	@make chlog-update
	@echo "Preparing changelog for $(VERSION)..."
	@./.github/workflows/scripts/prepare-changelog.sh $(VERSION)
