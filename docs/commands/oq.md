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
| `--color` | | Enable colored text output | false | Use `--no-color` to disable |
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md) |
| `--host` | `-h` | Host to use for queries | `app.terraform.io` | Command-scoped |
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag |
| `--schema` | | Dump the schema | false | Command-specific helper |
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag |
| `--titles` | | Show titles with text output | false | Use `--no-titles` to disable |
| `--tldr` | | Show tldr page | false | Command-specific helper |

Quick examples

```
# List organizations
tfctl oq

# Show common attributes
tfctl oq --schema

# Show tldr page
tfctl oq --tldr

# Filter organizations by name
tfctl oq --filter "name=my-org"

# Output as JSON
tfctl oq --output json
```

Notes

- Use `--schema` to discover attributes available to `--attrs` for this command.
- For automation, prefer `--output json`.

See also

- [Quickstart](../quickstart.md)
