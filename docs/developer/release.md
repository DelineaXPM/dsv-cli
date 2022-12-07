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

Update [CHANGELOG.md](CHANGELOG.md) by running `changie merge` which will rebuild the changelog file with all the documented notes.

## FAQ

### What drives the version number for the release?

Changie notes are named like `v1.0.4.md`.
This version number will be used to set the version of the release, so the docs in essence will be the version source of truth.
