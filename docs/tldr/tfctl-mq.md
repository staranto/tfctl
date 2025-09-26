# tfctl-mq

> Query modules in the HCP Terraform / Terraform Enterprise private registry.
> More information: https://github.com/staranto/tfctlgo.

- List modules in an organization:

`tfctl mq --org {{organization_name}}`

- Filter modules by name (contains):

`tfctl mq --org {{organization_name}} --filter "name@{{text}}" --output json`

- Filter modules by name (regex):

`tfctl mq --org {{organization_name}} --filter "name~{{pattern}}" --output json`

- Use a specific Terraform Enterprise host:

`tfctl mq --host {{tfe.example.com}} --org {{organization_name}}`