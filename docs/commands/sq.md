# tfctl sq — state query

Synopsis

```
tfctl sq [RootDir] [options]
```

Short description

Query Terraform/OpenTofu state files in a local IaC root. Supports listing resources, filtering, and comparing state versions.

Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Attributes: [Attributes](../attrs.md)
- Filtering: [Filters](../filters.md)

Flags

| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | (none) | Global flag |
| `--color` | `-c` | Enable colored text output | false | Global flag
| `--concrete` | `-k` | Only include concrete (managed) resources | false | sq-specific
| `--diff` |  | Show diff between state versions | false | sq-specific
| `--diff_filter` |  | Hidden, internal diff filter | `check_results` | Hidden / internal
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md)
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped via `NewHostFlag`
| `--limit` |  | Limit state versions returned | 99999 | sq-specific, hidden default
| `--noshort` |  | Include full resource name paths | false | sq-specific
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag
| `--org` | `-o` | Organization to query | (none) | Command-scoped via `NewOrgFlag`
| `--passphrase` |  | Passphrase for encrypted state | (none) | sq-specific; falls back to TF_VAR_passphrase or interactive prompt
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--sv` |  | State version to query | `0` | sq-specific
| `--titles` | `-t` | Show titles with text output | false | Global flag
| `--workspace` | `-w` | Workspace to use (overrides backend) | (none) | Command flag

Quick examples

```
# Query the current directory's state
 tfctl sq

# See state-specific flags (e.g., --concrete, --diff)
 tfctl sq --help
```

Notes

- `sq` operates against an IaC root directory (defaults to CWD when not provided).
- When using encrypted state, `sq` will prompt for a passphrase or use `TF_VAR_passphrase`.

See also

- [Quickstart](../quickstart.md)
