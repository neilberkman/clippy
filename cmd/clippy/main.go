package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/cmd/clippy/mcp"
	"github.com/neilberkman/clippy/cmd/internal/common"
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
	textMode        bool
	clearFlag       bool
	foldersFlag     []string
	defaultFolders  []string
	logger          *log.Logger
)

func main() {
	// Preprocess args to convert "-r 3" to "-r=3" for Cobra compatibility
	os.Args = preprocessArgs(os.Args)

	var rootCmd = &cobra.Command{
		Use:   "clippy [files...]",
		Short: "Smart clipboard tool for macOS",
		Args:  cobra.ArbitraryArgs,
		Long: `clippy - Smart clipboard tool for macOS

Copy files from your terminal that actually paste into GUI apps. No more switching to Finder.

Examples:
  # Copy text from stdin
  echo "Hello, World!" | clippy

  # Copy a file as file reference (default for all files)
  clippy document.txt
  clippy image.png

  # Copy text file content instead of reference
  clippy -t document.txt
  clippy --text README.md

  # Copy multiple files at once
  clippy *.jpg
  clippy file1.pdf file2.doc file3.png

  # Copy from curl
  curl -s https://example.com/image.jpg | clippy

  # Copy most recent file(s) from Downloads/Desktop/Documents
  clippy -r            # copy the most recent file
  clippy -r 3          # copy the 3 most recent files
  clippy -r 5m         # copy all recent files from last 5 minutes
  clippy -r 1h         # copy all recent files from last hour

  # Limit search to specific folders
  clippy -r --folders downloads        # only search Downloads
  clippy -r --folders downloads,desktop # search Downloads and Desktop only

  # Interactive picker for recent files
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

  # Clear clipboard
  clippy --clear               # empty the clipboard
  echo -n | clippy             # also clears the clipboard

Configuration:
  Create ~/.clippy.conf with:
    verbose = true        # Always show verbose output
    cleanup = false       # Disable automatic temp file cleanup
    temp_dir = /path      # Custom directory for temporary files
    absolute_time = true  # Show absolute timestamps in picker (default: relative)
    default_folders = downloads,desktop,documents  # Default folders to search (defaults to all three)`,
		Version: fmt.Sprintf("%s (%s) built on %s", common.Version, common.Commit, common.Date),
		Run: func(cmd *cobra.Command, args []string) {
			// Load config file
			loadConfig()

			// Initialize logger
			logger = common.SetupLogger(verbose, debug)

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

			// Handle --clear flag
			if clearFlag {
				if err := clearClipboard(); err != nil {
					logger.Error("Failed to clear clipboard: %v", err)
					os.Exit(1)
				}
				logger.Verbose("✅ Clipboard cleared")
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
	common.AddCommonFlags(rootCmd, &verbose, &debug)

	// Recent flag with optional value
	rootCmd.PersistentFlags().StringVarP(&recentFlag, "recent", "r", "", "Copy most recent file(s) from Downloads, Desktop, and Documents (defaults to 1, or specify number/duration like 3, 5m, 1h)")
	rootCmd.PersistentFlags().Lookup("recent").NoOptDefVal = " " // Allow -r without value

	// Interactive flag with optional value
	rootCmd.PersistentFlags().StringVarP(&interactiveFlag, "interactive", "i", "", "Show interactive picker for recent files from Downloads, Desktop, and Documents (optional: number/duration like 3, 5m, 1h)")
	rootCmd.PersistentFlags().Lookup("interactive").NoOptDefVal = " " // Allow -i without value

	rootCmd.PersistentFlags().BoolVar(&paste, "paste", false, "Also paste copied files to current directory")
	rootCmd.PersistentFlags().BoolVar(&cleanup, "cleanup", true, "Enable automatic temp file cleanup")
	rootCmd.PersistentFlags().BoolVarP(&textMode, "text", "t", false, "Copy text files as content instead of file reference")
	rootCmd.PersistentFlags().BoolVar(&clearFlag, "clear", false, "Clear the clipboard")
	rootCmd.PersistentFlags().StringSliceVar(&foldersFlag, "folders", nil, "Specific folders to search (e.g., --folders downloads,desktop). Options: downloads, desktop, documents")

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

// clearClipboard clears the clipboard (common function for DRY code)
func clearClipboard() error {
	return clippy.ClearClipboard()
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string, interactiveMode bool) {
	// Use Core function to parse the argument
	count, maxAge, err := recent.ParseRecentArgument(timeStr)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}

	// Get recent files based on criteria
	config := recent.PickerConfig{
		MaxAge:       maxAge,
		AbsoluteTime: absoluteTime,
	}

	// Pass count to Core layer for proper limiting
	// If interactive mode, get more files for the picker to show
	maxFiles := count
	if interactiveMode && count == 0 {
		maxFiles = 20 // Default for interactive picker
	}

	// Handle folder selection if specified
	var searchDirs []string
	if len(foldersFlag) > 0 {
		searchDirs = mapFoldersToDirectories(foldersFlag)
		if len(searchDirs) == 0 {
			logger.Error("Invalid folder selection. Use: downloads, desktop, documents")
			os.Exit(1)
		}
	} else if len(defaultFolders) > 0 {
		// Use config defaults if no command line folders specified
		searchDirs = mapFoldersToDirectories(defaultFolders)
		logger.Debug("Using default folders from config: %v", searchDirs)
	}

	files, err := getRecentDownloadsWithDirs(config, maxFiles, searchDirs)
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
		// Non-interactive mode: files are already limited by Core layer
		if len(files) == 1 {
			logger.Verbose("Copying most recent file: %s (modified %s ago)",
				files[0].Name, time.Since(files[0].Modified).Round(time.Second))
			handleFileMode(files[0].Path)
		} else {
			logger.Verbose("Copying %d most recent files:", len(files))
			var paths []string
			for _, file := range files {
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
		case "default_folders":
			defaultFolders = strings.Split(value, ",")
		}
	}
}

// Logic for when a filename is provided as an argument
func handleFileMode(filePath string) {
	logger.Debug("handleFileMode called with path: %s", filePath)

	// Use the library function for smart copying with result info
	logger.Debug("Calling clippy.CopyWithResultAndMode for: %s (textMode=%v)", filePath, textMode)
	result, err := clippy.CopyWithResultAndMode(filePath, textMode)
	if err != nil {
		logger.Error("Could not copy file %s: %v", filePath, err)
		os.Exit(1)
	}
	logger.Debug("clippy.CopyWithResultAndMode returned successfully")

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
	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin has data - read it
		var buf bytes.Buffer
		_, err := io.Copy(&buf, os.Stdin)
		if err != nil {
			logger.Error("Could not read from stdin: %v", err)
			os.Exit(1)
		}

		// Check if input is empty
		if buf.Len() == 0 {
			// Empty input - clear clipboard
			if err := clearClipboard(); err != nil {
				logger.Error("Failed to clear clipboard: %v", err)
				os.Exit(1)
			}
			logger.Verbose("✅ Clipboard cleared (empty input)")
		} else {
			// Non-empty input - copy to clipboard
			err := clippy.CopyDataWithTempDir(&buf, tempDir)
			if err != nil {
				logger.Error("Could not copy from stdin: %v", err)
				os.Exit(1)
			}
			logger.Verbose("✅ Copied content from stream using smart detection")
		}
	} else {
		// No stdin data and no arguments - show usage
		logger.Error("No input provided. Use --help for usage information.")
		os.Exit(1)
	}
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

// preprocessArgs converts "-r 3" to "-r=3" for better Cobra compatibility
func preprocessArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Check if this is -r or -i flag
		if (arg == "-r" || arg == "--recent" || arg == "-i" || arg == "--interactive") && i+1 < len(args) {
			// Check if next arg looks like a value (not another flag)
			nextArg := args[i+1]
			if !strings.HasPrefix(nextArg, "-") {
				// Combine into single arg with =
				result = append(result, arg+"="+nextArg)
				i++ // Skip next arg since we consumed it
				continue
			}
		}
		result = append(result, arg)
	}
	return result
}

// mapFoldersToDirectories converts folder names to actual directory paths
func mapFoldersToDirectories(folders []string) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var dirs []string
	for _, folder := range folders {
		switch strings.ToLower(strings.TrimSpace(folder)) {
		case "downloads", "download":
			dirs = append(dirs, filepath.Join(homeDir, "Downloads"))
		case "desktop":
			dirs = append(dirs, filepath.Join(homeDir, "Desktop"))
		case "documents", "docs":
			dirs = append(dirs, filepath.Join(homeDir, "Documents"))
		}
	}
	return dirs
}

// getRecentDownloadsWithDirs gets recent downloads with custom directory list
func getRecentDownloadsWithDirs(config recent.PickerConfig, maxFiles int, customDirs []string) ([]recent.FileInfo, error) {
	opts := recent.DefaultFindOptions()
	if config.MaxAge != 0 {
		opts.MaxAge = config.MaxAge
	}
	if maxFiles > 0 {
		opts.MaxCount = maxFiles
	} else {
		opts.MaxCount = 20 // Default to 20 if not specified
	}

	// Override directories if custom ones are provided
	if len(customDirs) > 0 {
		opts.Directories = customDirs
	}

	files, err := recent.FindRecentFiles(opts)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no recent files found")
	}

	return files, nil
}
