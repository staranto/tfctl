# tfctl oq — organization query

Synopsis

```
tfctl oq [RootDir] [options]
```

Short description

Query Organizations from your configured Terraform/HCP/TFE host.

Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Attributes: [Attributes](../attrs.md)
- Filtering: [Filters](../filters.md)

Flags

| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | (none) | Global flag |
| `--color` | `-c` | Enable colored text output | false | Global flag, configurable via config
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md)
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped via `NewHostFlag`
| `--output` | `-o` | Output format (`text`, `json`, `raw`) | `text` | Global flag
| `--schema` |  | Dump the schema | false | Command-specific helper
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--titles` | `-t` | Show titles with text output | false | Global flag
| `--tldr` |  | Show tldr page (if available) | false | Hidden unless `tldr` installed

Quick examples

```
# List organizations
 tfctl oq

# Show common attributes
 tfctl oq --schema

# Examples provided by the command
 tfctl oq --examples
```

Notes

- Use `--schema` to discover attributes available to `--attrs` for this command.
- Use `--examples` to view curated usage examples.
- For automation, prefer `--output json`.

See also

- [Quickstart](../quickstart.md)
