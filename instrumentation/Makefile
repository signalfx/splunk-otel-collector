ARCH?=$(shell go env GOARCH)

# Docker repository used.
DOCKER_REPO?=docker.io

INJECTOR_RELEASE=$(shell cat packaging/injector-release.txt | tr -d '[:space:]')
INJECTOR_DOWNLOAD_URL=https://github.com/open-telemetry/opentelemetry-injector/releases/download/$(INJECTOR_RELEASE)

.PHONY: all
all: so/libotelinject_$(ARCH).so

# make -d: Prerequisite 'obj' is newer than target 'obj/logger.o'.
so:
	@mkdir -p so

.PHONY: clean
clean:
	rm -f tests so/*

so/libotelinject_$(ARCH).so: so
	@wget -q -O so/libotelinject_$(ARCH).so $(INJECTOR_DOWNLOAD_URL)/libotelinject_$(ARCH).so

.PHONY: dist
dist: so/libotelinject_$(ARCH).so
	@mkdir -p dist
	@cp so/libotelinject_$(ARCH).so dist/libotelinject_$(ARCH).so

.PHONY: deb-rpm-package
%-package:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) dist
endif
	@mkdir -p dist
	docker build -t instrumentation-fpm packaging/fpm
	docker run --rm -v $(CURDIR)/../:/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) instrumentation-fpm

.PHONY: tests
tests: test-java test-nodejs test-dotnet

.PHONY: test-dotnet-java-nodejs
test-%: dist
	(cd tests/$* && ARCH=$(ARCH) ./test.sh)
