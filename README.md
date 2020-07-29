![Go](https://github.com/thycotic/dsv-cli/workflows/Go/badge.svg)
# Thycotic DevOps Secrets Vault CLI

Thycotic DevOps Secrets Vault CLI is an automation tool for the management and access of secret information.

## CLI Usage Documentation:
Current documentation on the CLI: http://docs.thycotic.com/dsv

Docs are hosted in https://github.com/thycotic/dsv-docs

To request a docs change fork and submit a PR.

## Setting up dev environment

* The makefile, and cicd-integration tests only run on linux. If developing on Windows install the WSL
* Also make sure that you are not converting line endings to CRLF automatically. The .gitattributes file should prevent committing CRLF line endings
but you might need to run some git commands to make sure everything is set to LF locally
	* Turn off autocrlf: `git config --global --unset core.autocrlf`
	* Re-checkout: `git checkout-index --force --all`
	* If that fails do: `git config core.eol lf` and re-checkout again or delete and reclone the repo.


## Installation

Installation is only required for autocomplete. Autocomplete is supported on bash, zsh, and fish shells via https://github.com/posener/complete.

To install:
```
go get -u github.com/posener/complete/gocomplete
gocomplete -install
thy -install
```
To uninstall:
```
thy -uninstall
gocomplete -uninstall
```

## Structure
Commands:
* secret [describe|read|create|update|delete]

```commandline
Flags  
   --auth-client-id            Client ID used for auth  (if AuthType='clientcred')  
   --auth-client-secret        Client Secret for auth (if AuthType='clientcred')  
   --auth-password, -p         Password used for auth  
   --auth-type, -a             Authentication type [default:password]  
   --auth-username, -u         User used for auth  
   --beautify, -b              Should beautify output  
   --config, -c                Config file path [default:%USERPROFILE%\.thy.yaml]  
   --encoding, -e              Output encoding (json|yaml) [default:json]  
   --filter, -f                Filter in jq (stedolan.github.io/jq)  
   --out, -o                   Output destination (stdout|clip|file:<fname>) [default:stdout]  
   --profile                   Configuration Profile [default:default]  
   --tenant, -t                Tenant used for auth  
   --verbose, -v               Verbose output [default:false]  
```


Config: Configuration is read by default from ~/.thy.yml. Alternatively, username, password and tenant may be specified as env variables or in the command line command


## Examples

Create a secret at the path resources/us-east-1/server1
```bash


thy secret create --path resources/us-east-1/server1 --data "{
                                                                 "data": {"foo":"test"},
                                                                 "description": "foo secret",
                                                                 "attributes": {
                                                                 	"tag1": "val1"
                                                                 }
                                                             }"
```
Or, as a shortcut for secret commands, the first two arguments are interpereted as --path and --data. Additional flags should follow
```bash

thy secret update resources/us-east-1/server1  --data "{
    "data": {"foo":"test"},
    "description": "foo secret",
    "attributes": {
    	"tag1": "val2"
    }
}"
```
Read a secret field
```bash
thy secret read resources/us-east-1/server1 -bf .data.foo
```



## Built With

* [Mitchellh / cli](https://github.com/thycotic-rd/cli)


## Developing DSV CLI
If you want to contribute or trying out for yourself, you will first need GO installed on your development environment. At least Go Version 1.12 is required.  

Since Go 1.11, a new recommended dependency management system is via modules. That is also our recommended approach.

Run formatting 
```bash
    make fmt 
```

Run unit test cases:
```bash
    make test
```

For running integration test cases locally 
Export the environment variable or put them in cicd-integration/script/.env.sh 

```bash
    export ADMIN_USER=''  
    export ADMIN_PASS=''
    export CLIENT_ID=''  
    export CLIENT_SECRET=''  
    export USER_NAME=''  
    export USER_PASSWORD=''  
    export DEV_DOMAIN=''
    export LOCAL_DOMAIN=''  
    export TEST_TENANT=''
    export USER1_NAME=''
    export USER1_PASSWORD=''
```
then just run
```bash
     cd cicd-integration/script      
     ./cicd-test.sh
```
   

To Build:
```bash
    make build
```



## Generated stubs 
We use [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) to generate stubs. 
#####Steps:
1. Install counterfeiter 
```bash
 GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter
```
   
2.  To generate mocks from an interface `Client` in a file `requests.go`:
```bash
   counterfeiter thy/requests/ Client
```

   
## Authors
* **Thycotic Software** - [Thycotic](https://thycotic.com)

## License
See LICENSE file.
