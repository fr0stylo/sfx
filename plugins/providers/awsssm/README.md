# AWS SSM Parameter Store Provider

Reads SecureString or String parameters from AWS Systems Manager Parameter Store.

## Authentication

Uses the standard AWS credential chain. Specify a profile if you want to source credentials from a named entry in `~/.aws/credentials` or `~/.aws/config`.

## Build

```bash
go -C plugins/providers/awsssm build
```

`make build` creates `bin/providers/awsssm`.

## Request Format

- **ref**: Full parameter name (for example, `/prod/payments/db/password`).

## Configuration Options

- **region** *(optional)*: Region override for the SSM client.
- **profile** *(optional)*: Shared config profile name.
- **with_decryption** *(optional, bool)*: Set to `false` to receive encrypted SecureString values. Defaults to `true`.
- **timeout** *(optional)*: Request timeout (Go duration).

## Example

```yaml
secrets:
  DB_PASSWORD:
    ref: /prod/payments/db/password
    provider: awsssm
    provider_options:
      region: us-west-2
      with_decryption: true
```
