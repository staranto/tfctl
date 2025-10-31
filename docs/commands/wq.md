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
| `--color` | | Enable colored text output | false | Use `--no-color` to disable |
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md)
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped |
| `--limit` | `-l` | Limit workspaces returned | 99999 | Command-specific |
| `--org` | | Organization to query | (none) | Command-scoped |
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag |
| `--schema` | | Dump the schema | false | Command-specific helper |
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag |
| `--titles` | | Show titles with text output | false | Use `--no-titles` to disable |
| `--tldr` | | Show tldr page | false | Command-specific helper

Quick examples

```
# List workspaces in the current org
tfctl wq

# Show common workspace attributes
tfctl wq --schema

# Show tldr page
tfctl wq --tldr

# Filter workspaces by name
tfctl wq --filter "name=production"

# Limit results and include custom attributes
tfctl wq --limit 10 --attrs "name,.vcs_repo"
```

Notes

- Use `--org` to scope to a specific organization when required.
- Use `--schema` to discover attributes available to `--attrs` for this command.

See also

- [Quickstart](../quickstart.md)
