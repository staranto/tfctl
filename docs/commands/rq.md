# tfctl rq — run query

Synopsis

```
tfctl rq [RootDir] [options]
```

Short description

Query Runs for a workspace. Useful for examining the execution history and status of Terraform runs.

Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Attributes: [Attributes](../attrs.md)
- Filtering: [Filters](../filters.md)

Flags

| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | `.id,created-at,status` | Global flag |
| `--color` | | Enable colored text output | false | Use `--no-color` to disable |
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md) |
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped |
| `--limit` | `-l` | Limit runs returned | 99999 | Command-specific |
| `--org` | | Organization to query | (none) | Command-scoped |
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag |
| `--schema` | | Dump the schema | false | Command-specific helper |
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag |
| `--titles` | | Show titles with text output | false | Use `--no-titles` to disable |
| `--tldr` | | Show tldr page | false | Command-specific helper |
| `--workspace` | `-w` | Workspace to use for query | (none) | Command-scoped |

Quick examples

```
# List runs for the current workspace
tfctl rq

# Show common run attributes
tfctl rq --schema

# Show tldr page
tfctl rq --tldr

# Filter runs by status
tfctl rq --filter "status=applied"

# Limit results and include custom attributes
tfctl rq --limit 10 --attrs "created-at,status,message"
```

Notes

- Use `--workspace` to scope to a specific workspace when required.
- Use `--org` to specify the organization if not using the default.
- Use `--schema` to discover attributes available to `--attrs` for this command.

See also