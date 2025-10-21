//go:build darwin

package transform

import (
	"strings"
	"testing"
)

func TestMarkdownToRTF(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name:     "simple bold",
			markdown: "This is **bold** text",
			wantErr:  false,
		},
		{
			name:     "simple italic",
			markdown: "This is *italic* text",
			wantErr:  false,
		},
		{
			name:     "heading",
			markdown: "# Header\n\nSome text",
			wantErr:  false,
		},
		{
			name:     "complex markdown",
			markdown: "# Header\n\nSome **bold** and *italic* text.\n\n- Item 1\n- Item 2",
			wantErr:  false,
		},
		{
			name:     "empty string",
			markdown: "",
			wantErr:  false, // Empty markdown is valid, just produces empty RTF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rtf, err := MarkdownToRTF(tt.markdown)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkdownToRTF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(rtf) == 0 {
					t.Error("MarkdownToRTF() returned empty RTF data")
				}
				// RTF data should start with {\rtf
				rtfStr := string(rtf)
				if !strings.HasPrefix(rtfStr, "{\\rtf") {
					t.Errorf("MarkdownToRTF() RTF data doesn't start with {\\rtf, got: %s", rtfStr[:20])
				}
			}
		})
	}
}
