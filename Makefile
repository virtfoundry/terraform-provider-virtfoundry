.PHONY: build test install fmt vet lint

PROVIDER_NAME := virtfoundry
REGISTRY_HOST := registry.terraform.io
REGISTRY_NAMESPACE := virtfoundry
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -o bin/terraform-provider-$(PROVIDER_NAME) -ldflags "-X main.version=$(VERSION)"

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

install: build
	@mkdir -p ~/.terraform.d/plugins/$(REGISTRY_HOST)/$(REGISTRY_NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/linux_amd64
	cp bin/terraform-provider-$(PROVIDER_NAME) ~/.terraform.d/plugins/$(REGISTRY_HOST)/$(REGISTRY_NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/linux_amd64/

lint: fmt vet test
