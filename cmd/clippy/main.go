package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/internal/log"
)

var (
	verbose bool
	cleanup = true
	tempDir = ""
	version = "dev"
	commit  = "none"
	date    = "unknown"
	logger  *log.Logger
)

func main() {
	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `clippy - Smart clipboard tool for macOS

Usage:
  clippy [options] [file...]

Options:
  -v, --verbose    Enable verbose output
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

Configuration:
  Create ~/.clippy.conf with:
    verbose = true    # Always show verbose output
    cleanup = false   # Disable automatic temp file cleanup
    temp_dir = /path  # Custom directory for temporary files
`)
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

// getAbsPath returns the absolute path or exits with error
func getAbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger.Error("Invalid path %s", path)
	}
	return absPath
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
