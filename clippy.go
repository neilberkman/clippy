// clippy.go
package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/NSPasteboard.h>
#import <AppKit/NSApplication.h>
#import <CoreServices/CoreServices.h>

// Function to copy a file reference to the clipboard
void copyFile(const char *path) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSURL *fileURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:path]];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard writeObjects:@[fileURL]];
    }
}

// Function to copy multiple file references to the clipboard
void copyFiles(const char **paths, int count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSMutableArray *fileURLs = [NSMutableArray arrayWithCapacity:count];

        for (int i = 0; i < count; i++) {
            NSURL *fileURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:paths[i]]];
            [fileURLs addObject:fileURL];
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard writeObjects:fileURLs];
    }
}

// Function to copy plain text content to the clipboard
void copyText(const char *text) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSString *nsText = [NSString stringWithUTF8String:text];
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard setString:nsText forType:NSPasteboardTypeString];
    }
}

// Get current clipboard file paths if any
char** getClipboardFiles(int *count) {
    @autoreleasepool {
        [NSApplication sharedApplication]; // Initialize the app context
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *files = [pasteboard readObjectsForClasses:@[[NSURL class]]
                                                   options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];

        *count = (int)[files count];
        if (*count == 0) return NULL;

        char **paths = (char**)malloc(sizeof(char*) * (*count));
        for (int i = 0; i < *count; i++) {
            NSURL *url = files[i];
            const char *path = [[url path] UTF8String];
            paths[i] = strdup(path);
        }

        return paths;
    }
}

// Free the file paths array
void freeFilePaths(char **paths, int count) {
    if (!paths) return;
    for (int i = 0; i < count; i++) {
        free(paths[i]);
    }
    free(paths);
}
*/
import "C"
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
	"unsafe"

	"github.com/gabriel-vasile/mimetype"
)

var (
	verbose bool
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse flags
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")

	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *showVersion {
		fmt.Printf("clippy version %s (%s) built on %s\n", version, commit, date)
		os.Exit(0)
	}

	// Load config file
	loadConfig()

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
	cleanupOldTempFiles()
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
		}
	}
}

// Logic for when a filename is provided as an argument
func handleFileMode(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid path %s\n", filePath)
		os.Exit(1)
	}

	// Detect MIME type to check if it's text
	mtype, err := mimetype.DetectFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not read file %s\n", absPath)
		os.Exit(1)
	}

	// Special case for text files: copy content, not file reference
	if strings.HasPrefix(mtype.String(), "text/") {
		content, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not read file content %s\n", absPath)
			os.Exit(1)
		}
		cContent := C.CString(string(content))
		defer C.free(unsafe.Pointer(cContent))
		C.copyText(cContent)
		if verbose {
			fmt.Printf("✅ Copied text content from '%s'\n", filepath.Base(absPath))
		}
	} else {
		// For all other file types, copy as a file reference
		cPath := C.CString(absPath)
		defer C.free(unsafe.Pointer(cPath))
		C.copyFile(cPath)
		if verbose {
			fmt.Printf("✅ Copied file reference for '%s'\n", filepath.Base(absPath))
		}
	}
}

// Handle multiple files at once
func handleMultipleFiles(paths []string) {
	// Check if all files exist and collect absolute paths
	absPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not resolve path %s\n", path)
			os.Exit(1)
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Could not read file %s\n", absPath)
			os.Exit(1)
		}

		absPaths = append(absPaths, absPath)
	}

	// Convert paths to C strings
	cPaths := make([]*C.char, len(absPaths))
	for i, path := range absPaths {
		cPaths[i] = C.CString(path)
		defer C.free(unsafe.Pointer(cPaths[i]))
	}

	// Copy all files at once
	C.copyFiles(&cPaths[0], C.int(len(cPaths)))

	if verbose {
		fmt.Printf("✅ Copied %d file references\n", len(absPaths))
		for _, path := range absPaths {
			fmt.Printf("  - %s\n", filepath.Base(path))
		}
	}
}

// Logic for when data is piped via stdin
func handleStreamMode() {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from stdin")
		os.Exit(1)
	}
	data := buf.Bytes()
	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: Input stream was empty.")
		return
	}

	mtype := mimetype.Detect(data)

	// Special case for text streams: copy content
	if strings.HasPrefix(mtype.String(), "text/") {
		cContent := C.CString(string(data))
		defer C.free(unsafe.Pointer(cContent))
		C.copyText(cContent)
		if verbose {
			fmt.Println("✅ Copied text content from stream")
		}
	} else {
		// For binary streams, save to a temp file, then copy the reference
		tmpFile, err := os.CreateTemp("", "clippy-*"+mtype.Extension())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not create temporary file")
			os.Exit(1)
		}
		defer func() {
			if err := tmpFile.Close(); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to close temp file: %v\n", err)
			}
		}()

		if _, err := tmpFile.Write(data); err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not write to temporary file")
			os.Exit(1)
		}

		cPath := C.CString(tmpFile.Name())
		defer C.free(unsafe.Pointer(cPath))
		C.copyFile(cPath)

		if verbose {
			fmt.Printf("✅ Copied stream as temporary file: %s\n", tmpFile.Name())
		}
	}
}

// Clean up old temp files that are no longer in clipboard
func cleanupOldTempFiles() {
	// No need to wait - called after main operation completes

	// Get current clipboard files
	var count C.int
	clipboardFiles := C.getClipboardFiles(&count)
	defer C.freeFilePaths(clipboardFiles, count)

	// Build a map of clipboard files for quick lookup
	clipboardMap := make(map[string]bool)
	if clipboardFiles != nil {
		// Convert C array to Go slice
		cFiles := (*[1 << 30]*C.char)(unsafe.Pointer(clipboardFiles))[:count:count]
		for i := 0; i < int(count); i++ {
			path := C.GoString(cFiles[i])
			clipboardMap[path] = true
		}
	}

	// Scan temp directory for clippy files
	tempDir := os.TempDir()
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "clippy-") {
			fullPath := filepath.Join(tempDir, name)

			// Check if this file is in the clipboard
			if !clipboardMap[fullPath] {
				// Not in clipboard, safe to delete
				if verbose {
					info, _ := entry.Info()
					if info != nil {
						fmt.Fprintf(os.Stderr, "Cleaning up old temp file: %s (created %v ago)\n",
							name, time.Since(info.ModTime()).Round(time.Minute))
					}
				}
				if err := os.Remove(fullPath); err != nil && verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to remove temp file %s: %v\n", name, err)
				}
			}
		}
	}
}
