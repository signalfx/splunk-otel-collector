ARCH=amd64
INSTALL_DIR=/usr/lib/splunk-instrumentation

# Docker repository used.
DOCKER_REPO?=docker.io

.PHONY: all
all: so/libsplunk.so

# make -d: Prerequisite 'obj' is newer than target 'obj/logger.o'.
so:
	@mkdir -p so

obj:
	@mkdir -p obj

.PHONY: clean
clean:
	rm -f tests so/* obj/*

obj/main.o: obj src/main.c
	gcc -c -Wall -Werror -fpic -o obj/main.o src/main.c

so/libsplunk.so: obj so obj/main.o
	gcc -shared -o so/libsplunk.so obj/main.o

.PHONY: dist
dist:
	@mkdir -p dist
	docker buildx build --platform linux/$(ARCH) --build-arg DOCKER_REPO=$(DOCKER_REPO) -o type=image,name=libsplunk-builder:$(ARCH),push=false .
	docker rm -f libsplunk-builder 2>/dev/null || true
	docker run -d --platform linux/$(ARCH) --name libsplunk-builder libsplunk-builder:$(ARCH) sleep inf
	docker exec libsplunk-builder make all
	docker cp libsplunk-builder:/libsplunk/so/libsplunk.so dist/libsplunk_$(ARCH).so
	docker rm -f libsplunk-builder

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
	cp so/libsplunk.so $(INSTALL_DIR)
	echo $(INSTALL_DIR)/libsplunk.so > /etc/ld.so.preload

.PHONY: uninstall
uninstall:
	rm -f /etc/ld.so.preload
	rm -f $(INSTALL_DIR)/splunk-otel-javaagent.jar
	rm -f $(INSTALL_DIR)/libsplunk.so

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
tests: test-java test-nodejs

.PHONY: test-java
test-java: dist
	(cd tests/java && ./test.sh)

.PHONY: test-nodejs
test-nodejs: dist
	(cd tests/nodejs && ./test.sh)