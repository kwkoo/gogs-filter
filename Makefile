PREFIX=github.com/kwkoo
PACKAGE=gogsfilter
IMAGE_VERSION=0.1

BASE:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: install clean test dockerimage

install:
	@$(BASE)/install.sh

clean:
	@$(BASE)/clean.sh

test:
	@cd $(BASE)/src && go test -v $(PREFIX)/$(PACKAGE)/pkg

dockerimage:
	docker build -t $(PREFIX)/$(PACKAGE):$(IMAGE_VERSION) $(BASE)
	docker tag $(PREFIX)/$(PACKAGE):$(IMAGE_VERSION) ghcr.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION)
	docker tag ghcr.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION) ghcr.io/kwkoo/$(PACKAGE):latest
	docker login ghcr.io
	docker push ghcr.io/kwkoo/$(PACKAGE):$(IMAGE_VERSION)
	docker push ghcr.io/kwkoo/$(PACKAGE):latest