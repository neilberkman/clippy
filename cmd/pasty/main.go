// Pasty - Smart paste tool for macOS
// Companion to clippy, provides intelligent pasting from clipboard
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/clipboard"
)

var (
	verbose bool
	version = "dev"
	commit  = "none"
	date    = "unknown"
	logger  *log.Logger
)

func main() {
	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `pasty - Smart paste tool for macOS

Usage:
  pasty [options] [destination]

Options:
  -v, --verbose    Enable verbose output
  --version        Show version information
  -h, --help       Show this help message

Examples:
  # Paste clipboard content to stdout
  pasty

  # Paste to a specific file
  pasty output.txt

  # Paste and show what was pasted
  pasty -v

Description:
  Pasty intelligently pastes clipboard content:
  - Text content is written directly
  - File references are copied to destination
  - If no destination specified, outputs to stdout
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
		fmt.Printf("pasty version %s (%s) built on %s\n", version, commit, date)
		os.Exit(0)
	}

	// Initialize logger
	logger = log.New(log.Config{Verbose: verbose})

	// Get destination from args
	args := flag.Args()
	var destination string
	if len(args) > 0 {
		destination = args[0]
	}

	// Try to get text content first
	if text, ok := clipboard.GetText(); ok {
		handleTextContent(text, destination)
		return
	}

	// Try to get file references
	files := clipboard.GetFiles()
	if len(files) > 0 {
		handleFileReferences(files, destination)
		return
	}

	// Nothing on clipboard
	logger.Error("No text or file content found on clipboard")
}

// handleTextContent pastes text content
func handleTextContent(text, destination string) {
	if destination == "" {
		// Output to stdout
		fmt.Print(text)
		logger.Verbose("✅ Pasted text content to stdout")
	} else {
		// Write to file
		if err := os.WriteFile(destination, []byte(text), 0644); err != nil {
			logger.Error("Could not write to file %s: %v", destination, err)
		}
		logger.Verbose("✅ Pasted text content to '%s'", destination)
	}
}

// handleFileReferences copies files from clipboard
func handleFileReferences(files []string, destination string) {
	if destination == "" {
		// List files to stdout
		for _, file := range files {
			fmt.Println(file)
		}
		logger.Verbose("✅ Listed %d file references from clipboard", len(files))
		return
	}

	// Determine if destination should be a directory
	destIsDir := false
	if len(files) > 1 {
		destIsDir = true
	} else if strings.HasSuffix(destination, "/") {
		destIsDir = true
	} else if stat, err := os.Stat(destination); err == nil && stat.IsDir() {
		destIsDir = true
	}

	if destIsDir {
		// Ensure destination directory exists
		if err := os.MkdirAll(destination, 0755); err != nil {
			logger.Error("Could not create directory %s: %v", destination, err)
		}
	}

	// Copy each file
	for _, srcFile := range files {
		var destFile string
		if destIsDir {
			destFile = filepath.Join(destination, filepath.Base(srcFile))
		} else {
			destFile = destination
		}

		if err := copyFile(srcFile, destFile); err != nil {
			logger.Error("Could not copy %s to %s: %v", srcFile, destFile, err)
			continue
		}

		logger.Verbose("✅ Copied '%s' to '%s'", filepath.Base(srcFile), destFile)
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("could not read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("could not write destination file: %w", err)
	}

	return nil
}
