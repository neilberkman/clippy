package transform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// SlackMessage holds the representations Slack's composer understands. Delta
// is Quill Delta JSON, the format Slack's own editor puts on the clipboard
// under the slack/texty MIME type; pasting it reconstructs native formatting
// rather than round-tripping through HTML. PlainText is the fallback every
// other target receives.
type SlackMessage struct {
	Delta      string
	PlainText  string
	Formatting SlackFormatting
	Warnings   []SlackWarning
}

// SlackFormatting describes the generated document, not the result of a paste.
type SlackFormatting struct {
	CodeBlocks      int `json:"code_blocks"`
	CodeLines       int `json:"code_lines"`
	InlineCodeSpans int `json:"inline_code_spans"`
	Tables          int `json:"tables"`
}

// SlackWarning identifies Markdown that may lose the author's intended layout.
type SlackWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// slackOp is one Quill Delta operation. Attributes precedes Insert so the
// serialized shape matches what Slack itself writes.
type slackOp struct {
	Attributes map[string]any `json:"attributes,omitempty"`
	Insert     string         `json:"insert"`
}

// MarkdownToSlack converts CommonMark/GFM source to a Quill Delta document
// plus readable plain text.
//
// Quill's model splits formatting in two: inline attributes (bold, italic,
// strike, code, link) ride on the inserted text, while block attributes
// (header, list, blockquote, code-block, indent) attach to the newline that
// *terminates* the line they describe.
//
// Slack's composer has no table, image, or horizontal-rule construct, so those
// are flattened to text that still carries their meaning rather than dropped.
func MarkdownToSlack(source string) (*SlackMessage, error) {
	sourceBytes := []byte(source)
	document := markdownParser.Parser().Parse(text.NewReader(sourceBytes))

	renderer := &slackRenderer{source: sourceBytes}
	renderer.renderBlockChildren(document, nil, 0)
	renderer.trimTrailingBlankLines()

	delta, err := json.Marshal(map[string]any{"ops": mergeSlackOps(renderer.ops)})
	if err != nil {
		return nil, fmt.Errorf("encode Slack delta: %w", err)
	}

	return &SlackMessage{
		Delta: string(delta),
		// Soft breaks are preserved: chat messages are line-oriented, so an
		// authored line break should survive into the fallback text.
		PlainText:  renderPlainSegment(source, true),
		Formatting: renderer.formatting,
		Warnings:   renderer.warnings,
	}, nil
}

type slackRenderer struct {
	source     []byte
	ops        []slackOp
	formatting SlackFormatting
	warnings   []SlackWarning
}

func (r *slackRenderer) warn(code, message string) {
	for _, warning := range r.warnings {
		if warning.Code == code {
			return
		}
	}
	r.warnings = append(r.warnings, SlackWarning{Code: code, Message: message})
}

func (r *slackRenderer) pushInsert(value string, attrs map[string]any) {
	if value == "" {
		return
	}
	r.ops = append(r.ops, slackOp{Insert: value, Attributes: copyAttrs(attrs)})
}

// pushNewline ends the current line, applying the line's block attributes.
func (r *slackRenderer) pushNewline(attrs map[string]any) {
	r.ops = append(r.ops, slackOp{Insert: "\n", Attributes: copyAttrs(attrs)})
}

// trimTrailingBlankLines removes blank lines the block separator added after
// the final block, so a message never pastes with trailing empty lines.
func (r *slackRenderer) trimTrailingBlankLines() {
	for len(r.ops) > 1 {
		last := r.ops[len(r.ops)-1]
		if last.Insert != "\n" || last.Attributes != nil {
			break
		}
		prev := r.ops[len(r.ops)-2]
		if prev.Insert != "\n" || prev.Attributes != nil {
			break
		}
		r.ops = r.ops[:len(r.ops)-1]
	}
}

// renderBlockChildren renders sibling blocks, preserving the author's blank
// lines between them so paragraph spacing survives into the message.
func (r *slackRenderer) renderBlockChildren(parent ast.Node, blockAttrs map[string]any, listDepth int) {
	first := true
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if !first && child.HasBlankPreviousLines() {
			r.pushNewline(blockAttrs)
		}
		r.renderBlock(child, blockAttrs, listDepth)
		first = false
	}
}

func (r *slackRenderer) renderBlock(node ast.Node, blockAttrs map[string]any, listDepth int) {
	switch n := node.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		r.renderInlineChildren(node, nil, blockAttrs, listDepth)
		r.pushNewline(blockAttrs)

	case *ast.Heading:
		// Slack's composer has no heading construct and drops Quill's header
		// attribute silently, so heading text is emitted bold to survive the
		// paste. The header attribute is kept alongside it: it costs nothing
		// here and carries the real structure to any editor that honours it.
		r.renderInlineChildren(n, withAttr(nil, "bold", true), blockAttrs, listDepth)
		r.pushNewline(withAttr(blockAttrs, "header", n.Level))

	case *ast.CodeBlock:
		r.renderCodeLines(n.Lines().Value(r.source), blockAttrs)

	case *ast.FencedCodeBlock:
		r.renderCodeLines(n.Lines().Value(r.source), blockAttrs)

	case *ast.List:
		r.renderList(n, blockAttrs, listDepth)

	case *ast.Blockquote:
		r.renderBlockChildren(n, withAttr(blockAttrs, "blockquote", true), listDepth)

	case *ast.ThematicBreak:
		// Slack has no rule element; a divider line keeps the visual break.
		r.pushInsert("────────", nil)
		r.pushNewline(blockAttrs)

	case *ast.HTMLBlock:
		for _, line := range splitLines(string(htmlBlockValue(n, r.source))) {
			r.pushInsert(line, nil)
			r.pushNewline(blockAttrs)
		}

	case *extast.Table:
		r.renderTable(n, blockAttrs, listDepth)

	case *ast.LinkReferenceDefinition:
		// Definitions render through their references, never on their own.

	default:
		if node.Type() == ast.TypeInline {
			r.renderInline(node, nil, blockAttrs, listDepth)
			return
		}
		r.renderBlockChildren(node, blockAttrs, listDepth)
	}
}

// renderCodeLines emits one Quill line per source line, each terminated by a
// code-block newline — Quill marks code blocks line by line.
func (r *slackRenderer) renderCodeLines(value []byte, blockAttrs map[string]any) {
	lines := splitLines(string(value))
	if len(lines) == 0 {
		return
	}
	if len(r.ops) == 0 || r.ops[len(r.ops)-1].Attributes["code-block"] != true {
		r.formatting.CodeBlocks++
	}
	r.formatting.CodeLines += len(lines)
	codeAttrs := withAttr(blockAttrs, "code-block", true)
	for _, line := range lines {
		r.pushInsert(line, nil)
		r.pushNewline(codeAttrs)
	}
}

func (r *slackRenderer) renderList(list *ast.List, blockAttrs map[string]any, listDepth int) {
	listType := "bullet"
	if list.IsOrdered() {
		listType = "ordered"
	}
	itemAttrs := withAttr(blockAttrs, "list", listType)
	if listDepth > 0 {
		itemAttrs = withAttr(itemAttrs, "indent", listDepth)
	}

	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		item, ok := child.(*ast.ListItem)
		if !ok {
			continue
		}
		for content := item.FirstChild(); content != nil; content = content.NextSibling() {
			// A nested list belongs one indent level deeper; anything else is
			// this item's own content and carries the item's list attributes.
			if nested, ok := content.(*ast.List); ok {
				r.renderList(nested, blockAttrs, listDepth+1)
				continue
			}
			r.renderBlock(content, itemAttrs, listDepth)
		}
	}
}

// renderTable flattens a GFM table: Slack has no table construct, so each row
// becomes a line of cells joined by a separator, header cells kept bold.
func (r *slackRenderer) renderTable(table *extast.Table, blockAttrs map[string]any, listDepth int) {
	r.formatting.Tables++
	r.warn("table_flattened", "Markdown tables become proportional-font rows separated by pipes; columns are not padded or aligned. For aligned columns, use space-padded text inside a fenced code block.")
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		_, isHeader := row.(*extast.TableHeader)
		first := true
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			if !first {
				r.pushInsert("  |  ", nil)
			}
			cellAttrs := map[string]any{}
			if isHeader {
				cellAttrs["bold"] = true
			}
			r.renderInlineChildren(cell, cellAttrs, blockAttrs, listDepth)
			first = false
		}
		r.pushNewline(blockAttrs)
	}
}

func (r *slackRenderer) renderInlineChildren(parent ast.Node, inlineAttrs, blockAttrs map[string]any, listDepth int) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderInline(child, inlineAttrs, blockAttrs, listDepth)
	}
}

func (r *slackRenderer) renderInline(node ast.Node, inlineAttrs, blockAttrs map[string]any, listDepth int) {
	switch n := node.(type) {
	case *ast.Text:
		r.pushInsert(string(n.Segment.Value(r.source)), inlineAttrs)
		// Both break kinds start a new line inside the same block, so they
		// carry the block's attributes (a break inside a quote stays quoted).
		if n.SoftLineBreak() || n.HardLineBreak() {
			r.pushNewline(blockAttrs)
		}

	case *ast.String:
		r.pushInsert(string(n.Value), inlineAttrs)

	case *ast.CodeSpan:
		r.formatting.InlineCodeSpans++
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok && strings.ContainsAny(string(t.Segment.Value(r.source)), "\r\n") {
				r.warn("multiline_inline_code", "A code span crosses source lines but renders as inline code, not a code block. For a table or multiline code, put an opening fence of three backticks and a closing fence on their own lines, with real newlines between them.")
				break
			}
		}
		r.pushInsert(r.inlineText(n), withAttr(inlineAttrs, "code", true))

	case *ast.Emphasis:
		attr := "italic"
		if n.Level >= 2 {
			attr = "bold"
		}
		r.renderInlineChildren(n, withAttr(inlineAttrs, attr, true), blockAttrs, listDepth)

	case *extast.Strikethrough:
		r.renderInlineChildren(n, withAttr(inlineAttrs, "strike", true), blockAttrs, listDepth)

	case *ast.Link:
		r.renderInlineChildren(n, withAttr(inlineAttrs, "link", string(n.Destination)), blockAttrs, listDepth)

	case *ast.AutoLink:
		url := string(n.URL(r.source))
		r.pushInsert(url, withAttr(inlineAttrs, "link", url))

	case *ast.Image:
		// Slack uploads images separately, so a pasted image survives as a
		// link to its source rather than disappearing.
		label := r.inlineText(n)
		url := string(n.Destination)
		if label == "" {
			label = url
		}
		r.pushInsert(label, withAttr(inlineAttrs, "link", url))

	case *ast.RawHTML:
		r.pushInsert(string(n.Segments.Value(r.source)), inlineAttrs)

	case *extast.TaskCheckBox:
		// Slack has no task-list construct; the box character carries the state.
		if n.IsChecked {
			r.pushInsert("☑ ", inlineAttrs)
		} else {
			r.pushInsert("☐ ", inlineAttrs)
		}

	default:
		r.renderInlineChildren(node, inlineAttrs, blockAttrs, listDepth)
	}
}

// inlineText collects the literal text under a node, used where Slack needs a
// plain label (code spans, image alt text).
func (r *slackRenderer) inlineText(node ast.Node) string {
	var out strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Text:
			out.Write(c.Segment.Value(r.source))
		case *ast.String:
			out.Write(c.Value)
		default:
			out.WriteString(r.inlineText(child))
		}
	}
	return out.String()
}

// mergeSlackOps joins neighbouring operations that share attributes, matching
// the compact shape Slack produces.
func mergeSlackOps(ops []slackOp) []slackOp {
	merged := make([]slackOp, 0, len(ops))
	for _, op := range ops {
		if len(merged) > 0 {
			last := &merged[len(merged)-1]
			// Only attribute-free runs merge: block attributes describe one
			// specific line and must stay on their own newline.
			if last.Attributes == nil && op.Attributes == nil {
				last.Insert += op.Insert
				continue
			}
		}
		merged = append(merged, op)
	}
	if len(merged) == 0 {
		return []slackOp{{Insert: "\n"}}
	}
	return merged
}

func copyAttrs(attrs map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}

// withAttr returns attrs plus one more entry, leaving the original untouched
// so sibling nodes never inherit each other's formatting.
func withAttr(attrs map[string]any, key string, value any) map[string]any {
	out := make(map[string]any, len(attrs)+1)
	for k, v := range attrs {
		out[k] = v
	}
	out[key] = value
	return out
}

// splitLines splits a block's text into lines, dropping the trailing empty
// element a final newline produces.
func splitLines(value string) []string {
	lines := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}
