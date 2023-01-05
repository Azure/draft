#!/usr/bin/env bash

mkdir -p ./test/temp
./draft info > ./test/temp/info.json
echo "Draft Info JSON schema:"
cat ./test/info_schema.json
echo "Draft Info JSON:"
cat ./test/temp/info.json

npm install -g ajv-cli
echo "Validating Draft Info JSON against schema"
ajv validate -s ./test/info_schema.json -d ./test/temp/info.json