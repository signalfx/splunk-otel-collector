ARCH=amd64

# Docker repository used.
DOCKER_REPO?=docker.io

.PHONY: all
all: so/libotelinject_$(ARCH).so

# make -d: Prerequisite 'obj' is newer than target 'obj/logger.o'.
so:
	@mkdir -p so

.PHONY: clean
clean:
	rm -f tests so/*

so/libotelinject_$(ARCH).so: so
	@wget -q -O so/libotelinject_$(ARCH).so https://github.com/open-telemetry/opentelemetry-injector/releases/download/v0.1.0/libotelinject_$(ARCH).so

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
