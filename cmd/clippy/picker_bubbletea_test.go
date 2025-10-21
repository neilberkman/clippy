package main

import (
	"testing"
	"time"

	"github.com/neilberkman/clippy/pkg/recent"
)

func TestPickerModel(t *testing.T) {
	// Create test files
	files := []recent.FileInfo{
		{
			Name:     "test1.txt",
			Path:     "/tmp/test1.txt",
			Size:     1024,
			Modified: time.Now(),
		},
		{
			Name:     "test2.png",
			Path:     "/tmp/test2.png",
			Size:     2048,
			Modified: time.Now().Add(-5 * time.Minute),
		},
		{
			Name:     "test3.pdf",
			Path:     "/tmp/test3.pdf",
			Size:     3072,
			Modified: time.Now().Add(-10 * time.Minute),
		},
	}

	// Create picker model
	m := pickerModel{
		files:        files,
		cursor:       0,
		selected:     make(map[int]bool),
		absoluteTime: false,
		refreshFunc:  nil, // No refresh in tests
	}

	// Test initial state
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Test view rendering
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Test item rendering
	item := pickerItem{
		file:     files[0],
		index:    0,
		selected: false,
		focused:  true,
	}
	rendered := m.renderItem(item)
	if rendered == "" {
		t.Error("Expected non-empty rendered item")
	}

	// Test truncation
	long := "This is a very long string that should be truncated"
	truncated := truncateString(long, 10)
	if len(truncated) != 10 {
		t.Errorf("Expected truncated string length 10, got %d", len(truncated))
	}
}
