# TFVARS Exporter

Formats secrets as Terraform variable assignments.

## Build

```bash
go -C plugins/exporters/tfvars build
```

`make build` produces `bin/exporters/tfvars`.

## Request Format

Receives the full secret map; values are rendered as strings unless they parse cleanly as booleans or numbers. Multi-line strings are emitted using heredoc syntax.

## Configuration Options

- **type**: Set to `tfvars` in the `.sfx.yaml` output section.
- **order** *(optional)*: Array specifying the key order; unspecified keys are appended alphabetically.

## Example

```yaml
output:
  type: tfvars
  options:
    order: [db_password, api_key]
```
