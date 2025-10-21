package clippy

import (
	"os"
	"strings"
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

func TestIsTextualMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		// Text types
		{"plain text", "text/plain", true},
		{"HTML", "text/html", true},
		{"CSS", "text/css", true},
		{"JavaScript text", "text/javascript", true},
		{"XML text", "text/xml", true},
		{"CSV text", "text/csv", true},

		// Application types that are actually text
		{"JSON", "application/json", true},
		{"XML app", "application/xml", true},
		{"JavaScript app", "application/javascript", true},
		{"YAML", "application/x-yaml", true},
		{"SQL", "application/sql", true},
		{"GraphQL", "application/graphql", true},
		{"XHTML", "application/xhtml+xml", true},
		{"JSON-LD", "application/ld+json", true},
		{"Atom feed", "application/atom+xml", true},
		{"RSS", "application/rss+xml", true},

		// Binary types (should return false)
		{"PDF", "application/pdf", false},
		{"ZIP", "application/zip", false},
		{"JPEG", "image/jpeg", false},
		{"PNG", "image/png", false},
		{"MP3", "audio/mpeg", false},
		{"MP4", "video/mp4", false},
		{"Binary", "application/octet-stream", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTextualMimeType(tt.mimeType); got != tt.want {
				t.Errorf("isTextualMimeType(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestMimeToUTI(t *testing.T) {
	tests := []struct {
		name string
		mime string
		want string
	}{
		{"HTML", "text/html", "public.html"},
		{"JSON", "application/json", "public.json"},
		{"XML text", "text/xml", "public.xml"},
		{"XML app", "application/xml", "public.xml"},
		{"Plain text", "text/plain", "public.plain-text"},
		{"RTF text", "text/rtf", "public.rtf"},
		{"RTF app", "application/rtf", "public.rtf"},
		{"Markdown", "text/markdown", "net.daringfireball.markdown"},
		{"Unknown type", "application/unknown", "application/unknown"}, // Returns as-is
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mimeToUTI(tt.mime); got != tt.want {
				t.Errorf("mimeToUTI(%q) = %v, want %v", tt.mime, got, tt.want)
			}
		})
	}
}

func TestCopyTextWithAutoDetection(t *testing.T) {
	// Note: These tests check the detection logic but can't test actual clipboard operations
	// without mocking the clipboard package

	tests := []struct {
		name        string
		content     string
		wantUTI     string // Expected UTI that would be set
		description string
	}{
		{
			name:        "JSON content",
			content:     `{"key": "value", "number": 123}`,
			wantUTI:     "public.json",
			description: "Should detect JSON and set public.json",
		},
		{
			name:        "HTML content",
			content:     `<!DOCTYPE html><html><body><h1>Test</h1></body></html>`,
			wantUTI:     "public.html",
			description: "Should detect HTML and set public.html",
		},
		{
			name:        "XML content",
			content:     `<?xml version="1.0"?><root><item>test</item></root>`,
			wantUTI:     "public.xml",
			description: "Should detect XML and set public.xml",
		},
		{
			name:        "Plain text",
			content:     "This is just plain text without any special format.",
			wantUTI:     "", // Would use default CopyText
			description: "Should fall back to plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Detect MIME type as the function would
			mtype := mimetype.Detect([]byte(tt.content))
			mimeStr := mtype.String()

			// Check what UTI would be selected
			var expectedUTI string
			switch {
			case strings.HasPrefix(mimeStr, "text/html"):
				expectedUTI = "public.html"
			case mimeStr == "application/json":
				expectedUTI = "public.json"
			case strings.HasPrefix(mimeStr, "text/xml") || mimeStr == "application/xml":
				expectedUTI = "public.xml"
			default:
				expectedUTI = ""
			}

			if expectedUTI != tt.wantUTI {
				t.Errorf("Content detection failed for %s\nDetected MIME: %s\nExpected UTI: %s\nGot UTI: %s",
					tt.name, mimeStr, tt.wantUTI, expectedUTI)
			}
		})
	}
}

func TestCopyTextWithType(t *testing.T) {
	// Test MIME to UTI conversion in CopyTextWithType
	tests := []struct {
		name           string
		typeIdentifier string
		wantUTI        string
	}{
		{"MIME type HTML", "text/html", "public.html"},
		{"MIME type JSON", "application/json", "public.json"},
		{"Direct UTI", "public.xml", "public.xml"},
		{"Unknown MIME", "application/custom", "application/custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check conversion logic
			result := tt.typeIdentifier
			if strings.Contains(result, "/") {
				result = mimeToUTI(result)
			}

			if result != tt.wantUTI {
				t.Errorf("Type conversion failed\nInput: %s\nExpected: %s\nGot: %s",
					tt.typeIdentifier, tt.wantUTI, result)
			}
		})
	}
}

func TestConvertImageFormat(t *testing.T) {
	// Verify the function handles errors gracefully
	_, err := convertImageFormat([]byte("not an image"), ".png")
	if err == nil {
		t.Error("Expected error for invalid image data")
	}

	// Test unsupported format
	_, err = convertImageFormat([]byte("fake image"), ".bmp")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestFindAvailableFilename(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		existingFiles []string
		inputPath     string
		want          string
	}{
		{
			name:          "no conflict",
			existingFiles: []string{},
			inputPath:     tmpDir + "/photo.png",
			want:          tmpDir + "/photo.png",
		},
		{
			name:          "one conflict with extension",
			existingFiles: []string{"photo.png"},
			inputPath:     tmpDir + "/photo.png",
			want:          tmpDir + "/photo 2.png",
		},
		{
			name:          "two conflicts with extension",
			existingFiles: []string{"photo.png", "photo 2.png"},
			inputPath:     tmpDir + "/photo.png",
			want:          tmpDir + "/photo 3.png",
		},
		{
			name:          "no conflict without extension",
			existingFiles: []string{},
			inputPath:     tmpDir + "/README",
			want:          tmpDir + "/README",
		},
		{
			name:          "one conflict without extension",
			existingFiles: []string{"README"},
			inputPath:     tmpDir + "/README",
			want:          tmpDir + "/README 2",
		},
		{
			name:          "multiple conflicts without extension",
			existingFiles: []string{"README", "README 2", "README 3"},
			inputPath:     tmpDir + "/README",
			want:          tmpDir + "/README 4",
		},
		{
			name:          "multi-part extension",
			existingFiles: []string{"archive.tar.gz"},
			inputPath:     tmpDir + "/archive.tar.gz",
			want:          tmpDir + "/archive.tar 2.gz",
		},
		{
			name:          "gaps in numbering",
			existingFiles: []string{"file.txt", "file 3.txt"},
			inputPath:     tmpDir + "/file.txt",
			want:          tmpDir + "/file 2.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.existingFiles {
				if err := os.WriteFile(tmpDir+"/"+f, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			got := findAvailableFilename(tt.inputPath, false)
			if got != tt.want {
				t.Errorf("findAvailableFilename(%q, false)\n  got:  %q\n  want: %q", tt.inputPath, got, tt.want)
			}

			for _, f := range tt.existingFiles {
				_ = os.Remove(tmpDir + "/" + f)
			}
		})
	}
}

func TestFindAvailableFilenameWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(tmpDir+"/existing.txt", []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	path := tmpDir + "/existing.txt"
	got := findAvailableFilename(path, true)
	want := path

	if got != want {
		t.Errorf("findAvailableFilename(%q, true) should return original path when force=true\n  got:  %q\n  want: %q", path, got, want)
	}
}