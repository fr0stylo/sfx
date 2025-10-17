GO ?= go
PROTOC ?= protoc

BIN_DIR := bin
PLUGIN_BIN := $(BIN_DIR)/providers
EXPORTER_BIN := $(BIN_DIR)/exporters

PROVIDERS := file vault sops awssecrets awsssm gcpsecrets azurevault
EXPORTERS := env tfvars template shell k8ssecret ansible

.PHONY: all build build-sfx build-providers build-exporters fmt test proto clean

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
	$(GO) build -o $@ ./cmd/sfx

$(PLUGIN_BIN)/%: | $(PLUGIN_BIN)
	$(GO) -C plugins/providers/$* build -o $(abspath $@)

$(EXPORTER_BIN)/%: | $(EXPORTER_BIN)
	$(GO) -C plugins/exporters/$* build -o $(abspath $@)

fmt:
	$(GO)fmt ./...

test:
	$(GO) test ./...

proto:
	$(PROTOC) --go_out=paths=source_relative:. proto/secret.proto
	$(PROTOC) --go_out=paths=source_relative:. proto/export.proto
	@mv proto/*.pb.go internal/rpc/ 2>/dev/null || true

clean:
	rm -rf $(BIN_DIR)
