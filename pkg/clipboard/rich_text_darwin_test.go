package clipboard

import (
	"bytes"
	"testing"
)

func TestCopyRichTextWritesRepresentations(t *testing.T) {
	htmlContent := "<div><strong>Hello, 世界 👋</strong><br>line two</div>"
	rtfContent := []byte("{\\rtf1\\ansi{\\fonttbl{\\f0 Helvetica;}}\\f0 Hello \\u19990?\\u30028?}\\x00\\xff")
	plainText := "Hello, 世界 👋\nline two"

	if err := CopyRichText(htmlContent, rtfContent, plainText); err != nil {
		t.Fatalf("CopyRichText() error = %v", err)
	}

	wantTypes := []string{"public.rtf", "public.html", "public.utf8-plain-text"}
	availableTypes := GetClipboardTypes()
	for _, wantType := range wantTypes {
		found := false
		for _, availableType := range availableTypes {
			if availableType == wantType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("clipboard is missing canonical type %q (all types: %v)", wantType, availableTypes)
		}
	}

	tests := []struct {
		name     string
		typeName string
		want     []byte
	}{
		{name: "RTF", typeName: "public.rtf", want: rtfContent},
		{name: "HTML", typeName: "public.html", want: []byte(htmlContent)},
		{name: "plain text", typeName: "public.utf8-plain-text", want: []byte(plainText)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetClipboardDataForType(tt.typeName)
			if !ok {
				t.Fatalf("clipboard is missing %q (all types: %v)", tt.typeName, availableTypes)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("data for %q = %q, want exact bytes %q", tt.typeName, got, tt.want)
			}
		})
	}

	gotPlainText, ok := GetText()
	if !ok {
		t.Fatal("GetText() did not find the plain-text fallback")
	}
	if gotPlainText != plainText {
		t.Errorf("GetText() = %q, want %q", gotPlainText, plainText)
	}
}
