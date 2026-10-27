include ../../Makefile.Common

# The goldens were captured from this UF build, so bumping it means regenerating them.
SPLUNK_UF_VERSION ?= 9.4.0
SPLUNK_UF_BUILD_HASH ?= 6b4ebe426ca6
SPLUNK_UF_PASSWORD ?= parityTest123!

UF_ROOT := $(CURDIR)/.local
UF_DIR := $(UF_ROOT)/splunkforwarder

.PHONY: update-goldens
update-goldens: install-uf
	PARITY_UF_DIR=$(UF_DIR) $(GOTEST) -update -v -timeout 30m .

.PHONY: install-uf
install-uf: $(UF_DIR)

$(UF_DIR):
	@set -eu; \
	case "$$(uname -s)/$$(uname -m)" in \
		Linux/x86_64) dir=linux; file=linux-amd64.tgz ;; \
		Linux/aarch64) dir=linux; file=linux-arm64.tgz ;; \
		Darwin/*) dir=osx; file=darwin-universal2.tgz ;; \
		*) echo "no pinned UF tarball for $$(uname -s)/$$(uname -m); set PARITY_UF_DIR to an existing install" >&2; exit 1 ;; \
	esac; \
	url="https://download.splunk.com/products/universalforwarder/releases/$(SPLUNK_UF_VERSION)/$$dir/splunkforwarder-$(SPLUNK_UF_VERSION)-$(SPLUNK_UF_BUILD_HASH)-$$file"; \
	echo "Downloading $$url"; \
	mkdir -p "$(UF_ROOT)"; \
	curl -fsSL "$$url" | tar -xz -C "$(UF_ROOT)"; \
	mkdir -p "$(UF_DIR)/etc/system/local"; \
	printf '[user_info]\nUSERNAME = admin\nPASSWORD = %s\n' '$(SPLUNK_UF_PASSWORD)' \
		> "$(UF_DIR)/etc/system/local/user-seed.conf"
