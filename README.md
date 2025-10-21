# Clippy 📎

[![Homebrew](https://img.shields.io/homebrew/v/clippy?color=FBB040)](https://formulae.brew.sh/formula/clippy)
[![Release](https://img.shields.io/github/v/release/neilberkman/clippy)](https://github.com/neilberkman/clippy/releases)
[![CI](https://github.com/neilberkman/clippy/actions/workflows/release.yml/badge.svg)](https://github.com/neilberkman/clippy/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/neilberkman/clippy)](https://go.dev/)
[![License](https://img.shields.io/github/license/neilberkman/clippy)](https://github.com/neilberkman/clippy/blob/main/LICENSE)

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

**Installation:**

```bash
brew install clippy
```

**The Terminal-First Clipboard Suite:** [Clippy](#core-features) copies files to clipboard, includes an [MCP server](#mcp-server) for AI assistants, [Pasty](#pasty---intelligent-clipboard-pasting) pastes intelligently, and [Draggy](#draggy---visual-clipboard-companion) (optional GUI) bridges drag-and-drop workflows. Use as a [Go library](#library) for custom integrations. All designed to minimize context switching from your terminal.

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
clippy ~/Downloads/report.pdf --paste  # Copy to clipboard AND paste here
clippy -r --paste          # Copy recent download and paste here
clippy -i --paste           # Pick file, copy it, and paste here
```

### 5. Clear Clipboard

```bash
clippy --clear         # Empty the clipboard
echo -n | clippy       # Also clears the clipboard
```

### 6. Content Type Detection

A nice bonus: clippy auto-detects content types (JSON, HTML, XML) so receiving apps handle them properly - something `pbcopy` can't do. This means when you paste into apps that support rich content, they'll handle it correctly - JSON viewers will syntax highlight, HTML will render, etc.

```bash
echo '{"key": "value"}' | clippy          # Recognized as JSON
clippy -t page.html                       # Recognized as HTML
clippy -t file.txt --mime application/json  # Manual override when needed
```

### 7. Helpful Flags

```bash
clippy -v file.txt     # Show what happened
clippy --debug file.txt # Technical details for debugging
```

## Why "Clippy"?

Because it's a helpful clipboard assistant that knows what you want to do! 📎

---

## MCP Server

Clippy includes a built-in MCP (Model Context Protocol) server that lets AI assistants copy generated content directly to your clipboard.

Ask Claude to generate any text - code, emails, documents - and have it instantly available to paste anywhere:

- "Write a Python script to process CSV files and copy it to my clipboard"
- "Draft an email about the meeting and put it on my clipboard"
- "Generate that regex and copy it so I can paste into my editor"

No more manual selecting and copying from the chat interface.

### Setup

**Claude Code:**

```bash
# Install for all your projects (recommended)
claude mcp add --scope user clippy $(which clippy) mcp-server

# Or for current project only
claude mcp add clippy $(which clippy) mcp-server
```

Note: `$(which clippy)` finds the clippy binary on your system. On Apple Silicon Macs this is typically `/opt/homebrew/bin/clippy`, on Intel Macs it's `/usr/local/bin/clippy`.

**Claude Desktop:**

Add to your config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

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

#### System Clipboard Tools

- **clipboard_copy** - Copy text or files to system clipboard
- **clipboard_paste** - Paste clipboard content to files/directories
- **get_recent_downloads** - List recently downloaded files

#### Agent Buffer Tools

- **buffer_copy** - Copy file bytes (with optional line ranges) to agent's private buffer
- **buffer_cut** - Cut lines from file to buffer (copy and delete from source)
- **buffer_paste** - Paste bytes to file with append/insert/replace modes
- **buffer_list** - Show buffer metadata (lines, source file, range)

**Why buffer tools?** Solves the LLM "remember and re-emit" problem. The MCP server reads/writes file bytes directly - agents never generate tokens for copied content. Enables surgical refactoring (copy lines 17-32, paste to replace lines 5-8) with byte-for-byte accuracy, without touching your system clipboard.

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

**3. Save browser images**

```bash
# Right-click "Copy Image" in any browser, then:
pasty photo.png          # Saves the image (auto-converts TIFF to PNG)
pasty --preserve-format  # Keep original format if needed
```

Also handles rich text with embedded images (`.rtfd` bundles from TextEdit/Notes).

**4. Debugging and plain text extraction**

```bash
pasty --inspect          # Show what's on clipboard and what pasty will use
pasty --plain notes.txt  # Force plain text, strip all formatting
```

---

## Draggy - Visual Clipboard Companion

Draggy is a menu bar app that brings visual functionality to your clipboard workflow. While clippy handles copying from the terminal, Draggy provides a visual interface for dragging files to applications and viewing recent downloads.

**Important:** Draggy is a separate, optional tool. It's not automatically installed with clippy.

### Features

#### Core Functionality

- **Drag & Drop Bridge** - Makes clipboard files draggable to web browsers, Slack, and other apps
- **Recent Downloads Viewer** - Toggle between clipboard and recent downloads with one click
- **File Thumbnails** - Visual previews for images and PDFs right in the file list
- **Quick Preview** - Hold ⌥ Option while hovering to see larger previews
- **Zero Background Activity** - No polling or battery drain, only activates on demand

#### User Experience

- **Double-Click to Open** - Quick access to files without leaving the menu
- **Keyboard Shortcuts** - ESC to close, Space to refresh

#### Design Philosophy

- **Not a clipboard manager** - No history, no database, no complexity
- **Terminal-first workflow** - Designed to complement terminal usage, not replace it
- **Minimal but complete** - Every feature serves a specific workflow need

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
# Copy files in terminal:
clippy ~/Downloads/*.pdf    # Copy PDFs with clippy
curl -sL image.jpg | clippy # Pipe downloads to clipboard
clippy -r                   # Copy most recent download

# Use Draggy GUI:
# 1. Click Draggy icon in menu bar
# 2. Drag files to browser upload fields, Slack, etc.
# 3. Toggle to Recent Downloads view with clock icon
# 4. Hold ⌥ Option to preview files
# 5. Double-click to open files
```

### Workflow Examples

**Upload screenshots to GitHub:**

```bash
# Take screenshot (macOS saves to Desktop)
# In terminal: clippy ~/Desktop/Screenshot*.png
# In Draggy: Drag to GitHub comment box
```

**Quick file sharing:**

```bash
# Terminal: clippy ~/Downloads/report.pdf
# Draggy: Shows thumbnail, drag to Slack or email
```

**Recent downloads workflow:**

```bash
# Download file in browser
# Click Draggy → Click clock icon → See your download
# Drag where needed or double-click to open
```

### Philosophy

Draggy is intentionally not a clipboard manager. No history, no search, no database. It's a visual bridge between your terminal clipboard workflow and GUI applications. For terminal users who occasionally need to see what's on their clipboard or drag files somewhere, then get back to work.

---

## Build from Source

```bash
# Clone and build
git clone https://github.com/neilberkman/clippy.git
cd clippy
go build -o clippy ./cmd/clippy
go build -o pasty ./cmd/pasty
sudo mv clippy pasty /usr/local/bin/

# Or use go install
go install github.com/neilberkman/clippy/cmd/clippy@latest
go install github.com/neilberkman/clippy/cmd/pasty@latest
```

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
