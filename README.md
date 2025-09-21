# tfctl

> **Supercharge your Terraform workflow with powerful CLI queries**

[![Go Version](https://img.shields.io/github/go-mod/go-version/staranto/tfctlgo)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/staranto/tfctlgo-release?include_prereleases)](https://github.com/staranto/tfctlgo-release/releases)

**tfctl** is a command-line tool for querying Terraform and OpenTofu infrastructure. Query workspaces, organizations, modules, and state data across multiple backend types from a single interface.

## Key Features

**Cross-Backend Support** - Works with HCP Terraform, Terraform Enterprise, local state files, S3 backends, and module registries

**Fast Performance** - Built in Go with concurrent operations and intelligent caching

**Flexible Output** - Filter, sort, and format results as JSON, YAML, or formatted tables

**Secure** - Supports encrypted state files and multiple authentication methods

**Comprehensive** - Query any attribute available through the Terraform APIs

## Common Examples

```bash
# Find all workspaces containing "prod" across your organization
tfctl wq --filter 'name@prod'

# Compare state versions to see what changed
tfctl sq --diff

# List modules by popularity across registries
tfctl mq --sort downloads --reverse

# Export workspace data for reporting
tfctl wq --attrs created-at,updated-at --output json
```

## Available Commands

| Command | Purpose | Example |
|---------|---------|---------|
| **`mq`** | Module Registry queries | `tfctl mq --filter 'name@aws'` |
| **`oq`** | Organization management | `tfctl oq --attrs email` |
| **`pq`** | Project oversight | `tfctl pq --sort created-at` |
| **`sq`** | State file analysis | `tfctl sq --diff` |
| **`svq`** | State version history | `tfctl svq --limit 10` |
| **`wq`** | Workspace operations | `tfctl wq --filter 'status@applied'` |

## Installation

### Install tfctl

**Download pre-built binary (recommended)**
```bash
# Download for your platform from releases
curl -L https://github.com/staranto/tfctlgo-release/releases/latest/download/tfctl_linux_amd64.tar.gz | tar xz
sudo mv tfctl /usr/local/bin/
```

<details>
<summary>Other installation methods</summary>

**Build from source**
```bash
git clone https://github.com/staranto/tfctlgo.git
cd tfctlgo && go build -o tfctl
```

**Build with GoReleaser**
```bash
goreleaser build --snapshot --clean --single-target
```
</details>

### Setup Authentication

For HCP Terraform / Terraform Enterprise:
```bash
# Generate tokens at: https://app.terraform.io/app/settings/tokens
export TFE_TOKEN="your-hcp-token-here"
```

### First Steps

```bash
# List all your workspaces
tfctl wq

# Get help anytime
tfctl --help
tfctl sq --help --examples
```

## Documentation

- **[Quick Start Tutorial](docs/quickstart.md)** - Detailed walkthrough with examples
- **[Command Reference](docs/flags.md)** - Complete flag documentation
- **[Attribute Guide](docs/attributes.md)** - Advanced filtering techniques
- **[Filter Expressions](docs/filters.md)** - Query syntax reference## Roadmap

**tfctl** is currently read-only and focused on querying. Version 1.x provides stable query functionality across all major Terraform backends.

**Planned features:**
- Workspace and state manipulation
- Enhanced S3 backend configuration options
- Real-time state monitoring
- Advanced reporting and dashboards

*Want a feature? [Open an issue](https://github.com/staranto/tfctlgo/issues) and help us prioritize!*

## Contributing

Contributions are welcome! Whether it's:
- Bug reports and fixes
- Feature requests and implementations
- Documentation improvements
- Ideas and feedback

**Get started:** Fork the repo, make your changes, and submit a PR. Check out our [issues](https://github.com/staranto/tfctlgo/issues) for good first contributions.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

*Questions? Contact: [staranto@gmail.com](mailto:staranto@gmail.com)*

## Verify releases

We sign release artifacts with Sigstore `cosign`. Two common verification methods:

- Keyless (recommended): verify signatures published to the transparency log:

```bash
cosign verify-blob --keyless path/to/artifact
```

- Key-based (if a `cosign.pub` is provided with the release):

```bash
curl -L https://github.com/staranto/tfctlgo-release/releases/latest/download/cosign.pub -o cosign.pub
cosign verify-blob --key cosign.pub path/to/artifact
```

If a `cosign.pub` file is published with the release, it will be available alongside other assets on the release page.