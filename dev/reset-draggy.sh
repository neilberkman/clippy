#!/bin/bash

echo "🔄 Resetting Draggy..."

# Kill any running Draggy process
echo "  → Killing Draggy process..."
pkill -f Draggy || true

# Delete all preferences
echo "  → Deleting preferences..."
defaults delete com.neilberkman.draggy 2>/dev/null || true
rm -rf ~/Library/Preferences/com.neilberkman.draggy* 2>/dev/null || true

# Kill preferences daemon to ensure clean slate
echo "  → Restarting preferences daemon..."
killall cfprefsd

# Wait for everything to settle
sleep 2

# Navigate to draggy directory
DRAGGY_DIR="$(dirname "$0")/../gui/draggy"
cd "$DRAGGY_DIR"

# Rebuild if requested
if [ "$1" == "--build" ]; then
    echo "  → Building Go library..."
    cd ../.. && ./build-clib.sh && cd gui/draggy
    echo "  → Building Draggy..."
    ./build-app.sh
fi

# Open Draggy
echo "  → Opening Draggy..."
open Draggy.app

echo "✅ Draggy reset complete!"
