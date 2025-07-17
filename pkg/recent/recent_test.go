package recent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultFindOptions(t *testing.T) {
	opts := DefaultFindOptions()
	
	if opts.MaxAge != 5*time.Minute {
		t.Errorf("Expected MaxAge to be 5 minutes, got %v", opts.MaxAge)
	}
	
	if opts.MaxCount != 10 {
		t.Errorf("Expected MaxCount to be 10, got %v", opts.MaxCount)
	}
	
	if !opts.ExcludeTemp {
		t.Error("Expected ExcludeTemp to be true")
	}
	
	if !opts.SmartUnarchive {
		t.Error("Expected SmartUnarchive to be true")
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"", 5 * time.Minute},
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
		{"30s", 30 * time.Second},
		{"10", 10 * time.Minute}, // Just numbers assume minutes
	}
	
	for _, test := range tests {
		result, err := ParseDuration(test.input)
		if err != nil {
			t.Errorf("ParseDuration(%q) returned error: %v", test.input, err)
			continue
		}
		
		if result != test.expected {
			t.Errorf("ParseDuration(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGetDefaultDownloadDirs(t *testing.T) {
	dirs := GetDefaultDownloadDirs()
	
	if len(dirs) == 0 {
		t.Error("Expected at least one default download directory")
	}
	
	// Should include Downloads directory
	homeDir, _ := os.UserHomeDir()
	expectedDownloads := filepath.Join(homeDir, "Downloads")
	
	found := false
	for _, dir := range dirs {
		if dir == expectedDownloads {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected Downloads directory %q to be in default directories", expectedDownloads)
	}
}

func TestIsArchive(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.zip", true},
		{"test.tar.gz", true},
		{"test.pdf", false},
		{"test.txt", false},
		{"test.dmg", true},
		{"test.pkg", true},
		{"test.tar", true},
		{"test.7z", true},
	}
	
	for _, test := range tests {
		result := IsArchive(test.filename)
		if result != test.expected {
			t.Errorf("IsArchive(%q) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestIsTemporaryFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.txt", false},
		{"test.tmp", true},
		{"test.download", true},
		{"test.crdownload", true},
		{"test.part", true},
		{"normal-file.pdf", false},
	}
	
	for _, test := range tests {
		result := isTemporaryFile(test.filename)
		if result != test.expected {
			t.Errorf("isTemporaryFile(%q) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}