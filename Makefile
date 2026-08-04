.PHONY: build test install fmt vet lint test-integration

PROVIDER_NAME := virtfoundry
REGISTRY_HOST := registry.terraform.io
REGISTRY_NAMESPACE := virtfoundry
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

build:
	go build -o bin/terraform-provider-$(PROVIDER_NAME) -ldflags "-X main.version=$(VERSION)"

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

install: build
	@mkdir -p ~/.terraform.d/plugins/$(REGISTRY_HOST)/$(REGISTRY_NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(GOOS)_$(GOARCH)
	cp bin/terraform-provider-$(PROVIDER_NAME) ~/.terraform.d/plugins/$(REGISTRY_HOST)/$(REGISTRY_NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(GOOS)_$(GOARCH)/

test-integration: build
	@chmod +x scripts/test-vm.sh
	./scripts/test-vm.sh

lint: fmt vet test
