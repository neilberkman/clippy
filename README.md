# Clippy 📎

Copy files from your terminal that actually paste into GUI apps. No more switching to Finder.

**macOS only** - built specifically for the Mac clipboard system.

## Why Clippy?

`pbcopy` copies file _contents_, but GUI apps need file _references_. When you `pbcopy < image.png`, you can't paste it into Slack or email - those apps expect files, not raw bytes. Your only option is to leave the terminal and drag files from Finder.

Clippy bridges this gap by detecting what you want and using the right clipboard format:

```bash
# Copy files as references (for pasting into GUI apps)
clippy report.pdf         # ⌘V into any app - uploads the file
clippy *.jpg             # Multiple files at once

# Copy piped/streamed data as files
curl -s https://picsum.photos/300 | clippy       # Download → clipboard as file
cat archive.tar.gz | clippy                      # Pipe → paste as file

# Smart detection for text files
clippy notes.txt          # Copies text content
echo "Hello" | clippy     # Works like pbcopy for piped text
```

Stay in your terminal. Copy anything. Paste anywhere.

## Key Features

- **Smart detection**: Automatically decides whether to copy content or file references
- **File support**: `clippy document.pdf` copies the PDF as a file (not its raw content)
- **Multiple files**: `clippy *.jpg` copies all JPGs as file references
- **Binary streaming**: Pipe images/PDFs and paste them as files in other apps
- **Text handling**: Text files are copied as content, just like pbcopy
- **Automatic cleanup**: Temporary files from piped data are cleaned up intelligently

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
go build -o clippy .
sudo mv clippy /usr/local/bin/

# Or use go install
go install github.com/neilberkman/clippy@latest
```

## Usage

### Basic Usage

```bash
# Copy a file
clippy myfile.txt           # Text content
clippy photo.jpg            # File reference

# Pipe data
echo "Hello, World!" | clippy
cat image.png | clippy
```

### Flags

```bash
clippy --verbose file.txt   # Show success messages
clippy -v file.txt          # Short version
```

### Configuration

Create `~/.clippy.conf` for persistent settings:

```ini
# Enable verbose output by default
verbose = true

```

## Features

### Smart Detection

- **Text files** (`.txt`, `.md`, code files) → copies actual text
- **Binary files** (images, PDFs, zips) → copies as file reference
- Works with both file arguments and piped input

### Silent by Default

- No output unless `--verbose` flag is used
- Perfect for scripts and automation
- Errors still shown on stderr

### Automatic Cleanup

- Temporary files created for piped binary data are managed automatically
- Background process scans for orphaned temp files on each run
- Checks if temp files are still referenced in clipboard before removal
- Non-blocking - cleanup happens while main operation proceeds

## How It Works

1. **File Mode** (when you pass a filename):
   - Detects MIME type
   - Text files → reads content, copies as text
   - Other files → copies file path reference

2. **Stream Mode** (when you pipe data):
   - Detects MIME type from content
   - Text data → copies as text
   - Binary data → saves to temp file, copies reference

## Examples

```bash
# Copy code file as text
clippy main.go

# Copy PDF as file reference
clippy report.pdf

# Copy command output
ls -la | clippy

# Copy image from curl
curl -s https://picsum.photos/300 | clippy

# Silent operation for scripts
clippy data.txt && echo "Copied!"

# Verbose mode to see what happened
clippy -v recording.mp4
```

## Testing

```bash
# Run tests
go test -v

# Run with race detector
go test -race
```

## Alternatives

### gcopy

[gcopy](https://github.com/TheDen/gcopy) is a cross-platform clipboard tool that also works with files. Key differences:

- **Cross-platform**: gcopy works on Linux/Windows, clippy is macOS-only
- **Implementation**: clippy uses native macOS APIs, gcopy uses AppleScript
- **File handling**: clippy copies file references (like Finder), gcopy copies file contents - this makes clippy work for pasting images/PDFs into apps, while gcopy doesn't
- **Multiple files**: clippy supports `*.jpg`, gcopy handles one file at a time
- **Cleanup**: clippy auto-cleans temp files, gcopy doesn't

Choose gcopy if you need cross-platform support. Choose clippy for better macOS integration and multi-file support.

## Why "Clippy"?

Because it's a helpful clipboard assistant that knows what you want to do! 📎

## License

MIT
