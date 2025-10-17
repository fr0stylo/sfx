# Template Exporter

Renders secrets through a Go `text/template`, enabling custom output formats such as Kubernetes manifests, JSON, or shell scripts.

## Build

```bash
go -C plugins/exporters/template build
```

`make build` produces `bin/exporters/template`.

## Request Format

The template receives the following data structure:

```go
map[string]any{
    "Values": map[string]string, // secret values coerced to strings
    "Raw":    map[string][]byte,  // original byte slices
}
```

Sprig functions are available inside the template.

## Configuration Options

- **type**: Set to `template` in `.sfx.yaml`.
- **template** *(optional)*: Inline template content.
- **template_path** *(optional)*: Path to a template file (used when `template` is omitted).
- **delims.left / delims.right** *(optional)*: Override template delimiters.

Exactly one of `template` or `template_path` must be provided.

## Example (inline template)

```yaml
output:
  type: template
  options:
    template: |
      API_TOKEN={{ index .Values "api_token" }}
      DB_PASSWORD={{ index .Values "db_password" | quote }}
```

## Example (template file)

```yaml
output:
  type: template
  options:
    template_path: ./templates/.env.tmpl
```
