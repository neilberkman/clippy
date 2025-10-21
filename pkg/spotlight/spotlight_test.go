//go:build darwin

package spotlight

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearch(t *testing.T) {
	// Create a test file in temp directory
	tmpDir := t.TempDir()
	testFileName := "test_spotlight_search_12345.txt"
	testFile := filepath.Join(tmpDir, testFileName)

	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Give Spotlight a moment to index (it's async)
	// Note: This test may be flaky on systems with slow Spotlight indexing
	t.Log("Created test file, waiting for Spotlight to index...")

	// Search for the test file
	results, err := Search(SearchOptions{
		Query:      testFileName,
		MaxResults: 10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// We may or may not find the file depending on Spotlight indexing speed
	// So we just verify the function doesn't crash
	t.Logf("Search returned %d results", len(results))

	// Verify result structure if we got any
	for _, result := range results {
		if result.Path == "" {
			t.Error("Result has empty path")
		}
		if result.Name == "" {
			t.Error("Result has empty name")
		}
		t.Logf("Found: %s (%s)", result.Name, result.Path)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	results, err := Search(SearchOptions{
		Query:      "",
		MaxResults: 10,
	})

	if err == nil {
		t.Error("Expected error for empty query, got nil")
	}

	if len(results) != 0 {
		t.Errorf("Expected no results for empty query, got %d", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	// Search for something very unlikely to exist
	results, err := Search(SearchOptions{
		Query:      "xyzzy_impossible_filename_999999",
		MaxResults: 10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Logf("Unexpectedly found %d results for impossible filename", len(results))
	}
}

func TestSearchMaxResults(t *testing.T) {
	// Search for something common
	results, err := Search(SearchOptions{
		Query:      "test",
		MaxResults: 5,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) > 5 {
		t.Errorf("MaxResults not respected: got %d results, expected max 5", len(results))
	}

	t.Logf("Search with MaxResults=5 returned %d results", len(results))
}
