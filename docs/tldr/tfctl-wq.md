# tfctl-wq

> Query Workspaces for an organization or project. Useful for inventorying workspaces and extracting attributes like VCS information and Terraform version.
> More information: https://github.com/staranto/tfctl.

- List workspaces in the current org:

`tfctl wq`

- Show common workspace attributes:

`tfctl wq --schema`

- Show tldr page:

`tfctl wq --tldr`

- Filter workspaces by name:

`tfctl wq --filter "name=production"`

- Limit results and include custom attributes:

`tfctl wq --limit 10 --attrs "name,.vcs_repo"`
