#!/bin/sh
set -eu
golangci-lint run --new-from-merge-base="${LINT_BASE_REF:-main}" "$@"
