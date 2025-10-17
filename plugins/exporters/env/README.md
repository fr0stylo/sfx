# ENV Exporter Configuration

- **type**: Must be set to `env`.
- **key_template** *(optional)*: Go template (Sprig enabled) evaluated for each secret name; defaults to `{{ .Value | upper }}`.
