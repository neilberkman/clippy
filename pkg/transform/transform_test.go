package transform

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkdownToRichTextInlineSemanticsAndParagraphs(t *testing.T) {
	rich, err := MarkdownToRichText("Hello **bold**, *italic*, and ~~gone~~.\n\nSecond paragraph with \\*literal\\*.")
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `font-family: -apple-system;`)
	assertContains(t, rich.HTML, `font-size: 13px;`)
	assertContains(t, rich.HTML, `<b>bold</b>`)
	assertContains(t, rich.HTML, `<i>italic</i>`)
	assertContains(t, rich.HTML, `<s>gone</s>`)
	assertContains(t, rich.HTML, `<div><br></div>`)
	if got, want := rich.PlainText, "Hello bold, italic, and gone.\n\nSecond paragraph with *literal*."; got != want {
		t.Fatalf("PlainText mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestMarkdownToRichTextPreservesMarkdownBlankLines(t *testing.T) {
	rich, err := MarkdownToRichText("Hi Acme API team,\n\nI am following up on the issue.\n\nThe second paragraph has intentional spacing.")
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<div>Hi Acme API team,</div><div><br></div><div>I am following up on the issue.</div>`)
	assertContains(t, rich.HTML, `<div>The second paragraph has intentional spacing.</div>`)
	if got, want := rich.PlainText, "Hi Acme API team,\n\nI am following up on the issue.\n\nThe second paragraph has intentional spacing."; got != want {
		t.Fatalf("PlainText mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestMarkdownToRichTextTreatsSoftBreaksAsSpaces(t *testing.T) {
	rich, err := MarkdownToRichText("Thanks for taking a look.\n\nBest,\nNeil")
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<div>Best, Neil</div>`)
	if got, want := rich.PlainText, "Thanks for taking a look.\n\nBest, Neil"; got != want {
		t.Fatalf("PlainText mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestEmailToRichTextUsesExplicitRegions(t *testing.T) {
	rich, err := EmailToRichText(EmailParts{
		Salutation:   "Bonjour,",
		BodyMarkdown: "Voici le problème.\n\n**Détails** ci-dessous.",
		Signoff:      "Cordialement,\nNeil",
	})
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<div>Bonjour,</div><div>Voici le problème.`)
	assertContains(t, rich.HTML, `<div><br></div><div>Cordialement,<br>Neil</div>`)
	if got, want := rich.PlainText, "Bonjour,\nVoici le problème.\n\nDétails ci-dessous.\n\nCordialement,\nNeil"; got != want {
		t.Fatalf("PlainText mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestMarkdownToRichTextListsAndLinks(t *testing.T) {
	source := "- first\n- [the site](https://example.com/search?a=1&b=2)\n\n3. three\n4. four"
	rich, err := MarkdownToRichText(source)
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<ul style="margin: 0px; padding-inline-start: 2em;">`)
	assertContains(t, rich.HTML, `<ol style="margin: 0px; padding-inline-start: 2em;" start="3">`)
	assertContains(t, rich.HTML, `<li>first</li>`)
	assertContains(t, rich.HTML, `<a href="https://example.com/search?a=1&amp;b=2">the site</a>`)
	if got, want := rich.PlainText, "- first\n- the site\n\n3. three\n4. four"; got != want {
		t.Fatalf("PlainText mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestMarkdownToRichTextCodeUnicodeAndEscaping(t *testing.T) {
	source := "Use `<tag> & stuff` with café ☕.\n\n```go\nfmt.Println(\"héllo\")\n\tif x < 2 && ready {\n}\n```\n\n<em onclick=\"bad()\">unsafe</em>"
	rich, err := MarkdownToRichText(source)
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<code style="`)
	assertContains(t, rich.HTML, `&lt;tag&gt; &amp; stuff`)
	assertContains(t, rich.HTML, `<pre spellcheck="false" style="`)
	for _, style := range []string{
		`white-space: pre-wrap;`,
		`font-family: ui-monospace;`,
		`font-size: 13px;`,
		`border: 1px solid rgb(206, 206, 206);`,
		`background-color: rgb(248, 248, 248);`,
		`padding: 10px;`,
		`border-radius: 4px;`,
		`margin: 0px;`,
		`tab-size: 4;`,
	} {
		assertContains(t, rich.HTML, style)
	}
	assertContains(t, rich.HTML, `fmt.Println(&quot;héllo&quot;)`)
	assertContains(t, rich.HTML, `fmt.Println(&quot;héllo&quot;)<br>`)
	assertContains(t, rich.HTML, "\tif x &lt; 2 &amp;&amp; ready {")
	if strings.Contains(rich.HTML, `<br></pre>`) {
		t.Fatalf("fenced code gained a structural trailing blank line: %s", rich.HTML)
	}
	assertContains(t, rich.HTML, `&lt;em onclick=&quot;bad()&quot;&gt;unsafe&lt;/em&gt;`)
	if strings.Contains(rich.HTML, `<em onclick=`) {
		t.Fatalf("raw HTML was emitted as active markup: %s", rich.HTML)
	}
	assertContains(t, rich.PlainText, "Use <tag> & stuff with café ☕.")
	assertContains(t, rich.PlainText, "fmt.Println(\"héllo\")")
}

func TestMarkdownToRichTextDangerousLinkAndRTF(t *testing.T) {
	rich, err := MarkdownToRichText("[safe](https://example.com) [unsafe](javascript:alert(1))")
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, rich.HTML, `<a href="https://example.com">safe</a>`)
	if strings.Contains(strings.ToLower(rich.HTML), "javascript:") {
		t.Fatalf("dangerous URL was emitted: %s", rich.HTML)
	}
	assertContains(t, rich.HTML, "unsafe")
	if !bytes.HasPrefix(rich.RTF, []byte(`{\rtf`)) {
		t.Fatalf("RTF has an unexpected signature: %q", rich.RTF[:min(len(rich.RTF), 32)])
	}
}

func assertContains(t *testing.T, value, substring string) {
	t.Helper()
	if !strings.Contains(value, substring) {
		t.Fatalf("expected %q to contain %q", value, substring)
	}
}
