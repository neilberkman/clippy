# Changelog

Notable changes to clippy.

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
- Updated curl example to use real image service (picsum.photos)

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