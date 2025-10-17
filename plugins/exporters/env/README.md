# ENV Exporter

Produces `.env`-style key/value output from the collected secrets.

## Build

```bash
go -C plugins/exporters/env build
```

`make build` emits `bin/exporters/env`.

## Rendering Rules

- Values are sorted lexicographically by key.
- Keys are run through a configurable Go text/template.
- Values containing whitespace or shell-sensitive characters are quoted; empty values render as `""`.

## Configuration Options

- **type**: Set to `env` in `.sfx.yaml`.
- **key_template** *(optional)*: Go template (Sprig-enabled) applied to each key. Receives `TemplateValue{Value: <secret name>}` and defaults to `{{ .Value | upper }}`.

## Example

```yaml
output:
  type: env
  options:
    key_template: "{{ .Value | replace \"-\" \"_\" | upper }}"
```
