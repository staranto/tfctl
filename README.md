[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/staranto/tfctlgo)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/staranto/tfctlgo?include_prereleases)](https://github.com/staranto/tfctlgo/releases)
[![CodeQL](https://github.com/staranto/tfctlgo/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/staranto/tfctlgo/actions/workflows/github-code-scanning/codeql)
[![Go Report Card](https://goreportcard.com/badge/github.com/staranto/tfctlgo)](https://goreportcard.com/report/github.com/staranto/tfctlgo)

# tfctl

> **Supercharge your Terraform workflow with powerful CLI queries**


**tfctl** is a command-line tool for querying Terraform and OpenTofu infrastructure. You can query workspaces, organizations, modules, and state data across multiple backend types from a single interface.

## Why tfctl?

The native Terraform CLI provides essential infrastructure-as-code capabilities, but it lacks powerful state querying tools and offers no way to query other elements of the Terraform ecosystem like workspaces, organizations, or module registries. This is especially problematic for automation use cases, where you need programmatic access to infrastructure metadata, state history, or cross-workspace insights.

**tfctl fills these gaps** by providing a unified, high-performance CLI for deep querying and analysis of Terraform-managed infrastructure, enabling better automation, reporting, and operational workflows.

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
| **`mq`** | Module query | `tfctl mq --filter 'name@aws'` |
| **`oq`** | Organization query | `tfctl oq --attrs email` |
| **`pq`** | Project query | `tfctl pq --sort created-at` |
| **`rq`** | Run query | `tfctl rq --attrs status` |
| **`si`** | Interactive state inspection | `tfctl si` |
| **`sq`** | State query | `tfctl sq --attrs arn --sort arn` |
| **`svq`** | State version query | `tfctl svq --limit 10` |
| **`wq`** | Workspace query | `tfctl wq --filter 'status@applied'` |

## Installation

For a quick start:

- Homebrew (recommended):
	```bash
	brew install staranto/tfctlgo/tfctl
	```
- Debian/Ubuntu (.deb):
	```bash
	curl -LO https://github.com/staranto/tfctlgo/releases/latest/download/tfctl_amd64.deb \
		&& sudo dpkg -i tfctl_amd64.deb || sudo apt-get -f install
	```
- Other methods (tarball, build from source), plus installing man and TLDR pages:
	see the full guide at [docs/installation.md](docs/installation.md).

<details>
<summary>Other installation methods</summary>

...existing code...
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
- **[Filter Expressions](docs/filters.md)** - Query syntax reference

## Roadmap

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

We sign release artifacts with GPG. To verify the integrity and authenticity of downloaded artifacts:

**Download and verify**
```bash
# Download the artifact and its signature
curl -L https://github.com/staranto/tfctlgo/releases/latest/download/tfctl_linux_amd64.tar.gz -o tfctl_linux_amd64.tar.gz
curl -L https://github.com/staranto/tfctlgo/releases/latest/download/tfctl_linux_amd64.tar.gz.sig -o tfctl_linux_amd64.tar.gz.sig

# Import the public key (one-time setup)
curl -L https://raw.githubusercontent.com/staranto/tfctlgo/master/KEYS | gpg --import

# Verify the signature
gpg --verify tfctl_linux_amd64.tar.gz.sig tfctl_linux_amd64.tar.gz
```

**Expected output**
```
gpg: Signature made [date] using RSA key [key-id]
gpg: Good signature from "tfctl Release Key"
```

If the signature verification fails or shows warnings, do not use the artifact and report the issue.

## Trademarks

- Terraform, Terraform Enterprise, and HCP Terraform are trademarks or registered trademarks of HashiCorp, Inc.
- OpenTofu is a trademark of The Linux Foundation.

Use of third-party names in this project is for identification and descriptive purposes only and does not imply endorsement, sponsorship, or affiliation.