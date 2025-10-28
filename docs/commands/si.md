# tfctl si — state inspector

Synopsis

```
tfctl si [RootDir]
```

Short description

Start an interactive console to explore state data (search resources, run ad-hoc queries, and view results inline).

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
| `--output` | `-o` | Output format (`text`, `json`, `yaml`, `raw`) | `text` | Global flag |
| `--passphrase` | `-p` | Passphrase for encrypted state files | (none) | si-specific |
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag |
| `--sv` | | State version to query | current | si-specific |
| `--titles` | | Show titles with text output | false | Use `--no-titles` to disable |

Quick examples

```
# Start interactive console in CWD
tfctl si

# Exit the console
Type `exit` or press Ctrl+C

# Start console with specific state version
tfctl si --sv 5

# Query with filters
tfctl si --filter "type=aws_instance"
```

Notes

- `si` uses a terminal UI and stores history in `~/.tfctl_si_history`.
- For scripted/extractable output, prefer `sq --output json`.
- Use `--passphrase` or `TF_VAR_passphrase` for encrypted state files.

See also

- [Quickstart](../quickstart.md)

