#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

go build -o "$TMPDIR/sipid" "$DIR/../main.go"
go build -o "$TMPDIR/hard" "$DIR/../kill/fixtures/hard_kill/main.go"

"$TMPDIR/sipid" claim --pid "$$" --pid-file /tmp/sipid-example

exec "$TMPDIR/hard"
