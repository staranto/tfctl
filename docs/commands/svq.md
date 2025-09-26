# tfctl svq — state version query

Synopsis

```
tfctl svq [RootDir] [options]
```

Short description

List and inspect state versions for a workspace or state backend. Useful for audits and rollbacks.

Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Attributes: [Attributes](../attrs.md)
- Filtering: [Filters](../filters.md)

Flags

| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | (none) | Global flag |
| `--color` | `-c` | Enable colored text output | false | Global flag
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md)
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped via `NewHostFlag`
| `--limit` | `-l` | Limit state versions returned | 99999 | Command-scoped
| `--org` | `-o` | Organization to query | (none) | Command-scoped via `NewOrgFlag`
| `--output` | `-o` | Output format (`text`, `json`, `raw`) | `text` | Global flag
| `--schema` |  | Dump the schema | false | Command helper
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--titles` | `-t` | Show titles with text output | false | Global flag
| `--workspace` | `-w` | Workspace to use (overrides backend) | (none) | Command flag

Quick examples

```
# List state versions
 tfctl svq

# Limit number of versions returned
 tfctl svq --limit 10
```

Notes

- `svq` integrates with backends that support state versioning (remote/HCP/TFE).

See also

- [Quickstart](../quickstart.md)
