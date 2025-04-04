.DEFAULT_GOAL := check

NUM_CORES ?= $(shell getconf _NPROCESSORS_ONLN)

# build tags required by any component should be defined as an independent variables and later added to GO_BUILD_TAGS below
GOCMD?= go
GOTEST=$(GOCMD) test
GO_BUILD_TAGS?=""
GOTEST_OPT?= -race -v -timeout 180s --tags=$(GO_BUILD_TAGS)
GOTEST_OPT_WITHOUT_RACE?= -v -timeout 180s --tags=$(GO_BUILD_TAGS)

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
	@echo 'Linting LINUX code'
	CGO_ENABLED=0 GOGC=40 golangci-lint run --allow-parallel-runners --timeout 10m -j$(NUM_CORES)
	@echo 'Linting WINDOWS code'
	GOOS=windows CGO_ENABLED=0 GOGC=40 golangci-lint run --allow-parallel-runners --timeout 15m -j$(NUM_CORES)

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

.PHONY: impi
impi: impi-noop

.PHONY: misspell
misspell: misspell-noop

#############################################################
# Cover targets are to conform to top level targets.
# SignalFx agent packages are not included in codecov output.
#############################################################

.PHONY: test-with-codecov
test-with-codecov:
	$(GOTEST) $(GOTEST_OPT) ./...

# Currently, running CGO_ENABLED=0 go test -race is broken in gopsutil/v4
# This target should be removed in favor of "test" once the following issue is resolved:
# https://github.com/shirou/gopsutil/issues/1774
.PHONY: test-cover-without-race
test-cover-without-race:
	$(GOTEST) $(GOTEST_OPT_WITHOUT_RACE) ./...
