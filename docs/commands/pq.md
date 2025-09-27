# tfctl pq — project query

Synopsis

```
tfctl pq [RootDir] [options]
```

Short description

Query Projects within an organization. Use to enumerate and extract project-level metadata.

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
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag
| `--schema` |  | Dump the schema | false | Command helper
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--titles` | `-t` | Show titles with text output | false | Global flag
| `--tldr` |  | Show tldr page (if installed) | false | Hidden unless `tldr` present

Quick examples

```
# List projects in org
 tfctl pq --org my-org

# Show common project attributes
 tfctl pq --schema
```

Notes

- `pq` is useful for discovering projects that match naming patterns and for extracting VCS repo settings.

See also

- [Quickstart](../quickstart.md)
