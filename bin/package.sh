#!/bin/sh
set -eu
. ./bin/native-env.sh
go build ./cmd/clippy ./cmd/pasty "$@"
