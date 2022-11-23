#!/usr/bin/env bash

./draft info > ./info.json
echo "Draft Info JSON schema:"
cat ./test/info_schema.json
  echo "Draft Info JSON:"
cat info.json

npm install -g ajv-cli
echo "Validating Draft Info JSON against schema"
ajv validate -s ./test/info_schema.json -d info.json