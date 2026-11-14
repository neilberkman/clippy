#!/bin/sh
set -eu

bin/dev-environment.sh

token=${GH_TOKEN:-${GITHUB_TOKEN:-}}
if [ -z "$token" ] && command -v gh >/dev/null 2>&1; then
  token=$(gh auth token 2>/dev/null) || token=
fi
RUBIO_BUILD_GITHUB_TOKEN=$token
export RUBIO_BUILD_GITHUB_TOKEN

base=$(awk '$1 == "FROM" && $3 == "AS" && $4 == "dev" { print $2; exit }' Dockerfile)
if [ -z "$base" ]; then
  echo 'Cannot find the dev base image in Dockerfile.' >&2
  exit 1
fi
if ! docker pull "$base"; then
  if ! docker image inspect "$base" >/dev/null 2>&1; then
    echo "Cannot pull $base and no local base image is available. Authenticate to the registry and run just build again." >&2
    exit 1
  fi
  echo "Cannot pull $base; building with the base image already present locally." >&2
fi
docker compose build dev
