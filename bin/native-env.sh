#!/bin/sh
set -eu
export MISE_DATA_DIR="$REPO_DIR/.mise"
export MISE_SHIMS_DIR="$MISE_DATA_DIR/shims"
export MISE_CACHE_DIR="$MISE_DATA_DIR/cache"
export MISE_STATE_DIR="$MISE_DATA_DIR/state"
export MISE_CONFIG_DIR="$MISE_DATA_DIR/config"
export MISE_SYSTEM_CONFIG_DIR="$MISE_DATA_DIR/system"
export MISE_GLOBAL_CONFIG_FILE="$MISE_DATA_DIR/config/empty.toml"
export MISE_OVERRIDE_CONFIG_FILENAMES="$REPO_DIR/image/mise.toml"
export MISE_TRUSTED_CONFIG_PATHS="$REPO_DIR"
export MISE_YES=1
export PATH="$MISE_DATA_DIR/shims:$PATH"
export BUNDLE_PATH="$MISE_DATA_DIR/bundle"
export BUNDLE_USER_HOME="$MISE_DATA_DIR/bundle-user"
export BUNDLE_APP_CONFIG="$MISE_DATA_DIR/bundle-config"
export FOCUS_BUILD_DIR="$MISE_DATA_DIR/build"
export MISE_AUTO_INSTALL=0
if [ -f "$MISE_DATA_DIR/runtime-env.sh" ]; then
  # shellcheck source=/dev/null
  . "$MISE_DATA_DIR/runtime-env.sh"
fi
