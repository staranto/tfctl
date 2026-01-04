# tfctl-sq

> Query Terraform/OpenTofu state files in a local IaC root. Supports listing resources, filtering, and comparing state versions.
> More information: https://github.com/staranto/tfctl.

- Query the current directory's state:

`tfctl sq`

- Find resources with Hungarian notation naming convention:

`tfctl sq --filter hungarian=true`

- See state-specific flags (e.g., --concrete, --diff):

`tfctl sq --help`
