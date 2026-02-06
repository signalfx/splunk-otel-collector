ARCH=amd64
INSTALL_DIR=/usr/lib/splunk-instrumentation

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

# Run this to install and enable the auto-instrumentation files. Mostly intended for development.
.PHONY: install
install: all uninstall
	mkdir -p $(INSTALL_DIR)
	cp splunk-otel-javaagent.jar $(INSTALL_DIR)
	cp so/libotelinject.so $(INSTALL_DIR)
	echo $(INSTALL_DIR)/libotelinject.so > /etc/ld.so.preload

.PHONY: uninstall
uninstall:
	rm -f /etc/ld.so.preload
	rm -f $(INSTALL_DIR)/splunk-otel-javaagent.jar
	rm -f $(INSTALL_DIR)/libotelinject.so

# Run this from within this directory to create the devel image (just debian with gcc and a jdk). You only have to run
# this once-ish. Mostly intended for development.
.PHONY: docker-build
docker-build:
	docker build -t instr-devel -f devel.Dockerfile .

# Run this from with this directory to run and get a command line into the devel container created by dev-docker-build.
# Once you have a command line, you can run `make test`. Mostly intended for development.
.PHONY: docker-run
docker-run:
	docker run --rm -it -v `pwd`:/instr instr-devel

.PHONY: tests
tests: test-java test-nodejs test-dotnet

.PHONY: test-dotnet-java-nodejs
test-%: dist
	(cd tests/$* && ARCH=$(ARCH) ./test.sh)

