GO ?= go
PROTOC ?= protoc

BIN_DIR := bin
PLUGIN_BIN := $(BIN_DIR)/providers
EXPORTER_BIN := $(BIN_DIR)/exporters

PROVIDERS := file vault sops awssecrets awsssm gcpsecrets azurevault
EXPORTERS := env tfvars template shell k8ssecret ansible
PROVIDER_MODULE_DIRS := $(addprefix plugins/providers/, $(PROVIDERS))
EXPORTER_MODULE_DIRS := $(addprefix plugins/exporters/, $(EXPORTERS))
PLUGIN_MODULE_DIRS := $(PROVIDER_MODULE_DIRS) $(EXPORTER_MODULE_DIRS)

.PHONY: all build build-sfx build-providers build-exporters fmt lint test proto clean tidy-plugins

all: build

build: build-providers build-exporters build-sfx

build-sfx: $(BIN_DIR)/sfx

build-providers: $(addprefix $(PLUGIN_BIN)/, $(PROVIDERS))

build-exporters: $(addprefix $(EXPORTER_BIN)/, $(EXPORTERS))

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(PLUGIN_BIN):
	mkdir -p $(PLUGIN_BIN)

$(EXPORTER_BIN):
	mkdir -p $(EXPORTER_BIN)

$(BIN_DIR)/sfx: build-providers build-exporters | $(BIN_DIR)
	$(GO) build -o $@ ./cmd

$(PLUGIN_BIN)/%: | $(PLUGIN_BIN)
	$(GO) -C plugins/providers/$* build -o $(abspath $@)

$(EXPORTER_BIN)/%: | $(EXPORTER_BIN)
	$(GO) -C plugins/exporters/$* build -o $(abspath $@)

fmt:
	$(GO)fmt ./...

lint:
	$(GO) vet ./...
	golangci-lint run ./...

tidy-plugins:
	@for dir in $(PLUGIN_MODULE_DIRS); do \
		echo "Running go mod tidy in $$dir"; \
		$(GO) -C $$dir mod tidy || exit 1; \
	done
	$(GO) work sync

test:
	$(GO) test ./...
	@for dir in $(PLUGIN_MODULE_DIRS); do \
		$(GO) -C $$dir test ./... || exit 1; \
	done

proto:
	$(PROTOC) --go_out=paths=source_relative:. proto/secret.proto
	$(PROTOC) --go_out=paths=source_relative:. proto/export.proto
	@mv proto/*.pb.go internal/rpc/ 2>/dev/null || true

clean:
	rm -rf $(BIN_DIR)
