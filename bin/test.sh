#!/bin/sh
set -eu
. ./bin/macos-env.sh
mkdir -p reports/test
gotestsum --junitfile=reports/test/junit.xml -- -count=1 -p 1 ./...
bash scripts/picker_snapshot_tmux.sh
