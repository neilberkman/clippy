#!/bin/sh
set -eu
# shellcheck source=/dev/null
. ./bin/native-env.sh
golangci-lint run --new-from-merge-base="${LINT_BASE_REF:-main}" "$@"
