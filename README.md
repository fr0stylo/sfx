# sfx

`sfx` is a pluggable secret fetcher and exporter CLI. It loads configuration from `.sfx.yaml` (via [spf13/viper](https://github.com/spf13/viper)), executes provider plugins to retrieve secrets, and renders the collected data through exporter plugins (for example, to a `.env` template).

## Features
- Length-delimited protobuf protocol over stdio for stable host↔plugin communication.
- Lightweight helper packages for writing provider plugins (`sfx/plugin`) and exporters (`sfx/exporter`) without touching internal message types.
- Shared execution helper (`sfx/internal/client`) for launching any plugin with a protobuf request/response contract.
- Sample provider (`plugins/providers/file`) and exporter (`plugins/exporters/env`) illustrating the extension points. Option key summaries live in `plugins/providers/README.md` and `plugins/exporters/README.md`; each plugin folder includes a README describing its configuration pattern.

## Getting Started
Prerequisites:
- Go 1.22+ installed.
- `protoc` available on your PATH.
- `protoc-gen-go` installed (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`).

Clone the repository, then build the binaries:

```bash
make build
```

Create a `.sfx.yaml` that maps secrets to provider binaries and selects an exporter. Defaults are provided for the built-in providers (`file`, `vault`, `sops`, `awssecrets`, `awsssm`, `gcpsecrets`, `azurevault`) and exporters (`env`, `tfvars`, `template`, `shell`, `k8ssecret`, `ansible`), so you can omit those entries if the defaults suit you:

```yaml
providers:
  file: ./bin/providers/file

exporters:
  env: ./bin/exporters/env
  tfvars: ./bin/exporters/tfvars
  template: ./bin/exporters/template
  shell: ./bin/exporters/shell
  k8ssecret: ./bin/exporters/k8ssecret
  ansible: ./bin/exporters/ansible

output:
  type: env

secrets:
  SECRET_KEY:
    ref: SECRET_KEY
    provider: file
```

Running `./bin/sfx` will print the rendered export payload to stdout.

### HashiCorp Vault Provider
- Ref format: `<path>#<field>` (field optional when the secret map only contains one entry). For example: `secret/data/my-app#password`.
- Options (via `provider_options`): `address`, `token`, `namespace`, `field`, `timeout`. `address`/`token` fall back to `VAULT_ADDR` / `VAULT_TOKEN`.
- Requires network access to your Vault cluster and whichever auth style you prefer (token only by default).

Example:

```yaml
secrets:
  DB_PASSWORD:
    ref: secret/data/app/config#password
    provider: vault
    provider_options:
      address: https://vault.example.com
      namespace: teams/payments
```

### SOPS Provider
- Ref format: `<file>#<path>` where `<path>` uses dot or slash-separated segments (`.` can be escaped with `\.`). Arrays can be indexed numerically, for example `secrets.enc.yaml#deep.list.0`. Omitting the key path returns the fully decrypted document. When you provide `provider_options.path`, set the ref to `#<path>` to target keys.
- Options: `path` (when you prefer to set the file path outside the ref), `format` (overrides auto-detection), and `key_path` (default lookup when ref lacks one).

Example:

```yaml
secrets:
  API_TOKEN:
    ref: config/secrets.enc.yaml#integrations.api_token
    provider: sops
```

### AWS Secrets Manager Provider
- Ref format: `<secret-id>#<metadata>` where metadata optionally declares version or stage (`stage:AWSCURRENT`, `version:1234`, etc.). Metadata without a prefix defaults to a version stage.
- Options: `region`, `profile`, `version_id`, `version_stage`, `timeout`.
- Respects the default AWS credential chain; no additional auth wiring required.

Example:

```yaml
secrets:
  PAYMENT_KEY:
    ref: prod/payments/api-key#stage:AWSPREVIOUS
    provider: awssecrets
    provider_options:
      region: us-east-1
```

### AWS SSM Parameter Store Provider
- Ref format: the full parameter name (for example, `/prod/payments/db/password`).
- Options: `region`, `profile`, `with_decryption` (defaults to true), `timeout`.

Example:

```yaml
secrets:
  DB_PASSWORD:
    ref: /prod/payments/db/password
    provider: awsssm
    provider_options:
      region: us-west-2
```

### GCP Secret Manager Provider
- Ref format accepts `projects/<project>/secrets/<secret>/versions/<version>`, `<project>/<secret>#<version>` (defaults to `latest` when omitted), or `<secret>#<version>` when `provider_options.project` supplies the project.
- Options: `project`, `secret`, `version`, `timeout`.
- Uses Application Default Credentials (ADC) for authentication.

Example:

```yaml
secrets:
  API_TOKEN:
    ref: my-gcp-project/api-token#5
    provider: gcpsecrets
```

### Azure Key Vault Provider
- Ref format accepts `https://<vault>.vault.azure.net/secrets/<secret>/<version>`, `<vault-name>/<secret>#<version>`, or `#<version>` when `provider_options.secret` supplies the name.
- Options: `vault_url`, `vault_name`, `secret`, `version`, `timeout`.
- Uses `DefaultAzureCredential`, so configure whichever Azure auth mechanism you need (environment, managed identity, etc.).

Example:

```yaml
secrets:
  STORAGE_CONN_STRING:
    ref: finance-vault/storage-conn#latest
    provider: azurevault
```

### ENV Exporter
- Output type: `env`.
- Options: `key_template` (Sprig-enabled Go template applied per key).
- Renders lexicographically-sorted `KEY=value` lines, quoting values when necessary.

Example:

```yaml
output:
  type: env
  options:
    key_template: "{{ .Value | replace \"-\" \"_\" | upper }}"
```

### TFVARS Exporter
- Output type: `tfvars`.
- Options: `order` (list of keys to emit first; remaining keys are sorted alphabetically).
- Strings are quoted automatically; multi-line values use heredoc syntax; numeric/bool strings are left unquoted.

Example:

```yaml
output:
  type: tfvars
  options:
    order: [db_password, api_key]
```

### Template Exporter
- Output type: `template`.
- Options: `template` (inline Go template), `template_path` (file path alternative), `delims.left`, `delims.right` (custom delimiters).
- The template receives `.Values` (map of strings) and `.Raw` (map of byte slices); Sprig functions are preloaded.

Example:

```yaml
output:
  type: template
  options:
    template_path: ./templates/.env.tmpl
```

### Shell Exporter
- Output type: `shell`.
- Options: `shebang` (defaults to `#!/usr/bin/env bash`), `header` (array of comment lines), `export_format` (`export` or `assign`), `order` (explicit key order).
- Produces a shell script that exports/assigns each secret with proper quoting.

Example:

```yaml
output:
  type: shell
  options:
    header:
      - "Generated by sfx"
    export_format: export
```

### Kubernetes Secret Exporter
- Output type: `k8ssecret`.
- Options: `name` (required), `namespace`, `type`, `labels`, `annotations`.
- Emits a Kubernetes `Secret` manifest with base64-encoded data entries.

Example:

```yaml
output:
  type: k8ssecret
  options:
    name: app-secrets
    namespace: prod
    type: Opaque
```

### Ansible Exporter
- Output type: `ansible`.
- Options: `prefix` (prepend to each key), `order` (emit keys in a specific order).
- Renders a YAML mapping suitable for inclusion in Ansible variable files.

Example:

```yaml
output:
  type: ansible
  options:
    prefix: secret_
```

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
- Build the full toolchain with `make build` (uses the workspace-aware Makefile). Use `make clean` to clear `bin/`.
- Target a single module with `go build ./cmd/sfx`, `go -C plugins/providers build ./vault`, etc.—`go.work` keeps the modules linked locally.

## License
Distributed under the MIT License. See [LICENSE](LICENSE) for details.
