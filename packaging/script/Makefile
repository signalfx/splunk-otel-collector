ARCH?=amd64

.PHONY: docker
docker:
	docker build -t script:latest .

.PHONY: deb-rpm-tar-package
%-package:
	docker build -t otelcol-script-fpm-$(ARCH) --platform linux/$(ARCH) fpm
	docker run --rm -v $(CURDIR)/../..:/repo -e PACKAGE=$* -e VERSION=$(VERSION) -e ARCH=$(ARCH) otelcol-script-fpm-$(ARCH)

.PHONY: deb-rpm-tar-test
%-test: venv
	bash -c "source venv/bin/activate && pytest -m $* tests/package_test.py"

venv:
	virtualenv venv
	bash -c "source venv/bin/activate && pip install -r tests/requirements.txt"

.PHONY: package
package: tar-package deb-package rpm-package

.PHONY: test
test: tar-test rpm-test deb-test


