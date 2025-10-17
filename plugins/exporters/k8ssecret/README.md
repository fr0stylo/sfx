# Kubernetes Secret Exporter

Produces a Kubernetes `Secret` manifest containing the collected secrets.

## Build

```bash
go -C plugins/exporters/k8ssecret build
```

`make build` emits `bin/exporters/k8ssecret`.

## Manifest Layout

- `apiVersion: v1`, `kind: Secret`.
- `metadata` populated from configuration (name is required).
- `data` holds base64-encoded secret values sorted by key.

## Configuration Options

- **type**: Set to `k8ssecret`.
- **name** *(required)*: Secret name.
- **namespace** *(optional)*: Namespace for the Secret.
- **type** *(optional)*: Secret type (e.g. `Opaque`, `kubernetes.io/dockerconfigjson`).
- **labels** *(optional)*: Map of labels.
- **annotations** *(optional)*: Map of annotations.

## Example

```yaml
output:
  type: k8ssecret
  options:
    name: app-secrets
    namespace: prod
    type: Opaque
    labels:
      app: payments
```
