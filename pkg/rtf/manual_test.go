// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/neilberkman/clippy/pkg/clipboard"
	"github.com/neilberkman/clippy/pkg/rtf"
)

func main() {
	fmt.Println("Testing RTF to HTML conversion with actual clipboard data")
	fmt.Println("Please copy some colored terminal output to your clipboard")
	fmt.Println("Press Enter when ready...")
	fmt.Scanln()

	types, err := clipboard.GetClipboardTypes()
	if err != nil {
		log.Fatalf("Failed to get clipboard types: %v", err)
	}

	fmt.Println("\nClipboard types available:")
	for _, t := range types {
		fmt.Printf("  - %s\n", t)
	}

	hasRTF := false
	for _, t := range types {
		if t == "public.rtf" {
			hasRTF = true
			break
		}
	}

	if !hasRTF {
		log.Fatal("No RTF data on clipboard. Try copying from Terminal.app")
	}

	rtfData, err := clipboard.GetClipboardDataForType("public.rtf")
	if err != nil {
		log.Fatalf("Failed to get RTF data: %v", err)
	}

	fmt.Printf("\nRTF data length: %d bytes\n", len(rtfData))
	fmt.Println("\nFirst 500 characters of RTF:")
	if len(rtfData) > 500 {
		fmt.Println(string(rtfData[:500]) + "...")
	} else {
		fmt.Println(string(rtfData))
	}

	result, err := rtf.ToHTML(string(rtfData))
	if err != nil {
		log.Fatalf("RTF to HTML conversion failed: %v", err)
	}

	fmt.Println("\n=== Conversion Result ===")
	fmt.Printf("Background Color: %s\n", result.BackgroundColor)
	fmt.Printf("Default Text Color: %s\n", result.DefaultTextColor)
	fmt.Println("\nHTML Output:")
	fmt.Println(result.HTML)

	outFile := "/tmp/rtf-test-output.html"
	htmlDoc := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body {
	background: %s;
	color: %s;
	font-family: Monaco, Courier, monospace;
	padding: 20px;
	white-space: pre-wrap;
}
</style>
</head>
<body>%s</body>
</html>`, result.BackgroundColor, result.DefaultTextColor, result.HTML)

	if err := os.WriteFile(outFile, []byte(htmlDoc), 0644); err != nil {
		log.Fatalf("Failed to write HTML file: %v", err)
	}

	fmt.Printf("\nFull HTML written to: %s\n", outFile)
	fmt.Println("Open it in your browser to see the result!")
}
