ARCH=amd64

.PHONY: all
all: so/libsplunk.so

so:
	@mkdir -p so

obj:
	@mkdir -p obj

.PHONY: clean
clean:
	rm -f tests so/* obj/*

obj/logger.o: obj src/logger.c
	gcc -c -Wall -Werror -fpic -o obj/logger.o src/logger.c

obj/splunk.o: obj src/splunk.c
	gcc -c -Wall -Werror -fpic -o obj/splunk.o src/splunk.c

so/libsplunk.so: obj so obj/splunk.o obj/logger.o
	gcc -shared -o so/libsplunk.so obj/splunk.o obj/logger.o

tests: src/splunk.c src/test_main.c
	gcc -o tests src/test_main.c src/logger_test.c src/splunk.c

.PHONY: test
test: tests
	./tests

.PHONY: dist
dist:
	@mkdir -p dist
	docker buildx build --platform linux/$(ARCH) -o type=image,name=libsplunk-builder:$(ARCH),push=false .
	docker run --platform linux/$(ARCH) --rm -v $(CURDIR)/dist:/libsplunk/so libsplunk-builder:$(ARCH) make tests all
	cp $(CURDIR)/dist/libsplunk.so $(CURDIR)/dist/libsplunk_$(ARCH).so
