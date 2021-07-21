.PHONY: dev-image
dev-image:
	docker build -t splunk-otel-connector-puppet-dev .

.PHONY: run-dev-image
run-dev-image:
	docker exec -it splunk-otel-collector-puppet-dev bash 2>/dev/null || \
	docker run \
		--rm \
		--name splunk-otel-collector-puppet-dev \
		-it \
		-v $(CURDIR):/etc/puppetlabs/code/modules/splunk_otel_collector \
		splunk-otel-connector-puppet-dev \
		bash

.PHONY: lint
lint: dev-image
	docker run \
		--rm \
		splunk-otel-connector-puppet-dev \
		puppet-lint --fail-on-warnings .

.PHONY: rake-spec
rake-spec: dev-image
	docker run \
		--rm \
		splunk-otel-connector-puppet-dev \
		rake spec

.PHONY: package
package: dev-image
	docker run \
		--rm \
		-it \
		-v $(CURDIR):/etc/puppetlabs/code/modules/splunk_otel_collector \
		splunk-otel-connector-puppet-dev \
		bash -ec 'bundle exec rake module:clean && bundle exec rake module:build'
