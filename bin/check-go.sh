#!/bin/sh
set -eu
# shellcheck source=/dev/null
. ./bin/native-env.sh
mkdir -p "$MISE_DATA_DIR"
bin/native-check-config.sh --native golangci-lint >"$MISE_DATA_DIR/golangci-lint.yml.tmp"
test -s "$MISE_DATA_DIR/golangci-lint.yml.tmp"
mv "$MISE_DATA_DIR/golangci-lint.yml.tmp" "$MISE_DATA_DIR/golangci-lint.yml"
LEFTHOOK_CONFIG="$MISE_DATA_DIR/golangci-lint.yml" sh .standards/mise/scope.sh --tag golangci-lint --no-auto-install --no-stage-fixed --no-fail-on-changes
