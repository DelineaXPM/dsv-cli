#!/usr/bin/env bash

# Create an empty .thy.yml. If there already exists a working config, move it out temporarily and bring back later. 
config_exists=false
if [ -f "$HOME/.thy.yml" ]; then
	mv "$HOME/.thy.yml" "$HOME/.thy2.yml"
	config_exists=true
fi


touch "$HOME/.thy.yml"
if [[ ! -v BINARY_PATH ]]; then
	cd ..
	if [[ "$IS_SYSTEM_TEST" == "true" ]]; then
        make build-test
        mv thy.test inittests/thy
    else
        make
        mv thy inittests/thy
    fi
	cd inittests
	source .defaultvars
else
	cp $BINARY_PATH/*/thy-linux-x64 ./thy
	chmod +x thy
fi

source .env

python3 tests.py

deactivate
rm thy

# Return the original config, if it had existed.
if [ "$config_exists" == true ]; then
	mv "$HOME/.thy2.yml" "$HOME/.thy.yml"
fi
