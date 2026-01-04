# tfctl-flags

> Common flags shared across tfctl commands.
> More information: https://github.com/staranto/tfctl/docs/flags.md.

- Set output format to JSON:

`tfctl oq --output json`

- Add attribute columns to results:

`tfctl wq --attrs {{name,terraform-version}}`

- Filter results by attribute values:

`tfctl oq --filter 'name@{{prod}}'`

- Specify HCP/TFE host and organization:

`tfctl wq --host {{tfe.example.com}} --org {{organization_name}}`

- Show column titles in text mode:

`tfctl wq --titles`