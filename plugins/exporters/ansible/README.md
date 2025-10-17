# Ansible Exporter

Renders secrets as a YAML mapping suitable for Ansible variable files.

## Build

```bash
go -C plugins/exporters/ansible build
```

`make build` creates `bin/exporters/ansible`.

## Output Format

- Values are emitted as plain strings.
- Keys can be reordered or prefixed to match Ansible naming conventions.

## Configuration Options

- **type**: Set to `ansible`.
- **prefix** *(optional)*: String prepended to each key (for example, `secret_`).
- **order** *(optional)*: Array specifying key order; unspecified keys are appended alphabetically.

## Example

```yaml
output:
  type: ansible
  options:
    prefix: secret_
    order: [db_password, api_token]
```
