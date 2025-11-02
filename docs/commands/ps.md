# tfctl ps — plan summary

## Synopsis

```
tfctl ps [plan-file]
```

## Short description

Parse and summarize Terraform plan output, displaying resource actions in a formatted table. Accepts plan files or reads from stdin. Supports filtering and sorting to narrow down changes.

## Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Filtering: [Filters](../filters.md)

## Flags

| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--color` | | Enable colored text output | false | Use `--no-color` to disable |
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md) |
| `--output` | `-o` | Output format (`text`, `json`, `yaml`) | `text` | Global flag |
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag |
| `--titles` | | Show titles with text output | false | Use `--no-titles` to disable |

## Quick examples

```bash
# Parse a plan file and display summary
tfctl ps plan.out

# Read from stdin
terraform plan -out=tfplan && terraform show -json tfplan | tfctl ps

# Filter for only destroyed resources
tfctl ps plan.txt --filter 'action=destroyed'

# Filter for created or changed resources
tfctl ps --filter 'action@created,action@changed'

# JSON output for programmatic access
tfctl ps plan.txt --output json

# Sort by action type
tfctl ps --sort action
```

## Notes

- `ps` reads from stdin by default. Pass a filename to read from a file instead.
- The tool parses resource action lines in the format: `# <resource-path> will/must be <action>`
- ANSI color codes (from colored terminal output) are automatically stripped before parsing.
- Supported actions: `created`, `destroyed`, `updated`, `replaced`, `updated in-place`, and others from Terraform's plan output.
- The `--filter` flag works with the `action` and `resource` attributes. See [Filters](../filters.md) for filter syntax.
- Use `--filter 'action=created'` to focus on new resources being added.

## Attributes

The `ps` command exposes the following attributes for filtering and sorting:

- **`resource`** — The resource path (e.g., `aws_instance.example`, `module.network.aws_vpc.main`)
- **`action`** — The action to be taken (e.g., `created`, `destroyed`, `updated`, `replaced`)

## See also

- [Quickstart](../quickstart.md)
- [Filters](../filters.md)
