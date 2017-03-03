#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

go build -o "$TMPDIR/sipid" "$DIR/../main.go"

"$TMPDIR/sipid" kill --pid-file /tmp/sipid-example --show-stacks
