<h1 align="center">Delinea DevOps Secrets Vault CLI</h1>

An automation tool for the management of credentials for applications, databases, CI/CD tools, and services.

## Installation

Prebuilt binaries for Linux, macOS and Windows can be downloaded from https://dsv.secretsvaultcloud.com/downloads.

## Documentation

The documentation is available at https://docs.delinea.com/dsv/current.

## Development environment setup

* Tasks in the Makefile and cicd-integration tests only run on Linux. On Windows, use the Windows Subsystem for Linux (WSL).
* Make sure to not convert line endings to CRLF. The `.gitattributes` file should prevent committing CRLF line endings. Configure Git and a text editor as necessary.
* Turn off autocrlf: `git config --global --unset core.autocrlf`.
* Re-checkout: `git checkout-index --force --all`.
* If that fails, run: `git config core.eol lf` and re-checkout again or delete and reclone the repo.

## Building the CLI

To build the program or contribute to its development, use Go version 1.17 or later.

Build:

```bash
make build
```

Run tests:

```bash
make test
```

## Autocompletion

Autocomplete is a convenience feature to help type commands, subcommands, flags, and secret paths faster.
It is supported on bash, zsh, and fish shells via https://github.com/posener/complete.

To enable autocomplete:

```bash
dsv -install
```

To disable it:

```bash
dsv -uninstall
```

## Working with the CLI

All commands have help text with examples that are displayed when the `-h` or `--help` flag is provided.

Client-side configuration is read by default from `~/.dsv.yml`.

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
dsv secret read resources/us-east-1/server1 -f .data.password
```

## Authors

* **Delinea** - https://delinea.com/

## License

See [LICENSE](https://github.com/thycotic/dsv-cli/blob/master/LICENSE) for the full license text.
