# CLI End-to-End testing

`[first draft of the E2E related docs]`

To run E2E tests you need to have a tenant and an admin access to it.

## Tenant Setup

### Create a User

```
dsv user create --username <username> --password <password>
```

### Grant Permissions

Create policies which will allow actions required for testing to the created user.

- Create, Read, Update and Delete Roles with "e2e-cli-test" prefix

```
dsv policy create \
    --path "roles:e2e-cli-test" \
    --resources "roles:e2e-cli-test<.*>" \
    --effect allow \
    --actions "<create|read|update|delete>" \
    --subjects users:<username>
```

- List all Roles

```
dsv policy create \
    --path "roles" \
    --resources "roles" \
    --effect allow \
    --actions "list" \
    --subjects users:<username>
```

- Create, Read and Delete Pools with "e2e-cli-test" prefix

```
dsv policy create \
    --path "pools:e2e-cli-test" \
    --resources "pools:e2e-cli-test<.*>" \
    --effect allow \
    --actions "<create|read|delete>" \
    --subjects users:<username>
```

- List all Pools

```
dsv policy create \
    --path "pools" \
    --resources "pools" \
    --effect allow \
    --actions "list" \
    --subjects users:<username>
```

- Create, Read and Delete Engines with "e2e-cli-test" prefix

```
dsv policy create \
    --path "engines:e2e-cli-test" \
    --resources "engines:e2e-cli-test<.*>" \
    --effect allow \
    --actions "<create|read|delete>" \
    --subjects users:<username>
```

- List all Engines

```
dsv policy create \
    --path "engines" \
    --resources "engines" \
    --effect allow \
    --actions "list" \
    --subjects users:<username>
```

- Create, Read, Update and Delete SIEMs with "e2e-cli-test" prefix

```
dsv policy create \
    --path "config:siem:e2e-cli-test" \
    --effect allow \
    --actions "<create|delete|read|update>" \
    --resources "config:siem:e2e-cli-test<.*>" \
    --subjects users:<username>
```

### Create certificate data

- Create a role for authentication by certificate

```
dsv role create --name e2e-cli-test-certauth
```

- Create root certificate

```
dsv pki generate-root \
    --rootcapath e2e-cli-test-root-for-auth \
    --common-name root.auth \
    --domains root.system.a,root.system.b \
    --maxttl 500000
```

- Create leaf certificate with role name in description. Use `"certificate"` and `"privateKey"` as env variables.

```
dsv pki leaf \
    --common-name root.system.a \
    --rootcapath e2e-cli-test-root-for-auth \
    --description e2e-cli-test-certauth \
    --ttl 500000
```

## Code Coverage

Coverage file is created for each execution and stored in the `<project root>/tests/e2e/coverage` directory.

### Merge Into One File

If you don't need that granularity, you can merge coverage results into one file. To do that install "gocovmerge":

```
go install -v github.com/hansboder/gocovmerge@latest
```

And after it, run from the project root directory:

```
gocovmerge -dir ./tests/e2e/coverage -pattern "\.out" > ./coverage-e2e.out
```

### Inspect Results

```
go tool cover -html=./coverage-e2e.out
```
