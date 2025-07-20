# Clippy 📎

Copy files from your terminal that actually paste into GUI apps. No more switching to Finder.

**macOS only** - built specifically for the Mac clipboard system.

## Why Clippy?

`pbcopy` copies file _contents_, but GUI apps need file _references_. When you `pbcopy < image.png`, you can't paste it into Slack or email - those apps expect files, not raw bytes.

Clippy bridges this gap by detecting what you want and using the right clipboard format:

```bash
# Copy files as references (paste into any GUI app)
clippy report.pdf         # ⌘V into Slack/email - uploads the file
clippy *.jpg             # Multiple files at once

# Pipe data as files
curl -sL https://picsum.photos/300 | clippy  # Download → clipboard as file

# Copy your most recent download (immediate)
clippy -r                # Grabs the file you just downloaded
clippy -r 3              # Copy 3 most recent downloads

# Interactive picker for recent files
clippy -i               # Choose from list of recent downloads
clippy -i 5m            # Show picker for last 5 minutes only
```

Stay in your terminal. Copy anything. Paste anywhere.

**The Terminal-First Clipboard Suite:** [Clippy](#core-features) copies files to clipboard, [Pasty](#pasty---intelligent-clipboard-pasting) pastes them intelligently, and [Draggy](#draggy---drag-and-drop-bridge) (optional GUI) bridges drag-and-drop workflows. Use as a [Go library](#library) for custom integrations. All designed to minimize context switching from your terminal.

💡 **Bonus:** Clippy includes an [MCP server](#mcp-server) for AI assistants like Claude to copy generated content directly to your clipboard.

## Installation

### Homebrew (Recommended)

```bash
brew install neilberkman/clippy/clippy
```

### Build from Source

```bash
# Clone and build
git clone https://github.com/neilberkman/clippy.git
cd clippy
go build -o clippy ./cmd/clippy
sudo mv clippy /usr/local/bin/

# Or use go install
go install github.com/neilberkman/clippy/cmd/clippy@latest
```

## Core Features

### 1. Smart File Copying

```bash
clippy document.pdf    # Copies as file reference (paste into any app)
clippy notes.txt       # Also copies as file reference
clippy -t notes.txt    # Use -t flag to copy text content instead
clippy *.jpg          # Multiple files at once
```

### 2. Recent Downloads

```bash
# Immediate copy (no UI)
clippy -r              # Copy your most recent download
clippy -r 3            # Copy 3 most recent downloads
clippy -r 5m           # Copy all downloads from last 5 minutes

# Interactive picker
clippy -i              # Choose from list of recent downloads
clippy -i 3            # Show picker with 3 most recent files
clippy -i 5m           # Show picker for last 5 minutes only

# Copy and paste in one step
clippy -r --paste      # Copy most recent and paste here
clippy -i --paste      # Pick file, copy it, and paste here
```

### 3. Pipe Data as Files

```bash
curl -sL https://example.com/image.jpg | clippy
cat archive.tar.gz | clippy
```

### 4. Copy and Paste Together

```bash
clippy file.txt --paste     # Copy to clipboard AND paste here
clippy -r --paste          # Copy recent download and paste here
clippy -i --paste           # Pick file, copy it, and paste here
```

### 5. Clear Clipboard

```bash
clippy --clear         # Empty the clipboard
echo -n | clippy       # Also clears the clipboard
```

### 6. Helpful Flags

```bash
clippy -v file.txt     # Show what happened
clippy --debug file.txt # Technical details for debugging
```

## Why "Clippy"?

Because it's a helpful clipboard assistant that knows what you want to do! 📎

---

## Pasty - Intelligent Clipboard Pasting

When you copy a file in Finder and press ⌘V in terminal, you just get the filename as text. Pasty actually copies the file itself to your current directory.

### Core Use Cases

**1. Copy file in Finder → Paste actual file in terminal**

```bash
# 1. Copy any file in Finder (⌘C)
# 2. Switch to terminal and run:
pasty
# File gets copied to your current directory (not just the filename!)
```

**2. Smart text file handling**

```bash
# Copy a text file in Finder (⌘C), then:
pasty                    # Outputs the file's text content to stdout
pasty notes.txt          # Saves the file's text content to notes.txt
```

---

## Install & Use

```bash
# Install via Homebrew
brew install neilberkman/clippy/clippy

# Or build from source
go install github.com/neilberkman/clippy/cmd/clippy@latest
go install github.com/neilberkman/clippy/cmd/pasty@latest
```

## Draggy - Drag and Drop Bridge

Sometimes you need to drag files to web upload fields or native apps. Draggy is a minimal menu bar app that makes clipboard files draggable.

**Important:** Draggy is a separate, optional tool. It's not automatically installed with clippy.

### Features

- **Minimal menu bar UI** - Click to see clipboard files, drag them where needed
- **Zero background activity** - No polling, no battery drain. Only checks clipboard when activated
- **Not a clipboard manager** - Draggy explicitly avoids history, search, or management features. It's just a bridge for drag-and-drop
- **Shows file icons** - Native macOS file icons for easy recognition

### Installation

```bash
# Separate brew install (not included with clippy)
brew install --cask neilberkman/clippy/draggy
```

**⚠️ First Launch:** macOS may show a security warning since Draggy isn't code-signed. If you see "Draggy is damaged and can't be opened":

- The Homebrew cask automatically removes the quarantine flag during installation
- If the warning persists, run: `xattr -dr com.apple.quarantine /Applications/Draggy.app`
- Or right-click Draggy.app and select "Open" to bypass Gatekeeper

### Usage

```bash
# In terminal:
clippy *.png              # Copy files with clippy
curl -sL pic.jpg | clippy # Or pipe downloads

# In GUI:
# 1. Click Draggy icon in menu bar
# 2. Drag files to browser upload fields, Slack, etc.
```

### Philosophy

Draggy is intentionally minimal. If you want a full-featured clipboard manager with history, search, and organization, use something else. Draggy is for terminal users who occasionally need to drag a file somewhere and want to get back to their terminal as quickly as possible.

## MCP Server

Clippy includes a built-in MCP (Model Context Protocol) server that lets AI assistants copy generated content directly to your clipboard.

Ask Claude to generate any text - code, emails, documents - and have it instantly available to paste anywhere:

- "Write a Python script to process CSV files and copy it to my clipboard"
- "Draft an email about the meeting and put it on my clipboard"
- "Generate that regex and copy it so I can paste into my editor"

No more manual selecting and copying from the chat interface.

### Setup

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "clippy": {
      "command": "clippy",
      "args": ["mcp-server"]
    }
  }
}
```

### Available Tools

- **clipboard_copy** - Copy text or files to clipboard
- **clipboard_paste** - Paste clipboard content to files/directories
- **get_recent_downloads** - List recently downloaded files

Claude can generate content and put it directly on your clipboard, ready to paste wherever you need it.

## Library

[Clippy](#core-features) can be used as a Go library in your own applications:

```bash
go get github.com/neilberkman/clippy
```

### High-Level API

```go
import "github.com/neilberkman/clippy"

// Smart copy - automatically detects text vs binary files
err := clippy.Copy("document.pdf")

// Copy multiple files as references
err := clippy.CopyMultiple([]string{"image1.jpg", "image2.png"})

// Copy text content
err := clippy.CopyText("Hello, World!")

// Copy data from reader (handles text/binary detection)
reader := strings.NewReader("Some content")
err := clippy.CopyData(reader)

// Copy from stdin
err := clippy.CopyData(os.Stdin)

// Get clipboard content
text, ok := clippy.GetText()
files := clippy.GetFiles()
```

### Features

- **Smart Detection**: Automatically determines whether to copy as file reference or text content
- **Multiple Files**: Copy multiple files in one operation
- **Reader Support**: Copy from any io.Reader with automatic format detection
- **Clipboard Access**: Read current clipboard content (text or file paths)
- **Cross-Platform Types**: Uses standard Go types, handles platform-specific clipboard internally

## License

MIT
