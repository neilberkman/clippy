#!/bin/bash
set -e

cd "$(dirname "$0")"

# Build the executable
echo "Building for production..."
swift build -c release --arch arm64

# Create app bundle structure
APP_NAME="Draggy"
APP_BUNDLE="$APP_NAME.app"
CONTENTS_DIR="$APP_BUNDLE/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

# Clean up old app bundle
rm -rf "$APP_BUNDLE"

# Create directories
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Copy executable
cp .build/release/Draggy "$MACOS_DIR/$APP_NAME"

# Copy Info.plist
cp Resources/Info.plist "$CONTENTS_DIR/"

# Create a simple icon (we'll use the system icon for now)
# In a real app, you'd create a proper .icns file

echo "✅ Built $APP_BUNDLE"
echo "To run: open $APP_BUNDLE"
