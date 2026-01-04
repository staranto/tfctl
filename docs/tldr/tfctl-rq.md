# tfctl-rq

> Query Runs for a workspace. Useful for examining the execution history and status of Terraform runs.
> More information: https://github.com/staranto/tfctl.

- List runs for the current workspace:

`tfctl rq`

- Show common run attributes:

`tfctl rq --schema`

- Show tldr page:

`tfctl rq --tldr`

- Filter runs by status:

`tfctl rq --filter "status=applied"`

- Limit results and include custom attributes:

`tfctl rq --limit 10 --attrs "created-at,status,message"`
