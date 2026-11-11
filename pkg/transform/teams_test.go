package transform

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMarkdownToTeamsRichTextWrapsRichFragment(t *testing.T) {
	markdown := "# Update\n\n**naïve café — 世界 👋**\n\n- one\n  - two\n\n```go\n  hello()\n```\n\n| A | B |\n| - | - |\n| 1 | 2 |"
	rich, err := MarkdownToRichText(markdown)
	if err != nil {
		t.Fatal(err)
	}
	teams, err := MarkdownToTeamsRichText(markdown)
	if err != nil {
		t.Fatal(err)
	}
	want := `<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>` + rich.HTML + `</body></html>`
	if teams.HTML != want {
		t.Errorf("Teams HTML = %q, want exact document %q", teams.HTML, want)
	}
	if !utf8.ValidString(teams.HTML) || !strings.Contains(teams.HTML, "naïve café — 世界 👋") {
		t.Errorf("Teams HTML lost UTF-8 text: %s", teams.HTML)
	}
	if !bytes.Equal(teams.RTF, rich.RTF) || teams.PlainText != rich.PlainText {
		t.Error("Teams conversion changed the rich text fallbacks")
	}
}
