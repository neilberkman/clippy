# Changelog

Notable changes to clippy.

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