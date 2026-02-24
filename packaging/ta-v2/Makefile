SRC_ROOT := $(shell git rev-parse --show-toplevel)

# TA_TARGET_OS empty means single package with both Linux and Windows binaries
TA_TARGET_OS ?=
# Only x86_64 architecture is supported for Splunk modular inputs
TA_ARCH ?= x86_64

ifneq ($(TA_TARGET_OS),)
	TA_PLATFORM_SUFFIX ?= _$(TA_TARGET_OS)_$(TA_ARCH)
else
	TA_PLATFORM_SUFFIX ?=
endif

# Set tar exclude options based on the target OS for packaging
ifeq ($(TA_TARGET_OS),windows)
	TAR_EXCLUDE_OPT ?= --exclude='*/linux_$(TA_ARCH)/*'
else ifeq ($(TA_TARGET_OS),linux)
	TAR_EXCLUDE_OPT ?= --exclude='*/windows_$(TA_ARCH)/*'
endif

MOD_INPUT_NAME := Splunk_TA_OTel_Collector
PKG_NAME := $(MOD_INPUT_NAME)$(TA_PLATFORM_SUFFIX)

sync-configs:
	@echo "Syncing configuration files ..."
	@if [ ! -d ./assets/configs/ ]; then mkdir -p ./assets/configs/; fi
	cp $(SRC_ROOT)/cmd/otelcol/config/collector/agent_config.yaml ./assets/configs/agent_config.yaml

.PHONY: ta-inputs-conf
ta-inputs-conf:
	@echo "Generating inputs.conf and inputs.conf.spec ..."
	cd $(SRC_ROOT) && go run ./cmd/ta-inputs-from-schema \
		-scheme cmd/otelcol/ta_v2_scheme.xml \
		-global-settings packaging/ta-v2/inputs_conf_global_settings.txt \
		-name $(MOD_INPUT_NAME) \
		-assets packaging/ta-v2/assets

.PHONY: validate-target-os
validate-target-os:
ifneq ($(TA_TARGET_OS),)
	@if [ "$(TA_TARGET_OS)" != "linux" ] && [ "$(TA_TARGET_OS)" != "windows" ]; then \
		echo "Error: TA_TARGET_OS must be either empty, 'linux' or 'windows', got '$(TA_TARGET_OS)'"; \
		exit 1; \
	fi
	@echo "Targeting $(TA_TARGET_OS) in a OS specific package ..."
else
	@echo "Targeting both Linux and Windows in a single package ..."
endif

.PHONY: all-packages
all-packages: pack
	# Build packages targeting specific OSes
	$(MAKE) TA_TARGET_OS=linux pack
	$(MAKE) TA_TARGET_OS=windows pack

.PHONY: pack
pack: validate-target-os sync-configs ta-inputs-conf
	# Build a package according to the specified TA_TARGET_OS
	@echo "Packing add-on $(PKG_NAME) ..."
	@if [ ! -d ./out/distribution/ ]; then mkdir -p ./out/distribution/; fi
	@tmpdir=$$(mktemp -d) && \
	listfile=$$(mktemp) && \
	tr -d '\r' < $(PKG_NAME).fileslist.txt > $$listfile && \
	cp -rL ./assets/ $$tmpdir/$(MOD_INPUT_NAME) && \
	COPYFILE_DISABLE=1 tar -czf ./out/distribution/$(PKG_NAME).tgz --no-recursion -C $$tmpdir -T $$listfile && \
	rm -rf $$tmpdir $$listfile

.PHONY: clean-packages
clean-packages:
	@echo "Cleaning packages ..."
	@if [ -d ./out/distribution/ ]; then rm -rf ./out/distribution/; fi

.PHONY: clean-otelcol
clean-otelcol:
	@echo "Cleaning collector binaries ..."
	@if [ -d ./assets/linux_x86_64/bin/ ]; then rm -rf ./assets/linux_x86_64/bin/; fi
	@if [ -d ./assets/windows_x86_64/bin/ ]; then rm -rf ./assets/windows_x86_64/bin/; fi

.PHONY: test-ta-v2
test-ta-v2:
	@echo "Running package tests ..."
	cd tests && go test --tags=ta_v2 -v ./...

.PHONY: build-otelcol
build-otelcol: build-otelcol-linux build-otelcol-windows

.PHONY: build-otelcol-linux
build-otelcol-linux: ./assets/linux_x86_64/bin/$(MOD_INPUT_NAME)
./assets/linux_x86_64/bin/$(MOD_INPUT_NAME):
	@echo "Building Linux executable ..."
	if [ ! -d ./assets/linux_x86_64/bin/ ]; then mkdir -p ./assets/linux_x86_64/bin/; fi
	cd $(SRC_ROOT)/ && GOOS=linux GOARCH=amd64 $(MAKE) otelcol
	cp $(SRC_ROOT)/bin/otelcol_linux_amd64 ./assets/linux_x86_64/bin/$(MOD_INPUT_NAME)

.PHONY: build-otelcol-windows
build-otelcol-windows: ./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe
./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe:
	@echo "Building Windows executable ..."
	if [ ! -d ./assets/windows_x86_64/bin/ ]; then mkdir -p ./assets/windows_x86_64/bin/; fi
	cd $(SRC_ROOT)/ && GOOS=windows GOARCH=amd64 EXTENSION=.exe $(MAKE) otelcol
	cp $(SRC_ROOT)/bin/otelcol_windows_amd64.exe ./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe
