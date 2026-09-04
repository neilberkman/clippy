package clippy

import (
	"errors"
	"strings"
	"testing"
)

func TestSlackCopyValidatesBeforeWriting(t *testing.T) {
	for _, tc := range []struct {
		name     string
		markdown string
		expected int
		preview  bool
		wantErr  bool
	}{
		{"preview", "```\ncode\n```", 1, true, false},
		{"table instead of fence", "| a | b |\n|---|---|\n| 1 | 2 |", 1, false, true},
		{"inline instead of block", "``\na  b\nc  d\n``", 1, false, true},
		{"literal newline escapes", "```\\ncode\\n```", 1, false, true},
		{"missing second block", "```\ncode\n```", 2, false, true},
		{"zero is a real expectation", "```\ncode\n```", 0, false, true},
		{"negative expectation", "text", -1, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, err := copySlackWithOptions(tc.markdown, SlackCopyOptions{
				Preview: tc.preview, ExpectedCodeBlocks: &tc.expected,
			}, func(_, _ string) error {
				t.Fatal("clipboard writer called during preview or failed validation")
				return nil
			})
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error = %v", err, tc.wantErr)
			}
			if report != nil && report.Copied {
				t.Fatal("reported a copy without writing")
			}
			if tc.expected >= 0 && report == nil {
				t.Fatal("missing formatting report")
			}
		})
	}
}

func TestSlackCopyReportsWriteOutcome(t *testing.T) {
	for _, writeErr := range []error{nil, errors.New("clipboard unavailable")} {
		var calls int
		expected := 1
		report, err := copySlackWithOptions("```\n  a  b\n  c  d\n```", SlackCopyOptions{ExpectedCodeBlocks: &expected}, func(delta, plainText string) error {
			calls++
			if !strings.Contains(delta, `"code-block":true`) || strings.Contains(plainText, "```") {
				t.Errorf("incorrect clipboard representations: %s / %q", delta, plainText)
			}
			return writeErr
		})
		if !errors.Is(err, writeErr) {
			t.Errorf("error = %v, want %v", err, writeErr)
		}
		if calls != 1 || report == nil {
			t.Fatalf("writes = %d, report = %+v", calls, report)
		}
		if report.Copied != (writeErr == nil) || report.Formatting.CodeLines != 2 {
			t.Errorf("incorrect report: %+v", report)
		}
	}
}
