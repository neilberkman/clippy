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
clippy -r                # Grabs the file you just downloaded
clippy -r 5m             # Only last 5 minutes

# Interactive picker for recent files
clippy -r --pick         # Choose from multiple recent downloads

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
clippy -r              # Copy your most recent download
clippy -r --pick       # Interactive picker for recent files
clippy -r 5m           # Only last 5 minutes
clippy -r --paste      # Copy and paste in one step
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
clippy -r --pick --paste   # Pick file, copy it, and paste here
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

## License

MIT
