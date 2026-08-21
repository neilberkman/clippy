package clipboard

import (
	"encoding/binary"
	"unicode/utf16"
)

// ChromiumWebCustomDataType is the pasteboard type Chromium uses to carry web
// MIME types it has no native clipboard mapping for. Electron apps (Slack,
// Discord, VS Code) read their own custom formats back out of it, so writing
// this blob is how a native tool hands an Electron editor its own rich format.
const ChromiumWebCustomDataType = "org.chromium.web-custom-data"

// SlackDeltaMIME is the custom MIME type Slack's composer reads: a Quill Delta
// document describing the message exactly as its own editor would.
const SlackDeltaMIME = "slack/texty"

// EncodeWebCustomData serializes MIME-type/value pairs in Chromium's Pickle
// format, the encoding behind ChromiumWebCustomDataType:
//
//	uint32  payload size (bytes after this header)
//	uint32  entry count
//	per entry, twice (MIME type then value):
//	  uint32  length in UTF-16 code units
//	  bytes   UTF-16LE data, zero-padded to a 4-byte boundary
//
// Order is preserved so output can be compared byte-for-byte against what the
// source application writes.
func EncodeWebCustomData(entries [][2]string) []byte {
	payload := binary.LittleEndian.AppendUint32(nil, uint32(len(entries)))
	for _, entry := range entries {
		payload = appendPickleString(payload, entry[0])
		payload = appendPickleString(payload, entry[1])
	}

	out := binary.LittleEndian.AppendUint32(nil, uint32(len(payload)))
	return append(out, payload...)
}

// appendPickleString writes one length-prefixed UTF-16LE string, padded so the
// next field stays 4-byte aligned.
func appendPickleString(dst []byte, value string) []byte {
	units := utf16.Encode([]rune(value))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(len(units)))
	for _, unit := range units {
		dst = binary.LittleEndian.AppendUint16(dst, unit)
	}
	for len(dst)%4 != 0 {
		dst = append(dst, 0)
	}
	return dst
}
