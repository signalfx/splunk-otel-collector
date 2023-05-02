.DEFAULT_GOAL := check

NUM_CORES ?= $(shell getconf _NPROCESSORS_ONLN)

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
	GOOS=windows CGO_ENABLED=0 GOGC=40 golangci-lint run --allow-parallel-runners --timeout 10m -j$(NUM_CORES)

.PHONY: fmt
fmt:
	go generate ./...
	CGO_ENABLED=0 gofmt -w -l .

.PHONY: tidy
tidy:
	go mod tidy -compat=1.18

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
