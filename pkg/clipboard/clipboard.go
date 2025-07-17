package clipboard

// ClipboardManager defines the interface for platform-specific clipboard operations
type ClipboardManager interface {
	// CopyFile copies a single file reference to clipboard
	CopyFile(path string) error
	
	// CopyFiles copies multiple file references to clipboard
	CopyFiles(paths []string) error
	
	// CopyText copies text content to clipboard
	CopyText(text string) error
	
	// GetFiles returns file paths currently on clipboard
	GetFiles() []string
	
	// GetText returns text content from clipboard
	GetText() (string, bool)
	
	// GetUTIForFile returns the UTI (Uniform Type Identifier) for a file path
	GetUTIForFile(path string) (string, bool)
	
	// GetClipboardTypes returns all available types on clipboard
	GetClipboardTypes() []string
	
	// GetClipboardDataForType returns data for a specific type from clipboard
	GetClipboardDataForType(typeStr string) ([]byte, bool)
	
	// ContainsType checks if clipboard contains a specific type
	ContainsType(typeStr string) bool
	
	// UTIConformsTo checks if a UTI conforms to a parent type
	UTIConformsTo(uti, parentType string) bool
	
	// GetClipboardContent returns clipboard content with smart type detection
	GetClipboardContent() (*ClipboardContent, error)
}

// ClipboardContent represents the content and type information from clipboard
type ClipboardContent struct {
	Type     string // UTI or MIME type
	Data     []byte // Raw data
	IsText   bool   // Whether this is text content
	IsFile   bool   // Whether this is file reference
	FilePath string // File path if IsFile is true
}

// manager holds the platform-specific implementation
var manager ClipboardManager

// init initializes the platform-specific clipboard manager
func init() {
	manager = newClipboardManager()
}

// CopyFile copies a single file reference to clipboard
func CopyFile(path string) {
	manager.CopyFile(path)
}

// CopyFiles copies multiple file references to clipboard
func CopyFiles(paths []string) {
	manager.CopyFiles(paths)
}

// CopyText copies text content to clipboard
func CopyText(text string) {
	manager.CopyText(text)
}

// GetFiles returns file paths currently on clipboard
func GetFiles() []string {
	return manager.GetFiles()
}

// GetText returns text content from clipboard
func GetText() (string, bool) {
	return manager.GetText()
}

// GetUTIForFile returns the UTI (Uniform Type Identifier) for a file path
func GetUTIForFile(path string) (string, bool) {
	return manager.GetUTIForFile(path)
}

// GetClipboardTypes returns all available types on clipboard
func GetClipboardTypes() []string {
	return manager.GetClipboardTypes()
}

// GetClipboardDataForType returns data for a specific type from clipboard
func GetClipboardDataForType(typeStr string) ([]byte, bool) {
	return manager.GetClipboardDataForType(typeStr)
}

// ContainsType checks if clipboard contains a specific type
func ContainsType(typeStr string) bool {
	return manager.ContainsType(typeStr)
}

// UTIConformsTo checks if a UTI conforms to a parent type using platform UTI system
func UTIConformsTo(uti, parentType string) bool {
	return manager.UTIConformsTo(uti, parentType)
}

// GetClipboardContent returns clipboard content with smart type detection
func GetClipboardContent() (*ClipboardContent, error) {
	return manager.GetClipboardContent()
}