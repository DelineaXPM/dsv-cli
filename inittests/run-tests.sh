#!/usr/bin/env bash

# Create an empty .thy.yml. If there already exists a working config, move it out temporarily and bring back later.
config_exists=false
if [ -f "$HOME/.thy.yml" ]; then
	mv "$HOME/.thy.yml" "$HOME/.thy2.yml"
	config_exists=true
fi

if [[ -v CONSTANTS_CLINAME ]]; then
  CONSTANTS_CLINAME=$CONSTANTS_CLINAME
else
  CONSTANTS_CLINAME=dsv
fi

echo "CLI NAME: ${CONSTANTS_CLINAME}"

touch "$HOME/.thy.yml"
if [[ ! -v BINARY_PATH ]]; then
	cd ..
	if [[ "$IS_SYSTEM_TEST" == "true" ]]; then
        make build-test
        mv $CONSTANTS_CLINAME.test inittests/$CONSTANTS_CLINAME
    else
        make
        mv $CONSTANTS_CLINAME inittests/$CONSTANTS_CLINAME
    fi
	cd inittests
	source .defaultvars
else
	cp $BINARY_PATH/*/$CONSTANTS_CLINAME-linux-x64 ./$CONSTANTS_CLINAME
	chmod +x $CONSTANTS_CLINAME
fi

source .env

python3 tests.py

deactivate
rm $CONSTANTS_CLINAME

# Return the original config, if it had existed.
if [ "$config_exists" == true ]; then
	mv "$HOME/.thy2.yml" "$HOME/.thy.yml"
fi
