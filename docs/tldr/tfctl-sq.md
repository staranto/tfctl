# tfctl-sq

> Query Terraform state resources from a workspace or local directory.
> More information: https://github.com/staranto/tfctlgo.

- Query state from a workspace:

`tfctl sq --workspace {{workspace_name}}`

- Query state from a local Terraform directory (RootDir):

`tfctl sq {{path/to/terraform/directory}}`

- Only include concrete resources:

`tfctl sq --workspace {{workspace_name}} --concrete`

- Filter by resource type and output as JSON:

`tfctl sq --workspace {{workspace_name}} --filter "type=aws_instance" --output json`

- Show full resource name paths:

`tfctl sq --workspace {{workspace_name}} --noshort`