GOPATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOBIN=$(GOPATH)/bin

.PHONY: install clean test

install:
	@$(GOPATH)/install.sh

clean:
	@$(GOPATH)/clean.sh

test:
	@GOPATH=$(GOPATH) go test -v github.com/kwkoo/gogsfilter