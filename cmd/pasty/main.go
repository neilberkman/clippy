// Pasty - Smart paste tool for macOS
// Companion to clippy, provides intelligent pasting from clipboard
package main

import (
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
	recentFlag = ""
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	logger     *log.Logger
)

func main() {
	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `pasty - Smart paste tool for macOS

Usage:
  pasty [options] [destination]

Options:
  -v, --verbose       Enable verbose output
  --recent [time]     Copy most recent file from Downloads (e.g. --recent 5m)
  --version           Show version information
  -h, --help          Show this help message

Examples:
  # Paste clipboard content to stdout
  pasty

  # Paste to a specific file
  pasty output.txt

  # Paste and show what was pasted
  pasty -v

  # Copy most recent download to current directory
  pasty --recent
  pasty --recent 10m  # last 10 minutes

  # Copy most recent download to specific directory
  pasty --recent /path/to/dest/

Description:
  Pasty intelligently pastes clipboard content:
  - Text content is written directly
  - File references are copied to destination
  - If no destination specified, outputs to stdout
  - With --recent, finds and copies most recent downloads
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

	// Handle --recent flag
	if recentFlagSet {
		handleRecentMode(recentFlag, flag.Args())
		return
	}

	// Get destination from args
	args := flag.Args()
	var destination string
	if len(args) > 0 {
		destination = args[0]
	}

	// Use library functions to paste content
	var result *clippy.PasteResult
	var err error

	if destination == "" {
		// Check if clipboard has files - if so, default to current directory
		if files := clippy.GetFiles(); len(files) > 0 {
			destination = "."
			result, err = clippy.PasteToFile(destination)
		} else {
			// Paste text to stdout
			result, err = clippy.PasteToStdout()
		}
	} else {
		// Paste to file or directory
		result, err = clippy.PasteToFile(destination)
	}

	if err != nil {
		logger.Error("%v", err)
	}

	// Show verbose output
	if result != nil {
		if destination == "" {
			if result.Type == "text" {
				logger.Verbose("✅ Pasted text content to stdout")
			} else {
				logger.Verbose("✅ Listed %d file references from clipboard", len(result.Files))
			}
		} else {
			if result.Type == "text" {
				logger.Verbose("✅ Pasted text content to '%s'", destination)
			} else {
				logger.Verbose("✅ Copied %d files to '%s'", result.FilesRead, destination)
				if verbose {
					for _, file := range result.Files {
						fmt.Fprintf(os.Stderr, "  - %s\n", filepath.Base(file))
					}
				}
			}
		}
	}
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string, args []string) {
	// Parse time duration (library handles this)
	maxAge, err := recent.ParseDuration(timeStr)
	if err != nil {
		logger.Error("Invalid time duration: %v", err)
		os.Exit(1)
	}

	// Determine destination
	destination := "."
	if len(args) > 0 {
		destination = args[0]
	}

	// Use high-level library function
	file, err := recent.PasteMostRecentDownload(destination, maxAge)
	if err != nil {
		logger.Error("No recent files found: %v", err)
		os.Exit(1)
	}

	logger.Verbose("Found recent file: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))
	logger.Verbose("✅ Copied recent file to '%s'", destination)
}
