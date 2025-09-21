# tfctl mq

> Query a Terraform module registry
> More information: <https://github.com/staranto/tfctlgo>.

- Show all Modules in the default Organization.

`tfctl mq`

- Show all Modules in the 'production' Organization.

`tfctl mq --org {{production}}`

- Show Modules with 'storage' in their name.

`tfctl mq --filter {{'name@storage'}}`

- Show all Modules with their Last Updated timestamp.

`tfctl mq --attrs {{updated-at}}`
