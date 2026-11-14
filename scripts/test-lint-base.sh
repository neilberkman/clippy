#!/bin/sh
set -eu

event=$(mktemp)
trap 'rm -f "$event"' EXIT
for kind in pull_request merge_group; do
  if [ "$kind" = pull_request ]; then
    printf '%s\n' '{"pull_request":{"base":{"sha":"not-a-valid-ref"}}}' >"$event"
  else
    printf '%s\n' '{"merge_group":{"base_sha":"not-a-valid-ref"}}' >"$event"
  fi
  if env -u LINT_BASE_REF GITHUB_EVENT_PATH="$event" bin/lint-go.sh >/dev/null 2>&1; then
    echo "lint accepted an invalid $kind base" >&2
    exit 1
  fi
done
base=$(git rev-parse HEAD)
printf '{"merge_group":{"base_sha":"%s"}}\n' "$base" >"$event"
env -u LINT_BASE_REF GITHUB_EVENT_PATH="$event" bin/lint-go.sh
