# Azure Key Vault Provider

Retrieves secrets from Azure Key Vault using the Azure Identity `DefaultAzureCredential` chain.

## Authentication

Configure one of the credential sources supported by `DefaultAzureCredential` (environment variables, managed identity, Visual Studio/VS Code login, etc.). For local development, exporting `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, and `AZURE_CLIENT_SECRET` is a common approach.

## Build

```bash
go -C plugins/providers/azurevault build
```

`make build` outputs `bin/providers/azurevault`.

## Request Format

- **ref**: One of the following:
  - `https://<vault>.vault.azure.net/secrets/<secret>/<version>`
  - `<vault-name>/<secret>#<version>`
  - `#<version>` when the secret name is provided in configuration

## Configuration Options

- **vault_url** *(optional)*: Fully qualified vault URL when the reference omits it.
- **vault_name** *(optional)*: Vault name used to derive the URL when `vault_url` is absent.
- **secret** *(optional)*: Secret name when the reference omits it.
- **version** *(optional)*: Version identifier (defaults to latest).
- **timeout** *(optional)*: Request timeout (Go duration).

## Example

```yaml
secrets:
  STORAGE_CONN_STRING:
    ref: finance-vault/storage-conn#latest
    provider: azurevault
    provider_options:
      timeout: 5s
```
