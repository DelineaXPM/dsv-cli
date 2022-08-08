#!/usr/bin/env bash

echo "installing git credential manager core"
USER=microsoft
REPO=Git-Credential-Manager-Core
TAG=">=2.0.785"
BINARY="gcm-linux_amd64.*deb"
echo "Running $(fetch --version) to grab $USER/$REPO"
mkdir -p /tmp/installer-git-gcm
chmod +rw /tmp/installer-git-gcm
fetch --repo="https://github.com/$USER/$REPO" --tag="$TAG" --release-asset="$BINARY" --progress /tmp/installer-git-gcm
DOWNLOADED_FILE=$(find /tmp/installer-git-gcm -name $BINARY)
echo "Matched $DOWNLOADED_FILE successfully"
if ! [[ -f "$DOWNLOADED_FILE" ]]; then echo "$BINARY not found searching in /tmp/installer-git-gcm"; fi
sudo dpkg -i $DOWNLOADED_FILE
git-credential-manager-core configure
rm -rf /tmp/installer-git-gcm
echo "✔ git-gcm installed"
