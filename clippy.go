// Package clippy provides smart clipboard operations for macOS.
// It automatically detects whether to copy file content or file references
// using hybrid detection: UTI -> MIME -> mimetype fallback for maximum reliability.
package clippy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/neilberkman/clippy/pkg/clipboard"
)

// CopyResult contains information about what was copied and how
type CopyResult struct {
	Method   string // "UTI", "MIME", or "content"
	Type     string // The detected type (UTI or MIME)
	AsText   bool   // Whether content was copied as text
	FilePath string // The file path that was copied
}

// Copy intelligently copies a file to clipboard.
// Text files copy their content, binary files copy file references.
// Uses hybrid detection: UTI -> MIME -> mimetype fallback.
func Copy(path string) error {
	_, err := CopyWithResult(path)
	return err
}

// CopyWithResult is like Copy but returns information about the detection method used
func CopyWithResult(path string) (*CopyResult, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %s: %w", path, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", absPath)
	}

	// Try UTI detection first (more reliable for macOS)
	if uti, ok := clipboard.GetUTIForFile(absPath); ok {
		// For dynamic UTIs (unknown types), skip to MIME detection
		if strings.HasPrefix(uti, "dyn.") {
			// Fall through to MIME detection
		} else if isTextUTI(uti) {
			content, err := os.ReadFile(absPath)
			if err != nil {
				return nil, fmt.Errorf("could not read file content %s: %w", absPath, err)
			}
			clipboard.CopyText(string(content))
			return &CopyResult{
				Method:   "UTI",
				Type:     uti,
				AsText:   true,
				FilePath: absPath,
			}, nil
		} else {
			// Non-text UTI, copy as file reference
			clipboard.CopyFile(absPath)
			return &CopyResult{
				Method:   "UTI",
				Type:     uti,
				AsText:   false,
				FilePath: absPath,
			}, nil
		}
	}

	// Fallback to MIME type detection
	mtype, err := mimetype.DetectFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not detect file type for %s: %w", absPath, err)
	}

	// Text files: copy content, others: copy file reference
	if strings.HasPrefix(mtype.String(), "text/") {
		content, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("could not read file content %s: %w", absPath, err)
		}
		clipboard.CopyText(string(content))
		return &CopyResult{
			Method:   "MIME",
			Type:     mtype.String(),
			AsText:   true,
			FilePath: absPath,
		}, nil
	} else {
		clipboard.CopyFile(absPath)
		return &CopyResult{
			Method:   "MIME",
			Type:     mtype.String(),
			AsText:   false,
			FilePath: absPath,
		}, nil
	}
}

// CopyMultiple copies multiple files to clipboard as file references.
func CopyMultiple(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no files provided")
	}

	// Convert to absolute paths and verify all files exist
	absPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", absPath)
		}

		absPaths = append(absPaths, absPath)
	}

	clipboard.CopyFiles(absPaths)
	return nil
}

// CopyText copies text content to clipboard.
func CopyText(text string) {
	clipboard.CopyText(text)
}

// CopyData copies data from a reader to clipboard.
// Text data is copied as text, binary data is saved to a temp file.
// Uses MIME type detection for content analysis.
func CopyData(reader io.Reader) error {
	return CopyDataWithTempDir(reader, "")
}

// CopyDataWithTempDir is like CopyData but allows specifying a custom temp directory.
func CopyDataWithTempDir(reader io.Reader, tempDir string) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	data := buf.Bytes()
	if len(data) == 0 {
		return fmt.Errorf("input data was empty")
	}

	// Detect MIME type from content
	mtype := mimetype.Detect(data)

	// Text data: copy as text
	if strings.HasPrefix(mtype.String(), "text/") {
		clipboard.CopyText(string(data))
		return nil
	}

	// Binary data: save to temp file and copy reference
	tmpFile, err := os.CreateTemp(tempDir, "clippy-*"+mtype.Extension())
	if err != nil {
		return fmt.Errorf("could not create temporary file: %w", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close temporary file: %v\n", err)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("could not write to temporary file: %w", err)
	}

	clipboard.CopyFile(tmpFile.Name())
	return nil
}

// GetText returns text content from clipboard.
// Uses hybrid detection for better reliability.
func GetText() (string, bool) {
	// Try hybrid detection first
	if content, err := clipboard.GetClipboardContent(); err == nil {
		if content.IsText {
			return string(content.Data), true
		}
	}

	// Fallback to simple text detection
	return clipboard.GetText()
}

// GetFiles returns file paths from clipboard.
// Uses hybrid detection for better reliability.
func GetFiles() []string {
	// Try hybrid detection first
	if content, err := clipboard.GetClipboardContent(); err == nil {
		if content.IsFile {
			return []string{content.FilePath}
		}
	}

	// Fallback to simple file detection
	return clipboard.GetFiles()
}

// isTextUTI checks if a UTI represents text content using macOS UTI system
func isTextUTI(uti string) bool {
	// Use macOS UTI system to check if this UTI conforms to text types
	conformsToText := clipboard.UTIConformsTo(uti, "public.text") ||
		clipboard.UTIConformsTo(uti, "public.plain-text") ||
		clipboard.UTIConformsTo(uti, "public.source-code")

	// For dynamic UTIs (unknown types), let MIME detection handle it
	if strings.HasPrefix(uti, "dyn.") {
		return false
	}

	return conformsToText
}

// CleanupTempFiles removes old temporary files that are no longer in clipboard
func CleanupTempFiles(tempDir string, verbose bool) {
	// Get current clipboard files
	files := GetFiles()

	// Build a map of clipboard files for quick lookup
	clipboardMap := make(map[string]bool)
	for _, file := range files {
		clipboardMap[file] = true
	}

	// Find only clippy temp files using glob
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	pattern := filepath.Join(tempDir, "clippy-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, fullPath := range matches {
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		age := time.Since(info.ModTime())

		// Check if this file is in the clipboard
		if !clipboardMap[fullPath] {
			// Only delete files older than 5 minutes to avoid race conditions
			// with parallel clippy/pasty operations
			if age >= 5*time.Minute {
				if verbose {
					name := filepath.Base(fullPath)
					fmt.Fprintf(os.Stderr, "Cleaning up old temp file: %s (created %v ago)\n",
						name, age.Round(time.Minute))
				}
				if err := os.Remove(fullPath); err != nil {
					if verbose {
						fmt.Fprintf(os.Stderr, "Warning: Failed to remove temp file %s: %v\n", filepath.Base(fullPath), err)
					}
				}
			}
		}
	}
}

// PasteResult contains information about what was pasted
type PasteResult struct {
	Type      string   // "text" or "files"
	Content   string   // Text content if Type is "text"
	Files     []string // File paths if Type is "files"
	FilesRead int      // Number of files successfully read/copied
}

// PasteToStdout pastes clipboard content to stdout
func PasteToStdout() (*PasteResult, error) {
	// Try to get file references first (prioritize files over text)
	files := GetFiles()
	if len(files) > 0 {
		for _, file := range files {
			fmt.Println(file)
		}
		return &PasteResult{
			Type:  "files",
			Files: files,
		}, nil
	}

	// Try to get text content
	if text, ok := GetText(); ok {
		fmt.Print(text)
		return &PasteResult{
			Type:    "text",
			Content: text,
		}, nil
	}

	return nil, fmt.Errorf("no text or file content found on clipboard")
}

// PasteToFile pastes clipboard content to a file or directory
func PasteToFile(destination string) (*PasteResult, error) {
	// Try to get file references first (prioritize files over text)
	files := GetFiles()
	if len(files) > 0 {
		filesRead, err := copyFilesToDestination(files, destination)
		if err != nil {
			return nil, err
		}
		return &PasteResult{
			Type:      "files",
			Files:     files,
			FilesRead: filesRead,
		}, nil
	}

	// Try to get text content
	if text, ok := GetText(); ok {
		if err := os.WriteFile(destination, []byte(text), 0644); err != nil {
			return nil, fmt.Errorf("could not write to file %s: %w", destination, err)
		}
		return &PasteResult{
			Type:    "text",
			Content: text,
		}, nil
	}

	return nil, fmt.Errorf("no text or file content found on clipboard")
}

// copyFilesToDestination copies files from clipboard to destination
func copyFilesToDestination(files []string, destination string) (int, error) {
	if len(files) == 0 {
		return 0, fmt.Errorf("no files to copy")
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
			return 0, fmt.Errorf("could not create directory %s: %w", destination, err)
		}
	}

	// Copy each file
	filesRead := 0
	for _, srcFile := range files {
		var destFile string
		if destIsDir {
			destFile = filepath.Join(destination, filepath.Base(srcFile))
		} else {
			destFile = destination
		}

		if err := copyFile(srcFile, destFile); err != nil {
			return filesRead, fmt.Errorf("could not copy %s to %s: %w", srcFile, destFile, err)
		}

		filesRead++
	}

	return filesRead, nil
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
