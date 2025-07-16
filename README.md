# Clippy 📎

Copy files from your terminal that actually paste into GUI apps. No more switching to Finder.

**macOS only** - built specifically for the Mac clipboard system.

**Includes [pasty](#pasty---intelligent-clipboard-pasting)** - the companion paste tool for intelligent clipboard reading and automation workflows.

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
- **Library support**: Use clippy as a Go library in your own applications
- **Companion tool**: Includes [`pasty`](#pasty---intelligent-clipboard-pasting) for intelligent pasting and automation workflows

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
clippy --help               # Show usage and examples
clippy -h                   # Short version
clippy --version            # Show version information
```

### Configuration

Create `~/.clippy.conf` for persistent settings:

```ini
# Enable verbose output by default
verbose = true

# Disable automatic cleanup of temporary files
cleanup = false

# Use custom directory for temporary files
temp_dir = /path/to/custom/temp
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
   - Detects MIME type using content analysis (not just file extension)
   - Text files → reads content, copies as text
   - Other files → copies file path reference

2. **Stream Mode** (when you pipe data):
   - Detects MIME type from content
   - Text data → copies as text
   - Binary data → saves to temp file, copies reference

3. **MIME Type Detection**:
   - Uses content-based detection, not file extensions
   - Anything with MIME type `text/*` is treated as text
   - This means `.log`, `.conf`, `.json`, etc. are correctly identified as text
   - Binary files are identified by their actual content (e.g., PNG magic bytes)

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

## Pasty - Intelligent Clipboard Pasting

Pasty is clippy's companion tool for intelligent pasting from the clipboard:

```bash
# Paste clipboard content to stdout
pasty

# Paste to a specific file
pasty output.txt

# Paste with verbose output
pasty -v

# Copy files from clipboard to directory
pasty /path/to/destination/
```

How it works:

- **Text content**: Pastes directly to stdout or file
- **File references**: Lists file paths or copies files to destination
- **Smart detection**: Automatically handles text vs file clipboard content

## Real-World Examples

### 📸 Screenshots & In-Memory Content

The killer app: Handle clipboard content that has no file path.

```bash
# Take a screenshot (Cmd+Ctrl+Shift+4), then process it:
pasty | convert - -resize 50% screenshot-small.png

# Screenshot directly to S3
pasty | aws s3 cp - s3://my-bucket/screenshot-$(date +%s).png

# Extract text from screenshot using OCR
pasty | tesseract - - | pbcopy
```

### 🔄 Scripting with Multiple Files

Copy files in Finder, then process them with clean, scriptable output.

```bash
# Copy 10 files in Finder, then process each one:
for file in $(pasty); do
  echo "Processing: $(basename "$file")"
  # your processing here
done

# Parallel processing with xargs
pasty | xargs -I {} stat -f "%z bytes: %N" "{}"

# Archive all copied files
pasty | xargs tar -czf backup-$(date +%Y%m%d).tar.gz

# Find large files from copied selection
pasty | xargs -I {} find "{}" -size +10M -type f
```

### 🚀 Automation & Workflows

Bridge GUI → CLI workflows that are impossible otherwise.

```bash
# Copy log files in Finder, then analyze
pasty | xargs grep -l "ERROR" | head -5

# Batch convert images copied from Finder
pasty | xargs -I {} convert "{}" -quality 85 "{%.jpg}"

# Upload copied files to server
pasty | xargs -I {} scp "{}" user@server:/backup/

# Process copied code files
pasty | xargs -I {} prettier --write "{}"
```

## Testing

```bash
# Run tests
go test -v ./...
```

## Using as a Library

Clippy can be used as a Go library in your own applications:

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
clippy.CopyText("Hello, World!")

// Copy data from reader (handles text/binary detection)
reader := strings.NewReader("Some content")
err := clippy.CopyData(reader)

// Copy from stdin
err := clippy.CopyData(os.Stdin)

// Get clipboard content
text, ok := clippy.GetText()
files := clippy.GetFiles()
```

### Low-Level Clipboard API

```go
import "github.com/neilberkman/clippy/pkg/clipboard"

// Direct clipboard operations
clipboard.CopyText("Hello")
clipboard.CopyFile("/path/to/file.pdf")
clipboard.CopyFiles([]string{"/path/to/file1.jpg", "/path/to/file2.png"})

// Read clipboard
text, ok := clipboard.GetText()
files := clipboard.GetFiles()
```

The high-level API provides the same smart detection and temp file management as the CLI tool, while the low-level API gives you direct access to clipboard operations without any automatic behavior.

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
