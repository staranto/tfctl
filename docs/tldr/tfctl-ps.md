# tfctl-ps

> Parse and summarize Terraform plan output, displaying resource actions in a formatted table. Accepts plan files or reads from stdin. Supports filtering and sorting to narrow down changes.
> More information: https://github.com/staranto/tfctlgo.

- Example:

`bash`

- Parse a plan file and display summary:

`tfctl ps plan.out`

- Read from stdin:

`terraform plan -out=tfplan && terraform show -json tfplan | tfctl ps`

- Filter for only destroyed resources:

`tfctl ps plan.txt --filter 'action=destroyed'`

- Filter for created or changed resources:

`tfctl ps --filter 'action@created,action@changed'`

- JSON output for programmatic access:

`tfctl ps plan.txt --output json`

- Sort by action type:

`tfctl ps --sort action`
