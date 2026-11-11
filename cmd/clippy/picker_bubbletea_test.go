package main

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
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

// fakeWatcher returns a watcher whose channels the test feeds directly.
// waitForFSEvent only reads Events and Errors.
func fakeWatcher() *fsnotify.Watcher {
	return &fsnotify.Watcher{
		Events: make(chan fsnotify.Event, 8),
		Errors: make(chan error, 8),
	}
}

func waitForFSEventWithin(t *testing.T, m pickerModel) tea.Msg {
	t.Helper()
	done := make(chan tea.Msg, 1)
	go func() { done <- m.waitForFSEvent() }()
	select {
	case msg := <-done:
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("waitForFSEvent did not return")
		return nil
	}
}

func TestWaitForFSEventKeepsWatchingPastOtherEvents(t *testing.T) {
	watcher := fakeWatcher()
	refreshed := []recent.FileInfo{{Name: "new.zip", Path: "/tmp/new.zip"}}
	calls := 0
	m := pickerModel{
		watcher: watcher,
		refreshFunc: func() ([]recent.FileInfo, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("directory busy")
			}
			return refreshed, nil
		},
	}

	// A download writes and renames before the file the picker refreshes
	// for appears; a failed refresh and a watch error must not end the watch.
	watcher.Errors <- fsnotify.ErrEventOverflow
	watcher.Events <- fsnotify.Event{Name: "/tmp/new.zip.crdownload", Op: fsnotify.Write}
	watcher.Events <- fsnotify.Event{Name: "/tmp/new.zip.crdownload", Op: fsnotify.Rename}
	watcher.Events <- fsnotify.Event{Name: "/tmp/other", Op: fsnotify.Create}
	watcher.Events <- fsnotify.Event{Name: "/tmp/new.zip", Op: fsnotify.Create}

	msg := waitForFSEventWithin(t, m)
	refresh, ok := msg.(refreshMsg)
	if !ok {
		t.Fatalf("waitForFSEvent() = %#v; want refreshMsg", msg)
	}
	if len(refresh.files) != 1 || refresh.files[0].Name != "new.zip" {
		t.Errorf("refreshMsg.files = %+v; want new.zip", refresh.files)
	}
	if calls != 2 {
		t.Errorf("refreshFunc called %d times; want 2", calls)
	}
}

func TestWaitForFSEventStopsWhenWatcherCloses(t *testing.T) {
	for _, closeChannel := range []string{"events", "errors"} {
		t.Run(closeChannel, func(t *testing.T) {
			watcher := fakeWatcher()
			if closeChannel == "events" {
				close(watcher.Events)
			} else {
				close(watcher.Errors)
			}
			m := pickerModel{
				watcher:     watcher,
				refreshFunc: func() ([]recent.FileInfo, error) { return nil, nil },
			}
			if msg := waitForFSEventWithin(t, m); msg != nil {
				t.Errorf("waitForFSEvent() = %#v; want nil after close", msg)
			}
		})
	}
}
