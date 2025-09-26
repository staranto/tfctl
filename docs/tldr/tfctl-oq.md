# tfctl-oq

> Query organizations in HCP Terraform / Terraform Enterprise.
> More information: https://github.com/staranto/tfctlgo.

- List organizations on the default server:

`tfctl oq`

- List organizations on a specific Terraform Enterprise host:

`tfctl oq --host {{tfe.example.com}}`

- Filter organizations by name (contains):

`tfctl oq --filter "name@{{text}}"`

- Return results as JSON:

`tfctl oq --output json`

- Include additional attributes in the output:

`tfctl oq --attrs {{email}} --output json`

- Sort organizations by an attribute:

`tfctl oq --sort {{name}}`