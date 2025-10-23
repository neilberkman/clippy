package rtf

import (
	"strings"
	"testing"
)

func TestParseRTFColor(t *testing.T) {
	tests := []struct {
		name     string
		colorDef string
		want     string
	}{
		{
			name:     "red color",
			colorDef: `\red255\green0\blue0`,
			want:     "rgb(255, 0, 0)",
		},
		{
			name:     "green color",
			colorDef: `\red0\green255\blue0`,
			want:     "rgb(0, 255, 0)",
		},
		{
			name:     "blue color",
			colorDef: `\red0\green0\blue255`,
			want:     "rgb(0, 0, 255)",
		},
		{
			name:     "white color",
			colorDef: `\red255\green255\blue255`,
			want:     "rgb(255, 255, 255)",
		},
		{
			name:     "black color",
			colorDef: `\red0\green0\blue0`,
			want:     "rgb(0, 0, 0)",
		},
		{
			name:     "gray color",
			colorDef: `\red128\green128\blue128`,
			want:     "rgb(128, 128, 128)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRTFColor(tt.colorDef)
			if got != tt.want {
				t.Errorf("parseRTFColor(%q) = %q, want %q", tt.colorDef, got, tt.want)
			}
		})
	}
}

func TestParseRTFColorTable(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red242\green242\blue242;\red0\green0\blue0;\red204\green98\blue70;}
}`

	colorTable := parseRTFColorTable(rtf)

	expected := []string{
		"",                     // Index 0 is auto/default
		"rgb(255, 255, 255)",   // white
		"rgb(242, 242, 242)",   // light gray
		"rgb(0, 0, 0)",         // black
		"rgb(204, 98, 70)",     // reddish
	}

	if len(colorTable) != len(expected) {
		t.Fatalf("Expected %d colors, got %d", len(expected), len(colorTable))
	}

	for i, color := range expected {
		if colorTable[i] != color {
			t.Errorf("Color at index %d: got %q, want %q", i, colorTable[i], color)
		}
	}
}

func TestToHTML_SimpleText(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 Hello World}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "Hello World") {
		t.Errorf("Expected HTML to contain 'Hello World', got: %s", result.HTML)
	}
}

func TestToHTML_ColoredText(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red255\green0\blue0;\red0\green255\blue0;\red0\green0\blue255;}
{\*\expandedcolortbl;;\cssrgb\c100000\c0\c0;\cssrgb\c0\c100000\c0;\cssrgb\c0\c0\c100000;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb1 \CocoaLigature0 \cf2 Red\cf3  Green\cf4  Blue}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "Red") {
		t.Errorf("Expected HTML to contain 'Red', got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "Green") {
		t.Errorf("Expected HTML to contain 'Green', got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "Blue") {
		t.Errorf("Expected HTML to contain 'Blue', got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "color:") {
		t.Errorf("Expected HTML to contain color styling, got: %s", result.HTML)
	}
}

func TestToHTML_BoldText(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 Normal \b Bold\b0  Normal}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "font-weight: bold") {
		t.Errorf("Expected HTML to contain bold styling, got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "Bold") {
		t.Errorf("Expected HTML to contain 'Bold', got: %s", result.HTML)
	}
}

func TestToHTML_UnicodeCharacters(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 Arrow \uc0\u8594 ? here}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "→") {
		t.Errorf("Expected HTML to contain unicode arrow (→), got: %s", result.HTML)
	}
}

func TestToHTML_HexEscapes(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 Test\'a0space}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	// \'a0 is non-breaking space (character code 160)
	if !strings.Contains(result.HTML, string(rune(160))) {
		t.Errorf("Expected HTML to contain non-breaking space, got: %s", result.HTML)
	}
}

func TestToHTML_InvalidRTF(t *testing.T) {
	tests := []struct {
		name string
		rtf  string
	}{
		{
			name: "missing pard",
			rtf:  `{\rtf1\ansi\ansicpg1252 Some text}`,
		},
		{
			name: "missing content marker",
			rtf: `{\rtf1\ansi\ansicpg1252
\pard
Some text without proper markers}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToHTML(tt.rtf)
			if err == nil {
				t.Errorf("Expected error for invalid RTF, got nil")
			}
		})
	}
}

func TestToHTML_PreservesSpaces(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 Word   Multiple   Spaces}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "   ") {
		t.Errorf("Expected HTML to preserve multiple spaces, got: %s", result.HTML)
	}
}

func TestToHTML_HTMLEscaping(t *testing.T) {
	rtf := `{\rtf1\ansi\ansicpg1252\cocoartf2859
\cocoatextscaling0\cocoaplatform0{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;\red0\green0\blue0;\red242\green242\blue242;}
{\*\expandedcolortbl;;\cssrgb\c0\c0\c0;\cssrgb\c96078\c96078\c96078;}
\deftab720
\pard\pardeftab720\partightenfactor0

\f0\fs20 \cf2 \cb3 \CocoaLigature0 <tag> & "quotes"}
`

	result, err := ToHTML(rtf)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	if !strings.Contains(result.HTML, "&lt;tag&gt;") {
		t.Errorf("Expected HTML to escape <tag>, got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "&amp;") {
		t.Errorf("Expected HTML to escape &, got: %s", result.HTML)
	}

	if !strings.Contains(result.HTML, "&#34;") && !strings.Contains(result.HTML, "&quot;") {
		t.Errorf("Expected HTML to escape quotes, got: %s", result.HTML)
	}
}
