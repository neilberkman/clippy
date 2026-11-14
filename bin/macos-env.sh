#!/bin/sh
set -eu
unset GIT_COMMON_DIR
export REPO_DIR="${REPO_DIR:-$(git rev-parse --show-toplevel)}"
if [ -z "${RUBIO_JUST_BIN:-}" ]; then
  RUBIO_JUST_BIN="$(mise which just 2>/dev/null || command -v just)"
  export RUBIO_JUST_BIN
fi
export MISE_DATA_DIR="$REPO_DIR/.mise"
export MISE_SHIMS_DIR="$MISE_DATA_DIR/shims"
export MISE_CACHE_DIR="$MISE_DATA_DIR/cache"
export MISE_STATE_DIR="$MISE_DATA_DIR/state"
export MISE_CONFIG_DIR="$MISE_DATA_DIR/config"
export MISE_SYSTEM_CONFIG_DIR="$MISE_DATA_DIR/system"
export MISE_GLOBAL_CONFIG_FILE="$MISE_CONFIG_DIR/empty.toml"
export MISE_OVERRIDE_CONFIG_FILENAMES="$REPO_DIR/image/mise.toml"
export MISE_TRUSTED_CONFIG_PATHS="$REPO_DIR"
export MISE_YES=1
export MISE_AUTO_INSTALL=0
export PATH="$MISE_SHIMS_DIR:$PATH"
if [ -f "$REPO_DIR/bin/macos-settings.sh" ]; then
  # shellcheck source=/dev/null
  . "$REPO_DIR/bin/macos-settings.sh"
fi
if [ -f "$MISE_DATA_DIR/runtime-env.sh" ]; then
  # shellcheck source=/dev/null
  . "$MISE_DATA_DIR/runtime-env.sh"
fi
PATH="$(dirname "$RUBIO_JUST_BIN"):$PATH"
export PATH
