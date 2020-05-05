#!/usr/bin/env bash

set -e pipefail

export GO111MODULE=on

EXITCODE=0

if [[ -f .env.sh ]]; then
   source ./.env.sh
fi

cd "$(dirname "$0")"/..

if [[ -f out.txt ]]; then
    rm -f out.txt
fi

echo "integration tests"
CGO_ENABLED=1 go test -v -count 1 -p 1  ./...  2>&1 > out.txt && EXITCODE=$((EXITCODE+$?)) || EXITCODE=$((EXITCODE+$?))
echo "integration tests exit code: $EXITCODE"

pushd "$(pwd)" &>/dev/null

cd  ../cli-config
rm -rf .thy.yml

popd &>/dev/null

exit $EXITCODE

