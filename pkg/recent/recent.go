package recent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

// FileInfo represents a file with its metadata
type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsDir    bool
	MimeType string // MIME type of the file (empty for directories)
}

// FindOptions controls how recent files are discovered
type FindOptions struct {
	MaxAge         time.Duration
	MaxCount       int
	Directories    []string
	Extensions     []string
	ExcludeTemp    bool
	SmartUnarchive bool // Look inside auto-unarchived folders
}

// ArchiveInfo represents information about an auto-unarchived download
type ArchiveInfo struct {
	OriginalName string // e.g. "project.zip"
	FolderPath   string // e.g. "/Users/neil/Downloads/project/"
	Contents     []FileInfo
}

// DefaultFindOptions returns sensible defaults for finding recent files
func DefaultFindOptions() FindOptions {
	return FindOptions{
		MaxAge:         2 * 24 * time.Hour, // 2 days - reasonable for "recent" downloads
		MaxCount:       10,
		Directories:    GetDefaultDownloadDirs(),
		ExcludeTemp:    true,
		SmartUnarchive: true,
	}
}

// GetDefaultDownloadDirs returns common download directories on macOS
func GetDefaultDownloadDirs() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{"/tmp"}
	}

	return []string{
		filepath.Join(homeDir, "Downloads"),
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "Documents"),
	}
}

// GetBrowserDownloadDir attempts to detect browser-specific download directories
func GetBrowserDownloadDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback if we can't get home directory
		return os.TempDir()
	}

	// Default to ~/Downloads - most browsers use this
	defaultDir := filepath.Join(homeDir, "Downloads")

	// TODO: Could check browser preferences here
	// Chrome: ~/Library/Application Support/Google/Chrome/Default/Preferences
	// Safari: ~/Library/Safari/Downloads.plist
	// Firefox: ~/.mozilla/firefox/profiles.ini

	return defaultDir
}

// FindRecentFiles finds files matching the given criteria
func FindRecentFiles(opts FindOptions) ([]FileInfo, error) {
	var allFiles []FileInfo

	cutoff := time.Now().Add(-opts.MaxAge)

	for _, dir := range opts.Directories {
		if !dirExists(dir) {
			continue
		}

		files, err := findFilesInDir(dir, cutoff, opts)
		if err != nil {
			// Log error but continue with other directories
			continue
		}

		allFiles = append(allFiles, files...)
	}

	// Sort by modification time, newest first
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Modified.After(allFiles[j].Modified)
	})

	// Limit results
	if opts.MaxCount > 0 && len(allFiles) > opts.MaxCount {
		allFiles = allFiles[:opts.MaxCount]
	}

	return allFiles, nil
}

// FindMostRecentFile finds the single most recent file
func FindMostRecentFile(opts FindOptions) (*FileInfo, error) {
	opts.MaxCount = 1
	files, err := FindRecentFiles(opts)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no recent files found")
	}

	return &files[0], nil
}

// ParseDuration parses duration strings like "5m", "1h", "30s"
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 5 * time.Minute, nil
	}

	// Handle just numbers (assume minutes)
	if num, err := strconv.Atoi(s); err == nil {
		if num < 0 {
			return 0, fmt.Errorf("duration cannot be negative")
		}
		return time.Duration(num) * time.Minute, nil
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if duration < 0 {
		return 0, fmt.Errorf("duration cannot be negative")
	}
	return duration, nil
}

// ParseRecentArgument parses the argument to -r or -i flags
// Returns count (number of files) and maxAge (time duration)
// Examples:
//   - "" or " " -> count=1, maxAge=0 (default)
//   - "3" -> count=3, maxAge=0
//   - "5m" -> count=0, maxAge=5 minutes (0 means all files in period)
func ParseRecentArgument(arg string) (count int, maxAge time.Duration, err error) {
	// Default behavior for empty argument
	if arg == "" || arg == " " {
		return 1, 0, nil
	}

	// Try to parse as a number first
	if num, parseErr := strconv.Atoi(arg); parseErr == nil && num > 0 {
		return num, 0, nil
	}

	// Parse as duration
	duration, err := ParseDuration(arg)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid argument %q: use a number like '3' or duration like '5m'", arg)
	}

	return 0, duration, nil
}

// findFilesInDir recursively finds files in a directory
func findFilesInDir(dir string, cutoff time.Time, opts FindOptions) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip the root directory itself
		if path == dir {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip temporary files
		if opts.ExcludeTemp && isTemporaryFile(info.Name()) {
			return nil
		}

		// Check modification time
		if info.ModTime().Before(cutoff) {
			return nil
		}

		// Skip directories - we only want files
		if info.IsDir() {
			return nil
		}

		// Check extensions if specified
		if len(opts.Extensions) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if !contains(opts.Extensions, ext) {
				return nil
			}
		}

		// Detect MIME type
		mtype, _ := mimetype.DetectFile(path)
		mimeType := ""
		if mtype != nil {
			mimeType = mtype.String()
		}

		files = append(files, FileInfo{
			Path:     path,
			Name:     info.Name(),
			Size:     info.Size(),
			Modified: info.ModTime(),
			IsDir:    false,
			MimeType: mimeType,
		})

		return nil
	})

	return files, err
}

// isTemporaryFile checks if a file appears to be temporary
func isTemporaryFile(name string) bool {
	tempSuffixes := []string{
		".tmp", ".temp", ".download", ".partial", ".crdownload",
		".part", ".filepart", ".opdownload",
	}

	lowerName := strings.ToLower(name)
	for _, suffix := range tempSuffixes {
		if strings.HasSuffix(lowerName, suffix) {
			return true
		}
	}

	return false
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsArchive checks if a file is a common archive format
func IsArchive(filename string) bool {
	archiveExts := []string{
		".zip", ".tar", ".tar.gz", ".tgz", ".tar.bz2", ".tbz2",
		".tar.xz", ".txz", ".7z", ".rar", ".gz", ".bz2", ".xz",
		".dmg", ".pkg",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, archiveExt := range archiveExts {
		if ext == archiveExt {
			return true
		}
	}

	// Handle .tar.gz and similar
	if strings.Contains(strings.ToLower(filename), ".tar.") {
		return true
	}

	return false
}

// HIGH-LEVEL BUSINESS CASE FUNCTIONS

// CopyMostRecentDownload finds the most recent download and copies it to clipboard
// This is the primary use case: "I just downloaded something, copy it to clipboard"
func CopyMostRecentDownload(maxAge time.Duration) (*FileInfo, error) {
	opts := DefaultFindOptions()
	if maxAge != 0 {
		opts.MaxAge = maxAge
	}

	file, err := FindMostRecentFile(opts)
	if err != nil {
		return nil, err
	}

	// Handle auto-unarchived folders
	if file.IsDir && opts.SmartUnarchive {
		if unarchived := detectAutoUnarchived(file); unarchived != nil {
			// If it's a single file inside, use that
			if len(unarchived.Contents) == 1 && !unarchived.Contents[0].IsDir {
				return &unarchived.Contents[0], nil
			}
			// Otherwise return the folder itself
		}
	}

	return file, nil
}

// CopyRecentDownloads finds multiple recent downloads and copies them to clipboard
// This handles the case where multiple files were downloaded as a batch
func CopyRecentDownloads(maxAge time.Duration, maxCount int) ([]FileInfo, error) {
	opts := DefaultFindOptions()
	if maxAge != 0 {
		opts.MaxAge = maxAge
	}
	if maxCount > 0 {
		opts.MaxCount = maxCount
	}

	files, err := FindRecentFiles(opts)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no recent files found")
	}

	// Group files by their download time (within 30 seconds = batch)
	batches := groupFilesByDownloadTime(files, 30*time.Second)

	// Return the most recent batch
	if len(batches) > 0 {
		return batches[0], nil
	}

	return files, nil
}

// PasteRecentDownloads finds and copies multiple recent downloads to destination
// This handles batch downloads like "I downloaded 5 photos, paste them all"
func PasteRecentDownloads(destination string, maxAge time.Duration, maxCount int) ([]FileInfo, error) {
	files, err := CopyRecentDownloads(maxAge, maxCount)
	if err != nil {
		return nil, err
	}

	if destination == "" {
		destination = "."
	}

	// Copy all files to destination
	for _, file := range files {
		err = CopyFileToDestination(file.Path, destination)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %s: %w", file.Name, err)
		}
	}

	return files, nil
}

// PickerResult represents the result of an interactive file picker
type PickerResult struct {
	Files     []*FileInfo
	PasteMode bool // true if user pressed 'p' to copy & paste
}

// PickRecentDownload returns a single recent download
// This handles the case where you want to select from multiple recent files
type PickerConfig struct {
	MaxAge       time.Duration
	AbsoluteTime bool
}

func PickRecentDownload(maxAge time.Duration) (*FileInfo, error) {
	config := PickerConfig{
		MaxAge:       maxAge,
		AbsoluteTime: false,
	}
	return PickRecentDownloadWithConfig(config)
}

// GetRecentDownloads returns recent files for picker display
func GetRecentDownloads(config PickerConfig, maxCount int) ([]FileInfo, error) {
	opts := DefaultFindOptions()
	if config.MaxAge != 0 {
		opts.MaxAge = config.MaxAge
	}
	if maxCount > 0 {
		opts.MaxCount = maxCount
	} else {
		opts.MaxCount = 20 // Default to 20 if not specified
	}

	files, err := FindRecentFiles(opts)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no recent files found")
	}

	return files, nil
}

func PickRecentDownloadWithConfig(config PickerConfig) (*FileInfo, error) {
	files, err := GetRecentDownloads(config, 0) // Use default maxCount
	if err != nil {
		return nil, err
	}

	// If only one file, return it directly
	if len(files) == 1 {
		return &files[0], nil
	}

	// Return an error - selection must be handled at the interface layer
	return nil, fmt.Errorf("multiple files found, selection required")
}

// PickMultipleRecentDownloads is deprecated - use GetRecentDownloads instead
func PickMultipleRecentDownloads(config PickerConfig) ([]*FileInfo, error) {
	files, err := GetRecentDownloads(config, 0) // Use default maxCount
	if err != nil {
		return nil, err
	}
	// Convert to pointers for backward compatibility
	var result []*FileInfo
	for i := range files {
		result = append(result, &files[i])
	}
	return result, nil
}

// PickRecentDownloadsWithResult is deprecated - use GetRecentDownloads instead
func PickRecentDownloadsWithResult(config PickerConfig) (*PickerResult, error) {
	return nil, fmt.Errorf("selection functionality not available in core library - use GetRecentDownloads instead")
}

// PasteMostRecentDownload finds and copies the most recent download to destination
// This is the primary use case: "I just downloaded something, paste it here"
func PasteMostRecentDownload(destination string, maxAge time.Duration) (*FileInfo, error) {
	file, err := CopyMostRecentDownload(maxAge)
	if err != nil {
		return nil, err
	}

	if destination == "" {
		destination = "."
	}

	// Copy the file to destination
	err = CopyFileToDestination(file.Path, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return file, nil
}

// CopyFileToDestination copies a file or directory to the specified destination
func CopyFileToDestination(srcPath, destPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	// If destination is a directory, copy into it
	if destInfo, err := os.Stat(destPath); err == nil && destInfo.IsDir() {
		destPath = filepath.Join(destPath, filepath.Base(srcPath))
	}

	if srcInfo.IsDir() {
		return copyDir(srcPath, destPath)
	}

	return CopyFile(srcPath, destPath)
}

// copyFile copies a single file
// CopyFile copies a file from src to dst, preserving permissions and creating directories as needed
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	if info, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, info.Mode())
	}

	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return CopyFile(path, dstPath)
	})
}

// detectAutoUnarchived checks if a directory looks like an auto-unarchived download
func detectAutoUnarchived(dir *FileInfo) *ArchiveInfo {
	if !dir.IsDir {
		return nil
	}

	// Check if directory name suggests it was unarchived
	dirName := filepath.Base(dir.Path)

	// Look for common archive patterns in the name
	// e.g. "project" folder might have come from "project.zip"
	archiveExtensions := []string{".zip", ".tar.gz", ".tgz", ".tar"}

	for _, ext := range archiveExtensions {
		possibleArchiveName := dirName + ext

		// Check if this looks like an auto-unarchived folder
		// (created recently, contains files, name suggests archive origin)
		if time.Since(dir.Modified) < 10*time.Minute {
			contents, err := getDirectoryContents(dir.Path)
			if err == nil && len(contents) > 0 {
				return &ArchiveInfo{
					OriginalName: possibleArchiveName,
					FolderPath:   dir.Path,
					Contents:     contents,
				}
			}
		}
	}

	return nil
}

// getDirectoryContents returns the contents of a directory
func getDirectoryContents(dirPath string) ([]FileInfo, error) {
	var contents []FileInfo

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip the root directory
		if path == dirPath {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Detect MIME type for files
		mimeType := ""
		if !info.IsDir() {
			mtype, _ := mimetype.DetectFile(path)
			if mtype != nil {
				mimeType = mtype.String()
			}
		}

		contents = append(contents, FileInfo{
			Path:     path,
			Name:     info.Name(),
			Size:     info.Size(),
			Modified: info.ModTime(),
			IsDir:    info.IsDir(),
			MimeType: mimeType,
		})

		return nil
	})

	return contents, err
}

// groupFilesByDownloadTime groups files that were downloaded within the same time window
// This helps identify batch downloads (e.g., multiple files downloaded from the same page)
func groupFilesByDownloadTime(files []FileInfo, window time.Duration) [][]FileInfo {
	if len(files) == 0 {
		return nil
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Modified.After(files[j].Modified)
	})

	var batches [][]FileInfo
	var currentBatch []FileInfo

	for i, file := range files {
		if i == 0 {
			// First file starts the first batch
			currentBatch = []FileInfo{file}
		} else {
			// Check if this file is within the time window of the batch
			timeDiff := currentBatch[0].Modified.Sub(file.Modified)
			if timeDiff <= window {
				// Add to current batch
				currentBatch = append(currentBatch, file)
			} else {
				// Start new batch
				batches = append(batches, currentBatch)
				currentBatch = []FileInfo{file}
			}
		}
	}

	// Add the final batch
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}
