# sfx – Secret Fetcher & Exporter

`sfx` is a pluggable CLI that retrieves secrets from diverse backends and renders them into the formats your tooling expects. It is designed as a product-grade foundation for teams that want deterministic secret materialisation without wiring every integration by hand.

- Fetch from Vault, SOPS, AWS, GCP, Azure, and more
- Export to `.env`, Terraform `.tfvars`, templated files, shell scripts, Kubernetes Secrets, and Ansible configs
- Add new providers/exporters with minimal glue thanks to the lightweight plugin SDKs

---

## Table of Contents

1. [Key Features](#key-features)
2. [Architecture at a Glance](#architecture-at-a-glance)
3. [Installation](#installation)
4. [Quick Start](#quick-start)
5. [Configuration Primer](#configuration-primer)
6. [Provider Catalog](#provider-catalog)
7. [Exporter Catalog](#exporter-catalog)
8. [Plugin Authoring](#plugin-authoring)
9. [Build & Test Workflow](#build--test-workflow)
10. [Roadmap & Ideas](#roadmap--ideas)
11. [Contributing](#contributing)
12. [License](#license)

---

## Key Features

- **Polyglot secret ingestion** – pull from files, Vault, SOPS, AWS Secrets Manager, AWS SSM Parameter Store, GCP Secret Manager, and Azure Key Vault out of the box.
- **Format-rich exporters** – ship secrets to `.env`, TFVARs, templated files, shell scripts, Kubernetes Secrets, and Ansible YAML.
- **Composable plugin system** – add new providers or exporters without touching the host code via stable protobuf-based RPC helpers.
- **Environment-aware defaults** – override configuration using environment variables (`SFX_*`).
- **Batteries-included tooling** – Makefile, Go workspace, and per-plugin modules keep builds reproducible.

---

## Architecture at a Glance

```
┌─────────────┐       ┌───────────────┐        ┌───────────────┐
│ .sfx.yaml   │  ───► │   sfx CLI     │  ───►  │ Exporter Plugin│
│ secrets map │       │(Go executable)│        │ (standalone)   │
└─────────────┘       │               │        └──────┬────────┘
                      │               │               │ protobuf over stdio
                      │               │        ┌──────▼────────┐
                      │               │        │ Provider Plugin│
                      └──────┬────────┘        │   (standalone) │
                             │ protobuf over stdio └────────────┘
                             ▼
                          Secret stores
```

Plugins communicate with the host via length-prefixed protobuf envelopes over standard I/O, ensuring a stable boundary across languages and processes.

---

## Installation

### Prerequisites

- Go 1.22+ (the workspace targets Go 1.25.1)
- `protoc` plus the Go protobuf plugin (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`) when regenerating protobuf bindings

### Build Everything

```bash
make build
```

This compiles the host CLI (`bin/sfx`) and all provider/exporter binaries under `bin/providers/` and `bin/exporters/`. Use `make clean` to purge build artifacts.

### Build a Single Module

```bash
go build -o bin/providers/vault ./plugins/providers/vault
go build -o bin/exporters/env ./plugins/exporters/env
```

---

## Quick Start

1. **Create `.sfx.yaml`**

   ```yaml
   providers:
     vault: ./bin/providers/vault
     sops: ./bin/providers/sops
     file: ./bin/providers/file

   exporters:
     env: ./bin/exporters/env

   output:
     type: env
     options:
       key_template: "{{ .Value | replace \"-\" \"_\" | upper }}"

   secrets:
     DB_PASSWORD:
       ref: secret/data/app/config#password
       provider: vault
       provider_options:
         address: https://vault.example.com
         namespace: platform-team
         timeout: 5s

     API_TOKEN:
       ref: config/secrets.enc.yaml#integrations.api_token
       provider: sops

     SAMPLE_SECRET:
       ref: SAMPLE_SECRET
       provider: file
       provider_options:
         path: /tmp/
   ```

2. **Run sfx**

   ```bash
   ./bin/sfx > .env
   ```

   The exporter renders the aggregated secrets to stdout. Redirect or pipe the output into the desired workflow.

3. **Override via Environment**

   Any `.sfx.yaml` key can be overridden with `SFX_*` environment variables (`.` → `_`). For example:

   ```bash
   export SFX_PROVIDERS_VAULT=./custom/vault-provider
   ```

---

## Configuration Primer

- **providers** – map plugin name ➜ executable path.
- **exporters** – map exporter name ➜ executable path.
- **output** – choose the exporter (`type`) and pass plugin-specific `options`.
- **secrets** – describe each secret: `ref`, `provider`, and optional `provider_options`.

Consult the per-plugin documentation under `plugins/providers/<name>/README.md` and `plugins/exporters/<name>/README.md` for detailed option references.

---

## Provider Catalog

| Provider          | Reference Format                              | Key Options (subset)                          |
|-------------------|-----------------------------------------------|-----------------------------------------------|
| `file`            | `<logical-name>`                              | `path`                                        |
| `vault`           | `<path>#<field>`                              | `address`, `token`, `namespace`, `field`, `timeout` |
| `sops`            | `<file>#<path>`                               | `path`, `format`, `key_path`                  |
| `awssecrets`      | `<secret-id>#stage:NAME` / `#version:ID`      | `region`, `profile`, `version_id`, `version_stage`, `timeout` |
| `awsssm`          | `<parameter-name>`                            | `region`, `profile`, `with_decryption`, `timeout` |
| `gcpsecrets`      | `projects/<project>/secrets/<secret>#<version>` | `project`, `secret`, `version`, `timeout`   |
| `azurevault`      | `https://<vault>.vault.azure.net/secrets/...` | `vault_url`, `vault_name`, `secret`, `version`, `timeout` |

Each provider README includes build instructions, authentication notes, and advanced usage tips.

---

## Exporter Catalog

| Exporter    | Output                          | Key Options (subset)                                       |
|-------------|---------------------------------|-------------------------------------------------------------|
| `env`       | `.env` key/value list           | `key_template`                                             |
| `tfvars`    | Terraform `.tfvars`             | `order`                                                    |
| `template`  | Go text/template                | `template`, `template_path`, `delims.left`, `delims.right` |
| `shell`     | Shell export script             | `shebang`, `header`, `export_format`, `order`              |
| `k8ssecret` | Kubernetes Secret manifest      | `name`, `namespace`, `type`, `labels`, `annotations`       |
| `ansible`   | Ansible-compatible YAML mapping | `prefix`, `order`                                          |

The exporter READMEs contain ready-to-use configuration snippets for each format.

---

## Plugin Authoring

### Provider Skeleton

```go
func main() {
	plugin.Run(plugin.HandlerFunc(func(req plugin.Request) (plugin.Response, error) {
		// use req.Ref and req.Options
		secret := []byte("value")
		return plugin.Response{Value: secret}, nil
	}))
}
```

### Exporter Skeleton

```go
func main() {
	exporter.Run(exporter.HandlerFunc(func(req exporter.Request) (exporter.Response, error) {
		// req.Values map[string][]byte
		payload := []byte("rendered output")
		return exporter.Response{Payload: payload}, nil
	}))
}
```

Both helpers take care of the protobuf transport, error propagation, and process wiring so you can focus on business logic.

---

## Build & Test Workflow

```bash
# Format
make fmt

# Run unit tests
make test

# Regenerate protobuf bindings (when proto/ changes)
make proto

# Full rebuild
make clean && make build
```

### Go Workspace Tips

- `go.work` includes the root CLI and each plugin module; `go work sync` keeps the workspace tidy.
- Build individual plugins with `go -C plugins/providers/<name> build`.

---

## Roadmap & Ideas

- Additional providers (HashiCorp Consul, CyberArk, Google Secret Manager versions, etc.).
- Exporters for Docker/Kubernetes env injection, JSON/INI files, and CI services.
- First-class testing harness for plugin authors.

Contributions and feature requests are welcome—see the next section.

---

## Contributing

1. Fork the repo and create a branch (`git checkout -b feature/my-feature`).
2. Run `make fmt` and `make test` before submitting.
3. Follow Go best practices and keep README/CHANGELOG entries up to date.
4. Open a pull request with a clear description and testing notes.

Issues and discussions are encouraged if you have questions or ideas.

---

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for the full text.

---

**Need help?** Open an issue or start a discussion. Happy secret shipping!
