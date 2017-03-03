#!/usr/bin/env bash

set -e

export BUILT_BINARY_DIR=$PWD/built-binary
export GOPATH="$PWD/gopath"
export PATH="$GOPATH/bin:$PATH"

cd gopath/src/github.com/cloudfoundry/sipid

go build -o "$BUILT_BINARY_DIR/sipid" main.go
