# tfctl-rq

> Query runs for a workspace.
> More information: https://github.com/staranto/tfctlgo.

- List runs for a workspace:

`tfctl rq --workspace {{workspace_name}}`

- Limit the number of runs returned:

`tfctl rq --workspace {{workspace_name}} --limit {{20}}`

- Output results as JSON:

`tfctl rq --workspace {{workspace_name}} --output json`

- Use a specific Terraform Enterprise host:

`tfctl rq --host {{tfe.example.com}} --org {{organization_name}} --workspace {{workspace_name}}`