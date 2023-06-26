HELM_PLUGIN_NAME := $(shell cat plugin.yaml | grep "name" | cut -d '"' -f 2)
VERSION := $(shell cat plugin.yaml | grep "version" | cut -d '"' -f 2)
LDFLAGS := "-X main.version=${VERSION}"

.PHONY: build tag

build:
	export CGO_ENABLED=0 && \
	go build -o bin/${HELM_PLUGIN_NAME} -ldflags $(LDFLAGS) ./cmd/diff

tag:
	@scripts/tag.sh
