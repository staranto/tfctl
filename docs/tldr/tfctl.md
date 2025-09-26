# tfctl

> Command-line tool for querying Terraform and OpenTofu infrastructure across multiple backend types.
> More information: https://github.com/staranto/tfctlgo.

> Related pages: [tfctl-attrs](./tfctl-attrs.md), [tfctl-filters](./tfctl-filters.md), [tfctl-flags](./tfctl-flags.md), [tfctl-mq](./tfctl-mq.md), [tfctl-oq](./tfctl-oq.md), [tfctl-pq](./tfctl-pq.md), [tfctl-rq](./tfctl-rq.md), [tfctl-sq](./tfctl-sq.md), [tfctl-svq](./tfctl-svq.md), [tfctl-wq](./tfctl-wq.md)


- Search modules in registry:

`tfctl mq --org {{organization_name}} --filter "name={{module_name}}"`

- Query organizations with filtering:

`tfctl oq --filter "name~{{pattern}}" --output json`

- List resources from workspace state:

`tfctl sq --workspace {{workspace_name}}`

- Query state from local directory:

`tfctl sq {{path/to/terraform/directory}}`

- Filter results by resource type:

`tfctl sq --workspace {{workspace_name}} --filter "type=aws_instance" --output json`

- Query workspaces in an organization:

`tfctl wq --org {{organization_name}}`