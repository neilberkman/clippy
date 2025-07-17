package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/recent"
)

var (
	verbose     bool
	debug       bool
	cleanup     = true
	tempDir     = ""
	recentFlag  string
	recentBatch bool
	recentPick  bool
	paste       bool
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	logger      *log.Logger
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "clippy [files...]",
		Short: "Smart clipboard tool for macOS",
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
  
  # Copy most recent download
  clippy --recent
  clippy -r            # short form
  clippy -r 5m         # last 5 minutes only
  
  # Copy batch of recent downloads
  clippy -r --batch    # copy all files downloaded together
  
  # Interactive picker for recent downloads
  clippy -r --pick     # choose from list of recent files
  
  # Copy and paste in one step
  clippy file.txt --paste      # copy to clipboard AND paste to current dir
  clippy -r --paste            # copy recent file and paste here
  clippy -r --pick --paste     # pick a file, copy it, and paste here

Configuration:
  Create ~/.clippy.conf with:
    verbose = true    # Always show verbose output
    cleanup = false   # Disable automatic temp file cleanup
    temp_dir = /path  # Custom directory for temporary files`,
		Version: fmt.Sprintf("%s (%s) built on %s", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// Load config file
			loadConfig()

			// Initialize logger
			logger = log.New(log.Config{Verbose: verbose || debug, Debug: debug})

			// Handle --recent flag
			if recentFlag != "" {
				handleRecentMode(recentFlag, recentBatch, recentPick)
				return
			}

			// Decide between File Mode and Stream Mode
			if len(args) > 0 {
				if len(args) == 1 {
					handleFileMode(args[0])
				} else {
					handleMultipleFiles(args)
				}
			} else {
				handleStreamMode()
			}

			// Run cleanup after main operation completes
			if cleanup {
				cleanupOldTempFiles()
			}
		},
	}

	// Add flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (includes technical details)")
	rootCmd.PersistentFlags().StringVarP(&recentFlag, "recent", "r", "", "Copy most recent file from Downloads (optional: time duration like 5m, 1h)")
	rootCmd.PersistentFlags().BoolVar(&recentBatch, "batch", false, "Copy all files from most recent batch download")
	rootCmd.PersistentFlags().BoolVarP(&recentPick, "pick", "p", false, "Show interactive picker for recent downloads")
	rootCmd.PersistentFlags().BoolVar(&paste, "paste", false, "Also paste copied files to current directory")
	rootCmd.PersistentFlags().BoolVar(&cleanup, "cleanup", true, "Enable automatic temp file cleanup")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string, batch bool, pick bool) {
	// Parse time duration (library handles this)
	maxAge, err := recent.ParseDuration(timeStr)
	if err != nil {
		logger.Error("Invalid time duration: %v", err)
		os.Exit(1)
	}

	if pick {
		// Use picker mode - interactively select file
		file, err := recent.PickRecentDownload(maxAge)
		if err != nil {
			logger.Error("No file selected: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Selected: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))

		// Copy the selected file using existing clippy logic
		handleFileMode(file.Path)
	} else if batch {
		// Use batch mode - copy multiple files
		files, err := recent.CopyRecentDownloads(maxAge, 10)
		if err != nil {
			logger.Error("No recent files found: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Found %d files in recent batch:", len(files))
		for _, file := range files {
			logger.Verbose("  - %s (modified %s ago)", file.Name, time.Since(file.Modified).Round(time.Second))
		}

		// Copy all files using existing clippy logic
		var paths []string
		for _, file := range files {
			paths = append(paths, file.Path)
		}
		handleMultipleFiles(paths)
	} else {
		// Use single file mode
		file, err := recent.CopyMostRecentDownload(maxAge)
		if err != nil {
			logger.Error("No recent files found: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Found recent file: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))

		// Copy the file using existing clippy logic
		handleFileMode(file.Path)
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
		}
	}
}

// Logic for when a filename is provided as an argument
func handleFileMode(filePath string) {
	// Use the library function for smart copying with result info
	result, err := clippy.CopyWithResult(filePath)
	if err != nil {
		logger.Error("Could not copy file %s: %v", filePath, err)
	}

	// Show user-friendly verbose output
	if result.AsText {
		logger.Verbose("✅ Copied text content from '%s'", filepath.Base(filePath))
	} else {
		logger.Verbose("✅ Copied file reference for '%s'", filepath.Base(filePath))
	}
	
	// Show technical details in debug mode
	logger.Debug("Detection method: %s, Type: %s", result.Method, result.Type)
	
	// Handle paste flag
	pasteFiles([]string{filePath})
}

// Handle multiple files at once
func handleMultipleFiles(paths []string) {
	// Use the library function for multiple file copying
	err := clippy.CopyMultiple(paths)
	if err != nil {
		logger.Error("Could not copy files: %v", err)
	}

	logger.Verbose("✅ Copied %d file references", len(paths))
	if verbose {
		for _, path := range paths {
			fmt.Printf("  - %s\n", filepath.Base(path))
		}
	}
	
	// Handle paste flag
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