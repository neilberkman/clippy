package main

// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/neilberkman/clippy"
	"github.com/neilberkman/clippy/pkg/recent"
)

// ClippyGetRecentDownloads finds recent files and returns them as a C-style array of strings.
// It follows C conventions:
// - Returns a null-terminated array of C strings.
// - The caller is responsible for freeing the array and its contents by calling ClippyFreeStringArray.
// - On error, it returns nil and provides a descriptive error message via the outError parameter.
//
//export ClippyGetRecentDownloads
func ClippyGetRecentDownloads(maxCount C.int, durationSecs C.int, outError **C.char) **C.char {
	// Set up config
	config := recent.PickerConfig{}

	// If duration is specified (non-zero), use it
	if durationSecs > 0 {
		config.MaxAge = time.Duration(durationSecs) * time.Second
	}
	// Otherwise, MaxAge stays zero which means use library default (2 days)

	// Get recent downloads
	files, err := recent.GetRecentDownloads(config, int(maxCount))
	if err != nil {
		*outError = C.CString(fmt.Sprintf("Error getting recent downloads: %v", err))
		return nil
	}

	// Core layer already handled the limiting

	// Convert to C string array
	// Format: path|name|unix_timestamp
	cArray := C.malloc(C.size_t((len(files) + 1)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cStrings := (*[1<<30 - 1]*C.char)(cArray)

	for i, file := range files {
		str := fmt.Sprintf("%s|%s|%d", file.Path, file.Name, file.Modified.Unix())
		cStrings[i] = C.CString(str)
	}
	cStrings[len(files)] = nil // Null-terminate the array

	return (**C.char)(cArray)
}

// ClippyGetRecentDownloadsWithFolders finds recent files from specific folders
//
//export ClippyGetRecentDownloadsWithFolders
func ClippyGetRecentDownloadsWithFolders(maxCount C.int, durationSecs C.int, folders *C.char, outError **C.char) **C.char {
	// Parse folder preferences
	folderStr := C.GoString(folders)
	var customDirs []string

	if folderStr != "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			*outError = C.CString(fmt.Sprintf("Error getting home directory: %v", err))
			return nil
		}

		folderNames := strings.Split(folderStr, ",")
		for _, folder := range folderNames {
			folder = strings.ToLower(strings.TrimSpace(folder))
			switch folder {
			case "downloads", "download":
				customDirs = append(customDirs, filepath.Join(homeDir, "Downloads"))
			case "desktop":
				customDirs = append(customDirs, filepath.Join(homeDir, "Desktop"))
			case "documents", "docs":
				customDirs = append(customDirs, filepath.Join(homeDir, "Documents"))
			}
		}
	}

	// Set up options
	opts := recent.DefaultFindOptions()
	if durationSecs > 0 {
		opts.MaxAge = time.Duration(durationSecs) * time.Second
	}
	if int(maxCount) > 0 {
		opts.MaxCount = int(maxCount)
	}

	// Override directories if custom ones provided
	if len(customDirs) > 0 {
		opts.Directories = customDirs
	}

	// Find files
	files, err := recent.FindRecentFiles(opts)
	if err != nil {
		*outError = C.CString(fmt.Sprintf("Error finding recent files: %v", err))
		return nil
	}

	// Convert to C string array
	cArray := C.malloc(C.size_t((len(files) + 1)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cStrings := (*[1<<30 - 1]*C.char)(cArray)

	for i, file := range files {
		str := fmt.Sprintf("%s|%s|%d", file.Path, file.Name, file.Modified.Unix())
		cStrings[i] = C.CString(str)
	}
	cStrings[len(files)] = nil

	return (**C.char)(cArray)
}

// ClippyFreeStringArray frees the memory for the array and all the strings inside it.
// This MUST be called from Swift to prevent memory leaks.
//
//export ClippyFreeStringArray
func ClippyFreeStringArray(arr **C.char) {
	if arr == nil {
		return
	}

	// Convert to Go slice for easier handling
	cStrings := (*[1<<30 - 1]*C.char)(unsafe.Pointer(arr))

	// Free each string
	i := 0
	for cStrings[i] != nil {
		C.free(unsafe.Pointer(cStrings[i]))
		i++
	}

	// Free the array itself
	C.free(unsafe.Pointer(arr))
}

// ClippyGetClipboardFiles gets file paths from the clipboard
//
//export ClippyGetClipboardFiles
func ClippyGetClipboardFiles(outError **C.char) **C.char {
	// Get clipboard files
	paths := clippy.GetFiles()
	if len(paths) == 0 {
		// No error, just no files
		return nil
	}

	// Convert to C string array
	cArray := C.malloc(C.size_t((len(paths) + 1)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cStrings := (*[1<<30 - 1]*C.char)(cArray)

	for i, path := range paths {
		cStrings[i] = C.CString(path)
	}
	cStrings[len(paths)] = nil // Null-terminate

	return (**C.char)(cArray)
}

// ClippyCopyFile copies a file to clipboard
//
//export ClippyCopyFile
func ClippyCopyFile(path *C.char, outError **C.char) C.int {
	goPath := C.GoString(path)

	err := clippy.Copy(goPath)
	if err != nil {
		*outError = C.CString(fmt.Sprintf("Error copying file: %v", err))
		return 0
	}

	return 1
}

// ClippyCopyText copies text to clipboard
//
//export ClippyCopyText
func ClippyCopyText(text *C.char, outError **C.char) C.int {
	goText := C.GoString(text)

	err := clippy.CopyText(goText)
	if err != nil {
		*outError = C.CString(fmt.Sprintf("Error copying text: %v", err))
		return 0
	}

	return 1
}

func main() {
	// This is needed for cgo to generate the C library
	// The main function is not used when building as a library
}
