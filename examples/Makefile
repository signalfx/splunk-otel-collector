### VARIABLES

ALL_EXAMPLES_DOCKER_COMPOSE_TESTS := $(shell find ./ -name 'docker-compose.yml' | xargs dirname | xargs basename -a | awk '{print "example-"$$1}')

VERSION?=latest


### TARGETS
.DEFAULT_GOAL := examples-test-run

.PHONY: update-examples
# example usage: (will set version of splunk otel collector ot 0.59.1)
# make VERSION=0.59.1 update-example
# --in-place without a rename is not portable btw osx and windows/linux, reccomendation seems to be to give it a valid backup file and then remove them.
update-examples:
	find ./ -name docker-compose.yml -exec sed --in-place='.orig' -e 's;quay.io/signalfx/splunk-otel-collector:.*$$;quay.io/signalfx/splunk-otel-collector:$(VERSION);g' {} \; -exec rm {}.orig \;

.PHONY: examples-test-run
examples-test-run: $(ALL_EXAMPLES_DOCKER_COMPOSE_TESTS)

.PHONY: $(ALL_EXAMPLES_DOCKER_COMPOSE_TESTS)
$(ALL_EXAMPLES_DOCKER_COMPOSE_TESTS): example-%:
	cd ./$* && docker-compose up -d
	sleep 15
ifeq ($(shell docker ps | grep otelcollector | wc -l),0)
	docker logs otelcollector
	cd ./$* && docker-compose down && exit 1
endif
	cd ./$* && docker-compose down
