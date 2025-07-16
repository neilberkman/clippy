# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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