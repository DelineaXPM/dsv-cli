---
title: Using the CLI
tags: ['cli']
---

More detailed documentation is listed on [CLI Documentation](https://docs.delinea.com/dsv/current/cli-ref/syntax.md?ref-githubrepo)

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
