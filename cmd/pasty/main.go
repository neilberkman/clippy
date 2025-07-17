// Pasty - Smart paste tool for macOS
// Companion to clippy, provides intelligent pasting from clipboard
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/neilberkman/clippy/pkg/recent"
)

var (
	verbose     bool
	debug       bool
	recentFlag  string
	recentCount bool
	recentBatch bool
	recentPick  bool
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	logger      *log.Logger
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "pasty [destination]",
		Short: "Smart paste tool for macOS",
		Long: `pasty - Smart paste tool for macOS

Companion to clippy, provides intelligent pasting from clipboard.

Examples:
  # Paste clipboard content to stdout
  pasty

  # Paste to a specific file
  pasty output.txt

  # Paste and show what was pasted
  pasty -v

  # Copy most recent download to current directory
  pasty --recent
  pasty --recent --recent-time 10m  # last 10 minutes

  # Copy most recent download to specific directory
  pasty --recent /path/to/dest/
  
  # Copy batch of recent downloads
  pasty --recent --batch  # copy all files downloaded together
  
  # Interactive picker for recent downloads
  pasty --recent --pick   # choose from list of recent files

Description:
  Pasty intelligently pastes clipboard content:
  - Text content is written directly
  - File references are copied to destination
  - If no destination specified, outputs to stdout
  - With --recent, finds and copies most recent downloads`,
		Version: fmt.Sprintf("%s (%s) built on %s", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			logger = log.New(log.Config{Verbose: verbose || debug, Debug: debug})

			// Handle --recent flag
			if recentCount {
				handleRecentMode(recentFlag, recentBatch, recentPick, args)
				return
			}

			// Get destination from args
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
				}
			}

			if destination == "" {
				result, err = clippy.PasteToStdout()
			} else {
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
		},
	}

	// Add flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (includes technical details)")
	rootCmd.PersistentFlags().BoolVarP(&recentCount, "recent", "r", false, "Copy most recent file from Downloads")
	rootCmd.PersistentFlags().StringVar(&recentFlag, "recent-time", "", "Time duration for recent files (5m, 1h, etc.)")
	rootCmd.PersistentFlags().BoolVar(&recentBatch, "batch", false, "Copy all files from most recent batch download")
	rootCmd.PersistentFlags().BoolVarP(&recentPick, "pick", "p", false, "Show interactive picker for recent downloads")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleRecentMode handles the --recent flag
func handleRecentMode(timeStr string, batch bool, pick bool, args []string) {
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

	if pick {
		// Use picker mode - interactively select file
		file, err := recent.PastePickedRecentDownload(destination, maxAge)
		if err != nil {
			logger.Error("No file selected: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Selected: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))
		logger.Verbose("✅ Copied selected file to '%s'", destination)
	} else if batch {
		// Use batch mode - copy multiple files
		files, err := recent.PasteRecentDownloads(destination, maxAge, 10)
		if err != nil {
			logger.Error("No recent files found: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Found %d files in recent batch:", len(files))
		for _, file := range files {
			logger.Verbose("  - %s (modified %s ago)", file.Name, time.Since(file.Modified).Round(time.Second))
		}
		logger.Verbose("✅ Copied %d recent files to '%s'", len(files), destination)
	} else {
		// Use single file mode
		file, err := recent.PasteMostRecentDownload(destination, maxAge)
		if err != nil {
			logger.Error("No recent files found: %v", err)
			os.Exit(1)
		}

		logger.Verbose("Found recent file: %s (modified %s ago)", file.Path, time.Since(file.Modified).Round(time.Second))
		logger.Verbose("✅ Copied recent file to '%s'", destination)
	}
}