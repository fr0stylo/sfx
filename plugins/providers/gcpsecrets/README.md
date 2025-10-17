# GCP Secret Manager Provider

Accesses secret payloads stored in Google Cloud Secret Manager using Application Default Credentials (ADC).

## Authentication

Ensure ADC is configured (for example via `gcloud auth application-default login`, workload identity, or service account JSON pointed to by `GOOGLE_APPLICATION_CREDENTIALS`).

## Build

```bash
go -C plugins/providers/gcpsecrets build
```

`make build` writes the binary to `bin/providers/gcpsecrets`.

## Request Format

- **ref**: Accepts `projects/<project>/secrets/<secret>/versions/<version>`, `<project>/<secret>#<version>`, or `<secret>#<version>` when the project is supplied through options. Version defaults to `latest`.

## Configuration Options

- **project** *(optional)*: Project ID when the reference omits it.
- **secret** *(optional)*: Secret name when the reference is `#<version>`.
- **version** *(optional)*: Version identifier (default `latest`).
- **timeout** *(optional)*: Request timeout (Go duration).

## Example

```yaml
secrets:
  API_TOKEN:
    ref: my-project/api-token#5
    provider: gcpsecrets
    provider_options:
      timeout: 4s
```
