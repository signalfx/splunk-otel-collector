include ../../Makefile.common

GOCMD?=go

.PHONY := tgz
tgz: build
	tar --format ustar -czvf splunk-connect-for-otlp.tgz splunk-connect-for-otlp

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
