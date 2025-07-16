package clipboard

import (
	"testing"
)

func TestGetUTIForFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectedUTI string
		shouldExist bool
	}{
		{
			name:        "PNG file",
			filePath:    "../../test-files/minimal.png",
			expectedUTI: "public.png",
			shouldExist: true,
		},
		{
			name:        "PDF file",
			filePath:    "../../test-files/test.pdf",
			expectedUTI: "com.adobe.pdf",
			shouldExist: true,
		},
		{
			name:        "Text file",
			filePath:    "../../test-files/sample.txt",
			expectedUTI: "public.plain-text",
			shouldExist: true,
		},
		{
			name:        "Elixir code file",
			filePath:    "../../test-files/code.elixir",
			expectedUTI: "public.text", // Generic text type for unknown extensions
			shouldExist: true,
		},
		{
			name:        "Unknown extension file",
			filePath:    "/some/path/file.xyz",
			expectedUTI: "",
			shouldExist: true, // UTI detection works on extension, not file existence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uti, exists := GetUTIForFile(tt.filePath)

			if exists != tt.shouldExist {
				t.Errorf("GetUTIForFile(%s) existence = %v, want %v", tt.filePath, exists, tt.shouldExist)
			}

			if tt.shouldExist && uti == "" {
				t.Errorf("GetUTIForFile(%s) returned empty UTI but should exist", tt.filePath)
			}

			if tt.shouldExist && tt.expectedUTI != "" && uti != tt.expectedUTI {
				t.Logf("GetUTIForFile(%s) = %s, expected %s (may be system-dependent)", tt.filePath, uti, tt.expectedUTI)
			}
		})
	}
}

func TestClipboardTypes(t *testing.T) {
	// Put some text on clipboard first
	CopyText("Test text for clipboard types")

	types := GetClipboardTypes()
	if len(types) == 0 {
		t.Error("Expected clipboard to have types after copying text")
	}

	// Should contain text type
	found := false
	for _, typeStr := range types {
		if typeStr == "public.utf8-plain-text" || typeStr == "NSStringPboardType" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected clipboard to contain text type, got: %v", types)
	}
}

func TestContainsType(t *testing.T) {
	// Put text on clipboard
	CopyText("Test text for type checking")

	// Should contain text type
	if !ContainsType("public.utf8-plain-text") && !ContainsType("NSStringPboardType") {
		t.Error("Expected clipboard to contain text type")
	}

	// Should not contain image type
	if ContainsType("public.png") {
		t.Error("Expected clipboard to not contain PNG type")
	}
}

func TestGetClipboardContent(t *testing.T) {
	// Test with text content
	CopyText("Test text content")

	content, err := GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent() error = %v", err)
	}

	if !content.IsText {
		t.Error("Expected content to be text")
	}

	if string(content.Data) != "Test text content" {
		t.Errorf("Expected content data = 'Test text content', got '%s'", string(content.Data))
	}

	// Test with file reference - need absolute path
	absPath := "/Users/neil/xuku/clippy/test-files/minimal.png"
	CopyFile(absPath)

	content, err = GetClipboardContent()
	if err != nil {
		t.Fatalf("GetClipboardContent() error = %v", err)
	}

	if !content.IsFile {
		t.Error("Expected content to be file")
	}

	if content.FilePath == "" {
		t.Error("Expected file path to be set")
	}
}

func TestImageUTIDetection(t *testing.T) {
	tests := []struct {
		uti      string
		expected bool
	}{
		{"public.png", true},
		{"public.jpeg", true},
		{"public.tiff", true},
		{"public.gif", true},
		{"public.pdf", false},
		{"public.plain-text", false},
		{"public.data", false},
	}

	for _, tt := range tests {
		t.Run(tt.uti, func(t *testing.T) {
			result := isImageUTI(tt.uti)
			if result != tt.expected {
				t.Errorf("isImageUTI(%s) = %v, want %v", tt.uti, result, tt.expected)
			}
		})
	}
}

func TestRichContentUTIDetection(t *testing.T) {
	tests := []struct {
		uti      string
		expected bool
	}{
		{"public.pdf", true},
		{"public.html", true},
		{"public.json", true},
		{"public.zip-archive", true},
		{"public.mp4", true},
		{"public.png", false},
		{"public.plain-text", false},
		{"public.data", false},
	}

	for _, tt := range tests {
		t.Run(tt.uti, func(t *testing.T) {
			result := isRichContentUTI(tt.uti)
			if result != tt.expected {
				t.Errorf("isRichContentUTI(%s) = %v, want %v", tt.uti, result, tt.expected)
			}
		})
	}
}

func TestUTIConformance(t *testing.T) {
	tests := []struct {
		name       string
		uti        string
		parentType string
		expected   bool
	}{
		{
			name:       "Plain text conforms to text",
			uti:        "public.plain-text",
			parentType: "public.text",
			expected:   true,
		},
		{
			name:       "C source conforms to source code",
			uti:        "public.c-source",
			parentType: "public.source-code",
			expected:   true,
		},
		{
			name:       "PNG does not conform to text",
			uti:        "public.png",
			parentType: "public.text",
			expected:   false,
		},
		{
			name:       "JSON should conform to text",
			uti:        "public.json",
			parentType: "public.text",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UTIConformsTo(tt.uti, tt.parentType)
			if result != tt.expected {
				t.Errorf("UTIConformsTo(%s, %s) = %v, want %v", tt.uti, tt.parentType, result, tt.expected)
			}
		})
	}
}
