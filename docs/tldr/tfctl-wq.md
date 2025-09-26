# tfctl-wq

> Query workspaces in HCP Terraform / Terraform Enterprise.
> More information: https://github.com/staranto/tfctlgo.

- List workspaces in an organization:

`tfctl wq --org {{organization_name}}`

- Output results as JSON:

`tfctl wq --org {{organization_name}} --output json`

- Filter by name (regex):

`tfctl wq --org {{organization_name}} --filter "name~{{pattern}}" --output json`

- Sort results by attribute:

`tfctl wq --org {{organization_name}} --sort {{name}}`

- Query a Terraform Enterprise host:

`tfctl wq --host {{tfe.example.com}} --org {{organization_name}}`