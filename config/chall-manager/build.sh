#!/bin/bash

CGO_ENABLED=0 go build -o main main.go
REGISTRY=${REGISTRY:-"registry:5000/"}

# custom the tag
VERSION="latest"
if [ $# -ne 0 ]; then
  VERSION=$1
fi

oras push --insecure \
  "${REGISTRY}chall-manager/deploy:${VERSION}" \
  --artifact-type application/vnd.ctfer-io.scenario \
  main:application/vnd.ctfer-io.file \
  Pulumi.yaml:application/vnd.ctfer-io.file
