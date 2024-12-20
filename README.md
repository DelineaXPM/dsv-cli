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

- Use with Docker.

  ![Docker Image Version (latest semver)](https://img.shields.io/docker/v/delineaxpm/dsv-cli?style=for-the-badge)

Examples:

```shell
# Make sure these files exists already so they aren't created by docker with the incorrect permissions
mkdir $HOME/.thy/
touch $HOME/.dsv.yml

# Use CLI and have the credentials mounted to home
docker run --rm -it \
    -v ${HOME}/.thy/:/home/nonroot/.thy/ \
    -v ${HOME}/.dsv.yml:/home/nonroot/.dsv.yml \
    delineaxpm/dsv-cli:latest --version version
# Example reading config
docker run --rm -it \
    --user 65532 \
    -v ${HOME}/.thy/:/home/nonroot/.thy/ \
    -v ${HOME}/.dsv.yml:/home/nonroot/.dsv.yml \
    delineaxpm/dsv-cli:latest cli-config read

# Wrap in a shell function for easier invoking via your zsh or bash profile.
function dsv() {
  docker run --rm -it \
      -v ${HOME}/.thy/:/home/nonroot/.thy/ \
      -v ${HOME}/.dsv.yml:/home/nonroot/.dsv.yml \
      delineaxpm/dsv-cli:latest "$@"
}
```

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
- 🍺 Homebrew: `brew install DelineaXPM/tap/dsv-cli`.
  - Upgrade with: `brew update && brew upgrade dsv-cli`
- Via Go (this will take longer than a binary install since it will build it):
  - run:

```shell
go install github.com/DelineaXPM/dsv-cli@latest
mv $(go env GOPATH)/bin/dsv-cli $(go env GOPATH)/bin/dsv
echo "dsv is installed at: $(go env GOPATH)/bin"
echo "Add to your profile to ensure Go binaries are in path by using:\n\n"
echo "export PATH=\"\$(go env GOPATH)/bin:\${PATH}\"\n\n"
echo "Current DSV Binaries installed: \n$(which -a dsv)"
```

- Curl (requires Go installed):

  ```shell
  go install github.com/mikefarah/yq/v4@latest
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
	<tbody>
		<tr>
            <td align="center">
                <a href="https://github.com/thycotic-rd">
                    <img src="https://private-avatars.githubusercontent.com/u/45605025?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjEwMjAsIm5iZiI6MTczNDY1OTgyMCwicGF0aCI6Ii91LzQ1NjA1MDI1In0.l-p1vWddeoxPPic3zIfIlm6Zk_RGU65UtoEtbDBj1hI&v=4" width="100;" alt="thycotic-rd"/>
                    <br />
                    <sub><b>Thycotic-Bot</b></sub>
                </a>
            </td>
            <td align="center">
                <a href="https://github.com/sheldonhull">
                    <img src="https://private-avatars.githubusercontent.com/u/3526320?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjA3MjAsIm5iZiI6MTczNDY1OTUyMCwicGF0aCI6Ii91LzM1MjYzMjAifQ.14EKpfaTbTTKs4Lul8VSxSICMjMUERx3tq8UXwfIfTA&v=4" width="100;" alt="sheldonhull"/>
                    <br />
                    <sub><b>Sheldonhull</b></sub>
                </a>
            </td>
            <td align="center">
                <a href="https://github.com/andrii-zakurenyi">
                    <img src="https://private-avatars.githubusercontent.com/u/85106843?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjExNDAsIm5iZiI6MTczNDY1OTk0MCwicGF0aCI6Ii91Lzg1MTA2ODQzIn0.5AFpAcymebv99y4QuV2PTarB39JvWBLe581aH9XCBVg&v=4" width="100;" alt="andrii-zakurenyi"/>
                    <br />
                    <sub><b>Andrii Zakurenyi</b></sub>
                </a>
            </td>
            <td align="center">
                <a href="https://github.com/mariiatuzovska">
                    <img src="https://private-avatars.githubusercontent.com/u/41679258?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjEwMjAsIm5iZiI6MTczNDY1OTgyMCwicGF0aCI6Ii91LzQxNjc5MjU4In0.EwVUwdcG-e-nsJB8uZGyjhPmPg3YHKlNHvFkT3Cabts&v=4" width="100;" alt="mariiatuzovska"/>
                    <br />
                    <sub><b>Mariia</b></sub>
                </a>
            </td>
            <td align="center">
                <a href="https://github.com/pacificcode">
                    <img src="https://private-avatars.githubusercontent.com/u/918320?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjA5MDAsIm5iZiI6MTczNDY1OTcwMCwicGF0aCI6Ii91LzkxODMyMCJ9.fzJonqgx0Wxe7TLZ2anmgPlP-UyCZivLJiH1MqF0WIc&v=4" width="100;" alt="pacificcode"/>
                    <br />
                    <sub><b>Bill Hamilton</b></sub>
                </a>
            </td>
            <td align="center">
                <a href="https://github.com/tdillenbeck">
                    <img src="https://private-avatars.githubusercontent.com/u/21064520?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTEiLCJleHAiOjE3MzQ2NjEwODAsIm5iZiI6MTczNDY1OTg4MCwicGF0aCI6Ii91LzIxMDY0NTIwIn0.-zw04hRsAWYNrEUcWeoqY7mNfQMAW8N0BHYQS7o2VzQ&v=4" width="100;" alt="tdillenbeck"/>
                    <br />
                    <sub><b>Tom Dillenbeck</b></sub>
                </a>
            </td>
		</tr>
	<tbody>
</table>
<!-- readme: collaborators,contributors -end -->

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

[prebuilt-binaries]: https://dsv.secretsvaultcloud.com/downloads
[aqua-project]: https://aquaproj.github.io/
