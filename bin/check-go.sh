#!/bin/sh
set -eu
# shellcheck source=/dev/null
. ./bin/macos-env.sh
mkdir -p "$MISE_DATA_DIR"
bin/macos-check-config.sh --macos golangci-lint >"$MISE_DATA_DIR/golangci-lint.yml.tmp"
test -s "$MISE_DATA_DIR/golangci-lint.yml.tmp"
mv "$MISE_DATA_DIR/golangci-lint.yml.tmp" "$MISE_DATA_DIR/golangci-lint.yml"
LEFTHOOK_CONFIG="$MISE_DATA_DIR/golangci-lint.yml" sh .standards/mise/scope.sh --tag golangci-lint --no-auto-install --no-stage-fixed --no-fail-on-changes
bin/lint-go.sh
