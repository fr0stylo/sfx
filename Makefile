GO ?= go
PROTOC ?= protoc

BIN_DIR := bin
PLUGIN_BIN := $(BIN_DIR)/plugins
EXPORTER_BIN := $(BIN_DIR)/exporters

PLUGINS := file
EXPORTERS := env

.PHONY: all build build-sfx build-plugins build-exporters fmt test proto clean

all: build

build: build-plugins build-exporters build-sfx

build-sfx: $(BIN_DIR)/sfx

build-plugins: $(addprefix $(PLUGIN_BIN)/, $(PLUGINS))

build-exporters: $(addprefix $(EXPORTER_BIN)/, $(EXPORTERS))

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(PLUGIN_BIN):
	mkdir -p $(PLUGIN_BIN)

$(EXPORTER_BIN):
	mkdir -p $(EXPORTER_BIN)

$(BIN_DIR)/sfx: build-plugins build-exporters | $(BIN_DIR)
	$(GO) build -o $@ ./cmd/sfx

$(PLUGIN_BIN)/%: | $(PLUGIN_BIN)
	$(GO) build -o $@ ./cmd/plugins/$*

$(EXPORTER_BIN)/%: | $(EXPORTER_BIN)
	$(GO) build -o $@ ./cmd/exporters/$*

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
