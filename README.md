# Clippy 📎

A superset of macOS's `pbcopy` that intelligently handles both text and files - the clipboard tool that just works.

## Key Features

- **Drop-in pbcopy replacement**: All your existing scripts work unchanged
- **Smart file handling**: Automatically detects whether to copy content or file references
- **Multiple file support**: `clippy *.pdf` copies all PDFs as file references
- **Natural language**: `clippy "my file.txt"` works with spaces, wildcards, and shell expansion
- **Binary streaming**: Pipe images, PDFs, etc. and paste them directly into apps
- **Automatic cleanup**: Temporary files are cleaned up intelligently in the background

## The Problem

macOS has two different clipboard tools:
- `pbcopy` - copies text content to clipboard
- File dragging/`Cmd+C` - copies file references to clipboard

This creates confusion:
- Which tool should you use?
- `pbcopy < file.pdf` doesn't let you paste the PDF into apps
- Dragging files to terminal is clunky
- No unified command-line interface

## The Solution

Clippy automatically detects what you want to copy and does the right thing:

```bash
# Text files → copies content (like pbcopy)
clippy notes.txt
# Pastes: "These are my notes..."

# Binary files → copies file reference
clippy image.png
# Pastes: [image.png as attachment]

# Piped text → copies content
echo "Hello" | clippy
# Pastes: "Hello"

# Piped binary → saves to temp file, copies reference
cat document.pdf | clippy
# Pastes: [document.pdf as attachment]
```

## Installation

```bash
# Build from source
go build -o clippy .
sudo mv clippy /usr/local/bin/

# Or use go install
go install github.com/yourusername/clippy@latest
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
curl -s https://example.com/image.png | clippy

# Silent operation for scripts
clippy data.txt && echo "Copied!"

# Verbose mode to see what happened
clippy -v presentation.pptx
```

## Testing

```bash
# Run tests
go test -v

# Run with race detector
go test -race
```

## Why "Clippy"?

Because it's a helpful clipboard assistant that knows what you want to do! 📎

## License

MIT