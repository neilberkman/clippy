#!/bin/sh
# POSIX sh on purpose: bash 5.3 (Homebrew's default on macOS, and present on
# GitHub macOS runners) deadlocks on heredocs larger than ~512 bytes.
#
# Generates the server.json submitted to the official MCP registry
# (registry.modelcontextprotocol.io) for a release. Called from the
# publish-mcp-registry job in .github/workflows/release.yml with the
# release version and the SHA-256 of the published .mcpb asset.
#
# Not to be confused with the repo-root server.json, which is embedded
# into the binary as MCP server metadata (see server_metadata.go).
# Registry descriptions are limited to 100 characters.
set -eu

VERSION="$1"
SHA256="$2"

cat <<EOF
{
  "\$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.neilberkman/clippy",
  "title": "Clippy",
  "description": "Copy AI-generated content to the macOS clipboard: rich Gmail-ready email drafts, text, and files",
  "repository": {
    "url": "https://github.com/neilberkman/clippy",
    "source": "github"
  },
  "websiteUrl": "https://github.com/neilberkman/clippy",
  "version": "$VERSION",
  "packages": [
    {
      "registryType": "mcpb",
      "identifier": "https://github.com/neilberkman/clippy/releases/download/v${VERSION}/clippy_${VERSION}.mcpb",
      "fileSha256": "$SHA256",
      "transport": {
        "type": "stdio"
      }
    }
  ]
}
EOF
