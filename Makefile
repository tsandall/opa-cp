BIN 		:= opa-cp
PKG 		:= github.com/open-policy-agent/opa-cp
REPOSITORY 	:= openpolicyagent/opa-cp
VERSION 	:= 0.1
BUILD_IMAGE := golang:1.8-alpine

build:
	docker run -it \
		-v $$(pwd)/.go:/go \
		-v $$(pwd):/go/src/$(PKG) \
		-v $$(pwd)/bin/linux_amd64:/go/bin \
		-v $$(pwd)/.go/std/amd64:/usr/local/go/pkg/linux_amd64_static \
		-w /go/src/$(PKG) \
		$(BUILD_IMAGE) \
	 	/bin/sh -c "./build.sh"
	docker build -t $(REPOSITORY):$(VERSION) \
		-t $(REPOSITORY):latest \
		-f Dockerfile.run \
		.
	@echo Successfully built $(REPOSITORY):$(VERSION) $(REPOSITORY):latest
