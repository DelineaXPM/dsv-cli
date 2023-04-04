# End-to-End testing

This detail focuses on the tests located in `tests/e2e`.

To run E2E tests you need to have a tenant and an admin access to it.

## Tenant Setup

### Create a User

```shell
dsv user create --username <username> --password <password>
```

### Grant Permissions

This E2E testing suite covers most of the commands and actions available in the CLI.
It is more convenient to add the newly created user to the list of admins of the tenant.
Please do not use production tenants for testing purposes since it can mess with your data.

If there is something that must be protected, create a policy that will guard that resource.
For example:

```bash
dsv policy create \
    --path "secrets:protected" \
    --effect "deny" \
    --actions "<.*>" \
    --subjects "users:<username>"
```

### Create certificate data

- Create a role for authentication by certificate

```shell
dsv role create --name e2e-cli-test-certauth
```

- Create root certificate

```shell
dsv pki generate-root \
    --rootcapath e2e-cli-test-root-for-auth \
    --common-name root.auth \
    --domains root.system.a,root.system.b \
    --maxttl 500000
```

- Create leaf certificate with role name in description. Use `"certificate"` and `"privateKey"` as env variables.

```shell
dsv pki leaf \
    --common-name root.system.a \
    --rootcapath e2e-cli-test-root-for-auth \
    --description e2e-cli-test-certauth \
    --ttl 500000
```

## Code Coverage

Coverage file is created for each execution and stored in the `tests/e2e/coverage` directory.

### Merge Into One File

If you don't need that granularity, you can merge coverage results into one file. To do that install "gocovmerge":

```shell
go install -v github.com/hansboder/gocovmerge@latest
```

And after it, run from the project root directory:

```shell
gocovmerge -dir ./tests/e2e/coverage -pattern "\.out" > ./coverage-e2e.out
```

### Inspect Results

```shell
go tool cover -html=./coverage-e2e.out
```
