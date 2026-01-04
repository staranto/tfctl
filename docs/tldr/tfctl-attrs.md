# tfctl-attrs

> Select and transform fields in output using `--attrs`.
> More information: https://github.com/staranto/tfctl/docs/attrs.md.

- Select specific attributes:

`tfctl oq --attrs {{name,email,created-at}}`

- Use root-level attributes (prefix with a dot):

`tfctl oq --attrs {{.id,.type}}`

- Rename output columns:

`tfctl oq --attrs {{created-at:Created,email:Admin}}`

- Apply transformations (uppercase, lowercase, length, time):

`tfctl oq --attrs {{name::U,created-at::t,email::L10}}`

- Combine with filtering and JSON output:

`tfctl wq --filter 'name@{{prod}}' --attrs {{name:Workspace,working-directory:Path}} --output json`