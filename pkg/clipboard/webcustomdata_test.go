package clipboard

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// decodeWebCustomData is the inverse of EncodeWebCustomData, used to prove the
// encoding round-trips. It mirrors how Chromium reads the blob back.
func decodeWebCustomData(t *testing.T, data []byte) [][2]string {
	t.Helper()
	if len(data) < 8 {
		t.Fatalf("blob too short: %d bytes", len(data))
	}
	payloadSize := binary.LittleEndian.Uint32(data[:4])
	if int(payloadSize) != len(data)-4 {
		t.Fatalf("payload size %d does not match %d trailing bytes", payloadSize, len(data)-4)
	}

	offset := 4
	count := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	readString := func() string {
		units := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		encoded := make([]uint16, units)
		for i := range encoded {
			encoded[i] = binary.LittleEndian.Uint16(data[offset+i*2:])
		}
		offset += int(units) * 2
		for offset%4 != 0 {
			offset++
		}
		return string(utf16.Decode(encoded))
	}

	entries := make([][2]string, 0, count)
	for range count {
		mime := readString()
		entries = append(entries, [2]string{mime, readString()})
	}
	return entries
}

func TestEncodeWebCustomDataRoundTrip(t *testing.T) {
	entries := [][2]string{
		{"public.utf8-plain-text", "hello"},
		{SlackDeltaMIME, `{"ops":[{"insert":"hi\n"}]}`},
	}

	decoded := decodeWebCustomData(t, EncodeWebCustomData(entries))

	if len(decoded) != len(entries) {
		t.Fatalf("decoded %d entries, want %d", len(decoded), len(entries))
	}
	for i, want := range entries {
		if decoded[i] != want {
			t.Errorf("entry %d = %v, want %v", i, decoded[i], want)
		}
	}
}

func TestEncodeWebCustomDataKeepsFourByteAlignment(t *testing.T) {
	// Odd-length strings force padding; every field must still start on a
	// 4-byte boundary or Chromium misreads the following length prefix.
	blob := EncodeWebCustomData([][2]string{{"a", "abc"}, {"ab", "abcde"}})

	if len(blob)%4 != 0 {
		t.Errorf("blob length %d is not 4-byte aligned", len(blob))
	}
	decoded := decodeWebCustomData(t, blob)
	if len(decoded) != 2 || decoded[0][1] != "abc" || decoded[1][1] != "abcde" {
		t.Errorf("misaligned decode: %v", decoded)
	}
}

func TestEncodeWebCustomDataHandlesNonBMPCharacters(t *testing.T) {
	// Emoji are surrogate pairs in UTF-16; the length prefix counts code
	// units, not runes.
	const value = "ship it 🚀"
	decoded := decodeWebCustomData(t, EncodeWebCustomData([][2]string{{SlackDeltaMIME, value}}))

	if len(decoded) != 1 || decoded[0][1] != value {
		t.Errorf("decoded %v, want %q", decoded, value)
	}
}
