SHELL := /bin/bash

APP_NAME := loglint
GO ?= go
GOFLAGS ?=
LDFLAGS ?=

BIN_DIR := bin
CMD_PKG := ./cmd/loglint
PLUGIN_PKG := ./plugin
MODULE_PLUGIN_PKG := ./moduleplugin

CUSTOM_GCL_CONFIG := examples/module-plugin/.custom-gcl.yml
CUSTOM_GCL_BIN := custom-gcl

.PHONY: help fmt test vet lint build build-plugin build-module-plugin build-all clean custom-gcl

help:
	@echo "Targets:"
	@echo "  fmt                 - gofmt project"
	@echo "  test                - run unit tests"
	@echo "  vet                 - run go vet"
	@echo "  lint                - test + vet"
	@echo "  build               - build standalone checker"
	@echo "  build-plugin        - build golangci-lint go plugin (.so)"
	@echo "  build-module-plugin - build module plugin package"
	@echo "  build-all           - build all binaries/artifacts"
	@echo "  custom-gcl          - build custom golangci-lint binary"
	@echo "  clean               - remove local build artifacts"

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

lint: test vet

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(APP_NAME) $(CMD_PKG)

build-plugin:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -buildmode=plugin -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(APP_NAME).so $(PLUGIN_PKG)

build-module-plugin:
	$(GO) build $(GOFLAGS) $(MODULE_PLUGIN_PKG)

build-all: build build-plugin build-module-plugin

custom-gcl:
	golangci-lint custom -v --config $(CUSTOM_GCL_CONFIG) --output $(CUSTOM_GCL_BIN)

clean:
	rm -f $(BIN_DIR)/$(APP_NAME) $(BIN_DIR)/$(APP_NAME).so $(CUSTOM_GCL_BIN)
