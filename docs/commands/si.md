# tfctl si — interactive state console

Synopsis

```
tfctl si [RootDir]
```

Short description

Start an interactive console to explore state data (search resources, run ad-hoc queries, and view results inline).

Flags and related docs

- See the common flag reference: [Flags](../flags.md)
- Attributes: [Attributes](../attrs.md)

Quick examples

```
# Start interactive console in CWD
 tfctl si

# Exit the console
 Type `exit` or press Ctrl+C
```

Notes

- `si` uses a terminal UI and stores history in `~/.tfctl_si_history`.
- For scripted/extractable output, prefer `sq --output json`.

See also

- [Quickstart](../quickstart.md)

Flags

| Flag | Alias | Description | Notes |
|------|-------|-------------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | Global flag |
| `--color` | `-c` | Enable colored text output | Global flag
| `--filter` | `-f` | Comma-separated list of filters to apply | See [Filters](../filters.md)
| `--output` | `-o` | Output format (`text`, `json`, `raw`) | Global flag
| `--sort` | `-s` | Attributes to sort by | Global flag
| `--titles` | `-t` | Show titles with text output | Global flag
| Flag | Alias | Description | Default | Notes |
|------|-------|-------------|---------|-------|
| `--attrs` | `-a` | Comma-separated list of attributes to include | (none) | Global flag |
| `--color` | `-c` | Enable colored text output | false | Global flag
| `--filter` | `-f` | Comma-separated list of filters to apply | (none) | See [Filters](../filters.md)
| `--output` | `-o` | Output format (`text`, `json`, `raw`) | `text` | Global flag
| `--passphrase` | `-p` | Passphrase for encrypted state files | (none) | si-specific
| `--sort` | `-s` | Attributes to sort by | (none) | Global flag
| `--sv` |  | State version to query | `0` | si-specific
| `--titles` | `-t` | Show titles with text output | false | Global flag
