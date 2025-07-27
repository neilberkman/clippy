#!/bin/bash
set -e

cd "$(dirname "$0")"

# Build the executable with hardened runtime
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

# Sign the app with hardened runtime and proper entitlements
echo "Signing app..."
codesign --force --options runtime --deep --entitlements Draggy.entitlements --sign "-" "$APP_BUNDLE"

echo "✅ Built and signed $APP_BUNDLE"

# Verify the signature
echo "Verifying signature..."
codesign --verify --verbose "$APP_BUNDLE"

# Check if it's notarized (for future use)
echo ""
echo "To notarize this app for distribution:"
echo "1. Sign with a Developer ID certificate (not '-')"
echo "2. Create a zip: ditto -c -k --keepParent $APP_BUNDLE Draggy.zip"
echo "3. Submit for notarization: xcrun notarytool submit Draggy.zip --apple-id YOUR_EMAIL --team-id YOUR_TEAM_ID --wait"
echo "4. Staple the ticket: xcrun stapler staple $APP_BUNDLE"