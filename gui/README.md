# GUI Components

This directory contains optional GUI applications that complement the clippy CLI tools.

## Draggy

A minimal menu bar app that makes clipboard files draggable. See [draggy/README.md](draggy/README.md) for details.

### Installation

```bash
# Build locally
cd draggy
./build-app.sh

# Or install via Homebrew (coming soon)
brew install --cask neilberkman/clippy/draggy
```

## Philosophy

These GUI tools are:
- **Optional** - The CLI tools work perfectly without them
- **Minimal** - Just enough UI to bridge terminal-to-GUI workflows
- **Focused** - Each tool does one thing well

They are NOT:
- Automatically installed with clippy
- Full-featured clipboard managers
- Required for any CLI functionality
