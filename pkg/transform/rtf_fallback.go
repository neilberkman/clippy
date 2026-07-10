//go:build !darwin || !cgo

package transform

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

// materializeRTF provides a portable plain-text RTF representation. Darwin
// builds use AppKit's HTML importer instead so presentation attributes are
// materialized in the RTF itself.
func materializeRTF(_ string, plainText string) ([]byte, error) {
	var rtf strings.Builder
	rtf.WriteString(`{\rtf1\ansi\deff0{\fonttbl{\f0\fswiss Arial;}}\f0\fs26 `)
	for _, codeUnit := range utf16.Encode([]rune(plainText)) {
		switch codeUnit {
		case '\\', '{', '}':
			rtf.WriteByte('\\')
			rtf.WriteRune(rune(codeUnit))
		case '\n':
			rtf.WriteString(`\par `)
		case '\r':
			// Ignore CR; LF carries the line break.
		default:
			if codeUnit < 0x80 {
				rtf.WriteRune(rune(codeUnit))
			} else {
				signed := int16(codeUnit)
				fmt.Fprintf(&rtf, `\u%d?`, signed)
			}
		}
	}
	rtf.WriteByte('}')
	return []byte(rtf.String()), nil
}
