GO ?= go
PROTOC ?= protoc
BIN_DIR := bin
PLUGINS := file
EXPORTERS := env

.PHONY: all build build-sfx build-plugins build-exporters fmt test proto clean

all: build

build: build-sfx build-plugins build-exporters

build-sfx: $(BIN_DIR)/sfx

build-plugins: $(addprefix $(BIN_DIR)/, $(PLUGINS))

build-exporters: $(addprefix $(BIN_DIR)/, $(EXPORTERS))

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(BIN_DIR)/sfx: | $(BIN_DIR)
	$(GO) build -o $@ ./cmd/sfx

$(BIN_DIR)/%: | $(BIN_DIR)
ifneq (,$(filter $*, $(PLUGINS)))
	$(GO) build -o $@ ./cmd/plugins/$*
else ifneq (,$(filter $*, $(EXPORTERS)))
	$(GO) build -o $@ ./cmd/exporters/$*
else
	$(error Unknown binary $*)
endif

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
