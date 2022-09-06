#!/usr/bin/env bash
echo "⚙️ installing go" && mkdir -p ./tmpinstaller/ && wget "https://raw.githubusercontent.com/udhos/update-golang/master/update-golang.sh" -O ./tmpinstaller/update-golang.sh &&
    wget -q https://raw.githubusercontent.com/udhos/update-golang/master/update-golang.sh.sha256 -O ./tmpinstaller/hash.txt &&
    chmod +r ./tmpinstaller/hash.txt && pushd ./tmpinstaller && sha256sum --check hash.txt && popd &&
    chmod +x ./tmpinstaller/update-golang.sh && sudo bash ./tmpinstaller/update-golang.sh && rm -rf ./tmpinstaller && echo "✅ go installed"
