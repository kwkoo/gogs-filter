PREFIX=github.com/kwkoo
PACKAGE=gogsfilter
IMAGE_VERSION=0.1

GOPATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOBIN=$(GOPATH)/bin

.PHONY: install clean test dockerimage

install:
	@$(GOPATH)/install.sh

clean:
	@$(GOPATH)/clean.sh

test:
	@GOPATH=$(GOPATH) go test -v $(PREFIX)/$(PACKAGE)

dockerimage:
	docker build -t $(PREFIX)/$(PACKAGE):$(IMAGE_VERSION) .
	docker tag $(PREFIX)/$(PACKAGE):$(IMAGE_VERSION) quay.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION)
	docker tag quay.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION) quay.io/kwkoo/$(PACKAGE):latest
	docker login quay.io
	docker push quay.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION)
	docker push quay.io/kwkoo/$(PACKAGE):latest