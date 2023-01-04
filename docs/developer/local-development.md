---
title: Local Development
tags: ['development', 'tooling']
---

## Building the CLI

To build the program or contribute to its development, use Go version 1.18 or later.

Run `mage` command line tool to list tasks for building and testing.
If using aqua to install tooling, this is automatically added.
If not, you'll need to install [mage manually](https://magefile.org).

These tasks are documented via `mage -l` or `mage -h taskname` in more detail.

## Running Local Test Build on Darwin

- Manually run: `brew install snapcraft`.

## Troubleshooting

### Line Endings

- Tasks in the Makefile and cicd-integration tests only run on Linux. On Windows, use the Windows Subsystem for Linux (WSL).
- Make sure to not convert line endings to CRLF. The `.gitattributes` file should prevent committing CRLF line endings. Configure Git and a text editor as necessary.
- Turn off autocrlf: `git config --global --unset core.autocrlf`.
- Re-checkout: `git checkout-index --force --all`.
- If that fails, run: `git config core.eol lf` and re-checkout again or delete and reclone the repo.
