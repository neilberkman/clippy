#!/bin/sh
set -eu
# shellcheck source=/dev/null
. ./bin/macos-env.sh
base=${LINT_BASE_REF:-}
if [ -z "$base" ] && [ -n "${GITHUB_EVENT_PATH:-}" ]; then
  base=$(python3 -c 'import json, sys; e = json.load(open(sys.argv[1])); print((e.get("pull_request") or {}).get("base", {}).get("sha") or (e.get("merge_group") or {}).get("base_sha") or "")' "$GITHUB_EVENT_PATH")
fi
if [ -z "$base" ]; then
  if git show-ref --verify --quiet refs/heads/main; then base=main; else base=origin/main; fi
fi
golangci-lint run --new-from-merge-base="$base" "$@"
