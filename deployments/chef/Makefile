.PHONY: dev-image
dev-image:
	docker build -t splunk-otel-collector-chef-dev .

.PHONY: run-dev-image
run-dev-image:
	docker exec -it splunk-otel-collector-chef-dev bash 2>/dev/null || \
	docker run \
		--rm \
		--name splunk-otel-collector-chef-dev \
		-it \
		splunk-otel-collector-chef-dev \
		bash

.PHONY: lint
lint: dev-image
	docker run \
		--rm \
		splunk-otel-collector-chef-dev \
		cookstyle .

.PHONY: rake-spec
rake-spec: dev-image
	docker run \
		--rm \
		splunk-otel-collector-chef-dev \
		bundle exec rake spec
