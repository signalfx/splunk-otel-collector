.PHONY: all
all: so/libsplunk.so

.PHONY: clean
clean:
	rm tests so/* obj/*

obj/logger.o: src/logger.c
	gcc -c -Wall -Werror -fpic -o obj/logger.o src/logger.c

obj/splunk.o: src/splunk.c
	gcc -c -Wall -Werror -fpic -o obj/splunk.o src/splunk.c

so/libsplunk.so: obj/splunk.o obj/logger.o
	gcc -shared -o so/libsplunk.so obj/splunk.o obj/logger.o

tests: src/splunk.c src/test_main.c
	gcc -o tests src/test_main.c src/logger_test.c src/splunk.c

.PHONY: test
test: tests
	./tests
