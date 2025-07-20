// Pasty - Smart paste tool for macOS
// Companion to clippy, provides intelligent pasting from clipboard
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/cmd/internal/common"
	"github.com/neilberkman/clippy/internal/log"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	debug   bool
	logger  *log.Logger
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

Description:
  Pasty intelligently pastes clipboard content:
  - Text content is written directly
  - File references are copied to destination
  - If no destination specified, outputs to stdout`,
		Version: fmt.Sprintf("%s (%s) built on %s", common.Version, common.Commit, common.Date),
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			logger = common.SetupLogger(verbose, debug)

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
	common.AddCommonFlags(rootCmd, &verbose, &debug)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
