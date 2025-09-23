# Claude AI Context File

This file contains context and memory for Claude AI when working on the clippy project.

## CRITICAL: Never Credit Yourself

**NEVER** add Claude credits to commits, code, or any project files. This includes:
- No "Generated with Claude Code" messages
- No "Co-Authored-By: Claude" in commits
- No mentions of AI assistance in code or documentation
- Keep all contributions attribution-free

## CRITICAL: Never Do Manual Release Steps

**NEVER** manually publish releases, edit releases, or do one-off fixes. The goal is ALWAYS a repeatable automated process. This includes:
- No `gh release edit` commands to fix drafts
- No manual edits to published releases
- No manual updates to Homebrew taps
- No workarounds - FIX THE AUTOMATION
- The goal is THE PROCESS, not getting a single release out

## Project Overview

Clippy is a macOS clipboard tool that bridges the gap between terminal file operations and GUI applications. It includes:

- **clippy**: Smart clipboard copying tool
- **pasty**: Intelligent clipboard pasting tool

## Recent Major Features

### Recent Downloads Functionality (v0.8.0+)
- Added `--recent` flag to clippy command (removed from pasty for better separation of concerns)
- Interactive picker using Bubble Tea library (replaced promptui for better multi-select support)
- Simplified `-r` behavior:
  - `-r` alone copies the most recent download (no picker)
  - `-r 3` copies the 3 most recent downloads
  - `-r 5m` copies all downloads from last 5 minutes
  - `-ri` shows interactive picker (can combine with numbers/durations)
- Multi-select support in picker: Space to toggle, Enter to copy, p to copy & paste
- Removed `--batch` flag (behavior integrated into numbered copies)
- Time-based filtering (e.g., `-r 5m`, `-r 1h`)
- macOS Downloads folder detection with smart archive handling
- Separate `--debug` and `--verbose` flags for better UX
- Config option `absolute_time = true` for absolute timestamps in picker

### Key Implementation Details

#### Library Structure
- `pkg/recent/recent.go`: Core library for recent downloads detection
- `pkg/recent/recent_test.go`: Comprehensive tests
- `internal/log/log.go`: Enhanced logging with debug support
- Library-first architecture with high-level business functions

#### Commands
- **clippy**: Uses `recent.GetRecentDownloads()` for core functionality, picker UI in cmd/
- **pasty**: No longer has recent downloads functionality (moved to clippy for cleaner separation)

#### Technical Features
- Cobra CLI framework for professional command-line interface
- Smart auto-unarchive detection for macOS Downloads folder
- Time-based filtering with duration parsing
- Batch handling for files downloaded together (within 30 seconds)
- Platform-specific build constraints (darwin vs windows)

## Development Guidelines

### Design Principles

#### Core vs Interface Philosophy (from Saša Jurić)
- **Core**: Implements the desired behavior of the system (what must be done regardless of how the system is accessed)
- **Interface**: Contains logic specific to how clients access the system (REST, CLI, GraphQL, etc.)
- **Key principle**: "Core implements behavior, Interface exposes it"
- **Decision rule**: If something is protocol-specific, it's an interface concern. If it must run in all cases, it's a core concern.
- **Benefits**: Clear separation allows developers to focus on one layer without understanding the other
- **CRITICAL**: Business logic NEVER goes in the interface layer!
  - Filtering files vs directories? Core concern - it's part of what "recent downloads" means
  - Excluding temp files? Core concern - it's part of the business rule
  - Converting user input to core types? Interface concern
  - Presenting data to users? Interface concern
- **Example violations to avoid**:
  - Don't filter data in controllers/views - the core should provide the right data
  - Don't validate business rules in the interface - the core enforces all rules
  - Don't make the interface "smart" - keep it dumb and focused on translation

#### Library-First Architecture
- **Core principle**: Implement all business logic as library functions first
- **Command-line tools**: Keep cmd/ tools as thin wrappers around library functions
- **Example**: `clippy.Copy()` function in library, `clippy` command calls it
- **Benefits**: Enables programmatic use, easier testing, cleaner separation of concerns
- **Pattern**: High-level business functions exposed through simple library APIs
- **UI in interface only**: Interactive elements (like pickers) belong in cmd/, not pkg/

#### CLI Design Philosophy
- **Professional CLI**: Use Cobra framework for consistent, professional command-line interface
- **Smart defaults**: Commands should work intuitively without excessive configuration
- **Composability**: Tools should work well together and with other Unix tools
- **Discoverability**: Use clear flag names and helpful examples

#### Code Organization
- `pkg/`: Public library packages (e.g., `pkg/clipboard`, `pkg/recent`)
- `internal/`: Private packages not meant for external use (e.g., `internal/log`)
- `cmd/`: Command-line applications as thin wrappers around library functions
- Build constraints for platform-specific code (e.g., `//go:build darwin`)

### README Guidelines

#### Content Strategy
- **"Taste of features"**: Give users a taste of what's possible without overwhelming them
- **Progressive disclosure**: Start with core use cases, then show advanced features
- **Right flow**: Guide users from basic to advanced naturally
- **No hyperbole**: Avoid exaggerated claims or marketing language
- **Practical examples**: Show real-world use cases, not contrived demos

#### Structure Principles
- **Why section**: Explain the problem clippy solves clearly
- **Core examples**: Show the most important 3-4 use cases upfront
- **Installation**: Simple, clear installation instructions
- **Feature sections**: Organize by user workflow, not technical implementation
- **Full details**: Provide comprehensive information for power users at the end

#### Writing Style
- **Concise**: Respect the reader's time
- **Practical**: Focus on what users actually need to do
- **Clear**: Avoid jargon and technical complexity in introductory sections
- **Comprehensive**: Provide full details for those who need them

### Code Style
- Library-first architecture - implement features as library functions first
- Follow existing patterns and conventions
- Use existing dependencies (check go.mod before adding new ones)
- Never add comments unless explicitly requested

### Testing
- Run tests with `go test -v ./...`
- Check specific test framework by examining existing tests
- Always verify implementation with tests before considering complete

### Git Workflow
- Never commit changes unless explicitly requested
- Clean commit messages without Claude credits
- Use meaningful commit messages that explain the "why" not just the "what"

## Build and Release

### Commands
```bash
# Build
go build -o clippy ./cmd/clippy
go build -o pasty ./cmd/pasty

# Test
go test -v ./...

# Install
go install github.com/neilberkman/clippy/cmd/clippy@latest
go install github.com/neilberkman/clippy/cmd/pasty@latest
```

### Version Management
- Update CHANGELOG.md for new releases
- Use semantic versioning
- Tag releases appropriately

## Current Status

- Main branch has clean git history with Claude credits removed
- Recent downloads functionality moved to clippy only (cleaner separation of concerns)
- Interactive picker replaced with Bubble Tea for better multi-select support
- Picker supports both single and multi-select without mode switching
- Paste mode integrated into picker (p key to copy & paste)
- README updated with streamlined "taste of features" approach
- Fixed clipboard Heisenbug with proper changeCount polling (no more sleep hack)
- Simplified -r flag behavior (immediate copy by default, -i for interactive)

## Dependencies

Key dependencies in go.mod:
- github.com/spf13/cobra: CLI framework
- github.com/charmbracelet/bubbletea: TUI framework for interactive picker
- github.com/charmbracelet/lipgloss: Terminal styling for picker UI
- github.com/gabriel-vasile/mimetype: MIME type detection

## Architecture Notes

The project follows a library-first approach where:
1. Core functionality is implemented in pkg/ packages (business logic only)
2. Command-line tools in cmd/ are thin wrappers that handle interface concerns (UI, CLI parsing)
3. Platform-specific code is handled with build constraints
4. High-level business logic is exposed through simple library functions
5. Interactive UI elements (like the Bubble Tea picker) live in cmd/, not pkg/

## Tips for Working with Clippy MCP Tool

### Efficient File Editing Pattern
When you need to copy code that you're iteratively editing:
1. Write the code to a temp file (e.g., `/tmp/script.exs`)
2. Edit the file as needed using the Edit tool
3. Use `mcp__clippy__clipboard_copy` with `file: "/tmp/script.exs"` and `force_text: "true"`
4. This copies the file contents as text (not as a file reference)

**Why this is better**: Instead of regenerating the full text each time you make edits, you can:
- Edit the temp file incrementally with the Edit tool
- Let clippy handle converting the file to clipboard text
- Avoid regenerating entire code blocks for small changes
- Makes iterative development much more efficient, especially for large files

**Example workflow**:
```
1. Write initial code: Write(/tmp/debug_script.exs, content)
2. User runs it, finds issues
3. Edit specific parts: Edit(/tmp/debug_script.exs, old_string, new_string)
4. Copy to clipboard: mcp__clippy__clipboard_copy(file: "/tmp/debug_script.exs", force_text: "true")
5. Repeat steps 2-4 as needed
```

This pattern is especially useful when helping users debug code in production environments where they need to paste updated scripts repeatedly.