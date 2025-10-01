# tfctl mq — module registry query

Synopsis

```
tfctl mq [RootDir] [options]
```

Short description

Query the Module Registry (HCP/TFE) for available modules and metadata.

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
# List registry modules
 tfctl mq  --sort createdAt

# Show common module attributes
 tfctl mq --schema
```

Notes

- `mq` may surface VCS and provider metadata for modules; use `--attrs` to extract specific values.

See also

- [Quickstart](../quickstart.md)
