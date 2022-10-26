---
title: Local Development
tags: ['development', 'tooling']
---

## Building the CLI

To build the program or contribute to its development, use Go version 1.17 or later.

Build:

```shell
make build
```

Run tests:

```shell
make test
```

## Troubleshooting

### Line Endings

- Tasks in the Makefile and cicd-integration tests only run on Linux. On Windows, use the Windows Subsystem for Linux (WSL).
- Make sure to not convert line endings to CRLF. The `.gitattributes` file should prevent committing CRLF line endings. Configure Git and a text editor as necessary.
- Turn off autocrlf: `git config --global --unset core.autocrlf`.
- Re-checkout: `git checkout-index --force --all`.
- If that fails, run: `git config core.eol lf` and re-checkout again or delete and reclone the repo.
