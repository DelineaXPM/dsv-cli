# Delinea DevOps Secrets Vault CLI

An automation tool for the management of credentials for applications, databases, CI/CD tools, and services.

![landing-demo](docs/vhs/assets/landing-demo.gif)

## Getting Started

- [Pre-built Binaries][prebuilt-binaries]
- [Developer](docs/developer): instructions on running tests, local tooling, and other resources.
- [DSV Documentation](https://docs.delinea.com/dsv/current?ref=githubrepo)

## Quick Start Install

<details closed>
<summary>ℹ️ Any Platform</summary>

- 🔨 Download from [prebuilt-binaries] manually.
- [Aquaproject][aqua-project]: `aqua generate 'DelineaXPM/dsv-cli' -i` and update your `aqua.yml` file.
- PowerShell Cross-Platform (pwsh) with console selector (move to directory in `$ENV:PATH` for it to be universally discoverable):

  ```powershell
  if (-not (Get-InstalledModule Microsoft.PowerShell.ConsoleGuiTools -ErrorAction SilentlyContinue))
  {
      Install-Module Microsoft.PowerShell.ConsoleGuiTools -Force -Confirm:$false -Scope CurrentUser
  }

  $json=(Invoke-WebRequest -ContentType 'application/json' -Uri 'https://s3.amazonaws.com/dsv.secretsvaultcloud.com/cli-version.json' -UseBasicParsing).Content | ConvertFrom-Json
  $download=$json.Links | get-member -Type NoteProperty | ForEach-Object {
  [pscustomobject]@{
      FileName = $_.Name
      DownloadLink = $json.Links.$($_.Name)
      OutFileName = ($json.Links.$($_.Name) -split '/')[-1]
  }} | Out-ConsoleGridView  -Title 'Delinea DevOps Secrets Vault CLI'
  $download | ForEach-Object { Invoke-WebRequest -Uri $_.DownloadLink -OutFile $_.OutFileName -UseBasicParsing }
  ```

</details>

<details closed>
<summary>ℹ️ Mac & Linux</summary>

## Mac & Linux

- [aqua-project] provides a binary tool manager similar to Brew.
- [🍺 Brew][brew-install]: `brew install DelineaXPM/tap/dsv-cli`.
  - Upgrade with: `brew update && brew upgrade dsv-cli`
- Curl (requires Go installed):

  ```shell
  version=$(curl -sb -H "Accept: application/json" https://s3.amazonaws.com/dsv.secretsvaultcloud.com/cli-version.json | $(go env GOPATH)/bin/yq '.latest')
  echo "version: $version"
  curl -fSsl https://dsv.secretsvaultcloud.com/downloads/cli/$version/dsv-darwin-x64 -o dsv && chmod +x ./dsv && sudo mv ./dsv /usr/local/bin
  ```

- Curl (no Go required). Requires specifying the version:

  ```shell
  curl -fSsl https://dsv.secretsvaultcloud.com/downloads/cli/1.39.5/dsv-darwin-x64 -o dsv && chmod +x ./dsv && sudo mv ./dsv /usr/local/bin
  ```

> **note**: It is not required to install to `/usr/local/bin`. If you choose to install to another location you'll want to make sure it's added to your PATH for the tool to be found.

</details>

<details closed>
<summary>ℹ️ Linux Only</summary>

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/dsv-cli)

- Via cli: `snap install dsv-cli`.
  - At this time add alias to your profile with: `alias dsv='dsv-cli', as the snap name is not aliased.
- Note: Snaps update automatically (4 times a day as the [default behavior as of 2023-01](https://snapcraft.io/docs/keeping-snaps-up-to-date)), but this can be run manually via `snap refresh`.

</details>

<details closed>
<summary>ℹ️ Windows</summary>

### Windows

- Scoop:
  - First time setup: `scoop bucket add DelineaXPM https://github.com/DelineaXPM/scoop-bucket.git`.
  - Install: `scoop install DelineaXPM/dsv-cli`.
  - Update: `scoop update DelineaXPM/dsv-cli`.
- Using curl in Windows PowerShell (for cross-platform pwsh see top section) and move to whatever directory you want:

  ```powershell
  $json=(Invoke-WebRequest -ContentType 'application/json' -Uri 'https://s3.amazonaws.com/dsv.secretsvaultcloud.com/cli-version.json' -UseBasicParsing).Content | ConvertFrom-Json
  # Change this to windows/386 if required to install x86.
  Invoke-WebRequest -Uri $json.links.'windows/amd64' -OutFile 'dsv.exe' -UseBasicParsing
  ```

</details>

## License

See [LICENSE](https://github.com/DelineaXPM/dsv-cli/blob/main/LICENSE) for the full license text.

## Contributors

<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->

<!-- readme: collaborators,contributors -start -->
<table>
<tr>
    <td align="center">
        <a href="https://github.com/thycotic-rd">
            <img src="https://avatars.githubusercontent.com/u/45605025?v=4" width="100;" alt="thycotic-rd"/>
            <br />
            <sub><b>Thycotic-Bot</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/sheldonhull">
            <img src="https://avatars.githubusercontent.com/u/3526320?v=4" width="100;" alt="sheldonhull"/>
            <br />
            <sub><b>Sheldonhull</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/mariiatuzovska">
            <img src="https://avatars.githubusercontent.com/u/41679258?v=4" width="100;" alt="mariiatuzovska"/>
            <br />
            <sub><b>Mariia</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/andrii-zakurenyi">
            <img src="https://avatars.githubusercontent.com/u/85106843?v=4" width="100;" alt="andrii-zakurenyi"/>
            <br />
            <sub><b>Andrii Zakurenyi</b></sub>
        </a>
    </td></tr>
</table>
<!-- readme: collaborators,contributors -end -->

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

[prebuilt-binaries]: https://dsv.secretsvaultcloud.com/downloads
[aqua-project]: https://aquaproj.github.io/
[brew-install]: PENDING
