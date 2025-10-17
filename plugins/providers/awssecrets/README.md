# AWS Secrets Manager Provider

Retrieves string or binary secrets from AWS Secrets Manager using the AWS SDK for Go v2.

## Authentication

Leverages the default AWS credential chain (environment variables, shared credentials file, IAM roles, etc.). Use `profile` to target a specific shared config profile if needed.

## Build

```bash
go -C plugins/providers/awssecrets build
```

`make build` places the binary at `bin/providers/awssecrets`.

## Request Format

- **ref**: `<secret-id>#<metadata>` where metadata can be `stage:<name>`, `version:<id>`, or a bare version stage (for example, `prod/payments/api-key#stage:AWSCURRENT`).

## Configuration Options

- **region** *(optional)*: Region override for the AWS client.
- **profile** *(optional)*: Shared config profile to load.
- **version_id** *(optional)*: Explicit version ID (takes precedence over metadata).
- **version_stage** *(optional)*: Version stage (takes precedence over metadata).
- **timeout** *(optional)*: Request timeout (Go duration).

## Example

```yaml
secrets:
  PAYMENT_KEY:
    ref: prod/payments/api-key#stage:AWSPREVIOUS
    provider: awssecrets
    provider_options:
      region: us-east-1
      timeout: 3s
```
