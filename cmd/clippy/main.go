package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/recent"
)

var (
	verbose    bool
	cleanup    = true
	tempDir    = ""
	recentFlag = ""
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	logger     *log.Logger
)

func main() {
	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `clippy - Smart clipboard tool for macOS

Usage:
  clippy [options] [file...]

Options:
  -v, --verbose    Enable verbose output
  --recent [time]  Copy most recent file from Downloads (e.g. --recent 5m)
  --cleanup        Enable automatic temp file cleanup (default: true)
  --version        Show version information
  -h, --help       Show this help message

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
  clippy --recent 10m  # last 10 minutes
  clippy --recent 1h   # last hour

Configuration:
  Create ~/.clippy.conf with:
    verbose = true    # Always show verbose output
    cleanup = false   # Disable automatic temp file cleanup
    temp_dir = /path  # Custom directory for temporary files
`)
	}

	// Custom flag handling for --recent (can be used with or without value)
	var recentFlagSet bool
	
	// Pre-process args to handle --recent flag
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--recent" {
			recentFlagSet = true
			// Check if next arg is a time duration (not a flag)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				recentFlag = os.Args[i+1]
				// Remove both --recent and the time arg
				os.Args = append(os.Args[:i], os.Args[i+2:]...)
			} else {
				// Remove just --recent
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
			}
			break
		}
	}

	// Parse flags
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&cleanup, "cleanup", true, "Enable automatic temp file cleanup")

	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")
	flag.BoolVar(showHelp, "h", false, "Show help message (shorthand)")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("clippy version %s (%s) built on %s\n", version, commit, date)
		os.Exit(0)
	}

	// Load config file
	loadConfig()

	// Initialize logger
	logger = log.New(log.Config{Verbose: verbose})

	// Handle --recent flag
	if recentFlagSet {
		handleRecentMode(recentFlag)
		return
	}

	// Decide between File Mode and Stream Mode
	args := flag.Args()
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
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string) {
	// Parse time duration (library handles this)
	maxAge, err := recent.ParseDuration(timeStr)
	if err != nil {
		logger.Error("Invalid time duration: %v", err)
		os.Exit(1)
	}

	// Use high-level library function
	file, err := recent.CopyMostRecentDownload(maxAge)
	if err != nil {
		logger.Error("No recent files found: %v", err)
		os.Exit(1)
	}

	logger.Verbose("Found recent file: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))

	// Copy the file using existing clippy logic
	handleFileMode(file.Path)
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

	// Show verbose output with detection method
	if result.AsText {
		logger.Verbose("✅ Copied text content from '%s' (%s: %s)",
			filepath.Base(filePath), result.Method, result.Type)
	} else {
		logger.Verbose("✅ Copied file reference for '%s' (%s: %s)",
			filepath.Base(filePath), result.Method, result.Type)
	}
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
