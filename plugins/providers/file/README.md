# File Provider

Lightweight sample provider that constructs a deterministic secret value for a given reference. It is primarily intended for local development and integration testing.

## Build

```bash
go -C plugins/providers/file build
```

Running `make build` from the repository root will also emit `bin/providers/file`.

## Request Format

- **ref**: Logical identifier of the secret (for example, `SAMPLE_SECRET`). The provider uses this value when constructing the response.

## Configuration Options

- **path** *(optional)*: String prepended to the generated secret value.

## Example

```yaml
secrets:
  SAMPLE_SECRET:
    ref: SAMPLE_SECRET
    provider: file
    provider_options:
      path: /tmp/
```
