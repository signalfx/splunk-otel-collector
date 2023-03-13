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
	go generate ./...
	@echo 'Linting LINUX code'
	CGO_ENABLED=0 GOGC=40 golangci-lint run --timeout 10m
	@echo 'Linting WINDOWS code'
	GOOS=windows CGO_ENABLED=0 GOGC=40 golangci-lint run --timeout 10m

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
