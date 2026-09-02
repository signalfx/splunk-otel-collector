GOCMD?=go

VERSION ?= $(shell git describe --match "v[0-9]*" HEAD)
PKG_NAME ?= splunk-connect-for-otlp

.PHONY := tgz
tgz: build
	@mkdir -p out/distribution/ && \
	tmpdir=$$(mktemp -d) && \
	cp -rL splunk-connect-for-otlp $$tmpdir/splunk-connect-for-otlp && \
	sed -i.bak \
		-e 's/^version[[:space:]]*=[[:space:]]*.*/version = $(patsubst v%,%,$(VERSION))/' \
		-e 's/^id[[:space:]]*=[[:space:]]*.*/id = $(PKG_NAME)/' \
		-e 's/^name[[:space:]]*=[[:space:]]*.*/name = $(PKG_NAME)/' \
		$$tmpdir/$(PKG_NAME)/default/app.conf && \
	rm -f $$tmpdir/$(PKG_NAME)/default/app.conf.bak && \
	tar --format ustar -C $$tmpdir -czf out/distribution/splunk-connect-for-otlp.tgz splunk-connect-for-otlp

splunk-connect-for-otlp.tgz: tgz

splunk-connect-for-otlp/linux_x86_64/bin/splunk-connect-for-otlp:
	mkdir -p splunk-connect-for-otlp/linux_x86_64/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C ../../cmd/splunk-connect-for-otlp -trimpath -o ../../packaging/otlpinput/splunk-connect-for-otlp/linux_x86_64/bin/splunk-connect-for-otlp .

splunk-connect-for-otlp/windows_x86_64/bin/splunk-connect-for-otlp:
	mkdir -p splunk-connect-for-otlp/windows_x86_64/bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -C ../../cmd/splunk-connect-for-otlp -trimpath -o ../../packaging/otlpinput/splunk-connect-for-otlp/windows_x86_64/bin/splunk-connect-for-otlp .

.PHONY := build
build: splunk-connect-for-otlp/linux_x86_64/bin/splunk-connect-for-otlp splunk-connect-for-otlp/windows_x86_64/bin/splunk-connect-for-otlp

.PHONY := splunk
splunk: splunk-connect-for-otlp.tgz
	docker run --rm -it -v $(PWD)/splunk-connect-for-otlp.tgz:/tmp/splunk-connect-for-otlp.tgz \
		--platform linux/amd64 \
		-e "SPLUNK_PASSWORD=changeme" \
		-e "SPLUNK_APPS_URL=file:///tmp/splunk-connect-for-otlp.tgz" \
		-e "SPLUNK_GENERAL_TERMS=--accept-sgt-current-at-splunk-com" \
		-e "SPLUNK_START_ARGS=--accept-license" \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8000:8000 \
		splunk/splunk:10.0
