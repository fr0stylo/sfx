# Provider Options Summary

This table lists the `provider_options` keys understood by each provider binary. Refer to the provider-specific directories for reference patterns and any additional notes.

| Provider | Option Keys |
|----------|-------------|
| `file` | `path` |
| `vault` | `address`, `token`, `namespace`, `field`, `timeout` |
| `sops` | `path`, `format`, `key_path` |
| `awssecrets` | `region`, `profile`, `version_id`, `version_stage`, `timeout` |
| `awsssm` | `region`, `profile`, `with_decryption`, `timeout` |
| `gcpsecrets` | `project`, `secret`, `version`, `timeout` |
| `azurevault` | `vault_url`, `vault_name`, `secret`, `version`, `timeout` |
