# Thycotic DevOps Secrets Vault CLI

An automation tool for the management of credentials for applications, databases, CI/CD tools, and services.

## CLI Usage Documentation
The documentation is available at http://docs.thycotic.com/dsv

The documentation repository is available at https://github.com/thycotic/dsv-docs

To request a documentation change, fork and submit a pull request.

## Development environment setup

* Tasks in the Makefile and cicd-integration tests only run on Linux. On Windows, use the Windows Subsystem for Linux (WSL).
* Make sure to not convert line endings to CRLF. The `.gitattributes` file should prevent committing CRLF line endings. Configure Git and a text editor as necessary.
	* Turn off autocrlf: `git config --global --unset core.autocrlf`.
	* Re-checkout: `git checkout-index --force --all`.
	* If that fails, run: `git config core.eol lf` and re-checkout again or delete and reclone the repo.

## Building the CLI
To build the program or contribute to its development, use Go version 1.16 or later.

Build:
```bash
make build
```

Run tests:
```bash
make test
```

## Installation
Installation is only required for autocomplete to work on Unix-based systems.
Autocomplete is a convenience feature to help type commands, subcommands, flags, and secret paths faster.
It is supported on bash, zsh, and fish shells via https://github.com/posener/complete.

To enable autocomplete:
```
go get -u github.com/posener/complete/gocomplete
gocomplete -install
dsv -install
```
To disable it:
```
dsv -uninstall
gocomplete -uninstall
```

## Working with the CLI
All commands have help text with examples that are displayed when the `-h` or `--help` flag is provided.

Client-side configuration is read by default from `~/.thy.yml`.

Alternatively, username, password and tenant may be specified as environment variables or as global command-line flags.

## Examples
Create a secret at the path `resources/us-east-1/server1`:
```bash
dsv secret create \
  --path resources/us-east-1/server1 \
  --desc 'my important secret' \
  --data '{"password": "0cuJvsU3sY6Lc"}'
```

Read a secret field:
```bash
dsv secret read resources/us-east-1/server1 -bf .data.password
```

## Authors
* **Thycotic Software** - [Thycotic](https://thycotic.com)

## License
See LICENSE file.
