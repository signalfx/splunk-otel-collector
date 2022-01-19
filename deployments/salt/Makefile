.PHONY: dev-image
dev-image:
	docker build -t splunk-otel-collector-salt-dev .

.PHONY: run-dev-image
run-dev-image:
	docker exec -it splunk-otel-collector-salt-dev 2>/dev/null || \
	docker run \
		--rm \
		--name splunk-otel-collector-salt-dev \
		-it \
		splunk-otel-collector-salt-dev \
		bash

.PHONY: test
test:
	salt-lint -v /srv/salt/splunk-otel-collector/*.sls
	salt-call --local state.show_sls splunk-otel-collector
	salt-call --local pillar.items

.PHONY: lint
lint: dev-image
	docker run \
		--rm \
		--name splunk-otel-collector-salt-dev \
		splunk-otel-collector-salt-dev \
		bash -c \
		"salt-lint -v /srv/salt/splunk-otel-collector/*.sls; salt-call --local state.show_sls splunk-otel-collector; salt-call --local pillar.items;"
