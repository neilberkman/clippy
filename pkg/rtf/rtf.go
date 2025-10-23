package rtf

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

type ConversionResult struct {
	HTML              string
	BackgroundColor   string
	DefaultTextColor  string
}

func parseRTFColor(colorDef string) string {
	redMatch := regexp.MustCompile(`\\red(\d+)`).FindStringSubmatch(colorDef)
	greenMatch := regexp.MustCompile(`\\green(\d+)`).FindStringSubmatch(colorDef)
	blueMatch := regexp.MustCompile(`\\blue(\d+)`).FindStringSubmatch(colorDef)

	if redMatch == nil || greenMatch == nil || blueMatch == nil {
		return ""
	}

	red, _ := strconv.Atoi(redMatch[1])
	green, _ := strconv.Atoi(greenMatch[1])
	blue, _ := strconv.Atoi(blueMatch[1])

	return fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
}

func parseRTFColorTable(rtf string) []string {
	colorTablePattern := regexp.MustCompile(`\{\\colortbl;([^}]+)\}`)
	match := colorTablePattern.FindStringSubmatch(rtf)
	if match == nil {
		return []string{""}
	}

	colorTable := []string{""} // Index 0 is auto/default color
	colorDefs := strings.Split(match[1], ";")

	for _, colorDef := range colorDefs {
		trimmed := strings.TrimSpace(colorDef)
		if trimmed != "" {
			color := parseRTFColor(trimmed)
			colorTable = append(colorTable, color)
		}
	}

	return colorTable
}

func ToHTML(rtf string) (*ConversionResult, error) {
	colorTable := parseRTFColorTable(rtf)

	var backgroundColor, defaultTextColor string

	initialFormatPattern := regexp.MustCompile(`\\f0\\fs\d+\s+\\cf(\d+)\s+\\cb(\d+)`)
	initialMatch := initialFormatPattern.FindStringSubmatch(rtf)
	if initialMatch != nil {
		textColorIndex, _ := strconv.Atoi(initialMatch[1])
		bgColorIndex, _ := strconv.Atoi(initialMatch[2])
		if textColorIndex < len(colorTable) {
			defaultTextColor = colorTable[textColorIndex]
		}
		if bgColorIndex < len(colorTable) {
			backgroundColor = colorTable[bgColorIndex]
		}
	}

	pardIndex := strings.Index(rtf, "\\pard")
	if pardIndex == -1 {
		return nil, fmt.Errorf("could not find paragraph content")
	}

	contentPattern := regexp.MustCompile(`\\f0\\fs\d+\s+\\cf\d+\s+\\cb\d+\s+\\CocoaLigature\d+\s+`)
	rtfSubstring := rtf[pardIndex:]
	match := contentPattern.FindStringIndex(rtfSubstring)
	if match == nil {
		return nil, fmt.Errorf("could not find content start marker")
	}

	contentStartOffset := pardIndex + match[1]
	lastBrace := strings.LastIndex(rtf, "}")
	content := rtf[contentStartOffset:lastBrace]

	var htmlBuilder strings.Builder
	i := 0
	currentColor := defaultTextColor
	currentBgColor := backgroundColor
	isBold := false

	for i < len(content) {
		char := content[i]

		if char == '\\' {
			j := i + 1

			if j < len(content) && content[j] == '\'' {
				j++
				hexCode := ""
				for j < len(content) && len(hexCode) < 2 && isHexDigit(content[j]) {
					hexCode += string(content[j])
					j++
				}
				if len(hexCode) == 2 {
					charCode, _ := strconv.ParseInt(hexCode, 16, 32)
					htmlBuilder.WriteRune(rune(charCode))
					i = j
					continue
				}
			}

			controlWord := ""
			for j < len(content) && isLetter(content[j]) {
				controlWord += string(content[j])
				j++
			}

			numParam := ""
			negative := false
			if j < len(content) && content[j] == '-' {
				negative = true
				j++
			}
			for j < len(content) && isDigit(content[j]) {
				numParam += string(content[j])
				j++
			}

			if j < len(content) && content[j] == ' ' {
				j++
			}

			switch controlWord {
			case "cf":
				colorIndex, _ := strconv.Atoi(numParam)
				if colorIndex < len(colorTable) {
					currentColor = colorTable[colorIndex]
				} else {
					currentColor = defaultTextColor
				}
				i = j
			case "cb":
				colorIndex, _ := strconv.Atoi(numParam)
				if colorIndex < len(colorTable) {
					currentBgColor = colorTable[colorIndex]
				} else {
					currentBgColor = backgroundColor
				}
				i = j
			case "b":
				if numParam == "0" {
					isBold = false
				} else {
					isBold = true
				}
				i = j
			case "uc":
				i = j
			case "u":
				codePoint, _ := strconv.Atoi(numParam)
				if negative {
					codePoint = -codePoint
				}
				if codePoint < 0 {
					unsigned := 65536 + codePoint
					htmlBuilder.WriteRune(rune(unsigned))
				} else {
					htmlBuilder.WriteRune(rune(codePoint))
				}
				if j < len(content) && content[j] == '?' {
					j++
				}
				i = j
			case "":
				if i+1 < len(content) {
					nextChar := content[i+1]
					switch nextChar {
					case '\\':
						htmlBuilder.WriteString("\\")
						i += 2
					case '\n':
						htmlBuilder.WriteString("\n")
						i += 2
					default:
						i++
					}
				} else {
					i++
				}
			default:
				i = j
			}
		} else if char == '\n' {
			htmlBuilder.WriteString("\n")
			i++
		} else {
			text := ""
			for i < len(content) && content[i] != '\\' && content[i] != '\n' {
				text += string(content[i])
				i++
			}

			if len(text) > 0 {
				var styles []string
				if currentColor != "" {
					styles = append(styles, fmt.Sprintf("color: %s", currentColor))
				}
				if currentBgColor != "" && currentBgColor != backgroundColor {
					styles = append(styles, fmt.Sprintf("background: %s", currentBgColor))
				}
				if isBold {
					styles = append(styles, "font-weight: bold")
				}

				escapedText := html.EscapeString(text)
				if len(styles) > 0 {
					htmlBuilder.WriteString(fmt.Sprintf(`<span style="%s;">%s</span>`, strings.Join(styles, "; "), escapedText))
				} else {
					htmlBuilder.WriteString(escapedText)
				}
			}
		}
	}

	return &ConversionResult{
		HTML:             htmlBuilder.String(),
		BackgroundColor:  backgroundColor,
		DefaultTextColor: defaultTextColor,
	}, nil
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
