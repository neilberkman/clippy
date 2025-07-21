package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Build the pasty binary for testing
	cmd := exec.Command("go", "build", "-o", "pasty_test", ".")
	if err := cmd.Run(); err != nil {
		panic("Failed to build pasty test binary: " + err.Error())
	}

	// Build the clippy binary for testing (pasty tests need it)
	cmd = exec.Command("go", "build", "-o", "clippy_test", "../clippy")
	if err := cmd.Run(); err != nil {
		panic("Failed to build clippy test binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = os.Remove("pasty_test")
	_ = os.Remove("clippy_test")

	os.Exit(code)
}

func TestPastyHelp(t *testing.T) {
	cmd := exec.Command("./pasty_test", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty help failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "pasty - Smart paste tool for macOS") {
		t.Error("Help output should contain title")
	}
	if !strings.Contains(outputStr, "Usage:") {
		t.Error("Help output should contain usage")
	}
	if !strings.Contains(outputStr, "Examples:") {
		t.Error("Help output should contain examples")
	}
}

func TestPastyVersion(t *testing.T) {
	cmd := exec.Command("./pasty_test", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty version failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "pasty version") {
		t.Error("Version output should contain 'pasty version'")
	}
}

func TestPastyWithTextClipboard(t *testing.T) {
	// Clear clipboard first using --clear flag
	clearCmd := exec.Command("./clippy_test", "--clear")
	_, err := clearCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to clear clipboard: %v", err)
	}
	
	// Small delay to ensure clipboard is cleared
	time.Sleep(100 * time.Millisecond)

	// First, put some text on the clipboard using clippy
	clippyCmd := exec.Command("./clippy_test", "--verbose")
	clippyCmd.Stdin = strings.NewReader("Test text content for pasty")
	clippyOutput, err := clippyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to set clipboard with clippy: %v\nOutput: %s", err, clippyOutput)
	}

	// Verify clippy copied text content
	if !strings.Contains(string(clippyOutput), "Copied content from stream using smart detection") {
		t.Fatalf("Clippy should have copied text content, got: %s", clippyOutput)
	}

	// Small delay to ensure clipboard operation completes
	time.Sleep(100 * time.Millisecond)

	// Now test pasty
	cmd := exec.Command("./pasty_test", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Test text content for pasty") {
		t.Errorf("Pasty should output the text content, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "✅ Pasted text content to stdout") {
		t.Errorf("Pasty should show verbose success message, got: %s", outputStr)
	}
}

func TestPastyToFile(t *testing.T) {
	// Put text on clipboard
	clippyCmd := exec.Command("./clippy_test", "-v")
	clippyCmd.Stdin = strings.NewReader("Content for file test")
	clippyOutput, err := clippyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to set clipboard: %v\nOutput: %s", err, clippyOutput)
	}

	// Verify clippy copied text content
	if !strings.Contains(string(clippyOutput), "Copied content from stream using smart detection") {
		t.Fatalf("Clippy should have copied text content, got: %s", clippyOutput)
	}

	// Create temp file path
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_output.txt")

	// Use pasty to write to file
	cmd := exec.Command("./pasty_test", "-v", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty to file failed: %v\nOutput: %s", err, output)
	}

	// Check file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file should have been created")
	}

	// Check file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != "Content for file test" {
		t.Errorf("File content mismatch. Expected 'Content for file test', got '%s'", string(content))
	}

	// Check verbose output
	outputStr := string(output)
	if !strings.Contains(outputStr, "✅ Pasted text content to") {
		t.Error("Should show verbose success message for file output")
	}
}

func TestPastyWithFileClipboard(t *testing.T) {
	// Put a file reference on clipboard using clippy (use binary file so it copies as reference)
	testFile := "../../test-files/minimal.png"
	clippyCmd := exec.Command("./clippy_test", "-v", testFile)
	clippyOutput, err := clippyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to set clipboard with file: %v\nOutput: %s", err, clippyOutput)
	}

	// Verify clippy copied file reference
	if !strings.Contains(string(clippyOutput), "Copied file reference for 'minimal.png'") {
		t.Fatalf("Clippy should have copied file reference, got: %s", clippyOutput)
	}

	// Test pasty listing files
	cmd := exec.Command("./pasty_test", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty with file clipboard failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should list the file path (clippy stores absolute paths)
	if !strings.Contains(outputStr, "minimal.png") {
		t.Errorf("Should list the file path from clipboard, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "✅ Copied 1 files to '.'") {
		t.Errorf("Should show verbose message about copying files, got: %s", outputStr)
	}
}

func TestPastyCopyFileToDirectory(t *testing.T) {
	// Put a file reference on clipboard (use binary file so it copies as reference)
	testFile := "../../test-files/test.pdf"
	clippyCmd := exec.Command("./clippy_test", "-v", testFile)
	clippyOutput, err := clippyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to set clipboard with file: %v\nOutput: %s", err, clippyOutput)
	}

	// Verify clippy copied file reference
	if !strings.Contains(string(clippyOutput), "Copied file reference for 'test.pdf'") {
		t.Fatalf("Clippy should have copied file reference, got: %s", clippyOutput)
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Use pasty to copy to directory
	cmd := exec.Command("./pasty_test", "-v", tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pasty copy to directory failed: %v\nOutput: %s", err, output)
	}

	// Check file was copied
	copiedFile := filepath.Join(tempDir, "test.pdf")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Fatal("File should have been copied to directory")
	}

	// Check content matches
	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	copiedContent, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(originalContent) != string(copiedContent) {
		t.Error("Copied file content should match original")
	}

	// Check verbose output
	outputStr := string(output)
	if !strings.Contains(outputStr, "✅ Copied 1 files to") {
		t.Error("Should show verbose message about copying file")
	}
}
