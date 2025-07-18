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

💡 **Bonus:** Clippy includes an [MCP server](#mcp-server) for AI assistants like Claude to manage your clipboard.

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
clippy notes.txt       # Copies text content
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

### 5. Helpful Flags

```bash
clippy -v file.txt     # Show what happened
clippy --debug file.txt # Technical details for debugging
```

---

## Pasty - Intelligent Clipboard Pasting

Pasty is clippy's companion tool for intelligent pasting from the clipboard.

### Core Use Cases

**1. Copy file in Finder → Paste in terminal**

```bash
# 1. Copy any file in Finder (⌘C)
# 2. Switch to terminal and run:
pasty
# File gets copied to your current directory
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

## Why "Clippy"?

Because it's a helpful clipboard assistant that knows what you want to do! 📎

## MCP Server

Clippy includes a built-in MCP (Model Context Protocol) server that allows AI assistants like Claude to interact with your clipboard programmatically.

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

Now Claude can help you manage your clipboard, create and copy files, and work with your recent downloads!

## License

MIT
