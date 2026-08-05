#!/bin/sh
# POSIX sh on purpose: bash 5.3 (Homebrew's default on macOS, and present on
# GitHub macOS runners) deadlocks on heredocs larger than ~512 bytes.
#
# Assembles the MCPB bundle (https://github.com/anthropics/mcpb) from the
# darwin universal binary GoReleaser has already built. Runs as a GoReleaser
# universal_binaries post hook. Clippy is macOS-only, so the bundle carries
# a single universal binary.
set -eu

VERSION="$1"
DIST="${2:-dist}"

STAGE="$DIST/mcpb-stage"
rm -rf "$STAGE"
mkdir -p "$STAGE/bin"

cp "$DIST"/clippy*_darwin_all/clippy "$STAGE/bin/clippy"
chmod +x "$STAGE/bin/clippy"

sed "s/__VERSION__/$VERSION/" > "$STAGE/manifest.json" <<'EOF'
{
  "manifest_version": "0.3",
  "name": "clippy",
  "display_name": "Clippy",
  "version": "__VERSION__",
  "description": "Copy AI-generated content to the macOS clipboard: Gmail-ready rich email drafts, text, code, files, and recent downloads.",
  "author": {
    "name": "Neil Berkman",
    "email": "neil@xuku.com",
    "url": "https://github.com/neilberkman"
  },
  "homepage": "https://github.com/neilberkman/clippy",
  "repository": {
    "type": "git",
    "url": "https://github.com/neilberkman/clippy"
  },
  "license": "MIT",
  "keywords": ["clipboard", "macos", "email", "gmail", "files"],
  "server": {
    "type": "binary",
    "entry_point": "bin/clippy",
    "mcp_config": {
      "command": "${__dirname}/bin/clippy",
      "args": ["mcp-server"]
    }
  },
  "compatibility": {
    "platforms": ["darwin"]
  }
}
EOF

OUT="clippy_${VERSION}.mcpb"
(cd "$STAGE" && zip -qr "../$OUT" .)
rm -rf "$STAGE"
echo "built $DIST/$OUT"
