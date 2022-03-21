### VARIABLES

ADDLICENSE= addlicense

# All source code and documents. Used in spell check.
ALL_DOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) \
                                -type f | sort)

# All source code excluding any third party code and excluding the testbed.
# This is the code that we want to run tests for and lint, etc.
ALL_SRC := $(shell find . -name '*.go' \
							-not -path './examples/*' \
							-not -path './tests/*' \
							-type f | sort)

# ALL_PKGS is the list of all packages where ALL_SRC files reside.
ALL_PKGS := $(shell go list $(sort $(dir $(ALL_SRC))))

ALL_TESTS_DIRS := $(shell find tests -name '*_test.go' | xargs -L 1 dirname | uniq | sort -r)

# BUILD_TYPE should be one of (dev, release).
BUILD_TYPE?=release

GIT_SHA=$(shell git rev-parse --short HEAD)
GO_ACC=go-acc
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

# CircleCI runtime.NumCPU() is for host machine despite container instance only granting 2.
# If we are in a CI job, limit to 2 (and scale as we increase executor size).
NUM_CORES := $(shell if [ -z ${CIRCLE_JOB} ]; then echo `getconf _NPROCESSORS_ONLN` ; else echo 2; fi )
GOTEST=go test -p $(NUM_CORES)
GOTEST_OPT?= -v -race -timeout 180s

IMPI=impi
LINT=golangci-lint
MISSPELL=misspell -error
MISSPELL_CORRECTION=misspell -w

BUILD_INFO_IMPORT_PATH=go.opentelemetry.io/collector/internal/version
BUILD_INFO_IMPORT_PATH_CORE=github.com/open-telemetry/opentelemetry-collector/internal/version
VERSION=$(shell git describe --match "v[0-9]*" HEAD)
BUILD_X1=-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)
BUILD_X2=-X $(BUILD_INFO_IMPORT_PATH_CORE).Version=$(VERSION)
BUILD_INFO=-ldflags "${BUILD_X1} ${BUILD_X2}"

SMART_AGENT_RELEASE=$(shell cat internal/buildscripts/packaging/smart-agent-release.txt)
SKIP_COMPILE=false
ARCH=amd64

# For integration testing against local changes you can run
# SPLUNK_OTEL_COLLECTOR_IMAGE='otelcol:latest' make -e docker-otelcol integration-test
# for local docker build testing or
# SPLUNK_OTEL_COLLECTOR_IMAGE='' make -e otelcol integration-test
# for local binary testing (agent-bundle configuration required)
export SPLUNK_OTEL_COLLECTOR_IMAGE?=quay.io/signalfx/splunk-otel-collector-dev:latest

### FUNCTIONS

# Function to execute a command. Note the empty line before endef to make sure each command
# gets executed separately instead of concatenated with previous one.
# Accepts command to execute as first parameter.
define exec-command
$(1)

endef

### TARGETS

all-srcs:
	@echo $(ALL_SRC) | tr ' ' '\n' | sort

all-pkgs:
	@echo $(ALL_PKGS) | tr ' ' '\n' | sort

.DEFAULT_GOAL := all

.PHONY: all
all: checklicense impi lint misspell test otelcol

.PHONY: test
test: integration-vet
	$(GOTEST) $(GOTEST_OPT) $(ALL_PKGS)

.PHONY: integration-vet
integration-vet:
	cd tests && go vet ./...

.PHONY: integration-test
integration-test:
	@set -e; for dir in $(ALL_TESTS_DIRS); do \
	  echo "go test ./... in $${dir}"; \
	  (cd "$${dir}" && \
	   $(GOTEST) -v -timeout 5m -count 1 ./... ); \
	done

.PHONY: end-to-end-test
end-to-end-test:
	@set -e; cd tests/endtoend && $(GOTEST) -v -tags endtoend -timeout 5m -count 1 ./...

.PHONY: test-with-cover
test-with-cover:
	@echo Verifying that all packages have test files to count in coverage
	@echo pre-compiling tests
	@time go test -p $(NUM_CORES) -i $(ALL_PKGS)
	$(GO_ACC) $(ALL_PKGS)
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: addlicense
addlicense:
	$(ADDLICENSE) -y "" -c 'Splunk, Inc.' $(ALL_SRC)

.PHONY: checklicense
checklicense:
	@ADDLICENSEOUT=`$(ADDLICENSE) -check $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:\n"; \
			echo "$$ADDLICENSEOUT\n"; \
			echo "Use 'make addlicense' to fix this."; \
			exit 1; \
		else \
			echo "Check License finished successfully"; \
		fi

# ALL_MODULES includes ./* dirs (excludes . dir)
ALL_GO_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )
ALL_PYTHON_DEPS := $(shell find . -type f \( -name "setup.py" -o -name "requirements.txt" \) -exec dirname {} \; | sort | egrep  '^./')
ALL_DOCKERFILES := $(shell find . -type f -name Dockerfile -exec dirname {} \; | grep -v '^./tests' | sort)
DEPENDABOT_PATH=./.github/dependabot.yml
.PHONY: gendependabot
gendependabot:
	@echo "Recreate dependabot.yml file"
	@echo "# File generated by \"make gendependabot\"; DO NOT EDIT.\n" > ${DEPENDABOT_PATH}
	@echo "version: 2" >> ${DEPENDABOT_PATH}
	@echo "updates:" >> ${DEPENDABOT_PATH}
	@echo "Add entry for \"/\""
	@echo "  - package-ecosystem: \"gomod\"\n    directory: \"/\"\n    schedule:\n      interval: \"daily\"" >> ${DEPENDABOT_PATH}
	@set -e; for dir in $(ALL_GO_MODULES); do \
		(echo "Add entry for \"$${dir:1}\"" && \
		  echo "  - package-ecosystem: \"gomod\"\n    directory: \"$${dir:1}\"\n    schedule:\n      interval: \"daily\"" >> ${DEPENDABOT_PATH} ); \
	done
	@set -e; for dir in $(ALL_PYTHON_DEPS); do \
		(echo "Add entry for \"$${dir:1}\"" && \
		  echo "  - package-ecosystem: \"pip\"\n    directory: \"$${dir:1}\"\n    schedule:\n      interval: \"daily\"" >> ${DEPENDABOT_PATH} ); \
	done
	@set -e; for dir in $(ALL_DOCKERFILES); do \
		(echo "Add entry for \"$${dir:1}\"" && \
		  echo "  - package-ecosystem: \"docker\"\n    directory: \"$${dir:1}\"\n    schedule:\n      interval: \"daily\"" >> ${DEPENDABOT_PATH} ); \
	done

.PHONY: misspell
misspell:
	$(MISSPELL) $(ALL_DOC)

.PHONY: misspell-correction
misspell-correction:
	$(MISSPELL_CORRECTION) $(ALL_DOC)

.PHONY: lint
lint:
	$(LINT) run

.PHONY: impi
impi:
	@$(IMPI) --local github.com/signalfx/splunk-otel-collector --scheme stdThirdPartyLocal ./...

.PHONY: install-tools
install-tools:
	go install github.com/client9/misspell/cmd/misspell@v0.3.4
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
	go install github.com/google/addlicense@v0.0.0-20200906110928-a0294312aa76
	go install github.com/jstemmer/go-junit-report@v0.9.1
	go install github.com/ory/go-acc@v0.2.6
	go install github.com/pavius/impi/cmd/impi@v0.0.3
	go install github.com/tcnksm/ghr@v0.14.0
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

.PHONY: otelcol
otelcol:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -o ./bin/otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/otelcol
	ln -sf otelcol_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/otelcol

.PHONY: translatesfx
translatesfx:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -o ./bin/translatesfx_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/translatesfx
	ln -sf translatesfx_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/translatesfx

.PHONY: migratecheckpoint
migratecheckpoint:
	go generate ./...
	GO111MODULE=on CGO_ENABLED=0 go build -o ./bin/migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/migratecheckpoint
	ln -sf migratecheckpoint_$(GOOS)_$(GOARCH)$(EXTENSION) ./bin/migratecheckpoint

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
	cp ./bin/otelcol_linux_$(ARCH) ./cmd/otelcol/otelcol
	cp ./bin/translatesfx_linux_$(ARCH) ./cmd/otelcol/translatesfx
	cp ./bin/migratecheckpoint_linux_$(ARCH) ./cmd/otelcol/migratecheckpoint
	cp ./internal/buildscripts/packaging/collect-libs.sh ./cmd/otelcol/collect-libs.sh
	docker buildx build --platform linux/$(ARCH) -o type=image,name=otelcol:$(ARCH),push=false --build-arg ARCH=$(ARCH) --build-arg SMART_AGENT_RELEASE=$(SMART_AGENT_RELEASE) ./cmd/otelcol/
	docker tag otelcol:$(ARCH) otelcol:latest
	rm ./cmd/otelcol/otelcol
	rm ./cmd/otelcol/translatesfx
	rm ./cmd/otelcol/migratecheckpoint
	rm ./cmd/otelcol/collect-libs.sh

.PHONY: binaries-all-sys
binaries-all-sys: binaries-darwin_amd64 binaries-linux_amd64 binaries-linux_arm64 binaries-windows_amd64

.PHONY: binaries-darwin_amd64
binaries-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) otelcol
	GOOS=darwin  GOARCH=amd64 $(MAKE) translatesfx
	GOOS=darwin  GOARCH=amd64 $(MAKE) migratecheckpoint

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

.PHONY: deb-rpm-package
%-package:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-linux_$(ARCH)
endif
	docker build -t otelcol-fpm internal/buildscripts/packaging/fpm
	docker run --rm -v $(CURDIR):/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) -e SMART_AGENT_RELEASE=$(SMART_AGENT_RELEASE) otelcol-fpm

.PHONY: msi
msi:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) binaries-windows_amd64
endif
	./internal/buildscripts/packaging/msi/build.sh "$(VERSION)" "$(SMART_AGENT_RELEASE)"
