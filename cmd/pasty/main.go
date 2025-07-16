// Pasty - Smart paste tool for macOS
// Companion to clippy, provides intelligent pasting from clipboard
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/internal/log"
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

	// Use library functions to paste content
	var result *clippy.PasteResult
	var err error

	if destination == "" {
		// Paste to stdout
		result, err = clippy.PasteToStdout()
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
