# tfctl-pq

> Query projects in HCP Terraform / Terraform Enterprise.
> More information: https://github.com/staranto/tfctlgo.

- List projects in an organization:

`tfctl pq --org {{organization_name}}`

- Filter projects by name (contains):

`tfctl pq --org {{organization_name}} --filter "name@{{text}}" --output json`

- Include last updated timestamp:

`tfctl pq --org {{organization_name}} --attrs {{updated-at}} --output json`

- Include VCS repo information:

`tfctl pq --org {{organization_name}} --attrs {{vcs-repo.identifier}} --output json`

- Use a specific Terraform Enterprise host:

`tfctl pq --host {{tfe.example.com}} --org {{organization_name}}`