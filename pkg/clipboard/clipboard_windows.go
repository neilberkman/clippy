//go:build windows

package clipboard

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows clipboard format constants
const (
	CF_TEXT         = 1
	CF_UNICODETEXT  = 13
	CF_HDROP        = 15
	DROPFILES_SIZE  = 20
	GHND           = 0x0042
)

// DROPFILES structure for file clipboard operations
type DROPFILES struct {
	pFiles uintptr // Offset to file list
	pt     struct {
		x, y int32
	}
	fNC      uint32 // TRUE if file names are without drive letters
	fWide    uint32 // TRUE if file names are Unicode
}

// WindowsClipboardManager implements ClipboardManager for Windows using Win32 APIs
type WindowsClipboardManager struct {
	user32   *syscall.DLL
	kernel32 *syscall.DLL
}

// newClipboardManager creates a new clipboard manager for Windows
func newClipboardManager() ClipboardManager {
	user32, err := syscall.LoadDLL("user32.dll")
	if err != nil {
		panic(fmt.Sprintf("failed to load user32.dll: %v", err))
	}
	
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		panic(fmt.Sprintf("failed to load kernel32.dll: %v", err))
	}

	return &WindowsClipboardManager{
		user32:   user32,
		kernel32: kernel32,
	}
}

// CopyFile copies a single file reference to clipboard
func (m *WindowsClipboardManager) CopyFile(path string) error {
	return m.CopyFiles([]string{path})
}

// CopyFiles copies multiple file references to clipboard using CF_HDROP format
func (m *WindowsClipboardManager) CopyFiles(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no files provided")
	}

	// Convert paths to absolute paths
	absPaths := make([]string, len(paths))
	for i, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
		absPaths[i] = absPath
	}

	// Get clipboard procedures
	openClipboard, err := m.user32.FindProc("OpenClipboard")
	if err != nil {
		return fmt.Errorf("failed to find OpenClipboard: %w", err)
	}
	
	emptyClipboard, err := m.user32.FindProc("EmptyClipboard")
	if err != nil {
		return fmt.Errorf("failed to find EmptyClipboard: %w", err)
	}
	
	setClipboardData, err := m.user32.FindProc("SetClipboardData")
	if err != nil {
		return fmt.Errorf("failed to find SetClipboardData: %w", err)
	}
	
	closeClipboard, err := m.user32.FindProc("CloseClipboard")
	if err != nil {
		return fmt.Errorf("failed to find CloseClipboard: %w", err)
	}

	globalAlloc, err := m.kernel32.FindProc("GlobalAlloc")
	if err != nil {
		return fmt.Errorf("failed to find GlobalAlloc: %w", err)
	}
	
	globalLock, err := m.kernel32.FindProc("GlobalLock")
	if err != nil {
		return fmt.Errorf("failed to find GlobalLock: %w", err)
	}
	
	globalUnlock, err := m.kernel32.FindProc("GlobalUnlock")
	if err != nil {
		return fmt.Errorf("failed to find GlobalUnlock: %w", err)
	}

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard")
	}
	defer closeClipboard.Call()

	// Empty clipboard
	ret, _, _ = emptyClipboard.Call()
	if ret == 0 {
		return fmt.Errorf("failed to empty clipboard")
	}

	// Create DROPFILES structure
	data := m.createDropFilesData(absPaths)

	// Allocate global memory
	hMem, _, _ := globalAlloc.Call(GHND, uintptr(len(data)))
	if hMem == 0 {
		return fmt.Errorf("failed to allocate global memory")
	}

	// Lock memory and copy data
	pMem, _, _ := globalLock.Call(hMem)
	if pMem == 0 {
		return fmt.Errorf("failed to lock global memory")
	}

	// Copy data to global memory
	dest := (*[1 << 30]byte)(unsafe.Pointer(pMem))
	copy(dest[:len(data)], data)

	// Unlock memory
	globalUnlock.Call(hMem)

	// Set clipboard data
	ret, _, _ = setClipboardData.Call(CF_HDROP, hMem)
	if ret == 0 {
		return fmt.Errorf("failed to set clipboard data")
	}

	return nil
}

// CopyText copies text content to clipboard
func (m *WindowsClipboardManager) CopyText(text string) error {
	// Get clipboard procedures
	openClipboard, err := m.user32.FindProc("OpenClipboard")
	if err != nil {
		return fmt.Errorf("failed to find OpenClipboard: %w", err)
	}
	
	emptyClipboard, err := m.user32.FindProc("EmptyClipboard")
	if err != nil {
		return fmt.Errorf("failed to find EmptyClipboard: %w", err)
	}
	
	setClipboardData, err := m.user32.FindProc("SetClipboardData")
	if err != nil {
		return fmt.Errorf("failed to find SetClipboardData: %w", err)
	}
	
	closeClipboard, err := m.user32.FindProc("CloseClipboard")
	if err != nil {
		return fmt.Errorf("failed to find CloseClipboard: %w", err)
	}

	globalAlloc, err := m.kernel32.FindProc("GlobalAlloc")
	if err != nil {
		return fmt.Errorf("failed to find GlobalAlloc: %w", err)
	}
	
	globalLock, err := m.kernel32.FindProc("GlobalLock")
	if err != nil {
		return fmt.Errorf("failed to find GlobalLock: %w", err)
	}
	
	globalUnlock, err := m.kernel32.FindProc("GlobalUnlock")
	if err != nil {
		return fmt.Errorf("failed to find GlobalUnlock: %w", err)
	}

	// Convert text to UTF-16
	utf16Text := windows.StringToUTF16(text)
	dataSize := len(utf16Text) * 2 // 2 bytes per UTF-16 character

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard")
	}
	defer closeClipboard.Call()

	// Empty clipboard
	ret, _, _ = emptyClipboard.Call()
	if ret == 0 {
		return fmt.Errorf("failed to empty clipboard")
	}

	// Allocate global memory
	hMem, _, _ := globalAlloc.Call(GHND, uintptr(dataSize))
	if hMem == 0 {
		return fmt.Errorf("failed to allocate global memory")
	}

	// Lock memory and copy data
	pMem, _, _ := globalLock.Call(hMem)
	if pMem == 0 {
		return fmt.Errorf("failed to lock global memory")
	}

	// Copy UTF-16 text to global memory
	dest := (*[1 << 30]uint16)(unsafe.Pointer(pMem))
	copy(dest[:len(utf16Text)], utf16Text)

	// Unlock memory
	globalUnlock.Call(hMem)

	// Set clipboard data
	ret, _, _ = setClipboardData.Call(CF_UNICODETEXT, hMem)
	if ret == 0 {
		return fmt.Errorf("failed to set clipboard data")
	}

	return nil
}

// GetFiles returns file paths currently on clipboard
func (m *WindowsClipboardManager) GetFiles() []string {
	// Get clipboard procedures
	openClipboard, err := m.user32.FindProc("OpenClipboard")
	if err != nil {
		return nil
	}
	
	getClipboardData, err := m.user32.FindProc("GetClipboardData")
	if err != nil {
		return nil
	}
	
	closeClipboard, err := m.user32.FindProc("CloseClipboard")
	if err != nil {
		return nil
	}
	
	globalLock, err := m.kernel32.FindProc("GlobalLock")
	if err != nil {
		return nil
	}
	
	globalUnlock, err := m.kernel32.FindProc("GlobalUnlock")
	if err != nil {
		return nil
	}

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return nil
	}
	defer closeClipboard.Call()

	// Get clipboard data
	hData, _, _ := getClipboardData.Call(CF_HDROP)
	if hData == 0 {
		return nil
	}

	// Lock the data
	pData, _, _ := globalLock.Call(hData)
	if pData == 0 {
		return nil
	}
	defer globalUnlock.Call(hData)

	// Parse DROPFILES structure
	return m.parseDropFiles(pData)
}

// GetText returns text content from clipboard
func (m *WindowsClipboardManager) GetText() (string, bool) {
	// Get clipboard procedures
	openClipboard, err := m.user32.FindProc("OpenClipboard")
	if err != nil {
		return "", false
	}
	
	getClipboardData, err := m.user32.FindProc("GetClipboardData")
	if err != nil {
		return "", false
	}
	
	closeClipboard, err := m.user32.FindProc("CloseClipboard")
	if err != nil {
		return "", false
	}
	
	globalLock, err := m.kernel32.FindProc("GlobalLock")
	if err != nil {
		return "", false
	}
	
	globalUnlock, err := m.kernel32.FindProc("GlobalUnlock")
	if err != nil {
		return "", false
	}

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return "", false
	}
	defer closeClipboard.Call()

	// Try Unicode text first
	hData, _, _ := getClipboardData.Call(CF_UNICODETEXT)
	if hData != 0 {
		pData, _, _ := globalLock.Call(hData)
		if pData != 0 {
			defer globalUnlock.Call(hData)
			text := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(pData)))
			return text, true
		}
	}

	// Fallback to ANSI text
	hData, _, _ = getClipboardData.Call(CF_TEXT)
	if hData != 0 {
		pData, _, _ := globalLock.Call(hData)
		if pData != 0 {
			defer globalUnlock.Call(hData)
			text := (*[1 << 20]byte)(unsafe.Pointer(pData))
			// Find null terminator
			for i := 0; i < len(text); i++ {
				if text[i] == 0 {
					return string(text[:i]), true
				}
			}
		}
	}

	return "", false
}

// GetUTIForFile returns a Windows equivalent to UTI (MIME type based on extension)
func (m *WindowsClipboardManager) GetUTIForFile(path string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	
	// Map common extensions to MIME types (Windows equivalent to UTI)
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}
	
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType, true
	}
	
	return "", false
}

// GetClipboardTypes returns all available types on clipboard (Windows formats)
func (m *WindowsClipboardManager) GetClipboardTypes() []string {
	// Windows clipboard format enumeration would require more complex Win32 API calls
	// For now, return basic types that we can check
	var types []string
	
	if text, ok := m.GetText(); ok && text != "" {
		types = append(types, "CF_UNICODETEXT")
	}
	
	if files := m.GetFiles(); len(files) > 0 {
		types = append(types, "CF_HDROP")
	}
	
	return types
}

// GetClipboardDataForType returns data for a specific type from clipboard
func (m *WindowsClipboardManager) GetClipboardDataForType(typeStr string) ([]byte, bool) {
	switch typeStr {
	case "CF_UNICODETEXT":
		if text, ok := m.GetText(); ok {
			return []byte(text), true
		}
	case "CF_HDROP":
		// Could implement binary HDROP data retrieval here if needed
		return nil, false
	}
	return nil, false
}

// ContainsType checks if clipboard contains a specific type
func (m *WindowsClipboardManager) ContainsType(typeStr string) bool {
	switch typeStr {
	case "CF_UNICODETEXT", "text/plain":
		_, ok := m.GetText()
		return ok
	case "CF_HDROP":
		files := m.GetFiles()
		return len(files) > 0
	}
	return false
}

// UTIConformsTo checks if a type conforms to a parent type (simplified for Windows)
func (m *WindowsClipboardManager) UTIConformsTo(uti, parentType string) bool {
	// Simplified type checking for Windows
	if parentType == "public.text" || parentType == "text" {
		return strings.HasPrefix(uti, "text/")
	}
	if parentType == "public.image" || parentType == "image" {
		return strings.HasPrefix(uti, "image/")
	}
	return false
}

// GetClipboardContent returns clipboard content with smart type detection
func (m *WindowsClipboardManager) GetClipboardContent() (*ClipboardContent, error) {
	// Priority 1: Check for file references
	if files := m.GetFiles(); len(files) > 0 {
		filePath := files[0]
		mimeType, _ := m.GetUTIForFile(filePath)
		return &ClipboardContent{
			Type:     mimeType,
			IsFile:   true,
			FilePath: filePath,
		}, nil
	}

	// Priority 2: Check for text content
	if text, ok := m.GetText(); ok {
		return &ClipboardContent{
			Type:   "text/plain",
			Data:   []byte(text),
			IsText: true,
		}, nil
	}

	return nil, fmt.Errorf("no supported content found on clipboard")
}

// createDropFilesData creates the binary data for CF_HDROP format
func (m *WindowsClipboardManager) createDropFilesData(paths []string) []byte {
	// Calculate total size needed
	totalSize := DROPFILES_SIZE
	for _, path := range paths {
		// Convert to UTF-16 and add null terminator
		utf16Path := windows.StringToUTF16(path)
		totalSize += len(utf16Path) * 2 // 2 bytes per character
	}
	totalSize += 2 // Final null terminator

	// Create buffer
	data := make([]byte, totalSize)
	
	// Fill DROPFILES structure
	binary.LittleEndian.PutUint32(data[0:4], DROPFILES_SIZE)   // pFiles offset
	binary.LittleEndian.PutUint32(data[4:8], 0)               // pt.x
	binary.LittleEndian.PutUint32(data[8:12], 0)              // pt.y
	binary.LittleEndian.PutUint32(data[12:16], 0)             // fNC
	binary.LittleEndian.PutUint32(data[16:20], 1)             // fWide (Unicode)

	// Add file paths
	offset := DROPFILES_SIZE
	for _, path := range paths {
		utf16Path := windows.StringToUTF16(path)
		for i, char := range utf16Path {
			binary.LittleEndian.PutUint16(data[offset+i*2:offset+i*2+2], char)
		}
		offset += len(utf16Path) * 2
	}

	// Add final null terminator
	binary.LittleEndian.PutUint16(data[offset:offset+2], 0)

	return data
}

// parseDropFiles parses the DROPFILES structure to extract file paths
func (m *WindowsClipboardManager) parseDropFiles(pData uintptr) []string {
	data := (*[1 << 20]byte)(unsafe.Pointer(pData))
	
	// Read DROPFILES header
	pFiles := binary.LittleEndian.Uint32(data[0:4])
	fWide := binary.LittleEndian.Uint32(data[16:20])
	
	var files []string
	offset := int(pFiles)
	
	if fWide != 0 {
		// Unicode strings
		for offset < len(data)-1 {
			// Find null terminator
			start := offset
			for offset < len(data)-1 && (data[offset] != 0 || data[offset+1] != 0) {
				offset += 2
			}
			
			if offset > start {
				// Convert UTF-16 to string
				utf16Data := (*[1 << 20]uint16)(unsafe.Pointer(&data[start]))
				length := (offset - start) / 2
				if length > 0 {
					utf16Slice := utf16Data[:length]
					files = append(files, windows.UTF16ToString(utf16Slice))
				}
			}
			
			offset += 2 // Skip null terminator
			
			// Check for double null terminator (end of list)
			if offset < len(data)-1 && data[offset] == 0 && data[offset+1] == 0 {
				break
			}
		}
	} else {
		// ANSI strings
		for offset < len(data) {
			start := offset
			for offset < len(data) && data[offset] != 0 {
				offset++
			}
			
			if offset > start {
				files = append(files, string(data[start:offset]))
			}
			
			offset++ // Skip null terminator
			
			// Check for double null terminator (end of list)
			if offset < len(data) && data[offset] == 0 {
				break
			}
		}
	}
	
	return files
}