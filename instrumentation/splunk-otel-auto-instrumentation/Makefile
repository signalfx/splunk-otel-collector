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

obj/logger.o: obj src/logger.c src/logger.h
	gcc -c -Wall -Werror -fpic -o obj/logger.o src/logger.c

obj/config.o: obj src/config.c src/config.h
	gcc -c -Wall -Werror -fpic -o obj/config.o src/config.c

obj/metrics_client.o: obj src/metrics_client.c src/metrics_client.h
	gcc -c -Wall -Werror -fpic -o obj/metrics_client.o src/metrics_client.c

obj/splunk.o: obj src/splunk.c src/splunk.h
	gcc -c -Wall -Werror -fpic -o obj/splunk.o src/splunk.c

obj/args.o: obj src/args.c src/args.h
	gcc -c -Wall -Werror -fpic -o obj/args.o src/args.c

obj/cmdline_reader.o: obj src/cmdline_reader.c src/cmdline_reader.h
	gcc -c -Wall -Werror -fpic -o obj/cmdline_reader.o src/cmdline_reader.c

so/libsplunk.so: obj so obj/logger.o obj/config.o obj/metrics_client.o obj/cmdline_reader.o obj/args.o obj/splunk.o
	gcc -shared -o so/libsplunk.so obj/logger.o obj/config.o obj/metrics_client.o obj/cmdline_reader.o obj/args.o obj/splunk.o

tests: src/test_main.h src/test_main.c src/test_utils.h src/test_utils.c src/logger.h src/test_logger.c \
src/metrics_client.h src/cmdline_reader.h src/cmdline_reader_test.c obj/config.o obj/metrics_client.o obj/args.o obj/splunk.o
	gcc -g -o tests src/test_main.c src/test_utils.c src/test_logger.c src/cmdline_reader_test.c obj/config.o obj/metrics_client.o obj/args.o obj/splunk.o

.PHONY: test
test: tests
	./tests

.PHONY: dist
dist:
	@mkdir -p dist
	docker buildx build --platform linux/$(ARCH) --build-arg DOCKER_REPO=$(DOCKER_REPO) -o type=image,name=libsplunk-builder:$(ARCH),push=false .
	docker rm -f libsplunk-builder 2>/dev/null || true
	docker run -d --platform linux/$(ARCH) --name libsplunk-builder libsplunk-builder:$(ARCH) sleep inf
	docker exec libsplunk-builder make test all
	docker cp libsplunk-builder:/libsplunk/so/libsplunk.so dist/libsplunk_$(ARCH).so
	docker rm -f libsplunk-builder

.PHONY: deb-rpm-package
%-package:
ifneq ($(SKIP_COMPILE), true)
	$(MAKE) dist
endif
	@mkdir -p dist
	docker build -t instrumentation-fpm $(CURDIR)/../packaging/fpm
	docker run --rm -v $(CURDIR)/../../:/repo instrumentation-fpm \
		/repo/instrumentation/splunk-otel-auto-instrumentation/packaging/$*/build.sh "$(VERSION)" "$(ARCH)"

# Run this to install and enable the auto-instrumentation files. Mostly intended for development.
.PHONY: install
install: all uninstall
	mkdir -p $(INSTALL_DIR)
	cp install/instrumentation.conf $(INSTALL_DIR)
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
