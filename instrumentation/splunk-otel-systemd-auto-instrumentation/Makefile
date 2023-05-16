ARCH=amd64

.PHONY: all
all:
	$(MAKE) deb-package ARCH=amd64
	$(MAKE) deb-package ARCH=arm64
	$(MAKE) rpm-package ARCH=amd64
	$(MAKE) rpm-package ARCH=arm64

.PHONY: deb-rpm-package
%-package:
	@mkdir -p dist
	docker build -t instrumentation-fpm $(CURDIR)/../packaging/fpm
	docker run --rm -v $(CURDIR)/../../:/repo instrumentation-fpm \
		/repo/instrumentation/splunk-otel-systemd-auto-instrumentation/packaging/$*/build.sh "$(VERSION)" "$(ARCH)"
