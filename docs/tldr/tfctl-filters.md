# tfctl-filters

> Filter query results with `--filter` using a compact syntax.
> More information: https://github.com/staranto/tfctlgo/docs/filters.md.

- Contains and negation:

`tfctl oq --filter 'name@{{prod}}'`
`tfctl oq --filter 'name!@{{prod}}'`

- Exact, starts-with, and case-insensitive:

`tfctl wq --filter 'status={{applied}}'`
`tfctl wq --filter 'name^{{prod}}'`
`tfctl oq --filter 'email~{{admin}}'`

- Regular expression:

`tfctl wq --filter 'name/^{{prod-\d{3}}}$'`

- Multiple filters (comma-delimited):

`tfctl oq --filter 'name@{{prod}},created-at>{{2024-01-01}}'`

- With JSON output:

`tfctl sq --workspace {{workspace_name}} --filter 'type={{aws_instance}}' --output json`