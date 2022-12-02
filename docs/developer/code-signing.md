# Code Signing

This is automated as part of the build process via calls to Azure Pipelines.

This doc captures any specific debugging or other helpful notes to work through any issues.

## How Do I Sign?

If you have all the required certs setup, then you just run `mage release` and all of the required signing steps will be performed by goreleaser for all platforms.

## General Tooling

A single linux based agent is all that is required to build and sign for cross-platform, due to the power of Go and some amazing community projects.
You should be able to do this on any OS, but this is primarily tested in Darwin & run in CI on a Linux based agent.

- For mac signing: [Quill](https://github.com/anchore/quill) to allow signing and notarization of the binary without using a darwin based build system.
- For windows & linux binary signing: [cosign](https://github.com/sigstore/cosign)

## Mac Specific

### Validation

- On a darwin system, to get more helpful diagnostics on why a signature is invalid try: `codesign -vvv --deep --strict ./.artifacts/dsv`.
- Use the spctl utility to determine if the software to be notarized will run with the system policies currently in effect: `spctl -vvv --assess --type exec ./.artifacts/dsv`

For visually verifying signing you can install: `brew install whatsyoursign` and then run `open '/usr/local/Caskroom/whatsyoursign/2.0.1/WhatsYourSign Installer.app'` or whatever version you find.
This will add a Finder button when right-clicking that is called "Signing Info" and provide a visual way to look at the signing information on a Mac system.

### [DEPRECATED] Codesign

The following notes are more specific to working with Apple certificate signing.

The following certs from Apple are required to sign correctly.
To do this automatically on a Darwin based system just run `mage certs:init`.

If the dates are incorrect, update this to the latest from this page [Apple Public Certificates](https://www.apple.com/certificateauthority/) in `magefile/certs` to add it to the download and install list.

### Mac Resource Links

| Topic | Description |
| [Troubleshooting Common Issues][common-issues] | Working through notarization issues |

### Mac Error Notes

- `CSSMERR_TP_NOT_TRUSTED` build error (and sometimes but less common, is it's Archive 'Share' or 'Submit' manifestation) is
  the result of mistakenly modifying Trust Settings on one of your iOS Development-related certificates. [stack overflow answer][stack-error-help].
  [Apple support answer][apple-support-error-help] also provides more help.

[common-issues]: https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution/resolving_common_notarization_issues#3087735
[stack-error-help]: https://stackoverflow.com/a/8766966/68698
[apple-support-error-help]: https://developer.apple.com/library/archive/technotes/tn2250/_index.html#//apple_ref/doc/uid/DTS40009933-CH1-TNTAG19
