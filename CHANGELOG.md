# Changelog

Notable changes to clippy.

## [Unreleased]

## [1.6.3] - 2026-01-10

### Added

- Pasty uses Finder-style duplicate naming (photo.png → photo 2.png) instead of overwriting
- `--force` / `-f` flag to override and allow overwriting
- Interactive picker auto-refreshes when new files appear in Downloads/Desktop
- Visual highlighting for newly appeared files in picker

### Fixed

- Multi-part extensions now handled correctly (archive.tar.gz → archive 2.tar.gz)

## [1.6.2] - 2026-01-06

### Fixed

- MCP tool `get_recent_downloads` now supports natural language time expressions (7d, yesterday, 2 weeks ago)
- Uses `github.com/olebedev/when` library for robust relative time parsing

## [1.6.1] - 2025-10-20

### Added

- Spotlight search: `clippy -f <query>` finds files without leaving terminal
- Uses native macOS MDQuery APIs for fast indexed search
- Results filtered to last 90 days, sorted by modification time

## [1.6.0] - 2025-10-21

### Fixed

- Clipboard priority now checks images before text (fixes "Copy Image" saving URL instead of image)

### Added

- Pasty saves browser images (auto-converts Safari's TIFF to PNG, 74-84% smaller)
- Smart image format conversion: specify target format via file extension (`pasty photo.jpg` converts to JPEG)
- Pasty handles rich text with embedded images (RTFD bundles from TextEdit/Notes)
- `--inspect` flag shows clipboard types and sizes for debugging
- `--plain` flag forces plain text extraction, stripping all formatting

## [1.5.4] - 2025-10-20

### Changed

- Updated dependencies and Go version to 1.25
- Improved documentation organization and clarity
- Cleaned up test output

## [1.5.3] - 2025-10-20

### Added

- **Smart content type detection**: Clippy now automatically detects and sets proper clipboard types for structured text formats
  - JSON content → sets `public.json` type
  - HTML content → sets `public.html` type
  - XML content → sets `public.xml` type
  - Many other text formats are properly detected
  - Receiving apps now handle content correctly (syntax highlighting, rendering, etc.)
- **Manual MIME type override** with `--mime` flag
  - `echo "data" | clippy --mime text/html` - force content as HTML
  - `clippy -t file.txt --mime application/json` - treat file as JSON
  - Accepts standard MIME types (text/html, application/json) or macOS UTIs
- **Major advantage over pbcopy**: Unlike pbcopy which only sets plain text, clippy now provides rich type information

### Changed

- Text copying now uses auto-detection by default to set appropriate clipboard types
- Added comprehensive list of textual MIME types for better content detection

## [1.5.2] - 2025-10-18

### Added

- buffer_cut tool for atomic cut operations (copy to buffer + delete from source)
- MCP server installation instructions in --help output

### Changed

- Optimized MCP tool descriptions for token efficiency
- README reorganized to promote MCP server section

## [1.5.1] - 2025-10-09

### Fixed

- Updated server.json tool descriptions to accurately reflect byte-level file manipulation
- Clarified buffer tool parameters and modes for better LLM understanding

## [1.5.0] - 2025-10-09

### Added

- **BREAKING**: Complete rewrite of agent buffer tools for true byte-level copy/paste
  - `buffer_copy` - Copy file bytes (line ranges supported) to agent's private buffer
  - `buffer_paste` - Paste bytes to file with append/insert/replace modes
  - `buffer_list` - Show buffer metadata (lines, source file, range)
  - **Key improvement**: MCP server reads/writes file bytes directly - agent never generates tokens for copied content
  - Solves the LLM "remember and re-emit" problem with actual file manipulation, not token generation
  - Enables surgical refactoring: copy lines 17-32, paste to replace lines 5-8, etc.
  - No system clipboard interference

### Breaking Changes

- `buffer_copy` now requires `file` parameter (no longer accepts `text`)
- `buffer_paste` now requires `file` and `mode` parameters
- Buffer stores raw bytes, not text strings

## [1.4.0] - 2025-10-09

### Deprecated

- Initial agent buffer tools (replaced in 1.5.0 with file-based byte manipulation)

## [1.3.4] - 2025-09-22

### Enhanced

- Improved MCP server tool descriptions to advertise the efficient `force_text` pattern
  - Added PRO TIPS explaining how to use temp files with `force_text='true'` for iterative editing
  - Helps LLMs discover this pattern for more efficient code editing workflows
  - Particularly useful for production debugging scenarios

## [1.3.3] - 2025-01-26

### Fixed

- Fixed interactive picker issues (#6)
  - Resolved picker UI problems
  - Improved interaction handling

## [1.3.2] - 2025-01-25

### Fixed

- Fixed timezone handling in recent files
  - File ages now calculated using UTC to avoid timezone confusion
  - Prevents negative durations when files have future timestamps in local timezone
  - Added comprehensive timezone tests

## Draggy [0.14.0] - 2026-01-15

### Changed

- Unified clipboard and recent files into a single view with clear section headers
- Clipboard files now always appear at the top, with a divider before recent files
- Removed clipboard/recent toggle and auto-switch toast to reduce mode switching
- Added folder-specific icons alongside file source labels (Downloads, Desktop, Documents, other)
- Renamed UI copy from "Recent Downloads" to "Recent Files"

## Draggy [0.13.3] - 2025-01-25

### Fixed

- Fixed timezone handling in recent downloads (uses shared core library)
  - File ages now calculated using UTC to avoid timezone confusion
  - Prevents negative durations when files have future timestamps in local timezone

## Draggy [0.13.2] - 2025-01-24

### Fixed

- Folders are no longer shown in recent files list
  - Only actual files are displayed, not directories
  - Improves usability since folders can't be meaningfully copied/pasted

## [1.3.1] - 2025-01-24

### Fixed

- Fixed `-t`/`--text` flag not working with non-standard text files (e.g., .exs files)
  - Now falls through to MIME detection when UTI is not recognized as text
  - Ensures all text files can have their content copied with `-t` flag

## Draggy [0.13.1] - 2025-01-23

### Changed

- **Improved UI polish**:
  - Removed redundant info bar, keeping only footer hint
  - Shows "No files" instead of "0 files" when empty
  - Option key hint styled as blue badge for better visibility
  - Auto-switch toast now has dismissible X button

### Fixed

- Fixed macOS version mismatch warnings during build
- Set deployment target to macOS 14.0 for both Go and Swift
- Fixed critical bug where Option key previews would trigger anywhere on screen when popover was closed
  - Properly cleanup timers when popover closes
  - Stop monitoring modifier keys when app loses focus
  - Clear view models on popover dismissal

## [1.3.0] - 2025-01-22

### Added

- **MIME type detection for all files**: Files now have their MIME types detected
  - Uses gabriel-vasile/mimetype library for accurate detection
  - MIME types passed through C bridge to Draggy
- **Human-friendly file type display**: Shows readable types instead of technical MIME
  - Integrated neilberkman/mimedescription library for type descriptions
  - Shows "PDF document", "Word document", "PNG image" etc.
  - Smart fallback for unknown types

### Changed

- **Enhanced clippy picker display**:
  - Shows file type next to each file (e.g., [PDF], [PNG], [Word])
  - Dynamic terminal width adjustment for better layout
  - Middle truncation for long filenames (shows start and end)
  - Improved spacing calculations

### Draggy [0.13.0]

- **File type information in preview**: Hold Option to see complete file details
  - Preview window now shows filename, file type, size, folder, and time
  - Improved metadata layout with visual separator
  - Better typography and spacing for readability
- **Updated data format**: Now parses MIME types from bridge
- **Fixed deprecation warning**: Updated onChange handler for macOS 14+

## [1.2.6] - 2025-01-21

### Added

- **Folder selection for recent files**: Control which folders to search
  - New `--folders` flag to specify folders (e.g., `--folders downloads,desktop`)
  - Supports downloads, desktop, and documents folders
  - Config option `default_folders` to set default search folders
  - Updated help text to reflect broader search scope

### Draggy [0.12.0]

- **Folder preferences**: Configure which folders to search for recent files
  - New preferences section for folder selection
  - Checkboxes for Downloads, Desktop, and Documents
  - Settings persist across app restarts
- **Improved recent files display**: Better organization and UI
  - Enhanced file row display with better formatting
  - Improved view model for recent files handling
  - Better integration with folder preferences
- **UI refinements**: Various interface improvements
  - Better settings window layout
  - Improved content view organization
  - Enhanced file browsing experience

## [1.2.5] - 2025-01-21

### Fixed

- MCP server panic "Prompt capabilities not enabled"
  - Upgraded mcp-go from accidentally downgraded v0.10.0 back to v0.34.0
  - Restores MCP server functionality for AI/LLM integration

## [1.2.2] - 2025-01-20

### Draggy [0.11.3]

- Fixed auto-switch message behavior
  - X button now only dismisses the message, doesn't switch views
  - Info bar with tips now appears after dismissing message
- Fixed update notification button truncation
  - "Copy Command" button shortened to "Copy" with tooltip
- UI improvements for better clarity

## [1.2.2] - 2025-01-20

### Fixed

- Recent downloads flags (`-r 3`, `-i 5m`) now work correctly
  - Fixed argument parsing to support space-separated values
  - Previously required `=` syntax (e.g., `-r=3`)
- Help text now correctly shows usage examples

### Draggy [0.11.2]

- Fixed blank "Draggy Settings" window appearing at startup
  - Properly configured SwiftUI App scene to prevent default window
  - Added window closing on app launch as safety measure
  - Only affects release builds (not seen in development)
- Previous fixes from 0.11.1:
  - Fixed settings window content display
  - Fixed version display (now shows v0.11.0)
  - Fixed window title consistency

## [1.2.1] - 2025-01-20

### Fixed

- Version information now properly displayed in `--version` output
  - Fixed ldflags in GoReleaser to correctly set version variables

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

### Draggy [0.11.0]

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
