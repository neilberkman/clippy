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
	verbose         bool
	debug           bool
	preserveFormat  bool
	logger          *log.Logger
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

  # Save browser image (auto-converts TIFF to PNG)
  pasty photo.png

  # Paste to a file
  pasty output.txt

  # Show verbose output
  pasty -v

Description:
  Pasty intelligently pastes clipboard content:
  - Text content is written directly
  - Image data is saved (TIFF auto-converts to PNG)
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
				result, err = clippy.PasteToFileWithOptions(destination, clippy.PasteOptions{
					PreserveFormat: preserveFormat,
				})
			}

			if err != nil {
				logger.Error("%v", err)
			}

			// Show verbose output
			if result != nil {
				if destination == "" {
					if result.Type == "text" {
						logger.Verbose("Pasted text content to stdout")
					} else {
						logger.Verbose("Listed %d file references from clipboard", len(result.Files))
					}
				} else {
					switch result.Type {
					case "text":
						logger.Verbose("Pasted text content to '%s'", destination)
					case "image":
						logger.Verbose("Saved image data to '%s'", result.Files[0])
					case "rtfd":
						logger.Verbose("Saved rich text with embedded images to '%s'", result.Files[0])
					case "files":
						logger.Verbose("Copied %d files to '%s'", result.FilesRead, destination)
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
	rootCmd.Flags().BoolVar(&preserveFormat, "preserve-format", false, "Preserve original image format (skip TIFF to PNG conversion)")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
