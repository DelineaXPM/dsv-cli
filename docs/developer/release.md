---
title: Release
tags: ['github-release', 'dsv-cli']
---

## Required For Pipelines

- [Create Github Token](https://github.com/settings/tokens/new?description=org-goreleaser&scopes=admin:repo,read:org,write:discussion) and store this in Azure Pipeline Vault.
  Look for a `GORELEASER` variable group and replace the `GITHUB_TOKEN` value, no expiration date set.
- Certs (secure files).
- DSV Tenants and other details are noted in the azure pipeline files.

## Creating a New Release

### Release Notes

This project uses an different approach to release, driving it from changelog and versioned changelog notes instead of tagging.

> Use [changie](https://changie.dev/guide/quick-start/) quick start for basic review.

### Creating New Notes

- During development, new changes of note get tracked via `changie new`. This can span many pull requests, whatever makes sense as version to ship as changes to users.
- To release the changes into a version, `changie batch <major|minor|patch>` (unless breaking changes occur, you'll want to stick with minor for feature additions, and patch for fixes or non app work.

Keep your summary of changes that users would care about in the `.changes/` files it will create.

### Release

Update [CHANGELOG.md](../../CHANGELOG.md) by running `changie merge` which will rebuild the changelog file with all the documented notes.

## FAQ

### What drives the version number for the release?

Changie notes are named like `v1.0.4.md`.
This version number will be used to set the version of the release, so the docs in essence will be the version source of truth.

## Snap Login

- Run `snapcraft login`.
- After login: `snapcraft export-login snapcraft-login` to create a file `snapcraft-login` for the login to use for CI purposes.
  Upload this as a secure file in Azure DevOps Secure file vault, or if using a shared team DSV Vault, place it in there (that's pending implementation as of 2023-01).

### Version Prefix

- Referencing the version number diretly (not assets), is done via `1.0.0` with no prefix.
  - For example the latest version number in `cli-versions.json` would not have the prefix.
  - [Scoop](https://github.com/DelineaXPM/scoop-bucket/blob/89cc09954d090f0e5421230db51f8eaa40b63e18/dsv-cli.json#L2) is the same with version not having a prefix.
- Tags (per Go standard) include `v` prefix.
  This is created by goreleaser github process automatically.
- Assets:
  - GitHub assets include `v` prefix like `v1.0.0`.
  - Scoop & Brew use version number without prefix in version field, but the download assets uploaded to github do have prefix in the file name.
- S3 asset folder prefers no `v` prefix.
