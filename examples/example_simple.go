// Example showing how to use clippy as a library
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/neilberkman/clippy"
)

func main() {
	fmt.Println("Testing clippy library API...")

	// 1. Copy text to clipboard
	if err := clippy.CopyText("Hello from clippy library!"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Copied text to clipboard")

	// 2. Copy a file intelligently (detects text vs binary)
	if err := clippy.Copy("../test-files/sample.txt"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Copied file to clipboard")

	// 3. Copy data from reader (handles text/binary detection)
	reader := strings.NewReader("Some text data from reader")
	if err := clippy.CopyData(reader); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Copied data from reader")

	// Note: You can also use os.Stdin for interactive input:
	// if err := clippy.CopyData(os.Stdin); err != nil {
	//     log.Fatal(err)
	// }

	// 4. Get text from clipboard
	text, ok := clippy.GetText()
	if ok {
		fmt.Printf("✓ Got text from clipboard: %s\n", text)
	}

	// 5. Get files from clipboard
	files := clippy.GetFiles()
	if len(files) > 0 {
		fmt.Printf("✓ Got %d files from clipboard\n", len(files))
	}

	fmt.Println("Library API test complete!")
}
