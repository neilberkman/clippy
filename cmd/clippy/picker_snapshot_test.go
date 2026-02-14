package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/neilberkman/clippy/pkg/recent"
)

const (
	pickerSnapshotPath = "testdata/picker_snapshot.txt"
	beginMarker        = "===PICKER_SNAPSHOT_BEGIN==="
	endMarker          = "===PICKER_SNAPSHOT_END==="
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func TestPickerSnapshotGolden(t *testing.T) {
	snapshot := renderPickerSnapshot()

	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := os.WriteFile(pickerSnapshotPath, []byte(snapshot), 0644); err != nil {
			t.Fatalf("failed writing snapshot: %v", err)
		}
	}

	wantBytes, err := os.ReadFile(pickerSnapshotPath)
	if err != nil {
		t.Fatalf("failed reading snapshot %s: %v", pickerSnapshotPath, err)
	}

	want := strings.TrimSpace(string(wantBytes))
	got := strings.TrimSpace(snapshot)
	if got != want {
		t.Fatalf(
			"picker snapshot mismatch\nre-run with UPDATE_SNAPSHOTS=1 if change is intentional\n--- got ---\n%s\n--- want ---\n%s",
			got,
			want,
		)
	}
}

func TestPickerSnapshotPrint(t *testing.T) {
	if os.Getenv("CLIPPY_SNAPSHOT_PRINT") != "1" {
		t.Skip("snapshot print disabled")
	}

	fmt.Println(beginMarker)
	fmt.Println(renderPickerSnapshot())
	fmt.Println(endMarker)
}

func renderPickerSnapshot() string {
	baseTime := time.Date(2026, 2, 13, 9, 30, 0, 0, time.UTC)
	files := []recent.FileInfo{
		{
			Name:     "workflow-run-logs-2026-02-13.txt",
			Path:     "/Users/tester/Downloads/workflow-run-logs-2026-02-13.txt",
			Size:     1536,
			Modified: baseTime,
			MimeType: "text/plain",
		},
		{
			Name:     "incident-response-playbook-v3.pdf",
			Path:     "/Users/tester/Documents/incident-response-playbook-v3.pdf",
			Size:     987654,
			Modified: baseTime.Add(-15 * time.Minute),
			MimeType: "application/pdf",
		},
		{
			Name:     "database-backup-2026-02-13-0915.sql.gz",
			Path:     "/Users/tester/Downloads/database-backup-2026-02-13-0915.sql.gz",
			Size:     1843200,
			Modified: baseTime.Add(-45 * time.Minute),
			MimeType: "application/gzip",
		},
		{
			Name:     "screenshot-prod-error.png",
			Path:     "/Users/tester/Desktop/screenshot-prod-error.png",
			Size:     245760,
			Modified: baseTime.Add(-2 * time.Hour),
			MimeType: "image/png",
		},
	}

	model := pickerModel{
		files:          files,
		cursor:         1,
		selected:       make(map[int]bool),
		absoluteTime:   true,
		terminalWidth:  100,
		terminalHeight: 24,
		newFiles:       map[string]time.Time{files[3].Path: baseTime},
	}

	return normalizeSnapshotOutput(model.View())
}

func normalizeSnapshotOutput(view string) string {
	s := strings.ReplaceAll(view, "\r\n", "\n")
	s = ansiRegex.ReplaceAllString(s, "")

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
