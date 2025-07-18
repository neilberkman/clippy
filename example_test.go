package clippy_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/neilberkman/clippy"
)

func Example() {
	// Copy text to clipboard
	if err := clippy.CopyText("Hello, World!"); err != nil {
		log.Fatal(err)
	}

	// Copy a file (text files copy content, others copy reference)
	err := clippy.Copy("document.pdf")
	if err != nil {
		log.Fatal(err)
	}

	// Copy multiple files
	err = clippy.CopyMultiple([]string{"image1.jpg", "image2.png"})
	if err != nil {
		log.Fatal(err)
	}

	// Copy from a reader (e.g., from a download)
	reader := strings.NewReader("Some text content")
	err = clippy.CopyData(reader)
	if err != nil {
		log.Fatal(err)
	}

	// Get clipboard content
	if text, ok := clippy.GetText(); ok {
		fmt.Printf("Clipboard text: %s\n", text)
	}

	// Get files from clipboard
	files := clippy.GetFiles()
	for _, file := range files {
		fmt.Printf("File in clipboard: %s\n", file)
	}
}

func ExampleCopy() {
	// Copy a single file intelligently
	err := clippy.Copy("report.pdf")
	if err != nil {
		log.Printf("Failed to copy file: %v", err)
	}
}

func ExampleCopyText() {
	// Copy text to clipboard
	if err := clippy.CopyText("Hello from clippy library!"); err != nil {
		log.Printf("Failed to copy text: %v", err)
	}
}

func ExampleGetText() {
	// Get text from clipboard
	if text, ok := clippy.GetText(); ok {
		fmt.Printf("Clipboard contains: %s\n", text)
	} else {
		fmt.Println("No text in clipboard")
	}
}
