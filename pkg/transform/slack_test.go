package transform

import (
	"encoding/json"
	"strings"
	"testing"
)

// decodeOps unpacks the delta document so tests can assert on operations
// rather than on JSON text.
func decodeOps(t *testing.T, delta string) []slackOp {
	t.Helper()
	var doc struct {
		Ops []slackOp `json:"ops"`
	}
	if err := json.Unmarshal([]byte(delta), &doc); err != nil {
		t.Fatalf("delta is not valid JSON: %v", err)
	}
	return doc.Ops
}

func renderSlack(t *testing.T, markdown string) []slackOp {
	t.Helper()
	message, err := MarkdownToSlack(markdown)
	if err != nil {
		t.Fatalf("MarkdownToSlack() error = %v", err)
	}
	return decodeOps(t, message.Delta)
}

// findInsert returns the operation inserting the given text.
func findInsert(ops []slackOp, text string) (slackOp, bool) {
	for _, op := range ops {
		if op.Insert == text {
			return op, true
		}
	}
	return slackOp{}, false
}

// attrOf reports one attribute of the operation inserting text.
func attrOf(t *testing.T, ops []slackOp, text, key string) any {
	t.Helper()
	op, ok := findInsert(ops, text)
	if !ok {
		t.Fatalf("no operation inserts %q; ops: %+v", text, ops)
	}
	return op.Attributes[key]
}

func TestSlackInlineAttributes(t *testing.T) {
	ops := renderSlack(t, "plain **bold** *italic* ~~gone~~ `code` [label](https://example.com)")

	cases := []struct {
		text string
		key  string
		want any
	}{
		{"bold", "bold", true},
		{"italic", "italic", true},
		{"gone", "strike", true},
		{"code", "code", true},
		{"label", "link", "https://example.com"},
	}
	for _, tc := range cases {
		if got := attrOf(t, ops, tc.text, tc.key); got != tc.want {
			t.Errorf("%q attribute %s = %v, want %v", tc.text, tc.key, got, tc.want)
		}
	}

	// Plain text must not inherit formatting from its neighbours.
	if op, ok := findInsert(ops, "plain "); !ok || op.Attributes != nil {
		t.Errorf("leading plain text carried attributes: %+v", op)
	}
}

func TestSlackBlockAttributesRideOnNewline(t *testing.T) {
	// Quill marks a line by attributing the newline that ends it, so the text
	// itself must stay unattributed.
	ops := renderSlack(t, "> quoted line")

	if op, ok := findInsert(ops, "quoted line"); !ok || op.Attributes != nil {
		t.Errorf("quote text should be unattributed, got %+v", op)
	}
	var found bool
	for _, op := range ops {
		if op.Insert == "\n" && op.Attributes["blockquote"] == true {
			found = true
		}
	}
	if !found {
		t.Errorf("no newline carried blockquote attribute; ops: %+v", ops)
	}
}

func TestSlackHeadingUsesHeaderLevel(t *testing.T) {
	ops := renderSlack(t, "# One\n\n## Two")

	var levels []float64
	for _, op := range ops {
		if header, ok := op.Attributes["header"]; ok {
			levels = append(levels, header.(float64))
		}
	}
	if len(levels) != 2 || levels[0] != 1 || levels[1] != 2 {
		t.Errorf("header levels = %v, want [1 2]", levels)
	}
}

// Slack drops the header attribute, so heading text must also be bold or it
// pastes as ordinary body text with no emphasis at all.
func TestSlackHeadingTextIsBold(t *testing.T) {
	ops := renderSlack(t, "# One\n\n## Two")

	for _, text := range []string{"One", "Two"} {
		if got := attrOf(t, ops, text, "bold"); got != true {
			t.Errorf("heading %q bold = %v, want true", text, got)
		}
	}
}

func TestSlackNestedListIndent(t *testing.T) {
	ops := renderSlack(t, "- top\n  - nested\n")

	var indents []any
	for _, op := range ops {
		if op.Insert == "\n" && op.Attributes["list"] == "bullet" {
			indents = append(indents, op.Attributes["indent"])
		}
	}
	if len(indents) != 2 {
		t.Fatalf("got %d bullet lines, want 2; ops: %+v", len(indents), ops)
	}
	if indents[0] != nil {
		t.Errorf("top-level bullet has indent %v, want none", indents[0])
	}
	if indents[1] != float64(1) {
		t.Errorf("nested bullet indent = %v, want 1", indents[1])
	}
}

func TestSlackOrderedListAndCodeBlock(t *testing.T) {
	ops := renderSlack(t, "1. first\n2. second\n\n```go\nline one\nline two\n```")

	var ordered, codeLines int
	for _, op := range ops {
		if op.Attributes["list"] == "ordered" {
			ordered++
		}
		if op.Attributes["code-block"] == true {
			codeLines++
		}
	}
	if ordered != 2 {
		t.Errorf("ordered list lines = %d, want 2", ordered)
	}
	// Each code line is marked individually, which is how Quill models a block.
	if codeLines != 2 {
		t.Errorf("code-block lines = %d, want 2", codeLines)
	}
}

// The constructs below have no Slack equivalent. They must degrade to
// something readable rather than vanish.
func TestSlackFlattensUnsupportedConstructs(t *testing.T) {
	t.Run("table keeps every cell and marks the header", func(t *testing.T) {
		ops := renderSlack(t, "| a | b |\n|---|---|\n| 1 | 2 |")

		if got := attrOf(t, ops, "a", "bold"); got != true {
			t.Errorf("header cell not bold: %v", got)
		}
		var text strings.Builder
		for _, op := range ops {
			text.WriteString(op.Insert)
		}
		for _, cell := range []string{"a", "b", "1", "2"} {
			if !strings.Contains(text.String(), cell) {
				t.Errorf("table cell %q missing from %q", cell, text.String())
			}
		}
	})

	t.Run("image becomes a link to its source", func(t *testing.T) {
		ops := renderSlack(t, "![alt text](https://example.com/x.png)")
		if got := attrOf(t, ops, "alt text", "link"); got != "https://example.com/x.png" {
			t.Errorf("image link = %v", got)
		}
	})

	t.Run("horizontal rule becomes a divider", func(t *testing.T) {
		// Unattributed runs merge, so the divider arrives inside a combined
		// insert rather than as an operation of its own.
		ops := renderSlack(t, "above\n\n---\n\nbelow")
		var text strings.Builder
		for _, op := range ops {
			text.WriteString(op.Insert)
		}
		if !strings.Contains(text.String(), "────────") {
			t.Errorf("no divider line in %q", text.String())
		}
	})

	t.Run("task list keeps its checked state", func(t *testing.T) {
		ops := renderSlack(t, "- [x] done\n- [ ] todo")
		var text strings.Builder
		for _, op := range ops {
			text.WriteString(op.Insert)
		}
		if !strings.Contains(text.String(), "☑") || !strings.Contains(text.String(), "☐") {
			t.Errorf("task states missing from %q", text.String())
		}
	})
}

func TestSlackHardLineBreakStaysInBlock(t *testing.T) {
	// A break inside a quote must remain quoted.
	ops := renderSlack(t, "> first\n> second")

	var quotedNewlines int
	for _, op := range ops {
		if op.Insert == "\n" && op.Attributes["blockquote"] == true {
			quotedNewlines++
		}
	}
	if quotedNewlines != 2 {
		t.Errorf("quoted lines = %d, want 2; ops: %+v", quotedNewlines, ops)
	}
}

func TestSlackDocumentEndsWithoutTrailingBlankLines(t *testing.T) {
	message, err := MarkdownToSlack("one\n\ntwo\n")
	if err != nil {
		t.Fatalf("MarkdownToSlack() error = %v", err)
	}
	if strings.HasSuffix(message.Delta, `{"insert":"\n\n"}]}`) {
		t.Errorf("delta ends with blank lines: %s", message.Delta)
	}
	if message.PlainText == "" {
		t.Error("plain-text fallback is empty")
	}
}

func TestSlackEmptyInputProducesValidDocument(t *testing.T) {
	message, err := MarkdownToSlack("")
	if err != nil {
		t.Fatalf("MarkdownToSlack() error = %v", err)
	}
	ops := decodeOps(t, message.Delta)
	// Quill requires a document to end in a newline; an empty message is a
	// single empty line rather than an empty op list.
	if len(ops) != 1 || ops[0].Insert != "\n" {
		t.Errorf("empty document ops = %+v, want one newline", ops)
	}
}

func TestSlackFormattingDiagnostics(t *testing.T) {
	cases := []struct {
		name     string
		markdown string
		want     SlackFormatting
		warning  string
	}{
		{"plain fence", "Before\n\n```\na  b\nc  d\n```\n\nAfter", SlackFormatting{CodeBlocks: 1, CodeLines: 2}, ""},
		{"indented rows", "Before\n\n```\n  a  b\n  c  d\n```\n\nAfter", SlackFormatting{CodeBlocks: 1, CodeLines: 2}, ""},
		{"four-backtick fence with literal fences", "````md\n```\n| a | b |\n```\n````", SlackFormatting{CodeBlocks: 1, CodeLines: 3}, ""},
		{"tilde fence", "~~~go\ncode\n~~~", SlackFormatting{CodeBlocks: 1, CodeLines: 1}, ""},
		{"indented code", "Before\n\n    code\n", SlackFormatting{CodeBlocks: 1, CodeLines: 1}, ""},
		{"blank code line", "```\nfirst\n\nlast\n```", SlackFormatting{CodeBlocks: 1, CodeLines: 3}, ""},
		{"separate blocks", "```\none\n```\n\n```\ntwo\n```", SlackFormatting{CodeBlocks: 2, CodeLines: 2}, ""},
		{"adjacent fences", "```\none\n```\n```\ntwo\n```", SlackFormatting{CodeBlocks: 2, CodeLines: 2}, ""},
		{"empty fence emits no code lines", "```\n```", SlackFormatting{}, ""},
		{"inline code", "Use `one` and ``a ` b``.", SlackFormatting{InlineCodeSpans: 2}, ""},
		{"two backticks across lines", "``\na  b\nc  d\n``", SlackFormatting{InlineCodeSpans: 1}, "multiline_inline_code"},
		{"fence inside paragraph", "Before ```a  b\nc  d``` after", SlackFormatting{InlineCodeSpans: 1}, "multiline_inline_code"},
		{"pipe table", "| label | value |\n|---|---|\n| long label | x |", SlackFormatting{Tables: 1}, "table_flattened"},
		{"prose pipes are not a table", "Use a | b.", SlackFormatting{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message, err := MarkdownToSlack(tc.markdown)
			if err != nil {
				t.Fatal(err)
			}
			if message.Formatting != tc.want {
				t.Errorf("formatting = %+v, want %+v", message.Formatting, tc.want)
			}
			if tc.warning == "" {
				if len(message.Warnings) != 0 {
					t.Errorf("unexpected warnings: %+v", message.Warnings)
				}
			} else if len(message.Warnings) != 1 || message.Warnings[0].Code != tc.warning {
				t.Errorf("warnings = %+v, want %s", message.Warnings, tc.warning)
			}

			// Diagnostics must agree with the actual attributes sent to Slack.
			var codeLines, codeBlocks int
			var previousCodeLine bool
			for _, op := range decodeOps(t, message.Delta) {
				for range strings.Count(op.Insert, "\n") {
					isCode := op.Attributes["code-block"] == true
					if isCode {
						codeLines++
						if !previousCodeLine {
							codeBlocks++
						}
					}
					previousCodeLine = isCode
				}
			}
			if codeLines != message.Formatting.CodeLines || codeBlocks != message.Formatting.CodeBlocks {
				t.Errorf("delta has %d code blocks / %d lines, report claims %+v", codeBlocks, codeLines, message.Formatting)
			}
		})
	}
}

// Both variants used in the September 2-3 incident must preserve padding and
// format every table row. Neither requires changing the fence or copy tool.
func TestSlackFencedTablePreservesPadding(t *testing.T) {
	for _, indent := range []string{"", "  "} {
		rows := indent + "name   value\n" + indent + "alpha  10\n" + indent + "beta   20\n"
		message, err := MarkdownToSlack("Before\n\n```\n" + rows + "```\n\nAfter")
		if err != nil {
			t.Fatal(err)
		}
		var text strings.Builder
		for _, op := range decodeOps(t, message.Delta) {
			text.WriteString(op.Insert)
		}
		if !strings.Contains(text.String(), rows) {
			t.Errorf("padding lost for indent %q: %q", indent, text.String())
		}
		if message.Formatting.CodeBlocks != 1 || message.Formatting.CodeLines != 3 {
			t.Errorf("unexpected formatting: %+v", message.Formatting)
		}
		if strings.Contains(message.PlainText, "```") {
			t.Error("plain text unexpectedly contains fence markers")
		}
	}
}
