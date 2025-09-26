# tfctl-svq

> Query state versions for a workspace.
> More information: https://github.com/staranto/tfctlgo.

- List state versions for a workspace:

`tfctl svq --workspace {{workspace_name}}`

- Limit the number of state versions returned:

`tfctl svq --workspace {{workspace_name}} --limit {{10}}`

- Output results as JSON:

`tfctl svq --workspace {{workspace_name}} --output json`

- Use a specific Terraform Enterprise host:

`tfctl svq --host {{tfe.example.com}} --org {{organization_name}} --workspace {{workspace_name}}`