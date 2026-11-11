package clipboard

import (
	"bytes"
	"slices"
	"testing"
)

func TestCopyTeamsRichTextWritesRichRepresentations(t *testing.T) {
	htmlContent := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body><b>naïve café — 世界 👋</b></body></html>"
	rtfContent := []byte("{\\rtf1\\ansi Hello}\x00\xff")
	if err := CopyText("stale plain text"); err != nil {
		t.Fatal(err)
	}
	if err := CopyTeamsRichText(htmlContent, rtfContent); err != nil {
		t.Fatal(err)
	}
	availableTypes := GetClipboardTypes()
	for _, typeName := range []string{"public.html", "public.rtf"} {
		if !slices.Contains(availableTypes, typeName) {
			t.Errorf("missing %s in %v", typeName, availableTypes)
		}
	}
	for _, tc := range []struct {
		typeName string
		want     []byte
	}{
		{"public.html", []byte(htmlContent)},
		{"public.rtf", rtfContent},
	} {
		got, ok := GetClipboardDataForType(tc.typeName)
		if !ok || !bytes.Equal(got, tc.want) {
			t.Errorf("%s = %q (present=%v), want exact bytes %q", tc.typeName, got, ok, tc.want)
		}
	}
	// In-process reads may translate RTF to plain text without an explicit plain-text representation.
	if text, ok := GetText(); ok && text != "Hello" {
		t.Errorf("GetText() = %q; want only RTF-derived text, never stale text or HTML source", text)
	}
}
