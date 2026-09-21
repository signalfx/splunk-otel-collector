-include config.env
export

SHELL := /bin/bash

SSH := ssh -i $(SSH_KEY) -o StrictHostKeyChecking=no
SCP := scp -i $(SSH_KEY) -o StrictHostKeyChecking=no

.PHONY: help build setup setup-splunk setup-uf setup-otel add-ta compare \
	append-test-event clean-indexes

help:
	@echo "Targets:"
	@echo "  build          Build otelcol binary from main + add-folder-observer"
	@echo "  setup          Setup all three VMs (splunk, uf, otel) in order"
	@echo "  setup-splunk   Install Splunk Enterprise on VM C"
	@echo "  setup-uf       Install Universal Forwarder + TA(s) on VM A"
	@echo "  setup-otel     Install OTel collector + TA(s) on VM B"
	@echo "  add-ta         Deploy a TA to both VM A and VM B  (TA=<name>)"
	@echo "  append-test-event  Append one identical event to UF and OTel test files"
	@echo "  compare        Run comparison against Splunk indexes"
	@echo "  clean-indexes  Delete and recreate both Splunk indexes"

# ------------------------------------------------------------
# Build
# ------------------------------------------------------------

build:
	scripts/build_otel.sh

# ------------------------------------------------------------
# Setup
# ------------------------------------------------------------

setup: setup-splunk setup-uf setup-otel

setup-splunk:
	scripts/install_splunk.sh

setup-uf:
	scripts/install_uf.sh

setup-otel:
	scripts/install_otel.sh

# ------------------------------------------------------------
# TA management
# ------------------------------------------------------------

add-ta:
	@test -n "$(TA)" || (echo "Usage: make add-ta TA=Splunk_TA_nix" && exit 1)
	scripts/add_ta.sh "$(TA)"

# ------------------------------------------------------------
# Deterministic test input
# ------------------------------------------------------------

TEST_EVENT_FILE ?= /home/splunker/otel_uf_test.log

append-test-event:
	@test -n "$(EVENT)" || (echo "Usage: make append-test-event EVENT='key=value ...'" && exit 1)
	@printf '%s\n' "$(EVENT)" | $(SSH) $(SSH_USER)@$(UF_HOST) \
		"sudo tee -a $(TEST_EVENT_FILE) >/dev/null"
	@printf '%s\n' "$(EVENT)" | $(SSH) $(SSH_USER)@$(OTEL_HOST) \
		"sudo tee -a $(TEST_EVENT_FILE) >/dev/null"
	@echo "Appended one event to $(TEST_EVENT_FILE) on UF and OTel."

# ------------------------------------------------------------
# Compare
# ------------------------------------------------------------

compare:
	pytest tests/ \
		--splunk-host $(SPLUNK_HOST) \
		--port $(SPLUNK_MGMT_PORT) \
		--password $(SPLUNK_ADMIN_PASSWORD) \
		--uf-index $(UF_INDEX) \
		--otel-index $(OTEL_INDEX) \
		-v

# ------------------------------------------------------------
# Maintenance
# ------------------------------------------------------------

clean-indexes:
	$(SSH) $(SSH_USER)@$(SPLUNK_HOST) \
		"/opt/splunk/bin/splunk remove index $(UF_INDEX) -auth admin:$(SPLUNK_ADMIN_PASSWORD) || true"
	$(SSH) $(SSH_USER)@$(SPLUNK_HOST) \
		"/opt/splunk/bin/splunk remove index $(OTEL_INDEX) -auth admin:$(SPLUNK_ADMIN_PASSWORD) || true"
	$(SSH) $(SSH_USER)@$(SPLUNK_HOST) \
		"/opt/splunk/bin/splunk add index $(UF_INDEX) -auth admin:$(SPLUNK_ADMIN_PASSWORD)"
	$(SSH) $(SSH_USER)@$(SPLUNK_HOST) \
		"/opt/splunk/bin/splunk add index $(OTEL_INDEX) -auth admin:$(SPLUNK_ADMIN_PASSWORD)"

