#!/bin/sh
set -eu
if [ -f bin/macos-env.sh ]; then
  # shellcheck source=/dev/null
  . ./bin/macos-env.sh
fi
export REPO_DIR="${REPO_DIR:-$(git rev-parse --show-toplevel)}"
export MISE_DATA_DIR="${MISE_DATA_DIR:-$REPO_DIR/.mise}"
mkdir -p "$MISE_DATA_DIR"
# shellcheck disable=SC2016 # yq expands its own replacement captures
rules='del(.parent_config) | .included[]? |= sub("^([^/].*)$", strenv(NATIVE_REPO_ROOT) + "/${1}") | .excluded[]? |= sub("^([^/].*)$", strenv(NATIVE_REPO_ROOT) + "/${1}")'
if command -v docker >/dev/null 2>&1; then
  RUBIO_GIT_COMMON_DIR="$(git rev-parse --path-format=absolute --git-common-dir)" docker compose run --rm -T dev env NATIVE_REPO_ROOT="$REPO_DIR" yq "$rules" .swiftlint.yml </dev/null >"$MISE_DATA_DIR/swiftlint-rules.yml.tmp"
  RUBIO_GIT_COMMON_DIR="$(git rev-parse --path-format=absolute --git-common-dir)" docker compose run --rm -T dev bin/macos-check-config.sh </dev/null >"$MISE_DATA_DIR/swiftlint.yml.tmp"
else
  NATIVE_REPO_ROOT="$REPO_DIR" yq "$rules" .swiftlint.yml >"$MISE_DATA_DIR/swiftlint-rules.yml.tmp"
  bin/macos-check-config.sh --macos >"$MISE_DATA_DIR/swiftlint.yml.tmp"
fi
test -s "$MISE_DATA_DIR/swiftlint.yml.tmp"
mv "$MISE_DATA_DIR/swiftlint-rules.yml.tmp" "$MISE_DATA_DIR/swiftlint-rules.yml"
mv "$MISE_DATA_DIR/swiftlint.yml.tmp" "$MISE_DATA_DIR/swiftlint.yml"
LEFTHOOK_CONFIG="$MISE_DATA_DIR/swiftlint.yml" sh .standards/mise/scope.sh --tag swiftlint --no-auto-install --no-stage-fixed --no-fail-on-changes
