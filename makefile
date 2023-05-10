HELM_PLUGIN_NAME := diff
VERSION := $(shell cat plugin.yaml | grep "version" | cut -d '"' -f 2)
LDFLAGS := "-X main.version=${VERSION}"


.PHONY: build
build:
	export CGO_ENABLED=0 && \
	go build -o bin/${HELM_PLUGIN_NAME} -ldflags $(LDFLAGS) ./cmd/diff

.PHONY: tag
tag:
	@scripts/tag.sh
