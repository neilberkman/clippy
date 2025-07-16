package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/clipboard"
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
	absPath := getAbsPath(filePath)

	// Detect MIME type to check if it's text
	mtype, err := mimetype.DetectFile(absPath)
	if err != nil {
		logger.Error("Could not read file %s", absPath)
	}

	// Special case for text files: copy content, not file reference
	if strings.HasPrefix(mtype.String(), "text/") {
		content, err := os.ReadFile(absPath)
		if err != nil {
			logger.Error("Could not read file content %s", absPath)
		}
		clipboard.CopyText(string(content))
		logger.Verbose("✅ Copied text content from '%s'", filepath.Base(absPath))
	} else {
		// For all other file types, copy as a file reference
		clipboard.CopyFile(absPath)
		logger.Verbose("✅ Copied file reference for '%s'", filepath.Base(absPath))
	}
}

// Handle multiple files at once
func handleMultipleFiles(paths []string) {
	// Check if all files exist and collect absolute paths
	absPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		absPath := getAbsPath(path)

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			logger.Error("Could not read file %s", absPath)
		}

		absPaths = append(absPaths, absPath)
	}

	// Copy all files at once
	clipboard.CopyFiles(absPaths)

	logger.Verbose("✅ Copied %d file references", len(absPaths))
	if verbose {
		for _, path := range absPaths {
			fmt.Printf("  - %s\n", filepath.Base(path))
		}
	}
}

// Logic for when data is piped via stdin
func handleStreamMode() {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, os.Stdin); err != nil {
		logger.Error("reading from stdin")
	}
	data := buf.Bytes()
	if len(data) == 0 {
		logger.PrintErr("Warning: Input stream was empty.")
		return
	}

	mtype := mimetype.Detect(data)

	// Special case for text streams: copy content
	if strings.HasPrefix(mtype.String(), "text/") {
		clipboard.CopyText(string(data))
		logger.Verbose("✅ Copied text content from stream")
	} else {
		// For binary streams, save to a temp file, then copy the reference
		tmpFile, err := os.CreateTemp(tempDir, "clippy-*"+mtype.Extension())
		if err != nil {
			logger.Error("Could not create temporary file")
		}
		defer func() {
			if err := tmpFile.Close(); err != nil {
				logger.Warning("failed to close temp file: %v", err)
			}
		}()

		if _, err := tmpFile.Write(data); err != nil {
			logger.Error("Could not write to temporary file")
		}

		clipboard.CopyFile(tmpFile.Name())
		logger.Verbose("✅ Copied stream as temporary file: %s", tmpFile.Name())
	}
}

// Clean up old temp files that are no longer in clipboard
func cleanupOldTempFiles() {
	// No need to wait - called after main operation completes

	// Get current clipboard files
	files := clipboard.GetFiles()

	// Build a map of clipboard files for quick lookup
	clipboardMap := make(map[string]bool)
	for _, file := range files {
		clipboardMap[file] = true
	}

	// Find only clippy temp files using glob
	tempDir := os.TempDir()
	pattern := filepath.Join(tempDir, "clippy-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, fullPath := range matches {
		// Check if this file is in the clipboard
		if !clipboardMap[fullPath] {
			// Not in clipboard, safe to delete
			if verbose {
				info, err := os.Stat(fullPath)
				if err == nil {
					name := filepath.Base(fullPath)
					fmt.Fprintf(os.Stderr, "Cleaning up old temp file: %s (created %v ago)\n",
						name, time.Since(info.ModTime()).Round(time.Minute))
				}
			}
			if err := os.Remove(fullPath); err != nil {
				logger.Warning("Failed to remove temp file %s: %v", filepath.Base(fullPath), err)
			}
		}
	}
}
