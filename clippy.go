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
	"github.com/neilberkman/clippy/pkg/recent"
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
	return CopyWithResultAndMode(path, false)
}

// CopyWithResultAndMode is like CopyWithResult but allows forcing text mode
func CopyWithResultAndMode(path string, forceTextMode bool) (*CopyResult, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %s: %w", path, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", absPath)
	}

	// If forceTextMode is false (default), always copy as file reference
	if !forceTextMode {
		if err := clipboard.CopyFile(absPath); err != nil {
			return nil, fmt.Errorf("could not copy file to clipboard: %w", err)
		}

		// Still detect the type for informational purposes
		uti, _ := clipboard.GetUTIForFile(absPath)
		typeStr := uti
		method := "UTI"
		if typeStr == "" || strings.HasPrefix(typeStr, "dyn.") {
			mtype, _ := mimetype.DetectFile(absPath)
			if mtype != nil {
				typeStr = mtype.String()
				method = "MIME"
			}
		}

		return &CopyResult{
			Method:   method,
			Type:     typeStr,
			AsText:   false,
			FilePath: absPath,
		}, nil
	}

	// Force text mode is enabled, check if file is actually text
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
			// Use auto-detection for proper clipboard type
			if err := CopyTextWithAutoDetection(string(content)); err != nil {
				return nil, fmt.Errorf("could not copy text to clipboard: %w", err)
			}
			return &CopyResult{
				Method:   "UTI",
				Type:     uti,
				AsText:   true,
				FilePath: absPath,
			}, nil
		} else if !forceTextMode {
			// Non-text UTI and text mode not forced - copy as file
			if err := clipboard.CopyFile(absPath); err != nil {
				return nil, fmt.Errorf("could not copy file to clipboard: %w", err)
			}
			return &CopyResult{
				Method:   "UTI",
				Type:     uti,
				AsText:   false,
				FilePath: absPath,
			}, nil
		}
		// Non-text UTI but text mode forced - fall through to MIME detection
	}

	// Fallback to MIME type detection
	mtype, err := mimetype.DetectFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not detect file type for %s: %w", absPath, err)
	}

	// Text files with force text mode: copy content
	if forceTextMode && isTextualMimeType(mtype.String()) {
		content, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("could not read file content %s: %w", absPath, err)
		}
		// Use auto-detection for proper clipboard type
		if err := CopyTextWithAutoDetection(string(content)); err != nil {
			return nil, fmt.Errorf("could not copy text to clipboard: %w", err)
		}
		return &CopyResult{
			Method:   "MIME",
			Type:     mtype.String(),
			AsText:   true,
			FilePath: absPath,
		}, nil
	} else {
		// Binary files or text mode not forced: copy file reference
		if err := clipboard.CopyFile(absPath); err != nil {
			return nil, fmt.Errorf("could not copy file to clipboard: %w", err)
		}
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

	if err := clipboard.CopyFiles(absPaths); err != nil {
		return fmt.Errorf("could not copy files to clipboard: %w", err)
	}
	return nil
}

// CopyText copies text content to clipboard.
func CopyText(text string) error {
	return CopyTextWithAutoDetection(text)
}

// CopyTextWithAutoDetection copies text with auto-detected type
func CopyTextWithAutoDetection(text string) error {
	// Try to detect the content type
	mtype := mimetype.Detect([]byte(text))
	mimeStr := mtype.String()

	// Map common MIME types to UTI types for better macOS integration
	var utiType string
	switch {
	case strings.HasPrefix(mimeStr, "text/html"):
		utiType = "public.html"
	case mimeStr == "application/json":
		utiType = "public.json"
	case strings.HasPrefix(mimeStr, "text/xml") || mimeStr == "application/xml":
		utiType = "public.xml"
	case strings.HasPrefix(mimeStr, "text/markdown"):
		// Note: macOS doesn't have a standard markdown UTI, but some apps recognize this
		utiType = "net.daringfireball.markdown"
	case strings.HasPrefix(mimeStr, "text/rtf") || mimeStr == "application/rtf":
		utiType = "public.rtf"
	default:
		// Fall back to plain text for other text types
		return clipboard.CopyText(text)
	}

	// Use the detected type
	return clipboard.CopyTextWithType(text, utiType)
}

// CopyTextWithType copies text with a specific MIME type or UTI
func CopyTextWithType(text string, typeIdentifier string) error {
	// If it looks like a MIME type, try to convert to UTI
	if strings.Contains(typeIdentifier, "/") {
		typeIdentifier = mimeToUTI(typeIdentifier)
	}
	return clipboard.CopyTextWithType(text, typeIdentifier)
}

// CopyFileAsTextWithType copies a file's text content with a specific MIME type or UTI.
// This is a core function that handles file I/O - interface layer should not read files directly.
func CopyFileAsTextWithType(path string, typeIdentifier string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", path, err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", absPath)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", absPath, err)
	}

	return CopyTextWithType(string(content), typeIdentifier)
}

// mimeToUTI converts common MIME types to macOS UTI
func mimeToUTI(mime string) string {
	switch mime {
	case "text/html":
		return "public.html"
	case "application/json":
		return "public.json"
	case "text/xml", "application/xml":
		return "public.xml"
	case "text/plain":
		return "public.plain-text"
	case "text/rtf", "application/rtf":
		return "public.rtf"
	case "text/markdown":
		return "net.daringfireball.markdown"
	default:
		// Return as-is if we don't have a mapping
		return mime
	}
}

// isTextualMimeType checks if a MIME type represents textual content
// that should be copied as text rather than binary
func isTextualMimeType(mimeType string) bool {
	// All text/* types are textual
	if strings.HasPrefix(mimeType, "text/") {
		return true
	}

	// Common application/* types that are actually text
	textualApplicationTypes := []string{
		"application/json",
		"application/ld+json",     // JSON-LD
		"application/xml",
		"application/xhtml+xml",
		"application/javascript",
		"application/typescript",
		"application/ecmascript",
		"application/x-httpd-php",
		"application/x-sh",         // Shell scripts
		"application/x-csh",
		"application/x-python",
		"application/x-ruby",
		"application/x-perl",
		"application/sql",
		"application/graphql",
		"application/x-yaml",
		"application/toml",
		"application/x-tex",
		"application/x-latex",
		"application/rtf",
		"application/csv",
	}

	for _, appType := range textualApplicationTypes {
		if mimeType == appType {
			return true
		}
	}

	// Check for +xml or +json suffix (like application/atom+xml)
	if strings.HasSuffix(mimeType, "+xml") || strings.HasSuffix(mimeType, "+json") {
		return true
	}

	return false
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
	mimeStr := mtype.String()

	// Text data: copy as text with proper type
	if isTextualMimeType(mimeStr) {
		// Use our auto-detection to set proper clipboard type
		if err := CopyTextWithAutoDetection(string(data)); err != nil {
			return fmt.Errorf("could not copy text to clipboard: %w", err)
		}
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

	if err := clipboard.CopyFile(tmpFile.Name()); err != nil {
		return fmt.Errorf("could not copy file to clipboard: %w", err)
	}
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
	// Always use the proper GetFiles that returns all files
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

// ClearClipboard clears the clipboard
func ClearClipboard() error {
	return clipboard.Clear()
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
	// Priority 1: File references
	if files := GetFiles(); len(files) > 0 {
		return pasteFileReferences(files, destination)
	}

	// Priority 2: Image/rich content data
	if content, err := clipboard.GetClipboardContent(); err == nil && !content.IsText && !content.IsFile && len(content.Data) > 0 {
		return pasteImageData(content, destination)
	}

	// Priority 3: Text content
	if text, ok := GetText(); ok {
		return pasteTextContent(text, destination)
	}

	return nil, fmt.Errorf("no content found on clipboard")
}

// pasteFileReferences copies file references from clipboard to destination
func pasteFileReferences(files []string, destination string) (*PasteResult, error) {
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

// pasteImageData saves image/rich content data from clipboard to file
func pasteImageData(content *clipboard.ClipboardContent, destination string) (*PasteResult, error) {
	ext := getFileExtensionFromUTI(content.Type)
	if ext == "" {
		ext = ".dat"
	}
	defaultFilename := fmt.Sprintf("clipboard-%s%s", time.Now().Format("2006-01-02-150405"), ext)

	destPath := resolveDestinationPath(destination, defaultFilename, true)

	if err := os.WriteFile(destPath, content.Data, 0644); err != nil {
		return nil, fmt.Errorf("could not write to file %s: %w", destPath, err)
	}

	return &PasteResult{
		Type:  "image",
		Files: []string{destPath},
	}, nil
}

// pasteTextContent saves text content from clipboard to file
func pasteTextContent(text string, destination string) (*PasteResult, error) {
	defaultFilename := fmt.Sprintf("clipboard-%s.txt", time.Now().Format("2006-01-02-150405"))
	destPath := resolveDestinationPath(destination, defaultFilename, false)

	if err := os.WriteFile(destPath, []byte(text), 0644); err != nil {
		return nil, fmt.Errorf("could not write to file %s: %w", destPath, err)
	}

	return &PasteResult{
		Type:    "text",
		Content: text,
		Files:   []string{destPath},
	}, nil
}

// resolveDestinationPath determines the final file path for pasting content
// If destination is a directory or looks like one, joins it with defaultFilename
// If allowNoExtension is true, treats paths without extensions as directories
func resolveDestinationPath(destination string, defaultFilename string, allowNoExtension bool) string {
	destInfo, err := os.Stat(destination)

	// Existing directory
	if err == nil && destInfo.IsDir() {
		return filepath.Join(destination, defaultFilename)
	}

	// Path ends with /
	if strings.HasSuffix(destination, "/") {
		return filepath.Join(destination, defaultFilename)
	}

	// Path doesn't exist and has no extension (for image data)
	if allowNoExtension && err != nil && !strings.Contains(filepath.Base(destination), ".") {
		return filepath.Join(destination, defaultFilename)
	}

	// Use destination as-is (it's a file path)
	return destination
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

		if err := recent.CopyFile(srcFile, destFile); err != nil {
			return filesRead, fmt.Errorf("could not copy %s to %s: %w", srcFile, destFile, err)
		}

		filesRead++
	}

	return filesRead, nil
}

// getFileExtensionFromUTI returns the file extension for a UTI
// using macOS's canonical type database
func getFileExtensionFromUTI(uti string) string {
	ext := clipboard.GetPreferredExtensionForUTI(uti)
	if ext == "" {
		return ""
	}
	// Add the dot prefix if not present
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}
