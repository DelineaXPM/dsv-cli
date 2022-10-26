---
title: Devcontainer & Codespaces Usage
tags: ['development', 'tooling']
---

## Prerequisites For Devcontainer

- Docker
- Visual Studio Code
  - Run `code --install-extension ms-vscode-remote.remote-containers`
  - For supporting Codespaces: `code --install-extension GitHub.codespaces`

## Spin It Up

> **_NOTE_**
>
> 🐎 PERFORMANCE TIP: Using the directions provided for named container volume will optimize performance over trying to just "open in container" as there is no mounting files to your local filesystem.

Use command pallet with vscode (Control+Shift+P or F1) and type to find the command `Remote Containers: Clone Repository in Named Container`.

- Put the git clone url in.

Some extra features are included such as:

- Extensions for VSCode defined in `.devcontainers`, such as Go, Kubernetes & Docker, and some others.
- Initial placeholder `.zshrc` file included to help initialize usage of `direnv` for automatically loading default `.envrc` which contains local developement default environment variables.

### After Devcontainer Loads

1. Accept "Install Recommended Extensions" from popup, to automatically get all the preset tools, and you can choose do this without syncing so it's just for this development environment.
2. Open a new `zsh-login` terminal and allow the automatic setup to finish, as this will ensure all other required tools are setup.
   - Make sure to run `direnv allow` as it prompts you, to ensure all project and your personal environment variables (optional).
3. Run setup task:
   - Using CLI: Run `mage init`

## Troubleshooting

### Connecting to Services Outside of devcontainer

You are in an isolated, self-contained Docker setup.
The ports internally aren't the same as externally in your host OS.
If the port forward isn't discovered automatically, enable it yourself, by using the port forward tab (next to the terminal tab).

1. You should see a port forward once the services are up (next to the terminal button in the bottom pane).
   1. If the click to open url doesn't work, try accessing the path manually, and ensure it is `https`.
      Example: `https://127.0.0.1:9999`

You can choose the external port to access, or even click on it in the tab and it will open in your host for you.
