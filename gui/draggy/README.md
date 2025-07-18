# Draggy

A minimal macOS menu bar app that bridges the terminal-to-GUI gap by making clipboard files draggable. Perfect companion to [clippy](https://github.com/neilberkman/clippy).

## Philosophy

Draggy is **not** a full-featured clipboard manager. It's a focused tool for terminal users who need just enough GUI to handle drag-and-drop workflows. 

If you live in your terminal but occasionally need to drag files to web browsers or native apps, this is for you. No history, no search, no fancy features - just a simple bridge from `clippy` to GUI drag targets.

## Features

- **Minimal UI**: Click to see files, drag them where needed
- **Zero background activity**: No polling, no battery drain
- **Terminal-first**: Designed to work seamlessly with clippy
- **Native performance**: Pure Swift, no Electron

## Usage

```bash
# In your terminal:
clippy *.png              # Copy files with clippy
curl -sL pic.jpg | clippy # Or pipe downloads

# In the GUI:
# 1. Click Draggy in menu bar
# 2. Drag files to browser upload fields, Slack, etc.
```

## Configuration

Right-click the menu bar icon for preferences:
- **Launch at login**
- **Show full paths** - Display complete file paths instead of just directories
- **Max files shown** - Limit for performance

## Building

```bash
swift build -c release
./build-app.sh
```

## Installation

### Via Homebrew (Recommended)

```bash
brew install --cask neilberkman/clippy/draggy
```

**Note:** macOS may show a security warning on first launch since Draggy isn't code-signed. The Homebrew cask handles this automatically, but if you see "Draggy is damaged", run:
```bash
xattr -dr com.apple.quarantine /Applications/Draggy.app
```

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/neilberkman/clippy/releases)
2. Unzip and drag Draggy.app to your Applications folder
3. Right-click and select "Open" to bypass Gatekeeper on first launch

## Requirements

- macOS 13.0+
- Files must be copied to clipboard as file references (clippy does this automatically)