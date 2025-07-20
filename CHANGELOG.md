# Changelog

Notable changes to clippy.

## [1.2.0] - 2025-01-20

### Added

- **Clear clipboard**: New `--clear` flag to explicitly empty the clipboard
  - `clippy --clear` empties the clipboard
  - Empty input (e.g., `echo -n | clippy`) also clears the clipboard
- **MCP server improvements**: Better guidance for AI tools on when to copy text vs file references
  - Clearer tool descriptions to prevent incorrect parameter usage

### Changed

- **BREAKING**: Changed default behavior to always copy files as file references, not content
  - Text files (.txt, .md, etc.) now copy as files by default
  - Added `--text` / `-t` flag to force copying text file content
  - This allows pasting text files into Finder/GUI apps as files

### Draggy [0.10.1]

- **Version checking**: Check for updates when opening the app (max once every 2 hours)
  - Respects Homebrew installations with appropriate update instructions
  - Non-intrusive notification bar at the top of the window
- **Onboarding improvements**: Fixed permission dialog UI issues
  - Fixed text truncation in permission messages
  - Fixed buttons not hiding after clicking Enable
  - Fixed window expanding to massive size
  - Improved visual design with borders and better spacing
- **File previews**: Thumbnail generation for images and PDFs
  - Uses QuickLook, native image rendering, and PDF rendering
  - Smart caching to minimize performance impact
  - Toggle in preferences
- **Double-click to open**: Quick file access without leaving the menu bar
- **UI improvements**: Hand cursor on hover, helpful tooltips
- **Bug fixes**:
  - Fixed drag-and-drop using simple NSItemProvider approach
  - Fixed clipboard persistence (files no longer disappear on restart)
  - Added ESC key support to close popover menu
  - Auto-show recent downloads when clipboard is empty
  - Fixed permission dialog z-order issues

## [0.9.0] - 2025-01-18

### Added

- **Draggy**: New optional GUI companion app for drag-and-drop workflows
  - Menu bar app that makes clipboard files draggable
  - Zero background activity - no polling, no battery drain
  - Event-driven architecture following Core vs Interface principles
  - Located in `gui/draggy/` directory
  - Not automatically installed with clippy (will be separate brew cask)

### Changed

- Updated MCP server config to use Homebrew-installed clippy path

### Fixed

- Fixed linter warnings in examples and main code

## [0.8.0] - 2025-01-18

### Added

- **MCP (Model Context Protocol) server**: New `mcp-server` subcommand enables AI assistants to use clippy
  - Integrate with Claude Desktop, Cursor, or any MCP-compatible AI tool
  - Three powerful tools exposed:
    - `clipboard_copy`: Copy text or files to clipboard programmatically
    - `clipboard_paste`: Paste clipboard content to files/directories
    - `get_recent_downloads`: List and work with recently downloaded files
  - Built with github.com/mark3labs/mcp-go for standard MCP compliance
  - Includes pre-built prompts for common operations
  - Simple setup: just add to claude_desktop_config.json
- **New `-i` flag for interactive picker**: Clean separation between immediate copy (`-r`) and interactive mode (`-i`)
  - `clippy -i` - show interactive picker with recent files
  - `clippy -i 3` - show picker with 3 most recent files
  - `clippy -i 5m` - show picker for files from last 5 minutes
  - Both `-r` and `-i` accept the same arguments (DRY implementation)
- **Lowercase 'p' key in picker**: Changed from 'P' to 'p' for copy & paste mode (easier to type)
- **Text paste to directories**: Pasting text content to a directory now creates timestamped .txt files

### Changed

- **Major architectural refactor**: Following Saša Jurić's Core vs Interface philosophy
  - Moved Bubble Tea picker from `pkg/recent/` to `cmd/clippy/`
  - Core library (pkg/) now only provides data and business logic
  - Interface elements (UI/TUI) live in cmd/ where they belong
  - This creates proper separation of concerns
- **Simplified `-r` flag behavior**: Now always does immediate copy (no picker)
  - `clippy -r` - copies most recent download immediately
  - `clippy -r 3` - copies 3 most recent downloads immediately
  - Use `-i` flag for interactive picker mode
- **Bubble Tea picker improvements**: Now supports both single and multi-select seamlessly
  - Space to toggle selection
  - Enter to copy (selected items or current item if nothing selected)
  - 'p' to copy & paste in one operation

### Fixed

- **Clipboard synchronization bug**: Fixed "Heisenbug" where clipboard operations only worked with --debug flag
  - Replaced hacky sleep with proper NSPasteboard changeCount polling
  - Clipboard operations now wait for macOS to confirm completion
  - Uses NSRunLoop to properly handle asynchronous pasteboard operations
  - This is the correct way to handle clipboard operations in macOS CLI tools
- **MCP tool naming**: Fixed validation error by changing tool names from forward slashes to underscores
- **Directories in picker**: Fixed issue where directories appeared in the file picker

### Removed

- **Removed --batch flag**: Functionality integrated into numbered copies (e.g., `clippy -r 3`)
- **Removed --pick flag**: Use `-i` for interactive picker instead
- **Removed picker from pkg/recent**: UI components now properly live in cmd/

## [0.7.2] - 2025-07-17

### Changed

- **Major architectural refactor**: Moved all recent downloads functionality from pasty to clippy
  - Recent downloads is about file selection, which belongs in clippy (copy TO clipboard)
  - Pasty now focuses purely on pasting FROM clipboard
  - This creates cleaner separation of concerns
- **New `--paste` flag in clippy**: Copy and paste in one step
  - `clippy file.txt --paste` - copy to clipboard AND paste to current directory
  - `clippy -r --paste` - copy recent download and paste here
  - `clippy -r --pick --paste` - pick file, copy it, and paste here

### Fixed

- **Picker functionality**: Fixed `--pick` flag that was broken in v0.7.1 due to buggy wrapper function
- **DRY refactor**: Removed `PastePickedRecentDownload()` wrapper, both tools now use shared `PickRecentDownload()` function
- **Public API**: Made `CopyFileToDestination()` public for shared use between tools
- **Test fixes**: Fixed hardcoded test expectations and handling of filenames with spaces

### Removed

- **Recent downloads in pasty**: All `-r/--recent` functionality removed from pasty
  - Use `clippy -r --paste` instead of `pasty -r`
  - This is a breaking change but creates better architecture

## [0.7.1] - 2025-07-17

### Changed

- **Streamlined recent downloads UX**: `-r/--recent` now accepts optional time duration directly
  - `clippy -r 5m` instead of `clippy --recent --recent-time 5m`
  - `pasty -r 1h` instead of `pasty --recent --recent-time 1h`
  - Backward compatibility maintained with `--recent` long form
- **Updated documentation**: All examples now show the streamlined `-r` usage pattern

## [0.7.0] - 2025-07-17

### Added

- **Recent downloads functionality**: Copy your most recent downloads without leaving the terminal
  - `clippy -r` / `pasty -r` - copy most recent download
  - `clippy -r 5m` / `pasty -r 5m` - time-based filtering
  - `clippy -r --batch` / `pasty -r --batch` - copy all files from recent batch download
  - `clippy -r --pick` / `pasty -r --pick` - interactive picker using promptui
- **Smart Downloads folder detection**: Automatically finds and scans macOS Downloads folder
- **Batch download handling**: Groups files downloaded within 30 seconds for batch operations
- **Archive auto-detection**: Smart handling of common archive types (.zip, .tar.gz, .dmg, etc.)
- **Separate debug and verbose modes**: `--debug` for technical details, `--verbose` for user-friendly output
- **Enhanced logging**: Debug mode shows technical details while verbose shows user-friendly messages

### Technical

- **Cobra CLI framework**: Professional command-line interface with comprehensive help
- **Library-first architecture**: All functionality available as Go library functions
- **Platform-specific build constraints**: Proper macOS-only compilation

## [0.6.1] - 2025-07-17

### Fixed

- Pasty now correctly prioritizes file references over text content when both are present on clipboard
- When you copy a file from Finder, pasty now outputs the file path instead of just the filename

### Changed

- Pasty now defaults to copying files to current directory when no destination is specified (equivalent to `pasty .`)

## [0.6.0] - 2025-07-16

### Added

- **UTI detection**: Hybrid detection system using macOS UTI → MIME → mimetype fallback for maximum reliability
- **Enhanced verbose output**: Shows which detection method was used (UTI, MIME, or content analysis)
- **CI for pull requests**: GitHub Actions workflow runs tests, linting, and builds on all PRs
- **Pasty library architecture**: Refactored to use `clippy.PasteToStdout()` and `clippy.PasteToFile()` functions

### Changed

- More accurate type detection using macOS's native UTI system instead of relying solely on MIME types
- Pasty now uses library functions for DRY architecture (thin CLI over rich library functionality)
- Updated all tests to match new verbose output format
- Improved error handling and code quality

### Fixed

- Removed unused code and resolved linting issues
- Better temporary file cleanup with proper error handling

## [0.5.1] - 2025-07-16

### Changed

- Updated README with compelling real-world examples for pasty
- Screenshots & in-memory content processing workflows
- Multi-file scripting and automation examples
- GUI → CLI bridge use cases

### Fixed

- Code formatting improvements
- Removed redundant test cleanup flags

## [0.5.0] - 2025-07-16

### Added

- **Library support**: Clippy can now be used as a Go library
- High-level API with smart detection (`clippy.Copy()`, `clippy.CopyData()`, etc.)
- Low-level clipboard API (`pkg/clipboard` package)
- Examples directory with working code samples
- Support for stdin in library API (`clippy.CopyData(os.Stdin)`)
- **Pasty tool**: Companion paste tool for intelligent clipboard reading
- GoReleaser config now builds both clippy and pasty binaries

### Changed

- Moved clipboard package to `pkg/clipboard` for public library use
- Simplified clipboard package (removed low-level functions like `GetImage`, `GetTypes`)
- Updated README with library usage documentation and pasty tool section
- Fixed GoReleaser to use `brews` instead of `homebrew_casks` for CLI tools

## [0.4.0] - 2025-07-16

### Added

- Help flag (-h, --help) with comprehensive usage examples
- Config file options: `cleanup` and `temp_dir`
- Project restructured to support multiple binaries (preparing for pasty companion tool)

### Changed

- Code refactored to reduce duplication (DRY improvements)
- Shared clipboard functionality extracted to internal/clipboard package
- Updated README with new build instructions and config options

### Fixed

- Removed outdated race detector references

## [0.3.1] - 2025-07-15

### Changed

- Optimized temp file cleanup to use glob pattern matching instead of scanning entire temp directory
- Fixed race condition in cleanup by running it synchronously after main operation completes

### Fixed

- README examples now accurately describe pbcopy behavior (only accepts piped input)
- Updated curl example to use real image service (picsum.photos) with -L flag for redirects

## [0.3.0] - 2025-07-15

### Added

- Multi-file support: `clippy *.jpg` copies all matching files at once
- NSApplication initialization for reliable clipboard operations
- Comprehensive pipeline tests
- Comparison with gcopy alternative in README

### Fixed

- Clipboard operations now complete reliably without timers or delays
- Replaced invalid test PNG with proper 400x400 image

### Changed

- Simplified Homebrew installation to single command
- Improved README introduction to focus on core problem/solution

## [0.2.0] - 2025-07-15

### Added

- GitHub Actions CI/CD pipeline
- GoReleaser configuration for automated releases
- Homebrew tap support (neilberkman/homebrew-clippy)

## [0.1.0] - 2025-07-15

### Added

- Initial release
- Smart content detection (text vs binary files)
- File reference copying for GUI app compatibility
- Piped binary data support with temp file creation
- Automatic cleanup of temporary files
- Verbose mode (-v flag)
- Configuration file support (~/.clippy.conf)
- Version flag (--version)
