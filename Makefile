include ./Makefile.Common

### VARIABLES


# BUILD_TYPE should be one of (dev, release).
BUILD_TYPE?=release
VERSION?=latest

GIT_SHA=$(shell git rev-parse --short HEAD)
GO_ACC=go-acc
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

FIND_MOD_ARGS=-type f -name "go.mod" -not -path '*/signalfx-agent/pkg/apm/*'
TO_MOD_DIR=dirname {} \; | sort | egrep  '^./'
# NONROOT_MODS includes ./* dirs (excludes . dir)
NONROOT_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )

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

SKIP_COMPILE=false
ARCH=amd64
BUNDLE_SUPPORTED_ARCHS := amd64 arm64
SKIP_BUNDLE=false

# For integration testing against local changes you can run
# SPLUNK_OTEL_COLLECTOR_IMAGE='otelcol:latest' make -e docker-otelcol integration-test
# for local docker build testing or
# SPLUNK_OTEL_COLLECTOR_IMAGE='' make -e otelcol integration-test
# for local binary testing (agent-bundle configuration required)
export SPLUNK_OTEL_COLLECTOR_IMAGE?=quay.io/signalfx/splunk-otel-collector-dev:latest

# Docker repository used.
DOCKER_REPO?=docker.io

GOTESPLIT_TOTAL?=1
GOTESPLIT_INDEX?=0

### TARGETS

.DEFAULT_GOAL := all

.PHONY: all
all: checklicense impi lint misspell test otelcol

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(NONROOT_MODS); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

.PHONY: integration-vet
integration-vet:
	@set -e; cd tests && go vet -tags integration,testutilsintegration,endtoend,zeroconfig,testutils ./... && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) -tags testutils,testutilsintegration -v -timeout 5m -count 1 ./...

.PHONY: integration-test
integration-test: integration-vet
	@set -e; cd tests && $(GOTEST_SERIAL) $(BUILD_INFO_TESTS) --tags=integration -v -timeout 5m -count 1 ./...

.PHONY: end-to-end-test
end-to-end-test:
	@set -e; cd tests/endtoend && $(GOTEST) -v -tags endtoend -timeout 5m -count 1 ./...

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
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="tidy"
	go mod tidy -compat=1.19

.PHONY: install-tools
install-tools:
	cd ./internal/tools && go install github.com/client9/misspell/cmd/misspell
	cd ./internal/tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd ./internal/tools && go install github.com/google/addlicense
	cd ./internal/tools && go install github.com/jstemmer/go-junit-report
	cd ./internal/tools && go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen
	cd ./internal/tools && go install github.com/ory/go-acc
	cd ./internal/tools && go install github.com/pavius/impi/cmd/impi
	cd ./internal/tools && go install github.com/tcnksm/ghr
	cd ./internal/tools && go install golang.org/x/tools/cmd/goimports
	cd ./internal/tools && go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment


.PHONY: generate-metrics
generate-metrics:
	go generate -tags mdatagen ./...
	$(MAKE) fmt

.PHONY: otelcol
otelcol:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -o ./bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/otelcol
	ln -sf otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/otelcol

.PHONY: translatesfx
translatesfx:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -o ./bin/translatesfx_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/translatesfx
	ln -sf translatesfx_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/translatesfx

.PHONY: migratecheckpoint
migratecheckpoint:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -trimpath -o ./bin/migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/migratecheckpoint
	ln -sf migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/migratecheckpoint

.PHONY: bundle.d
bundle.d:
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
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-linux_$(ARCH)
endif
ifneq ($(filter $(ARCH), $(BUNDLE_SUPPORTED_ARCHS)),)
ifneq ($(SKIP_BUNDLE), true)
	$(MAKE) -C internal/signalfx-agent/bundle agent-bundle-linux ARCH=$(ARCH) DOCKER_REPO=$(DOCKER_REPO)
endif
endif
	rm -rf ./cmd/otelcol/dist
	mkdir -p ./cmd/otelcol/dist
	cp ./bin/otelcol_linux_$(ARCH) ./cmd/otelcol/dist/otelcol
	cp ./bin/translatesfx_linux_$(ARCH) ./cmd/otelcol/dist/translatesfx
	cp ./bin/migratecheckpoint_linux_$(ARCH) ./cmd/otelcol/dist/migratecheckpoint
	cp ./internal/buildscripts/packaging/collect-libs.sh ./cmd/otelcol/dist/collect-libs.sh
ifneq ($(filter $(ARCH), $(BUNDLE_SUPPORTED_ARCHS)),)
	cp ./dist/agent-bundle_linux_$(ARCH).tar.gz ./cmd/otelcol/dist/agent-bundle.tar.gz
endif
	docker buildx build --platform linux/$(ARCH) -o type=image,name=otelcol:$(ARCH),push=false --build-arg ARCH=$(ARCH) --build-arg DOCKER_REPO=$(DOCKER_REPO) ./cmd/otelcol/
	docker tag otelcol:$(ARCH) otelcol:latest
	rm -rf ./cmd/otelcol/dist

.PHONY: binaries-all-sys
binaries-all-sys: binaries-darwin_amd64 binaries-darwin_arm64 binaries-linux_amd64 binaries-linux_arm64 binaries-windows_amd64 binaries-linux_ppc64le

.PHONY: binaries-darwin_amd64
binaries-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) otelcol
	GOOS=darwin  GOARCH=amd64 $(MAKE) translatesfx
	GOOS=darwin  GOARCH=amd64 $(MAKE) migratecheckpoint

.PHONY: binaries-darwin_arm64
binaries-darwin_arm64:
	GOOS=darwin  GOARCH=arm64 $(MAKE) otelcol
	GOOS=darwin  GOARCH=arm64 $(MAKE) translatesfx
	GOOS=darwin  GOARCH=arm64 $(MAKE) migratecheckpoint

.PHONY: binaries-linux_amd64
binaries-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) otelcol
	GOOS=linux   GOARCH=amd64 $(MAKE) translatesfx
	GOOS=linux   GOARCH=amd64 $(MAKE) migratecheckpoint

.PHONY: binaries-linux_arm64
binaries-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) otelcol
	GOOS=linux   GOARCH=arm64 $(MAKE) translatesfx
	GOOS=linux   GOARCH=arm64 $(MAKE) migratecheckpoint

.PHONY: binaries-windows_amd64
binaries-windows_amd64:
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) otelcol
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) translatesfx
	GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) migratecheckpoint

.PHONY: binaries-linux_ppc64le
binaries-linux_ppc64le:
	GOOS=linux GOARCH=ppc64le $(MAKE) otelcol
	GOOS=linux GOARCH=ppc64le $(MAKE) translatesfx
	GOOS=linux GOARCH=ppc64le $(MAKE) migratecheckpoint

.PHONY: deb-rpm-tar-package
%-package:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-linux_$(ARCH)
endif
ifneq ($(filter $(ARCH), $(BUNDLE_SUPPORTED_ARCHS)),)
ifneq ($(SKIP_BUNDLE), true)
	$(MAKE) -C internal/signalfx-agent/bundle agent-bundle-linux ARCH=$(ARCH) DOCKER_REPO=$(DOCKER_REPO)
endif
endif
	docker build -t otelcol-fpm internal/buildscripts/packaging/fpm
	docker run --rm -v $(CURDIR):/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) otelcol-fpm

.PHONY: msi
msi:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-windows_amd64
endif
	test -f ./dist/agent-bundle_windows_amd64.zip || (echo "./dist/agent-bundle_windows_amd64.zip not found! Either download a pre-built bundle to ./dist/, or run './internal/signalfx-agent/bundle/scripts/windows/make.ps1 bundle' on a windows host and copy it to ./dist/." && exit 1)
	./internal/buildscripts/packaging/msi/build.sh "$(VERSION)" "$(DOCKER_REPO)"

.PHONY: update-examples
update-examples:
	cd examples && $(MAKE) update-examples

.PHONY: install-test-tools
install-test-tools:
	cd ./tests/tools && go install github.com/Songmu/gotesplit/cmd/gotesplit

.PHONY: integration-test-split
integration-test-split: install-test-tools
	@set -e; cd tests && gotesplit --total=$(GOTESPLIT_TOTAL) --index=$(GOTESPLIT_INDEX) ./... -- -p 1 $(BUILD_INFO_TESTS) --tags=integration -v -timeout 5m -count 1
