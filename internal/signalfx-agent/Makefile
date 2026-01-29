.DEFAULT_GOAL := check

NUM_CORES ?= $(shell getconf _NPROCESSORS_ONLN)

# Used to control whether race detection is enabled for tests
ENABLE_RACE_DETECTOR?=true

# build tags required by any component should be defined as an independent variables and later added to GO_BUILD_TAGS below
GOCMD?= go
GOTEST=$(GOCMD) test
GO_BUILD_TAGS?=""
ifeq ($(ENABLE_RACE_DETECTOR),true)
GOTEST_OPT?= -race -v -timeout 180s --tags=$(GO_BUILD_TAGS)
else
GOTEST_OPT?= -v -timeout 180s --tags=$(GO_BUILD_TAGS)
endif

# COVER_PKGS is the list of packages to include in the coverage
COVER_PKGS := $(shell $(GOCMD) list ./... | tr "\n" ",")
COVER_DIR_ABS?=$(PWD)/coverage
COVER_TESTING_OPTS?=-cover -covermode=atomic -coverpkg $(COVER_PKGS) -args -test.gocoverdir="$(COVER_DIR_ABS)"

.PHONY: clean
clean:
	rm -f pkg/core/constants/versions.go
	find pkg/monitors -name "genmetadata.go" -delete
	find pkg/monitors -name "template.go" -delete
	rm -f pkg/monitors/collectd/collectd.conf.go
	rm -f pkg/monitors/zcodegen/monitorcodegen

.PHONY: vet
vet:
	go generate ./...
	# Only consider it a failure if issues are in non-test files
	! CGO_ENABLED=0 go vet ./... 2>&1 | tee /dev/tty | grep '.go' | grep -v '_test.go'

.PHONY: vetall
vetall:
	go generate ./...
	CGO_ENABLED=0 go vet ./...

.PHONY: lint
lint:
	@echo "Skip liniting for singnalfx-agent"

.PHONY: fmt
fmt:
	go generate ./...
	CGO_ENABLED=0 gofmt -w -l .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go generate ./...
	CGO_ENABLED=0 go test -p $(NUM_CORES) ./...

.PHONY: test-race-detector
test-race-detector:
	go generate ./...
	CGO_ENABLED=1 go test -race -p $(NUM_CORES) ./...

.PHONY: check
check: lint vet test

.PHONY: %-noop
%-noop:
	@echo 'Intentionally not requiring $* for initial agent relocation'

.PHONY: checklicense
checklicense: checklicense-noop

.PHONY: misspell
misspell: misspell-noop

.PHONY: test-with-codecov
test-with-codecov:
	mkdir -p $(COVER_DIR_ABS)
	$(GOTEST) $(GOTEST_OPT) ./... $(COVER_TESTING_OPTS)
