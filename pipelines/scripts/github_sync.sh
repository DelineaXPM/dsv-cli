#!/bin/bash

# This script pushes all changes to the DSV CLI public repository.

set -o errexit
set -o nounset
set -o pipefail

PUBLIC_REPO="github.com/thycotic/dsv-cli.git"
SYNC_GIT_USER_NAME="thycotic-rd"
SYNC_GIT_USER_EMAIL="lightvaulttrack@thycotic.com"
SYNC_GIT_TAG=$(git describe --always --dirty --tags)
SYNC_GIT_COMMIT=$(git rev-parse HEAD)

WORKING_DIR=$(pwd)

echo "[---] Current working directory: ${WORKING_DIR}"
echo "[---] Current tag is ${SYNC_GIT_TAG}. Validating tag."

if [[ "${SYNC_GIT_TAG}" == *"-rc"* ]]; then
    echo "[---] Tag ${SYNC_GIT_TAG} is RC tag. Skipping sync."
    exit 0
fi

if [[ "${SYNC_GIT_TAG}" == *"-"* ]]; then
    echo "[---] Tag ${SYNC_GIT_TAG} contains dash(es). Skipping sync."
    exit 0
fi

echo "[---] Tag validation complete. Tag ${SYNC_GIT_TAG} is OK."

echo "[---] Removing untracked files before sync."
git clean -fdx

echo "[---] Running git status."
git status

echo "[---] Cloning public repository."
git clone --depth 1 https://${PUBLIC_REPO} public-github-repository

cleanup() {
    echo "[---] Cleaning sync directory."
    cd ${WORKING_DIR}
    rm -rf public-github-repository
}

trap 'cleanup' EXIT

echo "[---] Syncing changes."

find ./public-github-repository -mindepth 1 -maxdepth 1 -not -name '.git' -exec rm -rf "{}" \;
cp -r $(find . -maxdepth 1 -mindepth 1 -not -name .git -not -name public-github-repository) public-github-repository

echo "[---] Running git status to show changes which will be commited."
cd ./public-github-repository
git status

echo "[---] Configuring user name and email."
git config user.name ${SYNC_GIT_USER_NAME}
git config user.email ${SYNC_GIT_USER_EMAIL}

echo "[---] Staging all changes for commit"
git add --all

echo "[---] Creating commit."
git commit -m "Automated from: ${SYNC_GIT_COMMIT}"

echo "[---] Creating tag."
git tag ${SYNC_GIT_TAG} --force

echo "[---] Pushing changes to the public repository."
git push https://${SYNC_GIT_USER_NAME}:${githubPat}@${PUBLIC_REPO} HEAD:master

echo "[---] Pushing tag ${SYNC_GIT_TAG} to the public repository."
git push https://${SYNC_GIT_USER_NAME}:${githubPat}@${PUBLIC_REPO} ${SYNC_GIT_TAG}

echo "[---] Sync finished."