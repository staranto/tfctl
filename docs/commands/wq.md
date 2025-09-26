# tfctl wq — workspace query

Synopsis

```
tfctl wq [RootDir] [options]
```

Short description

Query Workspaces for an organization or project. Useful for inventorying workspaces and extracting attributes like VCS information and Terraform version.

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
| `--org` | `-o` | Organization to query | (none) | Command-scoped via `NewOrgFlag`
| `--output` | `-o` | Output format (`text`, `json`, `raw`) | `text` | Global flag
| `--schema` |  | Dump the schema | false | Command-specific helper
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--titles` | `-t` | Show titles with text output | false | Global flag

Quick examples

```
# List workspaces in the current org
 tfctl wq

# Show common workspace attributes
 tfctl wq --schema

# Examples provided by the command
 tfctl wq --examples
```

Notes

- Use `--org` to scope to a specific organization when required.
- Use `--schema` to discover attributes available to `--attrs` for this command.

See also

- [Quickstart](../quickstart.md)
