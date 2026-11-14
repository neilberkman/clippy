#!/bin/sh
set -eu

repo_dir=$(git rev-parse --show-toplevel)
git_common_dir=$(git rev-parse --path-format=absolute --git-common-dir)
project_name="repo-$(printf %s "$repo_dir" | shasum -a 256 | cut -d ' ' -f 1)"

# Compose and the editor must use the same paths and dependency volumes as just.
umask 077
printf 'REPO_DIR=%s\nRUBIO_GIT_COMMON_DIR=%s\nCOMPOSE_PROJECT_NAME=%s\nRUBIO_BUILD_GITHUB_TOKEN=\n' \
  "$repo_dir" "$git_common_dir" "$project_name" >"$repo_dir/.env"
chmod 600 "$repo_dir/.env"
for volume in rubio-uv-cache rubio-npm-cache rubio-gomod-cache rubio-cargo-registry rubio-coursier-cache; do
  docker volume create "$volume" >/dev/null
done
