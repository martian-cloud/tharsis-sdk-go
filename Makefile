# Makefile for Tharsis SDK for Go

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "1.0.0")
PACKAGES := $(shell go list ./... | grep -v /vendor/)
LDFLAGS := -ldflags "-X main.Version=${VERSION}"

.PHONY: lint
lint: ## run golint on all Go package
	@revive $(PACKAGES)

.PHONY: vet
vet: ## run golint on all Go package
	@go vet $(PACKAGES)

.PHONY: fmt
fmt: ## run "go fmt" on all Go packages
	@go fmt $(PACKAGES)

.PHONY: test
test: ## run unit tests
	go test ./...

.PHONY: integration
integration: ## run integration (and unit) tests
	go test -v -tags=integration ./...

# The End.
