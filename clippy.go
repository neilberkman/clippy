// Package clippy provides smart clipboard operations for macOS.
// It automatically detects whether to copy file content or file references
// based on MIME type detection.
package clippy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/neilberkman/clippy/pkg/clipboard"
)

// Copy intelligently copies a file to clipboard.
// Text files copy their content, binary files copy file references.
func Copy(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", path, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", absPath)
	}

	// Detect MIME type
	mtype, err := mimetype.DetectFile(absPath)
	if err != nil {
		return fmt.Errorf("could not detect file type for %s: %w", absPath, err)
	}

	// Text files: copy content, others: copy file reference
	if strings.HasPrefix(mtype.String(), "text/") {
		content, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("could not read file content %s: %w", absPath, err)
		}
		clipboard.CopyText(string(content))
	} else {
		clipboard.CopyFile(absPath)
	}

	return nil
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
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("could not write to temporary file: %w", err)
	}

	clipboard.CopyFile(tmpFile.Name())
	return nil
}

// GetText returns text content from clipboard.
func GetText() (string, bool) {
	return clipboard.GetText()
}

// GetFiles returns file paths from clipboard.
func GetFiles() []string {
	return clipboard.GetFiles()
}