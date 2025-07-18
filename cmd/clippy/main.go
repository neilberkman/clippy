package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/cmd/clippy/mcp"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/recent"
	"github.com/spf13/cobra"
)

var (
	verbose         bool
	debug           bool
	cleanup         = true
	tempDir         = ""
	recentFlag      string
	interactiveFlag string
	paste           bool
	absoluteTime    bool
	version         = "dev"
	commit          = "none"
	date            = "unknown"
	logger          *log.Logger
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "clippy [files...]",
		Short: "Smart clipboard tool for macOS",
		Args:  cobra.ArbitraryArgs,
		Long: `clippy - Smart clipboard tool for macOS

Copy files from your terminal that actually paste into GUI apps. No more switching to Finder.

Examples:
  # Copy text from stdin
  echo "Hello, World!" | clippy
  
  # Copy a file (text files copy content, others copy reference)
  clippy document.txt
  clippy image.png
  
  # Copy multiple files at once
  clippy *.jpg
  clippy file1.pdf file2.doc file3.png
  
  # Copy from curl
  curl -s https://example.com/image.jpg | clippy
  
  # Copy most recent download(s)
  clippy -r            # copy the most recent download
  clippy -r 3          # copy the 3 most recent downloads
  clippy -r 5m         # copy all downloads from last 5 minutes
  clippy -r 1h         # copy all downloads from last hour
  
  # Interactive picker for recent downloads
  clippy -i            # show interactive picker with recent files
  clippy -i 3          # show picker with 3 most recent files
  clippy -i 5m         # show picker for files from last 5 minutes
  # Picker supports both single and multi-select:
  # - Space to toggle selection
  # - Enter to copy (selected items or current item)
  # - p to copy & paste (selected items or current item)
  
  # Copy and paste in one step
  clippy file.txt --paste      # copy to clipboard AND paste to current dir
  clippy -r --paste            # copy most recent file and paste here
  clippy -i --paste            # pick recent file interactively and paste here

Configuration:
  Create ~/.clippy.conf with:
    verbose = true        # Always show verbose output
    cleanup = false       # Disable automatic temp file cleanup
    temp_dir = /path      # Custom directory for temporary files
    absolute_time = true  # Show absolute timestamps in picker (default: relative)`,
		Version: fmt.Sprintf("%s (%s) built on %s", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// Load config file
			loadConfig()

			// Initialize logger
			logger = log.New(log.Config{Verbose: verbose || debug, Debug: debug})

			// If files are provided as arguments, handle them (takes precedence)
			if len(args) > 0 {
				if len(args) == 1 {
					handleFileMode(args[0])
				} else {
					handleMultipleFiles(args)
				}
				// Run cleanup and return
				if cleanup {
					cleanupOldTempFiles()
				}
				return
			}

			// Handle -i flag (interactive mode)
			if cmd.Flags().Changed("interactive") {
				handleRecentMode(interactiveFlag, true)
				// Run cleanup and return
				if cleanup {
					cleanupOldTempFiles()
				}
				return
			}

			// Handle -r flag (immediate copy)
			if cmd.Flags().Changed("recent") {
				handleRecentMode(recentFlag, false)
				// Run cleanup and return
				if cleanup {
					cleanupOldTempFiles()
				}
				return
			}

			// Default: handle stream mode (stdin)
			handleStreamMode()

			// Run cleanup after main operation completes
			if cleanup {
				cleanupOldTempFiles()
			}
		},
	}

	// Add flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (includes technical details)")

	// Recent flag with optional value
	rootCmd.PersistentFlags().StringVarP(&recentFlag, "recent", "r", "", "Copy most recent file(s) from Downloads (defaults to 1, or specify number/duration like 3, 5m, 1h)")
	rootCmd.PersistentFlags().Lookup("recent").NoOptDefVal = " " // Allow -r without value

	// Interactive flag with optional value
	rootCmd.PersistentFlags().StringVarP(&interactiveFlag, "interactive", "i", "", "Show interactive picker for recent files (optional: number/duration like 3, 5m, 1h)")
	rootCmd.PersistentFlags().Lookup("interactive").NoOptDefVal = " " // Allow -i without value

	rootCmd.PersistentFlags().BoolVar(&paste, "paste", false, "Also paste copied files to current directory")
	rootCmd.PersistentFlags().BoolVar(&cleanup, "cleanup", true, "Enable automatic temp file cleanup")

	// Add MCP server subcommand
	var mcpCmd = &cobra.Command{
		Use:   "mcp-server",
		Short: "Start MCP server for AI/LLM integration",
		Long: `Start a Model Context Protocol (MCP) server that exposes clippy's functionality to AI assistants.

The MCP server allows AI assistants like Claude to interact with your clipboard programmatically.

Available tools:
- clipboard_copy: Copy text or files to clipboard
- clipboard_paste: Paste clipboard content to files
- get_recent_downloads: List recently downloaded files

Example usage with Claude Desktop:
Add to ~/Library/Application Support/Claude/claude_desktop_config.json:
{
  "mcpServers": {
    "clippy": {
      "command": "clippy",
      "args": ["mcp-server"]
    }
  }
}`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stderr, "Starting Clippy MCP server...")
			if err := mcp.StartServer(); err != nil {
				fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(mcpCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string, interactiveMode bool) {
	var maxAge time.Duration
	var count int = 1 // Default to 1 most recent file
	var err error

	// First check if timeStr is empty or space (just -r or -i)
	if timeStr == "" || timeStr == " " {
		// Default behavior: copy the most recent file
		maxAge = 24 * time.Hour * 30 // Look back up to 30 days
	} else {
		// Try to parse as a number first (e.g., "3" for 3 most recent files)
		if num, err := strconv.Atoi(timeStr); err == nil && num > 0 {
			count = num
			maxAge = 24 * time.Hour * 30 // Look back up to 30 days
		} else {
			// Parse as duration (e.g., "5m", "1h")
			maxAge, err = recent.ParseDuration(timeStr)
			if err != nil {
				logger.Error("Invalid argument: %v (use a number like '3' or duration like '5m')", err)
				os.Exit(1)
			}
			count = 0 // 0 means get all files within the time period
		}
	}

	// Get recent files based on criteria
	config := recent.PickerConfig{
		MaxAge:       maxAge,
		AbsoluteTime: absoluteTime,
	}

	files, err := recent.GetRecentDownloads(config)
	if err != nil {
		logger.Error("Failed to find recent files: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		logger.Error("No recent files found")
		os.Exit(1)
	}

	// If interactive mode is requested, show the picker
	if interactiveMode {
		logger.Debug("Showing bubble tea picker with %d files", len(files))
		result, err := showBubbleTeaPickerWithResult(files, config.AbsoluteTime)
		if err != nil {
			if err.Error() == "cancelled" {
				fmt.Println("Cancelled.")
				os.Exit(0)
			}
			logger.Error("No files selected: %v", err)
			os.Exit(1)
		}

		if len(result.Files) == 0 {
			logger.Error("No files selected")
			os.Exit(1)
		}

		// Override paste flag if user pressed 'p' in picker
		if result.PasteMode {
			paste = true
		}

		// Handle selected files
		if len(result.Files) == 1 {
			logger.Verbose("Selected: %s (modified %s ago)", result.Files[0].Path, time.Since(result.Files[0].Modified).Round(time.Second))
			handleFileMode(result.Files[0].Path)
		} else {
			logger.Verbose("Selected %d files:", len(result.Files))
			var paths []string
			for _, file := range result.Files {
				logger.Verbose("  - %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))
				paths = append(paths, file.Path)
			}
			handleMultipleFiles(paths)
		}
	} else {
		// Non-interactive mode: copy the requested number of files
		filesToCopy := files

		// If count is specified and we have more files than needed, slice it
		if count > 0 && len(files) > count {
			filesToCopy = files[:count]
		}

		if len(filesToCopy) == 1 {
			logger.Verbose("Copying most recent file: %s (modified %s ago)",
				filesToCopy[0].Name, time.Since(filesToCopy[0].Modified).Round(time.Second))
			handleFileMode(filesToCopy[0].Path)
		} else {
			logger.Verbose("Copying %d most recent files:", len(filesToCopy))
			var paths []string
			for _, file := range filesToCopy {
				logger.Verbose("  - %s (modified %s ago)", file.Name, time.Since(file.Modified).Round(time.Second))
				paths = append(paths, file.Path)
			}
			handleMultipleFiles(paths)
		}
	}
}

// Load configuration from ~/.clippy.conf
func loadConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configPath := filepath.Join(homeDir, ".clippy.conf")
	file, err := os.Open(configPath)
	if err != nil {
		return // No config file is fine
	}
	defer func() {
		if err := file.Close(); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "verbose":
			if value == "true" || value == "1" {
				verbose = true
			}
		case "cleanup":
			if value == "false" || value == "0" {
				cleanup = false
			}
		case "temp_dir":
			tempDir = value
		case "absolute_time":
			if value == "true" || value == "1" {
				absoluteTime = true
			}
		}
	}
}

// Logic for when a filename is provided as an argument
func handleFileMode(filePath string) {
	logger.Debug("handleFileMode called with path: %s", filePath)

	// Use the library function for smart copying with result info
	logger.Debug("Calling clippy.CopyWithResult for: %s", filePath)
	result, err := clippy.CopyWithResult(filePath)
	if err != nil {
		logger.Error("Could not copy file %s: %v", filePath, err)
		os.Exit(1)
	}
	logger.Debug("clippy.CopyWithResult returned successfully")

	// Show user-friendly verbose output
	if result.AsText {
		logger.Verbose("✅ Copied text content from '%s'", filepath.Base(filePath))
	} else {
		logger.Verbose("✅ Copied file reference for '%s'", filepath.Base(filePath))
	}

	// Show technical details in debug mode
	logger.Debug("Detection method: %s, Type: %s, AsText: %v", result.Method, result.Type, result.AsText)

	// Handle paste flag
	logger.Debug("Paste flag is: %v", paste)
	pasteFiles([]string{filePath})
}

// Handle multiple files at once
func handleMultipleFiles(paths []string) {
	logger.Debug("handleMultipleFiles called with %d paths", len(paths))
	for i, path := range paths {
		logger.Debug("  Path[%d]: %s", i, path)
	}

	// Use the library function for multiple file copying
	logger.Debug("Calling clippy.CopyMultiple")
	err := clippy.CopyMultiple(paths)
	if err != nil {
		logger.Error("Could not copy files: %v", err)
		os.Exit(1)
	}
	logger.Debug("clippy.CopyMultiple returned successfully")

	logger.Verbose("✅ Copied %d file references", len(paths))
	if verbose {
		for _, path := range paths {
			fmt.Printf("  - %s\n", filepath.Base(path))
		}
	}

	// Handle paste flag
	logger.Debug("Paste flag is: %v", paste)
	pasteFiles(paths)
}

// Logic for when data is piped via stdin
func handleStreamMode() {
	// Use the library function for stream copying
	err := clippy.CopyDataWithTempDir(os.Stdin, tempDir)
	if err != nil {
		logger.Error("Could not copy from stdin: %v", err)
	}

	logger.Verbose("✅ Copied content from stream using smart detection")
}

// Clean up old temp files that are no longer in clipboard
func cleanupOldTempFiles() {
	// Use the library function for cleanup
	clippy.CleanupTempFiles(tempDir, verbose)
}

// pasteFiles handles pasting files to current directory if --paste flag is set
func pasteFiles(files []string) {
	if !paste {
		return
	}

	for _, file := range files {
		err := recent.CopyFileToDestination(file, ".")
		if err != nil {
			logger.Error("Failed to paste file %s: %v", filepath.Base(file), err)
			continue
		}
	}
	logger.Verbose("✅ Also pasted %d files to current directory", len(files))
}
