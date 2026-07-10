// Package transform converts source text into clipboard-ready representations.
package transform

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	rootStyle = "font-family: -apple-system; caret-color: rgb(0, 0, 0); color: rgb(0, 0, 0); font-size: 13px; font-style: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; line-height: normal; text-align: start; text-indent: 0px; text-transform: none; word-spacing: 0px; margin: 0px;"
	listStyle = "margin: 0px; padding-inline-start: 2em;"
	preStyle  = "white-space: pre-wrap; font-family: ui-monospace; caret-color: rgb(0, 0, 0); color: rgb(0, 0, 0); font-size: 13px; font-style: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; orphans: 2; text-align: start; text-indent: 0px; text-transform: none; widows: 2; word-spacing: 0px; -webkit-text-stroke-width: 0px; text-decoration-line: none; text-decoration-thickness: auto; text-decoration-style: solid; border: 1px solid rgb(206, 206, 206); background-color: rgb(248, 248, 248); padding: 10px; border-radius: 4px; margin: 0px; tab-size: 4;"
	codeStyle = "font-family: ui-monospace; font-size: 13px; border: 1px solid rgb(206, 206, 206); background-color: rgb(248, 248, 248); padding: 1px 3px; border-radius: 3px;"
)

// RichText contains equivalent representations of a Markdown document.
// HTML is an email-safe fragment rather than a complete HTML document.
type RichText struct {
	HTML      string
	RTF       []byte
	PlainText string
}

var markdownParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// MarkdownToRichText converts CommonMark/GFM source to email-oriented HTML,
// concrete RTF formatting, and readable plain text.
func MarkdownToRichText(source string) (*RichText, error) {
	sourceBytes := []byte(source)
	document := markdownParser.Parser().Parse(text.NewReader(sourceBytes))

	htmlFragment := newHTMLRenderer(sourceBytes).render(document)
	plainText := newPlainTextRenderer(sourceBytes).render(document)
	rtf, err := materializeRTF(htmlFragment, plainText)
	if err != nil {
		return nil, fmt.Errorf("convert rendered Markdown to RTF: %w", err)
	}

	return &RichText{
		HTML:      htmlFragment,
		RTF:       rtf,
		PlainText: plainText,
	}, nil
}

type htmlRenderer struct {
	source             []byte
	buf                directBuffer
	preserveSoftBreaks bool
}

// directBuffer satisfies Goldmark's buffered-writer contract while writing
// immediately. Immediate writes let renderer-owned tags and Goldmark-escaped
// text safely share one output stream.
type directBuffer struct {
	bytes.Buffer
}

func (*directBuffer) Available() int { return 0 }
func (*directBuffer) Buffered() int  { return 0 }
func (*directBuffer) Flush() error   { return nil }

func newHTMLRenderer(source []byte) *htmlRenderer {
	return &htmlRenderer{source: source}
}

func (r *htmlRenderer) render(document ast.Node) string {
	r.buf.WriteString(`<div dir="ltr" style="`)
	r.buf.WriteString(rootStyle)
	r.buf.WriteString(`">`)
	r.renderBlockChildren(document, false)
	r.buf.WriteString(`</div>`)
	return r.buf.String()
}

func (r *htmlRenderer) renderBlockChildren(parent ast.Node, compact bool) {
	first := true
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		previous := child.PreviousSibling()
		if !first && child.HasBlankPreviousLines() && !r.isSalutation(previous) {
			r.writeBlankLine()
		}
		r.renderBlock(child, compact)
		first = false
	}
}

// isSalutation identifies the conventional greeting at the start of an email.
// A blank Markdown line after it is rendered as a normal line break so prose
// reads like an ordinary email rather than having an extra empty paragraph.
func (r *htmlRenderer) isSalutation(node ast.Node) bool {
	paragraph, ok := node.(*ast.Paragraph)
	if !ok {
		return false
	}
	return looksLikeSalutation(string(paragraph.Lines().Value(r.source)))
}

func (r *htmlRenderer) renderBlock(node ast.Node, compact bool) {
	switch n := node.(type) {
	case *ast.Paragraph:
		r.renderParagraph(n, compact)
	case *ast.TextBlock:
		r.renderParagraph(n, compact)
	case *ast.Heading:
		size := 22 - (n.Level * 2)
		if size < 13 {
			size = 13
		}
		_, _ = fmt.Fprintf(&r.buf, `<div style="font-size: %dpx; font-weight: 600;"><b>`, size)
		r.renderInlineChildren(n)
		r.buf.WriteString(`</b></div>`)
	case *ast.CodeBlock:
		r.renderPre(n.Lines().Value(r.source))
	case *ast.FencedCodeBlock:
		r.renderPre(n.Lines().Value(r.source))
	case *ast.List:
		r.renderList(n)
	case *ast.ListItem:
		r.renderListItem(n)
	case *ast.Blockquote:
		r.buf.WriteString(`<blockquote style="margin: 0; padding-left: 1em; border-left: 2px solid rgb(206,206,206);">`)
		r.renderBlockChildren(n, false)
		r.buf.WriteString(`</blockquote>`)
	case *ast.ThematicBreak:
		r.buf.WriteString(`<hr style="border: 0; border-top: 1px solid rgb(206,206,206); margin: 0;">`)
	case *ast.HTMLBlock:
		r.buf.WriteString(`<div>`)
		r.writeEscapedWithBreaks(htmlBlockValue(n, r.source))
		r.buf.WriteString(`</div>`)
	case *extast.Table:
		r.buf.WriteString(`<table style="border-collapse: collapse; margin: 0;">`)
		r.renderBlockChildren(n, false)
		r.buf.WriteString(`</table>`)
	case *extast.TableHeader:
		r.buf.WriteString(`<thead>`)
		r.renderBlockChildren(n, false)
		r.buf.WriteString(`</thead>`)
	case *extast.TableRow:
		r.buf.WriteString(`<tr>`)
		r.renderBlockChildren(n, false)
		r.buf.WriteString(`</tr>`)
	case *extast.TableCell:
		tag := "td"
		if _, ok := n.Parent().(*extast.TableHeader); ok {
			tag = "th"
		}
		_, _ = fmt.Fprintf(&r.buf, `<%s style="border: 1px solid rgb(206,206,206); padding: 2px 4px; text-align: %s;">`, tag, tableAlignment(n.Alignment))
		r.renderInlineChildren(n)
		_, _ = fmt.Fprintf(&r.buf, `</%s>`, tag)
	case *ast.LinkReferenceDefinition:
		// Link definitions are metadata and have no visible representation.
	default:
		if node.Type() == ast.TypeInline {
			r.renderInline(node)
			return
		}
		r.renderBlockChildren(node, compact)
	}
}

func (r *htmlRenderer) renderParagraph(node ast.Node, compact bool) {
	if !compact {
		r.buf.WriteString(`<div>`)
	}
	previous := r.preserveSoftBreaks
	r.preserveSoftBreaks = isSignoff(node, r.source)
	r.renderInlineChildren(node)
	r.preserveSoftBreaks = previous
	if !compact {
		r.buf.WriteString(`</div>`)
	}
}

func (r *htmlRenderer) renderList(list *ast.List) {
	tag := "ul"
	if list.IsOrdered() {
		tag = "ol"
	}
	_, _ = fmt.Fprintf(&r.buf, `<%s style="%s"`, tag, listStyle)
	if list.IsOrdered() && list.Start != 1 {
		_, _ = fmt.Fprintf(&r.buf, ` start="%d"`, list.Start)
	}
	r.buf.WriteByte('>')
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderBlock(child, true)
	}
	_, _ = fmt.Fprintf(&r.buf, `</%s>`, tag)
}

func (r *htmlRenderer) renderListItem(item *ast.ListItem) {
	r.buf.WriteString(`<li>`)
	first := true
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		if !first && child.HasBlankPreviousLines() {
			r.writeBlankLine()
		}
		_, isParagraph := child.(*ast.Paragraph)
		_, isTextBlock := child.(*ast.TextBlock)
		r.renderBlock(child, isParagraph || isTextBlock)
		first = false
	}
	r.buf.WriteString(`</li>`)
}

func (r *htmlRenderer) renderPre(value []byte) {
	r.buf.WriteString(`<pre spellcheck="false" style="`)
	r.buf.WriteString(preStyle)
	r.buf.WriteString(`">`)
	// Goldmark's block segments include the structural newline immediately
	// before the closing fence. It is not part of the code block's visual
	// content; remove exactly that newline while preserving intentional blank
	// lines written before it.
	value = bytes.TrimSuffix(value, []byte("\n"))
	lines := bytes.Split(value, []byte("\n"))
	for i, line := range lines {
		if i > 0 {
			r.buf.WriteString(`<br>`)
		}
		goldhtml.DefaultWriter.RawWrite(&r.buf, line)
	}
	r.buf.WriteString(`</pre>`)
}

func (r *htmlRenderer) renderInlineChildren(parent ast.Node) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderInline(child)
	}
}

func (r *htmlRenderer) renderInline(node ast.Node) {
	switch n := node.(type) {
	case *ast.Text:
		if n.IsRaw() {
			goldhtml.DefaultWriter.RawWrite(&r.buf, n.Value(r.source))
		} else {
			goldhtml.DefaultWriter.Write(&r.buf, n.Value(r.source))
		}
		if n.HardLineBreak() {
			r.buf.WriteString(`<br>`)
		} else if n.SoftLineBreak() && r.preserveSoftBreaks {
			r.buf.WriteString(`<br>`)
		} else if n.SoftLineBreak() {
			r.buf.WriteByte(' ')
		}
	case *ast.String:
		if n.IsRaw() || n.IsCode() {
			goldhtml.DefaultWriter.RawWrite(&r.buf, n.Value)
		} else {
			goldhtml.DefaultWriter.Write(&r.buf, n.Value)
		}
	case *ast.Emphasis:
		tag := "i"
		if n.Level == 2 {
			tag = "b"
		}
		_, _ = fmt.Fprintf(&r.buf, `<%s>`, tag)
		r.renderInlineChildren(n)
		_, _ = fmt.Fprintf(&r.buf, `</%s>`, tag)
	case *extast.Strikethrough:
		r.buf.WriteString(`<s>`)
		r.renderInlineChildren(n)
		r.buf.WriteString(`</s>`)
	case *ast.CodeSpan:
		r.buf.WriteString(`<code style="`)
		r.buf.WriteString(codeStyle)
		r.buf.WriteString(`">`)
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			textNode := child.(*ast.Text)
			value := textNode.Value(r.source)
			if bytes.HasSuffix(value, []byte("\n")) {
				value = append(bytes.Clone(value[:len(value)-1]), ' ')
			}
			goldhtml.DefaultWriter.RawWrite(&r.buf, value)
		}
		r.buf.WriteString(`</code>`)
	case *ast.Link:
		r.renderLink(n)
	case *ast.AutoLink:
		r.renderAutoLink(n)
	case *ast.Image:
		// Avoid adding remote image loads to an email fragment. Preserve alt text.
		r.renderInlineChildren(n)
	case *ast.RawHTML:
		goldhtml.DefaultWriter.RawWrite(&r.buf, n.Segments.Value(r.source))
	case *extast.TaskCheckBox:
		if n.IsChecked {
			r.buf.WriteString("☑ ")
		} else {
			r.buf.WriteString("☐ ")
		}
	default:
		r.renderInlineChildren(node)
	}
}

func (r *htmlRenderer) renderLink(link *ast.Link) {
	destination := util.URLEscape(link.Destination, true)
	if goldhtml.IsDangerousURL(destination) {
		r.renderInlineChildren(link)
		return
	}

	r.buf.WriteString(`<a href="`)
	r.buf.Write(util.EscapeHTML(destination))
	r.buf.WriteByte('"')
	if len(link.Title) > 0 {
		r.buf.WriteString(` title="`)
		goldhtml.DefaultWriter.Write(&r.buf, link.Title)
		r.buf.WriteByte('"')
	}
	r.buf.WriteByte('>')
	r.renderInlineChildren(link)
	r.buf.WriteString(`</a>`)
}

func (r *htmlRenderer) renderAutoLink(link *ast.AutoLink) {
	destination := util.URLEscape(link.URL(r.source), false)
	if link.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(destination), []byte("mailto:")) {
		destination = append([]byte("mailto:"), destination...)
	}
	if goldhtml.IsDangerousURL(destination) {
		goldhtml.DefaultWriter.Write(&r.buf, link.Label(r.source))
		return
	}
	r.buf.WriteString(`<a href="`)
	r.buf.Write(util.EscapeHTML(destination))
	r.buf.WriteString(`">`)
	r.buf.Write(util.EscapeHTML(link.Label(r.source)))
	r.buf.WriteString(`</a>`)
}

func (r *htmlRenderer) writeBlankLine() {
	r.buf.WriteString(`<div><br></div>`)
}

func (r *htmlRenderer) writeEscapedWithBreaks(value []byte) {
	value = bytes.TrimSuffix(value, []byte("\n"))
	lines := bytes.Split(value, []byte("\n"))
	for i, line := range lines {
		if i > 0 {
			r.buf.WriteString(`<br>`)
		}
		goldhtml.DefaultWriter.RawWrite(&r.buf, line)
	}
}

func htmlBlockValue(block *ast.HTMLBlock, source []byte) []byte {
	var value []byte
	for i := 0; i < block.Lines().Len(); i++ {
		segment := block.Lines().At(i)
		value = append(value, segment.Value(source)...)
	}
	if block.HasClosure() {
		value = append(value, block.ClosureLine.Value(source)...)
	}
	return value
}

func tableAlignment(alignment extast.Alignment) string {
	switch alignment {
	case extast.AlignCenter:
		return "center"
	case extast.AlignRight:
		return "right"
	default:
		return "left"
	}
}

type plainTextRenderer struct {
	source             []byte
	preserveSoftBreaks bool
}

func newPlainTextRenderer(source []byte) *plainTextRenderer {
	return &plainTextRenderer{source: source}
}

func (r *plainTextRenderer) render(document ast.Node) string {
	var blocks []string
	for child := document.FirstChild(); child != nil; child = child.NextSibling() {
		if rendered := strings.TrimRight(r.renderBlock(child, 0), "\n"); rendered != "" {
			blocks = append(blocks, rendered)
		}
	}
	var rendered strings.Builder
	for i, block := range blocks {
		if i > 0 {
			separator := "\n\n"
			if i == 1 && looksLikeSalutation(blocks[0]) {
				separator = "\n"
			}
			rendered.WriteString(separator)
		}
		rendered.WriteString(block)
	}
	return rendered.String()
}

func looksLikeSalutation(value string) bool {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, "\r\n") {
		return false
	}
	lower := strings.ToLower(value)
	for _, prefix := range []string{"hi ", "hello ", "dear ", "hey "} {
		if strings.HasPrefix(lower, prefix) {
			return strings.HasSuffix(value, ",") || strings.HasSuffix(value, ":") || strings.HasSuffix(value, "!")
		}
	}
	return false
}

func isSignoff(node ast.Node, source []byte) bool {
	paragraph, ok := node.(*ast.Paragraph)
	if !ok {
		return false
	}
	value := strings.TrimSpace(string(paragraph.Lines().Value(source)))
	lower := strings.ToLower(value)
	for _, prefix := range []string{"best,", "thanks,", "thank you,", "regards,", "kind regards,", "best regards,", "sincerely,", "cheers,"} {
		if strings.HasPrefix(lower, prefix+"\n") || strings.HasPrefix(lower, prefix+"\r\n") {
			return true
		}
	}
	return false
}

func (r *plainTextRenderer) renderBlock(node ast.Node, indent int) string {
	switch n := node.(type) {
	case *ast.Paragraph, *ast.TextBlock, *ast.Heading:
		previous := r.preserveSoftBreaks
		r.preserveSoftBreaks = isSignoff(node, r.source)
		text := r.renderInlineChildren(node)
		r.preserveSoftBreaks = previous
		return text
	case *ast.CodeBlock:
		return strings.TrimSuffix(string(n.Lines().Value(r.source)), "\n")
	case *ast.FencedCodeBlock:
		return strings.TrimSuffix(string(n.Lines().Value(r.source)), "\n")
	case *ast.List:
		return r.renderList(n, indent)
	case *ast.ListItem:
		return r.renderListItem(n, indent)
	case *ast.Blockquote:
		content := r.renderBlockChildren(n, indent)
		lines := strings.Split(content, "\n")
		for i := range lines {
			lines[i] = "> " + lines[i]
		}
		return strings.Join(lines, "\n")
	case *ast.ThematicBreak:
		return "────────"
	case *ast.HTMLBlock:
		return strings.TrimRight(string(htmlBlockValue(n, r.source)), "\n")
	case *extast.Table:
		return r.renderBlockChildren(n, indent)
	case *extast.TableHeader, *extast.TableRow:
		var cells []string
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			cells = append(cells, r.renderBlock(child, indent))
		}
		return strings.Join(cells, " | ")
	case *extast.TableCell:
		return r.renderInlineChildren(n)
	case *ast.LinkReferenceDefinition:
		return ""
	default:
		if node.Type() == ast.TypeInline {
			return r.renderInline(node)
		}
		return r.renderBlockChildren(node, indent)
	}
}

func (r *plainTextRenderer) renderBlockChildren(parent ast.Node, indent int) string {
	var blocks []string
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if rendered := strings.TrimRight(r.renderBlock(child, indent), "\n"); rendered != "" {
			blocks = append(blocks, rendered)
		}
	}
	return strings.Join(blocks, "\n\n")
}

func (r *plainTextRenderer) renderList(list *ast.List, indent int) string {
	var items []string
	itemNumber := list.Start
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		item := child.(*ast.ListItem)
		marker := "- "
		if list.IsOrdered() {
			marker = fmt.Sprintf("%d. ", itemNumber)
			itemNumber++
		}
		content := r.renderListItem(item, indent+len(marker))
		lines := strings.Split(content, "\n")
		var rendered strings.Builder
		rendered.WriteString(strings.Repeat(" ", indent))
		rendered.WriteString(marker)
		if len(lines) > 0 {
			rendered.WriteString(lines[0])
		}
		for _, line := range lines[1:] {
			rendered.WriteByte('\n')
			if line != "" {
				rendered.WriteString(strings.Repeat(" ", indent+len(marker)))
				rendered.WriteString(line)
			}
		}
		items = append(items, rendered.String())
	}
	return strings.Join(items, "\n")
}

func (r *plainTextRenderer) renderListItem(item *ast.ListItem, indent int) string {
	var parts []string
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		if rendered := strings.TrimRight(r.renderBlock(child, indent), "\n"); rendered != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, "\n")
}

func (r *plainTextRenderer) renderInlineChildren(parent ast.Node) string {
	var rendered strings.Builder
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		rendered.WriteString(r.renderInline(child))
	}
	return rendered.String()
}

func (r *plainTextRenderer) renderInline(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Text:
		value := resolvePlainText(n.Value(r.source), n.IsRaw())
		if n.HardLineBreak() {
			return value + "\n"
		}
		if n.SoftLineBreak() && r.preserveSoftBreaks {
			return value + "\n"
		}
		if n.SoftLineBreak() {
			return value + " "
		}
		return value
	case *ast.String:
		return resolvePlainText(n.Value, n.IsRaw() || n.IsCode())
	case *ast.CodeSpan:
		var value strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			part := string(child.(*ast.Text).Value(r.source))
			part = strings.TrimSuffix(part, "\n")
			value.WriteString(part)
			if strings.HasSuffix(string(child.(*ast.Text).Value(r.source)), "\n") {
				value.WriteByte(' ')
			}
		}
		return value.String()
	case *ast.AutoLink:
		return string(n.Label(r.source))
	case *ast.RawHTML:
		return string(n.Segments.Value(r.source))
	case *extast.TaskCheckBox:
		if n.IsChecked {
			return "☑ "
		}
		return "☐ "
	default:
		return r.renderInlineChildren(node)
	}
}

func resolvePlainText(value []byte, raw bool) string {
	if raw {
		return string(value)
	}
	value = util.UnescapePunctuations(value)
	value = util.ResolveNumericReferences(value)
	value = util.ResolveEntityNames(value)
	return string(value)
}
