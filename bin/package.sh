#!/bin/sh
set -eu
. ./bin/macos-env.sh
go build ./cmd/clippy ./cmd/pasty "$@"
