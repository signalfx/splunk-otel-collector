include ../../Makefile.Common

MOD_INPUT_NAME = splunk_otel_dotnet_deployer

.PHONY: build-pack
build-pack: download-dotnet-assets build pack

.PHONY: download-dotnet-assets
download-dotnet-assets:
	@echo "Downloading .NET assets..."
	@for asset in Splunk.OTel.DotNet.psm1 splunk-opentelemetry-dotnet-windows.zip; do \
		if [ ! -f ./assets/windows_x86_64/bin/$${asset} ]; then \
			echo " > Downloading $${asset}..."; \
			curl -s -L https://github.com/signalfx/splunk-otel-dotnet/releases/latest/download/$${asset} -o ./assets/windows_x86_64/bin/$${asset}; \
		else \
			echo " > $${asset} already exists, skipping download."; \
		fi \
	done

.PHONY: build
build: ./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe
./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe: $(shell find ./ -name '*.go')
	@echo "Building executable ..."
	@GOOS=windows go build -o ./assets/windows_x86_64/bin/ ./cmd/$(MOD_INPUT_NAME)/...

.PHONY: pack
pack:
	@echo "Packing add-on ..."
	@if [ ! -d ./out/distribution/ ]; then mkdir -p ./out/distribution/; fi
	@tar -C ./assets/ --transform='s,^,$(MOD_INPUT_NAME)/,' -hcz -f ./out/distribution/$(MOD_INPUT_NAME).tgz .

.PHONY: clean
clean:
	@echo "Cleaning..."
	@if [ -f ./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe ]; then rm -rf ./assets/windows_x86_64/bin/$(MOD_INPUT_NAME).exe; fi
	@if [ -f ../out/distribution/ ]; then rm -rf ./out/distribution/; fi
