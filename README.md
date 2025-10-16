# sfx

`sfx` is a pluggable secret fetcher and exporter CLI. It loads configuration from `.sfx.yaml`, executes provider plugins to retrieve secrets, and renders the collected data through exporter plugins (for example, to a `.env` template).

## Features
- Length-delimited protobuf protocol over stdio for stable host↔plugin communication.
- Lightweight helper packages for writing provider plugins (`sfx/plugin`) and exporters (`sfx/exporter`) without touching internal message types.
- Shared execution helper (`sfx/internal/client`) for launching any plugin with a protobuf request/response contract.
- Sample provider (`cmd/plugins/file`) and exporter (`cmd/exporters/env`) illustrating the extension points.

## Getting Started
Prerequisites:
- Go 1.22+ installed.
- `protoc` available on your PATH.
- `protoc-gen-go` installed (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`).

Clone the repository, then build the binaries:

```bash
go build -o bin/sfx ./cmd/sfx
go build -o bin/file ./cmd/plugins/file
go build -o bin/env ./cmd/exporters/env
```

Create a `.sfx.yaml` that maps secrets to provider binaries and selects an exporter:

```yaml
providers:
  file: ./bin/file

output:
  type: env

secrets:
  SECRET_KEY:
    ref: SECRET_KEY
    provider: file
```

Running `./bin/sfx` will print the rendered export payload to stdout.

## Architecture
- `cmd/sfx`: CLI entrypoint; loads config, fetches secrets via `internal/client.Call`, and hands the map to an exporter.
- `internal/rpc`: Generated protobuf definitions and framing helpers (`io.go`).
- `plugin`: Public helper for provider plugins. Handlers operate on `plugin.Request` / `plugin.Response`.
- `exporter`: Public helper for exporter plugins. Handlers receive a `map[string][]byte` and return binary payloads.
- `internal/client`: Shared process runner that wraps `exec.Command`, writing and reading protobuf messages generically.

Communication uses a length-prefixed protobuf envelope, ensuring any consumer written in Go (or other languages) can participate.

## Writing Plugins
### Providers
Implement a handler using `plugin.Run`:

```go
func main() {
	plugin.Run(plugin.HandlerFunc(func(req plugin.Request) (plugin.Response, error) {
		secret := lookupSecret(req.Ref)
		return plugin.Response{Value: secret}, nil
	}))
}
```

### Exporters
Exporters transform the collected secrets into an output format:

```go
func main() {
	exporter.Run(exporter.HandlerFunc(func(req exporter.Request) (exporter.Response, error) {
		payload := renderTemplate(req.Values)
		return exporter.Response{Payload: payload}, nil
	}))
}
```

Both helpers propagate structured errors to the host without leaking protocol details to plugin authors.

## Regenerating Protobufs
When editing the `.proto` files under `proto/`, regenerate Go bindings:

```bash
protoc --go_out=paths=source_relative:. proto/secret.proto
protoc --go_out=paths=source_relative:. proto/export.proto
mv proto/*.pb.go internal/rpc/
```

## Development
- Format Go code with `gofmt`.
- Run tests/builds with `go build ./...` (plugins are small binaries and can be rebuilt individually).
- Keep repository binaries in `bin/` in sync (`go build -o bin/<name> ./cmd/...`).

## License
Distributed under the MIT License. See [LICENSE](LICENSE) for details.
