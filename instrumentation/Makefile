ARCH?=$(shell go env GOARCH)

# Docker repository used.
DOCKER_REPO?=docker.io

INJECTOR_GIT_REPO?=git@github.com:pjanotti/opentelemetry-injector.git
INJECTOR_GIT_BRANCH?=add-build-time-allowed-env-var-prefixes-option
# Allow Splunk-specific defaults in the shared injector env file when building the custom injector branch.
INJECTOR_ALLOWED_ENV_VAR_PREFIXES?=OTEL_,SPLUNK_

INJECTOR_SRC_DIR?=build/opentelemetry-injector

.PHONY: all
all: so/libotelinject_$(ARCH).so

# make -d: Prerequisite 'obj' is newer than target 'obj/logger.o'.
so:
	@mkdir -p so

.PHONY: clean
clean:
	rm -f tests so/*

# Checkout the injector source if not already present (dev convenience; in CI use the checkout target).
$(INJECTOR_SRC_DIR):
	@mkdir -p build
	git clone --quiet --depth 1 --branch "$(INJECTOR_GIT_BRANCH)" "$(INJECTOR_GIT_REPO)" "$(INJECTOR_SRC_DIR)"

.PHONY: checkout-injector
checkout-injector: $(INJECTOR_SRC_DIR)

so/libotelinject_$(ARCH).so: so $(INJECTOR_SRC_DIR)
	$(MAKE) -C "$(INJECTOR_SRC_DIR)" dist \
		ARCH="$(ARCH)" \
		DOCKER_REPO="$(DOCKER_REPO)" \
		ALLOWED_ENV_VAR_PREFIXES='$(INJECTOR_ALLOWED_ENV_VAR_PREFIXES)'
	cp "$(INJECTOR_SRC_DIR)/dist/libotelinject_$(ARCH).so" "$@"

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
