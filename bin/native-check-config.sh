#!/bin/sh
set -eu
if [ "${1:-}" = --native ] || [ ! -f /opt/standards/lefthook/standards.yml ]; then
  export REPO_DIR="${REPO_DIR:-$(git rev-parse --show-toplevel)}"
  export MISE_DATA_DIR="${MISE_DATA_DIR:-$REPO_DIR/.mise}"
  mkdir -p "$MISE_DATA_DIR"
  # shellcheck disable=SC2016 # yq expands its own replacement captures
  yq '.extends[]? |= (sub("^/opt/standards/", strenv(REPO_DIR) + "/.standards/") | sub("^([^/].*)$", strenv(REPO_DIR) + "/${1}"))' lefthook.yml >"$MISE_DATA_DIR/native-root.yml.tmp"
  mv "$MISE_DATA_DIR/native-root.yml.tmp" "$MISE_DATA_DIR/native-root.yml"
  export LEFTHOOK_CONFIG="$MISE_DATA_DIR/native-root.yml"
fi
# shellcheck disable=SC2016 # $hook is a yq variable
lefthook dump | yq '.["pre-commit"] as $hook | $hook.jobs[] | select(.name == "compilers") | .group.jobs[] | select(.name == "swiftlint") | .run = "swiftlint lint --quiet --strict --force-exclude --config .standards/config/swiftlint.yml --config .mise/swiftlint-rules.yml {staged_files}" | {"pre-commit": {"exclude_tags": ($hook.exclude_tags // []), "jobs": [.]}}'
