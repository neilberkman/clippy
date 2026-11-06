#!/bin/sh
set -eu
# shellcheck source=/dev/null
. ./bin/native-env.sh
mkdir -p "$MISE_CONFIG_DIR"
touch "$MISE_GLOBAL_CONFIG_FILE"
MISE_LOCKED=0 MISE_LOCKFILE=0 mise install aqua:jdx/mise@2026.11.1
native_mise="$(MISE_LOCKED=0 MISE_LOCKFILE=0 mise which mise)"
"$native_mise" install --locked aqua:jdx/mise aqua:realm/SwiftLint aqua:cpisciotta/xcbeautify aqua:evilmartians/lefthook aqua:mikefarah/yq
"$native_mise" reshim
"$native_mise" env --shell bash >"$MISE_DATA_DIR/runtime-env.sh.tmp"
sh -n "$MISE_DATA_DIR/runtime-env.sh.tmp"
mv "$MISE_DATA_DIR/runtime-env.sh.tmp" "$MISE_DATA_DIR/runtime-env.sh"
# shellcheck source=/dev/null
. "$MISE_DATA_DIR/runtime-env.sh"
