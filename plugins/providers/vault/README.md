# HashiCorp Vault Provider

Fetches secret values from a HashiCorp Vault cluster using the official Go SDK. Suitable for KV v1/v2 (and other logical backends that return key/value maps).

## Authentication

The provider relies on Vault tokens. Supply a token via the `token` option or export `VAULT_TOKEN`. Namespaces are supported for Vault Enterprise.

## Build

```bash
go -C plugins/providers/vault build
```

`make build` will produce `bin/providers/vault` automatically.

## Request Format

- **ref**: `<path>#<field>` (for example, `secret/data/app/config#password`). Omit `#<field>` when the secret map exposes a single entry.

## Configuration Options

- **address** *(optional)*: Vault address (defaults to `VAULT_ADDR`).
- **token** *(optional)*: Vault token (defaults to `VAULT_TOKEN`).
- **namespace** *(optional)*: Namespace string for Vault Enterprise.
- **field** *(optional)*: Default field when `ref` lacks `#<field>`.
- **timeout** *(optional)*: Request timeout (Go duration such as `15s`).

## Example

```yaml
secrets:
  DB_PASSWORD:
    ref: secret/data/app/config#password
    provider: vault
    provider_options:
      address: https://vault.example.com
      namespace: teams/payments
      timeout: 5s
```
