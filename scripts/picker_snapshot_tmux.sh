#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOCKET_NAME="clippy-snapshot-$RANDOM-$RANDOM"
SESSION_NAME="picker-snapshot"
BEGIN_MARKER="===PICKER_SNAPSHOT_BEGIN==="
END_MARKER="===PICKER_SNAPSHOT_END==="
EXPECTED_FILE="$ROOT_DIR/cmd/clippy/testdata/picker_snapshot.txt"

cleanup() {
  tmux -L "$SOCKET_NAME" kill-server >/dev/null 2>&1 || true
}
trap cleanup EXIT

tmux -L "$SOCKET_NAME" -f /dev/null new-session -d -s "$SESSION_NAME" -x 120 -y 40 \
  "cd '$ROOT_DIR' && CLIPPY_SNAPSHOT_PRINT=1 go test ./cmd/clippy -run TestPickerSnapshotPrint -count=1 -v; echo __CLIPPY_SNAPSHOT_DONE__; sleep 2"

for _ in $(seq 1 120); do
  pane="$(tmux -L "$SOCKET_NAME" capture-pane -p -t "$SESSION_NAME" || true)"
  if grep -q "__CLIPPY_SNAPSHOT_DONE__" <<<"$pane"; then
    break
  fi
  sleep 0.2
done

pane="$(tmux -L "$SOCKET_NAME" capture-pane -p -t "$SESSION_NAME" || true)"
captured="$(awk -v begin="$BEGIN_MARKER" -v end="$END_MARKER" '
  $0==begin {flag=1; next}
  $0==end {flag=0; exit}
  flag {print}
' <<<"$pane")"

if [[ -z "$captured" ]]; then
  echo "Failed to capture picker snapshot from tmux pane."
  echo "Captured pane:"
  echo "$pane"
  exit 1
fi

if [[ ! -f "$EXPECTED_FILE" ]]; then
  echo "Expected snapshot file not found: $EXPECTED_FILE"
  exit 1
fi

tmpfile="$(mktemp)"
printf '%s' "$captured" >"$tmpfile"

if ! diff -u "$EXPECTED_FILE" "$tmpfile"; then
  echo
  echo "Picker snapshot differs from golden file."
  echo "If expected, run: UPDATE_SNAPSHOTS=1 go test ./cmd/clippy -run TestPickerSnapshotGolden"
  rm -f "$tmpfile"
  exit 1
fi

rm -f "$tmpfile"
echo "Picker tmux snapshot matches golden output."
