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

# Copy your most recent download
clippy --recent          # Grabs the file you just downloaded

# Interactive picker for recent files
clippy --recent --pick   # Choose from multiple recent downloads

# Pipe data as files
curl -sL https://picsum.photos/300 | clippy  # Download → clipboard as file
```

Stay in your terminal. Copy anything. Paste anywhere.

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
clippy --recent        # Copy your most recent download
clippy --recent --pick # Interactive picker for recent files
clippy --recent 5m     # Only last 5 minutes
```

### 3. Pipe Data as Files
```bash
curl -sL https://example.com/image.jpg | clippy
cat archive.tar.gz | clippy
```

### 4. Helpful Flags
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
# File appears in your current directory
```

**2. Download in browser → Paste in terminal**
```bash
# 1. Download file in browser
# 2. Switch to terminal and run:
pasty --recent
# Most recent download appears in your current directory
```

**3. Interactive picker for recent files**
```bash
pasty --recent --pick    # Choose from multiple recent downloads
pasty --recent 5m        # Only last 5 minutes
```

**4. Text content handling**
```bash
# Copy text in any app, then:
pasty                    # Outputs text to stdout
pasty notes.txt          # Saves text to file
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

## License

MIT