# SOPS Provider

Decrypts documents encrypted with Mozilla SOPS and returns either the full payload or a specific field.

## Prerequisites

Ensure SOPS can locate the appropriate KMS/age/GPG keys for the encrypted file (for example via environment variables or key stores).

## Build

```bash
go -C plugins/providers/sops build
```

`make build` outputs `bin/providers/sops`.

## Request Format

- **ref**: `<file>#<path>` where `<path>` uses dot or slash notation (`data.api_key`, `sections/0/value`). Escape literal dots with `\.`. Omit `#<path>` to return the entire decrypted document.

## Configuration Options

- **path** *(optional)*: File path when `ref` omits it (e.g. `#data.api_key`).
- **format** *(optional)*: Override format detection (`yaml`, `json`, `ini`, `dotenv`, `binary`).
- **key_path** *(optional)*: Default lookup path used when the ref lacks `#<path>`.

## Example

```yaml
secrets:
  API_TOKEN:
    ref: config/secrets.enc.yaml#integrations.api_token
    provider: sops
    provider_options:
      format: yaml
```
