#!/bin/bash
set -e

echo "Building clippy C library..."

# Build for macOS (darwin) on arm64 (Apple Silicon)
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build \
    -buildmode=c-archive \
    -o gui/draggy/libclippy.a \
    ./cbridge

echo "✅ Built gui/draggy/libclippy.a and gui/draggy/libclippy.h"
echo ""
echo "To use in Xcode:"
echo "1. The files are already in gui/draggy/"
echo "2. Add them to your Xcode project"
echo "3. Create a bridging header that imports libclippy.h"
echo "4. Call the exported functions from Swift"
